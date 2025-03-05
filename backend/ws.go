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
		log.Printf("VOTE")
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

		} else if game.CurrentPhase == "night" {
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
				if player.Role == "Волчий провидец" {
					if target, exists := game.Players[msg.Target]; exists {
						if target.Aura == "bad" {
							log.Printf("Волчий провидец %s выбрал игрока %s, который является волком. Действие отклонено.", playerID, msg.Target)
						} else {
							if !player.CheckingWolfSeerUsed {
								log.Printf("Волчий провидец %s проверил игрока %s, роль: %s", playerID, target.ID, target.Role)
								message := fmt.Sprintf("Роль игрока %s: %s", target.ID, target.Role)
								teamCheckMessage, _ := json.Marshal(struct {
									Team string `json:"team"`
								}{
									Team: message,
								})
								player.Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
								// Разрешаем только одну проверку за ночь:
								player.Action = ""
								player.CheckingWolfSeerUsed = true
							}
						}
						game.Mutex.Unlock()
						broadcastGameStatus()
						return
					}
				} else {
					player.Action = msg.Target
					player.VotedFor = msg.Target
					log.Printf("Player %s targeted %s at night.", playerID, msg.Target)
				}
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
		log.Printf("SCREAM TARGET")
		// Обработка для ролей "Крикун" и "Дитя цветов"
		if player.Role == "Крикун" {
			player.TargetedScreamerPlayer = msg.Target
			log.Printf("Screamer selected target: %s", msg.Target)
		} else if player.Role == "Дитя цветов" {
			player.TargetedSunFlowerPlayer = msg.Target
			log.Printf("FlowerChild selected target: %s", msg.Target)
		} else if player.Role == "Медиум" {
			log.Printf("МЕДИУМ")
			targetPlayer, ok := game.Players[msg.Target]
			if ok && !targetPlayer.IsAlive {
				if !player.CheckingMediumUsed {
					log.Printf("НОРМ ЦЕЛЬ МЕДИУМ")
					player.TargetedMediumPlayer = msg.Target
					Message, _ := json.Marshal(struct {
						Error string `json:"message"`
					}{
						Error: msg.Target + " будет возрожден",
					})
					player.Conn.WriteMessage(websocket.TextMessage, Message)
					player.CheckingMediumUsed = true
					log.Printf("Medium selected target: %s", msg.Target)
				} else {
					Message, _ := json.Marshal(struct {
						Error string `json:"message"`
					}{
						Error: "Вы уже возродили игрока",
					})
					player.Conn.WriteMessage(websocket.TextMessage, Message)
				}
			} else {
				log.Printf("НЕНОРМ ЦЕЛЬ МЕДИУМ")
				Message, _ := json.Marshal(struct {
					Error string `json:"message"`
				}{
					Error: "Этого игрока нельзя возродить"})
				player.Conn.WriteMessage(websocket.TextMessage, Message)
			}
		}

		game.Mutex.Unlock()
		broadcastGameStatus()
	case "convert_to_werewolf":
		if player.Role == "Волчий провидец" {
			player.Role = "Обычный оборотень"
			player.Aura = "bad"
			log.Printf("Player %s converted to обычный оборотень", playerID)
			confirmation, _ := json.Marshal(struct {
				PlayerID string `json:"playerID"`
				Chat     string `json:"chat"`
			}{
				PlayerID: "[SERVER]",
				Chat:     "Вы стали обычным оборотнем",
			})
			player.Conn.WriteMessage(websocket.TextMessage, confirmation)
		}
		game.Mutex.Unlock()
		return
	case "jail_select":
		log.Printf("JAILER")
		// Действие для выбора цели тюремщиком в дневную фазу
		if player.Role == "Тюремщик" && game.CurrentPhase == "day" {
			if msg.Target == player.ID {
				log.Printf("Тюремщик не может выбрать себя.")
			} else if _, ok := game.Players[msg.Target]; ok {
				player.JailSelected = msg.Target
				log.Printf("Тюремщик %s выбрал игрока %s для ареста", player.ID, msg.Target)
				confirmation := map[string]string{
					"type":    "jail_select_confirm",
					"message": "Игрок " + msg.Target + " выбран для ареста.",
				}
				player.Conn.WriteJSON(confirmation)
			}
		}
		game.Mutex.Unlock()
		//broadcastGameStatus()
		return

	case "jail_kill":
		// Действие для убийства арестованного игрока ночью (один раз за игру)
		if player.Role == "Тюремщик" && game.CurrentPhase == "night" {
			if player.JailSelected == "" {
				log.Printf("Нет выбранного игрока для убийства.")
			} else if player.JailKillUsed {
				log.Printf("Способность убийства уже использована.")
			} else {
				// Фиксируем намерение убийства – сохраняем его в поле Action
				player.Action = "jail_kill"
				log.Printf("Тюремщик %s решил убить игрока %s", player.ID, player.JailSelected)
				confirmation := map[string]string{
					"type":    "jail_kill_confirm",
					"message": "Решение об убийстве игрока " + player.JailSelected + " зафиксировано.",
				}
				player.Conn.WriteJSON(confirmation)
			}
		}
		game.Mutex.Unlock()
		//broadcastGameStatus()
		return

	case "jail_chat":
		// Личный чат между тюремщиком и арестованным игроком
		// msg.Message содержит текст сообщения.
		var jailer *Player
		var arrestedID string
		// Ищем тюремщика, который уже выбрал цель
		for _, p := range game.Players {
			if p.Role == "Тюремщик" && p.JailSelected != "" {
				jailer = p
				arrestedID = p.JailSelected
				break
			}
		}
		if jailer == nil {
			log.Println("Нет активного тюремщика с выбранным игроком для чата.")
			return
		}
		// Разрешаем чат только для тюремщика и арестованного
		if player.ID != jailer.ID && player.ID != arrestedID {
			log.Println("Пользователь не имеет доступа к чату тюрьмы.")
			return
		}
		chatMsg := map[string]string{
			"type":    "jail_chat",
			"from":    player.ID,
			"message": msg.Message,
		}
		if jailer.Conn != nil {
			jailer.Conn.WriteJSON(chatMsg)
		}
		if arrested, ok := game.Players[arrestedID]; ok && arrested.Conn != nil {
			arrested.Conn.WriteJSON(chatMsg)
		}
		game.Mutex.Unlock()
		//broadcastGameStatus()
		return

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
