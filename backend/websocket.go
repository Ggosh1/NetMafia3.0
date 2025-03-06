package backend

import (
	"NetMafia3/backend/GameFiles"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		log.Println("id игрока не задан", err)
		return
	}

	player, err := roomManager.GetPlayer(playerID)

	if player == nil {
		log.Println("игрок с данным id не существует", err)
		return
	}

	if player.InRoom == false {
		log.Println("игрок с данным id не находится в комнате", err)
		return
	}

	roomManager.Mutex.Lock()
	log.Printf("Player %s connected", playerID)
	player.Conn = conn
	roomManager.Mutex.Unlock()
	var roomID string = player.RoomID
	room, err := roomManager.GetRoom(roomID)
	if err != nil {
		log.Printf("Ошибка %s", err)
		return
	}

	room.Game.BroadcastGameStatusToAllPlayers()
	room.Game.BroadcastGameStatusToAllPlayers()
	// Отправляем начальное состояние (например, список игроков)
	game.Mutex.Lock()
	// История чата
	if err := conn.WriteJSON(struct {
		Type    string                  `json:"type"`
		History []GameFiles.ChatMessage `json:"history"`
	}{
		Type:    "chatHistory",
		History: room.Game.GetChatHistory(playerID),
	}); err != nil {
		log.Printf("Ошибка отправки истории чата игроку %s: %v", playerID, err)
	}
	game.Mutex.Unlock()

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

			room.Game.BroadcastGameStatus(playerID)
			break
		}
		log.Printf("Message from %s: %s", playerID, string(message))

		room.Game.ProcessMessage(playerID, message)
	}
}
