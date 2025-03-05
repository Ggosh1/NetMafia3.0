package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var disconnectedPlayers = make(map[string]bool)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_id")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		} else {
			fmt.Println("Ошибка при получении куки:", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		playerID = fmt.Sprintf("player-%d", len(game.Players)+1)
	}

	game.Mutex.Lock()
	player, exists := game.Players[playerID]
	if exists {
		// При переподключении обновляем соединение
		player.Conn = conn
		log.Printf("Player %s reconnected", playerID)
	} else {
		// Создаём нового игрока
		player = &Player{
			ID:      playerID,
			Conn:    conn,
			IsAlive: true,
		}
		game.Players[playerID] = player
		log.Printf("New player %s connected", playerID)
	}
	// Готовим снимок списка игроков для отправки клиенту
	playersSnapshot := make(map[string]bool)
	for id, p := range game.Players {
		playersSnapshot[id] = p.IsAlive
	}
	game.Mutex.Unlock()

	// Отправляем начальное состояние (например, список игроков)
	initialStatus := struct {
		Type    string          `json:"type"`
		Players map[string]bool `json:"players"`
	}{
		Type:    "playerList",
		Players: playersSnapshot,
	}
	if err := conn.WriteJSON(initialStatus); err != nil {
		log.Printf("Ошибка отправки начального состояния игроку %s: %v", playerID, err)
	}

	// Отправляем историю чата и, если есть, роль игрока
	game.Mutex.Lock()
	// История чата
	if err := conn.WriteJSON(struct {
		Type    string        `json:"type"`
		History []ChatMessage `json:"history"`
	}{
		Type:    "chatHistory",
		History: chatHistory,
	}); err != nil {
		log.Printf("Ошибка отправки истории чата игроку %s: %v", playerID, err)
	}
	// Если у игрока уже установлена роль, отправляем её
	if player.Role != "" {
		roleMsg := struct {
			Type string `json:"type"`
			Role string `json:"role"`
		}{
			Type: "role",
			Role: player.Role,
		}
		if err := conn.WriteJSON(roleMsg); err != nil {
			log.Printf("Ошибка отправки роли игроку %s: %v", playerID, err)
		}
	}
	game.Mutex.Unlock()

	broadcastPlayerList()

	log.Printf("Player %s connected. Total players: %d", playerID, len(game.Players))

	// Чтение сообщений от игрока
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %s disconnected: %v", playerID, err)
			game.Mutex.Lock()
			// При разрыве соединения сбрасываем Conn, но не удаляем игрока
			if p, ok := game.Players[playerID]; ok {
				p.Conn = nil
			}
			game.Mutex.Unlock()

			broadcastPlayerList()
			break
		}
		log.Printf("Message from %s: %s", playerID, string(message))

		processMessage(playerID, message)
	}
}

func broadcastPlayerList() {
	game.Mutex.Lock()
	playersSnapshot := make(map[string]bool)
	for id, p := range game.Players {
		if p.Conn != nil {
			playersSnapshot[id] = p.IsAlive
		}
	}
	game.Mutex.Unlock()

	update := struct {
		Type    string          `json:"type"`
		Players map[string]bool `json:"players"`
	}{
		Type:    "playerList",
		Players: playersSnapshot,
	}

	game.Mutex.Lock()
	for _, p := range game.Players {
		if p.Conn != nil {
			if err := p.Conn.WriteJSON(update); err != nil {
				log.Printf("Ошибка отправки списка игрока %s: %v", p.ID, err)
			}
		}
	}
	game.Mutex.Unlock()
}

func processMessage(playerID string, message []byte) {
	var msg struct {
		Action  string `json:"action"`
		Target  string `json:"vote"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}

	if msg.Action == "start_game" {
		log.Printf("Player %s requested to start the game", playerID)
		go startGame(playerID)
		return
	}

	game.Mutex.Lock()
	player, exists := game.Players[playerID]
	if !exists || !player.IsAlive {
		game.Mutex.Unlock()
		return
	}

	switch msg.Action {

	case "vote":
		// Голосуем только в дневной фазе и если игрок не взломан
		if game.CurrentPhase == "day" && !player.Hacked {
			// Запрещаем голосовать за самого себя
			if msg.Target == playerID {
				log.Printf("Player %s attempted to vote for themselves. Vote ignored.", playerID)
				game.Mutex.Unlock()
				return
			}

			var needSetVote bool = true
			if player.VotedFor != "" && player.VotedFor == msg.Target {
				log.Printf("Player %s attempted to vote for the same target twice. Vote ignored.", playerID)
				needSetVote = false
			}

			if player.VotedFor != "" {
				log.Printf("Удаление голоса игрока %s.", playerID)
				game.Votes[player.VotedFor]--
				if game.Votes[player.VotedFor] < 0 {
					game.Votes[player.VotedFor] = 0
				}
				player.VotedFor = ""
			}

			if _, ok := game.Players[msg.Target]; ok && needSetVote {
				player.VotedFor = msg.Target
				game.Votes[msg.Target]++
				log.Printf("Player %s voted for %s", playerID, msg.Target)
			}

			//log.Printf("Итоговый голос игрока %s - %s", playerID, player.VotedFor)
			broadcastGameStatus()

		} else if game.CurrentPhase == "night" && !player.Hacked {
			if msg.Target == playerID {
				log.Printf("Player %s attempted to target themselves at night. Action ignored.", playerID)
				game.Mutex.Unlock()
				return
			}

			var needSetVote bool = true
			// Если игрок уже выбрал эту же цель, игнорируем повторное действие
			if player.VotedFor != "" && player.VotedFor == msg.Target {
				log.Printf("Player %s attempted to target the same player twice at night. Action ignored.", playerID)
				needSetVote = false
			}

			// Если уже было выбрано что-то ранее, удаляем предыдущий выбор
			if player.Action != "" {
				log.Printf("Removing previous night action for player %s.", playerID)
				// Если ведется счет голосов/выборов ночью, можно его уменьшить
				player.Action = ""
				player.VotedFor = ""
			}

			// Если цель существует и новое действие нужно установить
			if _, ok := game.Players[msg.Target]; ok && needSetVote {
				player.Action = msg.Target
				player.VotedFor = msg.Target
				log.Printf("Player %s targeted %s at night.", playerID, msg.Target)
			}

			broadcastGameStatus()

			broadcastGameStatus()
		}

		game.Mutex.Unlock()

	/*case "cancel_vote":
	// Обработка отмены голоса в дневной фазе (если голос был поставлен)
	if game.CurrentPhase == "day" && !player.Hacked {
		// Обрабатываем отмену только если передан целевой ID совпадающий с текущим голосом
		if player.VotedFor != "" && msg.Target == player.VotedFor {
			game.Votes[player.VotedFor]--
			if game.Votes[player.VotedFor] < 0 {
				game.Votes[player.VotedFor] = 0
			}
			player.VotedFor = ""
		}
	} else if game.CurrentPhase == "night" {
		player.Action = ""
	}
	game.Mutex.Unlock()*/

	case "chat":
		// Сообщения в чат обрабатываем без блокировки, чтобы не держать мьютекс во время сетевых операций
		game.Mutex.Unlock()
		if !player.Hacked {
			broadcastChatMessage(playerID, msg.Message)
		}

	case "scream_target":
		// Обработка для ролей "Крикун" и "Дитя цветов"
		if player.Role == "Крикун" {
			player.TargetedScreamerPlayer = msg.Target
			log.Printf("Screamer selected target: %s", msg.Target)
		} else if player.Role == "Дитя цветов" {
			player.TargetedSunFlowerPlayer = msg.Target
			log.Printf("FlowerChild selected target: %s", msg.Target)
		}
		game.Mutex.Unlock()
		broadcastGameStatus()

	default:
		game.Mutex.Unlock()
	}

}

func broadcastChatMessage(playerID, chatMessage string) {
	// Создаём сообщение
	msg := ChatMessage{
		PlayerID: playerID,
		Chat:     chatMessage,
	}
	// Сохраняем сообщение в истории (с защитой мьютекса, если требуется)
	game.Mutex.Lock()
	chatHistory = append(chatHistory, msg)
	game.Mutex.Unlock()

	// Сериализуем сообщение в JSON
	message, _ := json.Marshal(msg)
	// Рассылаем сообщение всем подключенным игрокам
	for _, player := range game.Players {
		if player.Conn != nil {
			if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send chat message to player %s: %v", player.ID, err)
			}
		}
	}
}

func startPhaseTimer(duration int, endPhaseFunc func()) {
	game.TimeRemaining = duration
	broadcastGameStatus() // Отправляем начальное значение таймера
	go func() {
		for game.TimeRemaining > 0 {
			time.Sleep(1 * time.Second)
			game.Mutex.Lock()
			game.TimeRemaining--
			game.Mutex.Unlock()
			broadcastGameStatus() // Обновляем таймер у всех клиентов
		}
		endPhaseFunc()
	}()
}

func broadcastGameStatus() {
	for _, player := range game.Players {
		// Базовый статус, который отправляется всем
		status := struct {
			Phase                   string          `json:"phase"`
			Players                 map[string]bool `json:"players"`
			Day                     int             `json:"day"`
			TargetedScreamPlayer    string          `json:"targeted_scream_player,omitempty"`
			TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player,omitempty"`
			TimeRemaining           int             `json:"time_remaining"`
			Votes                   map[string]int  `json:"votes"`
			PlayerVote              string          `json:"player_vote"`
		}{
			Phase: game.CurrentPhase,
			Players: func() map[string]bool {
				players := make(map[string]bool)
				for id, p := range game.Players {
					players[id] = p.IsAlive
				}
				return players
			}(),
			Day:           game.DayNumber,
			TimeRemaining: game.TimeRemaining,
			Votes:         game.Votes,
			PlayerVote:    player.VotedFor,
		}

		//log.Printf("Игрок %s голосует за %s", player.ID, player.VotedFor)

		// Добавляем информацию о цели только для "Крикуна"
		if player.Role == "Крикун" && player.TargetedScreamerPlayer != "" {
			status.TargetedScreamPlayer = player.TargetedScreamerPlayer
		}
		if player.Role == "Дитя цветов" && player.TargetedSunFlowerPlayer != "" {
			status.TargetedSunFlowerPlayer = player.TargetedSunFlowerPlayer
		}

		// Сериализация в JSON
		data, err := json.Marshal(status)
		if err != nil {
			log.Printf("Failed to marshal game status for player %s: %v", player.ID, err)
			continue
		}

		// Отправка данных игроку
		if player.Conn == nil {
			log.Printf("Соединение для игрока %s не установлено", player.ID)
			return
		}

		err = player.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Не удалось отправить статус игры игроку %s: %v", player.ID, err)
			// Здесь можно добавить дополнительную логику обработки, если нужно
		}
	}
}
