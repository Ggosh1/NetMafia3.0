package GameFiles

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

type ChatMessage struct {
	PlayerID string `json:"playerID"`
	Chat     string `json:"chat"`
}

func (g *Game) broadcastChatMessage(playerID, chatMessage string) {
	for _, player := range g.Players {
		g.broadcastChatMessageToPlayer(playerID, player.ID, chatMessage)
	}
}

func (g *Game) broadcastChatMessageToPlayer(fromID, toID, chatMessage string) {
	// Создаём сообщение
	msg := ChatMessage{
		PlayerID: fromID,
		Chat:     chatMessage,
	}

	toPlayer, err2 := g.GetPlayer(toID)

	message, _ := json.Marshal(msg)

	if err2 == nil {
		toPlayer.AddChatMessage(msg)
		if err := toPlayer.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send chat message to player %s: %v", toPlayer.ID, err)
		}
	}
}

func (g *Game) GetChatHistory(playerID string) []ChatMessage {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Printf("Не удалось получить историю чата игрока")
		return nil
	}
	return player.ChatHistory
}
