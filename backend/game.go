package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	//"net/http"

	"github.com/gorilla/websocket"
)

var PHASE_TIME int = 10

// запускает игру, назначает роли и начинает фазу дня.
func startGame(playerID string) {
	//log.Println("Обработка запроса на запуск игры")

	//log.Println("0тест")

	game.Mutex.Lock()
	game.Mutex.Unlock()

	//log.Println("1тест")

	if game.GameStarted {
		errorMessage, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: "Game already started",
		})
		game.Players[playerID].Conn.WriteMessage(websocket.TextMessage, errorMessage)
		//http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	//log.Println("2тест")

	if len(game.Players) < 4 {
		//log.Println("зашел сюда")
		errorMessage, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: "Not enough players to start the game",
		})
		game.Players[playerID].Conn.WriteMessage(websocket.TextMessage, errorMessage)
		//http.Error(w, "Not enough players to start the game", http.StatusBadRequest)
		return
	}

	//log.Println("3тест")

	game.Roles = generateRoles(len(game.Players))
	log.Println("Starting game...")
	assignRoles()
	game.GameStarted = true
	game.DayNumber = 1
	startDayPhase()
}

func assignRoles() {
	roles := shuffleRoles(game.Roles)
	index := 0
	for _, player := range game.Players {
		player.Role = roles[index]
		if player.Role == "Альфа оборотень" || player.Role == "Волчий провидец" ||
			player.Role == "Малыш оборотень" || player.Role == "Волчий страж" {
			player.Aura = "bad"
		} else if player.Role == "Шут" || player.Role == "Хакер" ||
			player.Role == "Тюремщик" || player.Role == "Медиум" {
			player.Aura = "unknown"
		} else {
			player.Aura = "good"
		}
		index++
		log.Printf("Assigned role %s to player %s", player.Role, player.ID)
	}
	broadcastRoles()
}

func generateRoles(playerCount int) []string {
	var roles []string
	switch playerCount {
	case 4:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор"}
	case 5:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун"}
	case 6:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов"}
	case 7:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер"}
	case 8:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец"}
	case 9:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум"}
	case 10:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик"}
	case 11:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель"}
	case 12:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень"}
	case 13:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры"}
	case 14:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей"}
	case 15:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей", "Купидон"}
	case 16:
		roles = []string{"Альфа оборотень", "Провидец", "Шут", "Доктор", "Крикун", "Дитя цветов", "Хакер", "Волчий провидец", "Медиум", "Тюремщик", "Линчеватель", "Малыш оборотень", "Провидец ауры", "Охотник на зверей", "Купидон", "Волчий страж"}
	}
	return roles
}

func shuffleRoles(roles []string) []string {
	shuffled := make([]string, len(roles))
	copy(shuffled, roles)
	for i := range shuffled {
		j := i + rand.Intn(len(shuffled)-i)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func broadcastRoles() {
	for _, player := range game.Players {
		roleMessage, _ := json.Marshal(struct {
			Role string `json:"role"`
		}{
			Role: player.Role,
		})
		player.Conn.WriteMessage(websocket.TextMessage, roleMessage)
	}
}

func startDayPhase() {
	game.Mutex.Lock()
	game.CurrentPhase = "day"
	game.Votes = make(map[string]int)
	log.Println("Day phase started.")
	broadcastGameStatus() // Рассылаем обновление о фазе всем клиентам
	game.Mutex.Unlock()

	startPhaseTimer(10, endDayPhase)
}

func startNightPhase() {
	game.Mutex.Lock()
	game.CurrentPhase = "night"
	log.Println("Night phase started.")
	broadcastGameStatus() // Рассылаем обновление о фазе всем клиентам
	game.Mutex.Unlock()
	startPhaseTimer(10, func() {
		log.Println("Night phase timer ended.")
		processNightActions()
		endNightPhase()
	})
}

func processNightActions() {
	log.Println("Processing night actions...")
	// Блокируем ночные действия у арестованных игроков
	for _, player := range game.Players {
		if player.Arrested {
			player.Action = ""
		}
	}

	// Собираем действия игроков
	werewolfVotes := make(map[string]int)
	nightActions := make(map[string]string)
	mediumTarget := ""
	game.Mutex.Lock()
	log.Println("#5")
	for _, player := range game.Players {
		player.CheckingWolfSeerUsed = false
		if player.Action != "" && player.IsAlive {
			nightActions[player.ID] = player.Action
			log.Println("####!!!", player.ID, player.Action)
			log.Println("#6")
		}

		if player.Role == "Медиум" {
			log.Println("МЕДИУМ111")
			if player.TargetedMediumPlayer != "" {
				log.Println("МЕДИУМ222")
				mediumTarget = player.TargetedMediumPlayer
				log.Println(mediumTarget)
				player.TargetedMediumPlayer = ""
			}
		}
		if player.Role == "Тюремщик" && player.JailSelected != "" {
			// Получаем цель тюремщика
			log.Println("JAILER 111")
			target, exists := game.Players[player.JailSelected]
			if exists && target.IsAlive {
				log.Println("JAILER 222", player.Action, player.JailKillUsed)

				// Отмечаем цель как арестованную
				target.Arrested = true

				// Если тюремщик решил убить арестованного игрока (action == "jail_kill")
				if player.Action == "jail_kill" && !player.JailKillUsed {
					log.Println("JAILER 333")
					target.IsAlive = false
					player.JailKillUsed = true
				}
			}
			// Сбрасываем выбор и действие тюремщика для следующей ночи
			player.JailSelected = ""
			player.Action = ""
		}
		player.Action = "" // Сбрасываем действия после обработки
		player.VotedFor = ""
	}

	aliveWerewolves := 0
	for _, player := range game.Players {
		if player.IsAlive && player.Aura == "bad" {
			aliveWerewolves++
		}
	}
	doctorTarget := ""
	hackerTarget := ""
	game.Mutex.Unlock()
	log.Println("#7")
	// Обработка действий
	for id, targetID := range nightActions {
		p := game.Players[id]
		log.Println("####id-targetid", id, targetID)
		if p != nil && p.IsAlive && p.Aura == "bad" {
			werewolfVotes[targetID]++
			log.Println("####", targetID, werewolfVotes[targetID])
		}
		if p != nil && p.IsAlive && p.Role == "Доктор" {
			doctorTarget = targetID
			log.Println("####doctorTarget", doctorTarget)
		}
		if p != nil && p.IsAlive && p.Role == "Хакер" {
			hackerTarget = targetID
			log.Printf("Hacker targeted %s", hackerTarget)
		}
	}

	if hackerTarget != "" {
		if target, exists := game.Players[hackerTarget]; exists {
			target.Hacked = true
			log.Printf("Player %s has been hacked and will lose voting/chat rights", target.ID)
			message, _ := json.Marshal(struct {
				PlayerID string `json:"playerID"`
				Chat     string `json:"chat"`
			}{
				PlayerID: "[SERVER]",
				Chat:     "Вы были взломаны! Вы не можете голосовать и писать в чат. Вы погибните в конце дня.",
			})
			target.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}

	voteThreshold := aliveWerewolves / 2
	// Определяем жертву
	maxVotes := 0
	candidates := []string{}
	for targetID, count := range werewolfVotes {
		if count > maxVotes {
			maxVotes = count
			candidates = []string{targetID}
		} else if count == maxVotes {
			candidates = append(candidates, targetID)
		}
	}
	if len(candidates) == 1 && maxVotes >= voteThreshold {
		targetID := candidates[0]
		if target, ok := game.Players[targetID]; ok && target.IsAlive && targetID != doctorTarget {
			// убиваем цель
			if target.Role == "Крикун" && target.TargetedScreamerPlayer != "" {
				// эффект Крикуна – раскрытие роли выбранного игрока
				if scrTarget, ok2 := game.Players[target.TargetedScreamerPlayer]; ok2 {
					broadcastChatMessage("[SERVER]",
						fmt.Sprintf("Крикун погиб и раскрывает роль игрока %s: %s",
							scrTarget.ID, scrTarget.Role))
				}
			}
			target.IsAlive = false
			log.Printf("[Night] Werewolves killed player %s", targetID)
		}
	} else {
		log.Println("[Night] No one was killed by werewolves this night.")
	}

	log.Printf("[Night] Werewolf votes: %v, threshold=%d, maxVotes=%d, candidates=%v",
		werewolfVotes, voteThreshold, maxVotes, candidates,
	)

	if len(candidates) == 1 && maxVotes >= voteThreshold {
		targetID := candidates[0]
		targetPlayer, ok := game.Players[targetID]
		if ok && targetPlayer.IsAlive && targetID != doctorTarget {
			if targetPlayer.Role == "Крикун" {
				log.Println("##Крикун1")
				if targetPlayer.TargetedScreamerPlayer != "" {
					targetPlayer := game.Players[targetPlayer.TargetedScreamerPlayer]
					if targetPlayer != nil {
						log.Println("##Крикун2")
						broadcastChatMessage("[SERVER]", fmt.Sprintf("Крикун раскрыл роль игрока %s - %s", targetPlayer.ID, targetPlayer.Role))
					}
				}
			}
			targetPlayer.IsAlive = false
			log.Printf("[Night] Werewolves killed player %s", targetID)
		}
	} else {
		log.Println("[Night] No one was killed by werewolves this night.")
	}

	// Действия Провидца и Провидца ауры
	for id, action := range nightActions {
		if game.Players[id].Role == "Провидец" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Detective checked player %s, role: %s", target.ID, target.Role)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Role,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
				log.Printf("сообщение игроку %s отправлено", id)
			}
		}
		if game.Players[id].Role == "Провидец ауры" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Aura seer checked player %s, aura: %s", target.ID, target.Aura)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Aura,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
			}
		}
	}
	if mediumTarget != "" {
		log.Printf("MEDIUM TARGET %s", mediumTarget)
		targetPlayer, ok := game.Players[mediumTarget]
		ok = ok
		if !targetPlayer.IsAlive {
			targetPlayer.IsAlive = true
			broadcastChatMessage("[SERVER]", fmt.Sprintf("Игрок %s был возрожден", targetPlayer.ID))
		}
		mediumTarget = ""
	}

}

func endDayPhase() {
	log.Println("Ending day phase. Processing votes...")
	processVotes()
	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false
		return
	}
	startNightPhase()
}

func endNightPhase() {
	log.Println("Ending night phase. Starting new day.")
	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false
		return
	}
	game.DayNumber++
	startDayPhase()
}

func checkGameOver() (bool, string) {
	aliveMafia := 0
	aliveVillagers := 0
	hackerAlive := false
	for _, player := range game.Players {
		if player.IsAlive {
			if player.Role == "Хакер" {
				hackerAlive = true
			}
			if player.Aura == "bad" {
				aliveMafia++
			} else {
				aliveVillagers++
			}
		}
	}

	if aliveMafia == 0 && !hackerAlive {
		return true, "Villagers win!"
	}

	if aliveMafia >= aliveVillagers && !hackerAlive {
		return true, "Mafia wins!"
	}
	if hackerAlive && aliveVillagers == 0 && aliveMafia == 0 {
		return true, "Hacker win"
	}

	return false, ""
}

func broadcastWinner(winner string) {
	message, _ := json.Marshal(struct {
		Winner string `json:"winner"`
	}{
		Winner: winner,
	})

	for _, player := range game.Players {
		player.IsAlive = true
		player.Hacked = false
		player.TargetedSunFlowerPlayer = ""
		player.TargetedScreamerPlayer = ""
		player.VotedFor = ""
		player.Action = ""
		player.Aura = ""
		player.Role = ""
		if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send winner message to player %s: %v", player.ID, err)
		}
	}
}

func processVotes() {
	flowerTarget := ""
	alivePlayers := 0
	for _, player := range game.Players {
		if player.IsAlive {
			alivePlayers++
		}
		if player.Role == "Дитя цветов" {
			if player.TargetedSunFlowerPlayer != "" {
				flowerTarget = game.Players[player.TargetedSunFlowerPlayer].ID
			}
		}
		if player.Hacked {
			player.IsAlive = false
			log.Printf("Player %s was killed by hacker", player.ID)
		}

		player.VotedFor = ""
	}

	voteThreshold := calculateVoteThreshold(alivePlayers)

	maxVotes := 0
	candidates := []string{}
	for playerID, votes := range game.Votes {
		if votes > maxVotes {
			maxVotes = votes
			candidates = []string{playerID}
		} else if votes == maxVotes {
			candidates = append(candidates, playerID)
		}
	}

	log.Printf("Vote threshold: %d, Max votes: %d, Candidates: %v", voteThreshold, maxVotes, candidates)

	if maxVotes >= voteThreshold && len(candidates) == 1 {
		excludedPlayerID := candidates[0]
		if player, exists := game.Players[excludedPlayerID]; exists {
			flag := true
			if player.ID == flowerTarget {
				broadcastChatMessage("[SERVER]", "Этого игрока нельзя казнить сегодня.")
				flag = false
			} else if player.Role == "Шут" {
				broadcastWinner("Шут победил!")
				game.GameStarted = false
				return
			} else if player.Role == "Крикун" {
				log.Println("##Крикун1")
				if player.TargetedScreamerPlayer != "" {
					targetPlayer := game.Players[player.TargetedScreamerPlayer]
					if targetPlayer != nil {
						log.Println("##Крикун2")
						broadcastChatMessage("[SERVER]", fmt.Sprintf("Крикун раскрыл роль игрока %s - %s", targetPlayer.ID, targetPlayer.Role))
					}
				}
			}
			if flag {
				player.IsAlive = false
				log.Printf("Player %s was voted out.", excludedPlayerID)
			}
		}
	} else {
		log.Println("No player was excluded.")
	}

	game.Votes = make(map[string]int)
	broadcastGameStatus()
}

func calculateVoteThreshold(alivePlayers int) int {
	if alivePlayers%2 == 0 {
		return alivePlayers / 2
	}
	return alivePlayers/2 + 1
}
