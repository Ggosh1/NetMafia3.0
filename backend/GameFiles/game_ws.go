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

	status := struct {
		Phase                   string          `json:"phase"`
		Players                 map[string]bool `json:"players"`
		TargetedScreamPlayer    string          `json:"targeted_scream_player,omitempty"`
		TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player,omitempty"`
		TimeRemaining           int             `json:"time_remaining"`
		Votes                   map[string]int  `json:"votes"`
		PlayerVote              string          `json:"player_vote"`
		ChosenPlayer            string          `json:"player_choise"`
		CanStartGame            bool            `json:"can_start_game"`
		IsHacked                bool            `json:"hacked"`
	}{
		Phase:         string(g.CurrentPhase),
		Players:       alivePlayerInfo,
		TimeRemaining: g.TimeRemaining,
		PlayerVote:    player.VotedFor,
		Votes:         g.GetVotesMap(),
		ChosenPlayer:  player.ChosenPlayer,
		CanStartGame:  player.CanStartGame,
		IsHacked:      player.IsHacked,
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

	log.Printf("Сообщение от сервера игроку %s : %s", player.ID, status)

	err = player.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("Не удалось отправить статус игры игроку %s: %v", player.ID, err)
		// Здесь можно добавить дополнительную логику обработки, если нужно
	}
}

func (g *Game) BroadcastGameStatusToAllPlayers() {
	for _, player := range g.Players {
		g.BroadcastGameStatus(player.ID)
	}
}

func (g *Game) ProcessMessage(playerID string, message []byte) {
	var msg struct {
		Action  string `json:"action"`
		Target  string `json:"vote"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %s", err)
		return
	}

	player, exists := g.Players[playerID]

	if exists == false {
		log.Printf("Player %s not found in game", playerID)
		return
	}

	switch msg.Action {
	case "start_game":
		log.Printf("Player %s requested to start the GameFiles", playerID)
		go g.StartGame(playerID)
	case "chat":
		g.broadcastChatMessage(playerID, msg.Message)

	case "vote":
		player.VoteForPlayer(msg.Target)

		log.Printf("Итоговый голос игрока %s - %s", playerID, player.VotedFor)
		g.BroadcastGameStatusToAllPlayers()

	}

	return
	/*if msg.Action == "start_game" {
		log.Printf("Player %s requested to start the GameFiles", playerID)
		go StartGame(playerID)
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
			BroadcastGameStatusToAllPlayers()

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

			BroadcastGameStatusToAllPlayers()

			BroadcastGameStatusToAllPlayers()
		}

		game.Mutex.Unlock()

	/*case "cancel_vote":
	// Обработка отмены голоса в дневной фазе (если голос был поставлен)
	if GameFiles.CurrentPhase == "day" && !player.Hacked {
		// Обрабатываем отмену только если передан целевой ID совпадающий с текущим голосом
		if player.VotedFor != "" && msg.Target == player.VotedFor {
			GameFiles.Votes[player.VotedFor]--
			if GameFiles.Votes[player.VotedFor] < 0 {
				GameFiles.Votes[player.VotedFor] = 0
			}
			player.VotedFor = ""
		}
	} else if GameFiles.CurrentPhase == "night" {
		player.Action = ""
	}
	GameFiles.Mutex.Unlock()*/

	/*case "chat":
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
		BroadcastGameStatusToAllPlayers()

	default:
		game.Mutex.Unlock()
	}
	*/
}

/*func (g *Game) SendMessage(playerID string, message []byte) {
	log.Printf("Отправка сообщения пользователю %s", playerID)

	if player, exists := g.Players[playerID]; exists {
		if player.Conn == nil {
			log.Printf("Соединение с пользователем %s", playerID)

		}
	}
	else{
		log.Printf("Пользователя с таким id не существует: %s", playerID)
	}
	g.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
	log.Printf("сообщение игроку %s отправлено", id)
}*/
