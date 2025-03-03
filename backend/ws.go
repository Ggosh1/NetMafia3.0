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
	defer conn.Close()

	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		playerID = fmt.Sprintf("player-%d", len(game.Players)+1)
	}

	game.Mutex.Lock()
	if disconnectedPlayers[playerID] {
		game.Mutex.Unlock()
		log.Printf("Reject connection: player %s is marked as disconnected permanently", playerID)
		conn.WriteMessage(websocket.CloseMessage, []byte("You have left the game permanently"))
		return
	}
	if existing, ok := game.Players[playerID]; ok && existing.Conn != nil {
		game.Mutex.Unlock()
		log.Printf("Reject connection: player %s is already connected", playerID)
		conn.WriteMessage(websocket.CloseMessage, []byte("Player already connected"))
		return
	}

	player := &Player{
		ID:      playerID,
		Conn:    conn,
		IsAlive: true,
	}
	game.Players[playerID] = player

	playersSnapshot := make(map[string]bool)
	for id, p := range game.Players {
		if p.Conn != nil {
			playersSnapshot[id] = p.IsAlive
		}
	}
	game.Mutex.Unlock()

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

	broadcastPlayerList()

	log.Printf("Player %s connected. Total active players: %d", playerID, len(game.Players))
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %s disconnected: %v", playerID, err)
			game.Mutex.Lock()
			delete(game.Players, playerID)
			disconnectedPlayers[playerID] = true
			game.Mutex.Unlock()

			roomLock.Lock()
			for _, room := range rooms {
				if _, exists := room.Players[playerID]; exists {
					delete(room.Players, playerID)
					log.Printf("Player %s removed from room %s", playerID, room.ID)
				}
			}
			roomLock.Unlock()

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
	game.Mutex.Lock()
	game.Mutex.Unlock()

	var msg struct {
		Action  string `json:"action"`
		Target  string `json:"vote"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}

	player, exists := game.Players[playerID]
	if !exists || !player.IsAlive {
		return
	}

	if game.CurrentPhase == "day" && msg.Action == "vote" && !player.Hacked {
		if msg.Target == player.VotedFor {
			game.Votes[player.VotedFor]--
			if game.Votes[player.VotedFor] < 0 {
				game.Votes[player.VotedFor] = 0
			}
			player.VotedFor = ""
		} else {
			if player.VotedFor != "" {
				game.Votes[player.VotedFor]--
				if game.Votes[player.VotedFor] < 0 {
					game.Votes[player.VotedFor] = 0
				}
			}
			if _, ok := game.Players[msg.Target]; ok {
				player.VotedFor = msg.Target
				game.Votes[msg.Target]++
				log.Printf("Player %s voted for %s", playerID, msg.Target)
			}
		}
	} else if game.CurrentPhase == "night" && (player.Aura == "bad" || player.Role == "Провидец" ||
		player.Role == "Провидец ауры" || player.Role == "Доктор" || player.Role == "Хакер") &&
		msg.Action != "cancel_vote" {
		player.Action = msg.Target
		log.Printf("Player %s (%s) targets %s", playerID, player.Role, msg.Target)
	} else if msg.Action == "start_game" {
		log.Printf("Player %s requested to start the game", playerID)
		startGame(nil, nil)
	} else if msg.Action == "chat" && !player.Hacked {
		broadcastChatMessage(playerID, msg.Message)
	} else if game.CurrentPhase == "day" && msg.Action == "cancel_vote" && !player.Hacked {
		game.Votes[msg.Target]--
	} else if game.CurrentPhase == "night" && msg.Action == "cancel_vote" {
		player.Action = ""
	} else if player.Role == "Крикун" && msg.Action == "scream_target" {
		game.Mutex.Lock()
		player.TargetedScreamerPlayer = msg.Target
		game.Mutex.Unlock()
		log.Printf("Screamer selected target: %s", msg.Target)
		broadcastGameStatus()
	} else if player.Role == "Дитя цветов" && msg.Action == "scream_target" {
		game.Mutex.Lock()
		player.TargetedSunFlowerPlayer = msg.Target
		game.Mutex.Unlock()
		log.Printf("FlowerChild selected target: %s", msg.Target)
		broadcastGameStatus()
	}
}

func broadcastChatMessage(playerID, chatMessage string) {
	message, _ := json.Marshal(struct {
		PlayerID string `json:"playerID"`
		Chat     string `json:"chat"`
	}{
		PlayerID: playerID,
		Chat:     chatMessage,
	})

	for _, player := range game.Players {
		if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send chat message to player %s: %v", player.ID, err)
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
		}

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
		err = player.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Failed to send game status to player %s: %v", player.ID, err)
		}
	}
}
