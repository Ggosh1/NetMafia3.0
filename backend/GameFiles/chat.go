package GameFiles

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

const SERVER string = "[SERVER]"

type ChatMessage struct {
	PlayerID string `json:"playerID"`
	Chat     string `json:"chat"`
}

func (g *Game) broadcastChatMessage(playerID, chatMessage string) {

	var playerArea SpeakArea = all

	if playerID != SERVER {
		player, err := g.GetPlayer(playerID)
		if err != nil {
			log.Printf("Player with id %s not found: %v", playerID, err)
			return
		}

		playerArea = player.GetSpeakArea(player, g.CurrentPhase)
		if playerArea == nobody {
			return
		}
	}

	for _, recipient := range g.Players {
		if playerArea == all || recipient.GetSpeakArea(recipient, g.CurrentPhase) == playerArea {
			g.broadcastChatMessageToPlayer(playerID, recipient.ID, chatMessage)
		}
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
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Printf("Не удалось получить историю чата игрока")
		return nil
	}
	return player.ChatHistory
}
