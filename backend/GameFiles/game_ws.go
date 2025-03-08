package GameFiles

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

func (g *Game) BroadcastGameStatus(playerID string) {
	player, exists := g.Players[playerID]
	if exists == false {
		return
	}

	alivePlayerInfo := func() map[string]bool {
		players := make(map[string]bool)
		for id, p := range g.Players {
			players[id] = p.IsAlive
		}
		return players
	}()

	votes := g.GetVotesMap()
	if g.CurrentPhase == night && player.GetTeam() != mafia {
		votes = make(map[string]int)
	}

	_, winnermsg := g.CheckGameOver()

	status := struct {
		Phase                   string          `json:"phase"`
		Winner                  string          `json:"winner"`
		Players                 map[string]bool `json:"players"`
		TargetedScreamPlayer    string          `json:"targeted_scream_player,omitempty"`
		TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player,omitempty"`
		TimeRemaining           int             `json:"time_remaining"`
		Votes                   map[string]int  `json:"votes"`
		PlayerVote              string          `json:"player_vote"`
		Target                  string          `json:"target"`
		CanStartGame            bool            `json:"can_start_game"`
		IsHacked                bool            `json:"hacked"`
		PlayerRole              string          `json:"role"`
		HaveNightAction         bool            `json:"have_night_action"`
		NeedToChooseTarget      bool            `json:"need_to_choose_target"`
	}{
		Phase:         string(g.CurrentPhase),
		Winner:        winnermsg,
		Players:       alivePlayerInfo,
		TimeRemaining: g.TimeRemaining,
		PlayerVote:    player.VotedFor,
		Votes:         votes,
		Target:        player.Target,
		//CanStartGame:  player.CanStartGame,
		IsHacked:           player.IsHacked,
		PlayerRole:         player.GetRussianName(),
		HaveNightAction:    player.HaveNightAction(),
		NeedToChooseTarget: player.NeedTarget(g.CurrentPhase),
	}

	//log.Printf("Игрок %s голосует за %s", player.ID, player.VotedFor)

	// Добавляем информацию о цели только для "Крикуна"
	/*if player.Role == "Крикун" && player.TargetedScreamerPlayer != "" {
		status.TargetedScreamPlayer = player.TargetedScreamerPlayer
	}
	if player.Role == "Дитя цветов" && player.TargetedSunFlowerPlayer != "" {
		status.TargetedSunFlowerPlayer = player.TargetedSunFlowerPlayer
	}*/

	// Сериализация в JSON
	data, err := json.Marshal(status)
	if err != nil {
		log.Printf("Failed to marshal GameFiles status for player %s: %v", player.ID, err)
		return
	}

	// Отправка данных игроку
	if player.Conn == nil {
		log.Printf("Соединение для игрока %s не установлено", player.ID)
		return
	}

	//og.Printf("Сообщение от сервера игроку %s : %s", player.ID, status)

	err = player.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("Не удалось отправить статус игры игроку %s: %v", player.ID, err)
		// Здесь можно добавить дополнительную логику обработки, если нужно
	}
}

func (g *Game) BroadcastGameStatusToAllPlayers() {
	for _, player := range g.Players {
		if player.Conn == nil || !player.InRoom {
			continue
		}
		g.BroadcastGameStatus(player.ID)
	}
}

func (g *Game) ProcessMessage(playerID string, message []byte) bool {
	var msg struct {
		Action  string `json:"action"`
		Vote    string `json:"vote"`
		Target  string `json:"target"`
		Message string `json:"message"`
	}
	log.Printf("Message from %s", playerID)
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %s", err)
		return false
	}

	player, exists := g.Players[playerID]

	if exists == false {
		log.Printf("Player %s not found in game", playerID)
		return false
	}

	switch msg.Action {
	case "start_game":
		log.Printf("Player %s requested to start the GameFiles", playerID)
		go g.StartGame(playerID)
	case "chat":
		g.broadcastChatMessage(playerID, msg.Message)
	case "vote":
		log.Printf("Player %s voted for %s", playerID, msg.Vote)
		if g.PlayerCanVote(playerID) == false {
			return false
		}
		player.VoteForPlayer(msg.Vote)

		log.Printf("Итоговый голос игрока %s - %s", playerID, player.VotedFor)
		g.BroadcastGameStatusToAllPlayers()
	case "choose_target":
		log.Printf("Player %s targeted %s", playerID, msg.Target)
		player.ChooseTarget(msg.Target)
		g.BroadcastGameStatusToAllPlayers()
	case "leave_room":
		log.Printf("Player %s requested to leave the Game", playerID)
		return true
	}

	return false
}
