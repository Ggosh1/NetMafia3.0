package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)
import _ "net/http/pprof"

type Player struct {
	ID                      string          `json:"id"`
	Conn                    *websocket.Conn `json:"-"`
	Role                    string          `json:"role"`
	IsAlive                 bool            `json:"is_alive"`
	VotedFor                string          `json:"voted_for"`
	Action                  string          `json:"action"` // Used for night actions
	Aura                    string          `json:"aura"`
	TargetedScreamerPlayer  string          `json:"targeted_screamer_player"`
	TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player"`
}

type Game struct {
	Players       map[string]*Player
	Roles         []string
	Phase         string
	Votes         map[string]int
	Mutex         sync.Mutex
	GameStarted   bool
	CurrentPhase  string
	DayNumber     int
	TimeRemaining int
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var game = Game{
	Players: make(map[string]*Player),
	Votes:   make(map[string]int),
	Roles:   []string{"mafia", "detective", "villager", "villager"}, // Example roles
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	//http.HandleFunc("/start", startGame)
	http.HandleFunc("/status", gameStatus)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server started on :8080")
	log.Println(http.ListenAndServe(":8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var player Player
	player.ID = r.URL.Query().Get("id")
	if player.ID == "" {
		player.ID = fmt.Sprintf("player-%d", len(game.Players)+1)
	}
	player.Conn = conn
	player.IsAlive = true

	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.Players[player.ID] = &player
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")

	log.Printf("Player %s connected. Total players: %d", player.ID, len(game.Players))

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Player %s disconnected.", player.ID)
			game.Mutex.Lock()
			//log.Println("Mutex Locked")
			delete(game.Players, player.ID)
			game.Mutex.Unlock()
			//log.Println("Mutex UNLocked")
			break
		}

		log.Printf("Message from %s: %s", player.ID, string(message))
		processMessage(player.ID, message)
	}
}

func startGame(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")

	if game.GameStarted {
		http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	if len(game.Players) < 4 {
		http.Error(w, "Not enough players to start the game", http.StatusBadRequest)
		return
	}
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
		if player.Role == "Альфа оборотень" || player.Role == "Волчий провидец" || player.Role == "Малыш оборотень" || player.Role == "Волчий страж" {
			player.Aura = "bad"
		} else if player.Role == "Шут" || player.Role == "Хакер" || player.Role == "Тюремщик" || player.Role == "Линчеватель" {
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
	roles := []string{}

	//// Добавляем мафию (1 мафия на каждые 4 игрока)
	//mafiaCount := playerCount / 4
	//for i := 0; i < mafiaCount; i++ {
	//	roles = append(roles, "mafia")
	//}
	//
	//// Добавляем детектива (1 детектив на каждые 6 игроков)
	//if playerCount >= 6 {
	//	roles = append(roles, "detective")
	//}
	//
	//// Добавляем доктора, если игроков больше 5
	//if playerCount >= 5 {
	//	roles = append(roles, "doctor")
	//}
	//
	//// Остальные роли - мирные жители
	//villagerCount := playerCount - len(roles)
	//for i := 0; i < villagerCount; i++ {
	//	roles = append(roles, "villager")
	//}

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
	//log.Println("1")
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	//log.Println("2")
	game.CurrentPhase = "day"
	game.Votes = make(map[string]int)
	log.Println("Day phase started.")
	broadcastGameStatus() // Отправить клиентам обновление о фазе
	//log.Println("3")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")
	//log.Println("4")
	startPhaseTimer(30, endDayPhase)
}

func startNightPhase() {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.CurrentPhase = "night"
	log.Println("Night phase started.")
	broadcastGameStatus() // Отправить клиентам обновление о фазе
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")
	startPhaseTimer(30, func() {
		log.Println("Night phase timer ended.")
		processNightActions()
		endNightPhase()
	})
}

func processNightActions() {
	log.Println("Processing night actions...")

	// Собираем голоса (действия) только от оборотней
	werewolfVotes := make(map[string]int)
	nightActions := make(map[string]string)
	game.Mutex.Lock()
	log.Println("#5")
	for _, player := range game.Players {
		if player.Action != "" && player.IsAlive {
			nightActions[player.ID] = player.Action
			log.Println("####!!!", player.ID, player.Action)
			log.Println("#6")
		}
		player.Action = "" // Reset actions after processing
	}

	aliveWerewolves := 0
	for _, player := range game.Players {
		if player.IsAlive && player.Aura == "bad" {
			aliveWerewolves++
		}
	}
	doctorTarget := ""
	game.Mutex.Unlock()
	log.Println("#7")
	// Mafia's action: eliminate a player
	for id, targetID := range nightActions {
		p := game.Players[id]
		// Если aura=bad и игрок жив, учитываем его голос
		log.Println("####id-targetid", id, targetID)
		if p != nil && p.IsAlive && p.Aura == "bad" {
			werewolfVotes[targetID]++
			log.Println("####", targetID, werewolfVotes[targetID])
		}
		if p != nil && p.IsAlive && p.Role == "Доктор" {
			doctorTarget = targetID
			log.Println("####doctorTarget", doctorTarget)

		}
	}

	// С threshold определяем, сколько нужно голосов
	// Для упрощения логики используем округление вверх: (aliveWerewolves/2 + 1), если нечётно
	voteThreshold := aliveWerewolves / 2
	if aliveWerewolves%2 != 0 {
		voteThreshold = aliveWerewolves/2 + 1
	}

	// Определяем лидера голосования среди оборотней
	maxVotes := 0
	var candidates []string
	for targetID, count := range werewolfVotes {
		if count > maxVotes {
			maxVotes = count
			candidates = []string{targetID}
		} else if count == maxVotes {
			candidates = append(candidates, targetID)
		}
	}

	log.Printf("[Night] Werewolf votes: %v, threshold=%d, maxVotes=%d, candidates=%v",
		werewolfVotes, voteThreshold, maxVotes, candidates,
	)

	// Убийство совершается, только если:
	// 1) Есть ровно один лидер (candidates имеет длину 1)
	// 2) Лидер набрал >= порога
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

	// Detective's action: check a player's role
	for id, action := range nightActions {
		if game.Players[id].Role == "Провидец" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Detective checked player %s, role: %s", target.ID, target.Role)
				message := fmt.Sprintf("Player %s is %s", target.ID, target.Role)
				log.Printf("Sending message to detective %s: %s", id, message)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Role,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
			}
		}
		if game.Players[id].Role == "Провидец ауры" {
			if target, exists := game.Players[action]; exists {
				log.Printf("Aura seer checked player %s, aura: %s", target.ID, target.Aura)
				message := fmt.Sprintf("Player %s is %s", target.ID, target.Role)
				log.Printf("Sending message to detective %s: %s", id, message)
				teamCheckMessage, _ := json.Marshal(struct {
					Team string `json:"team"`
				}{
					Team: target.Aura,
				})
				game.Players[id].Conn.WriteMessage(websocket.TextMessage, teamCheckMessage)
			}
		}
	}
}

func endDayPhase() {
	log.Println("Ending day phase. Processing votes...")
	processVotes()

	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false // Останавливаем игру
		return
	}

	startNightPhase()
}

func endNightPhase() {
	log.Println("Ending night phase. Starting new day.")

	if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false // Останавливаем игру
		return
	}

	game.DayNumber++
	startDayPhase()
}

func checkGameOver() (bool, string) {
	aliveMafia := 0
	aliveVillagers := 0

	for _, player := range game.Players {
		if player.IsAlive {
			if player.Aura == "bad" {
				aliveMafia++
			} else {
				aliveVillagers++
			}
		}
	}

	if aliveMafia == 0 {
		return true, "Villagers win!"
	}

	if aliveMafia >= aliveVillagers {
		return true, "Mafia wins!"
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
		if err := player.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send winner message to player %s: %v", player.ID, err)
		}
	}
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

func processMessage(playerID string, message []byte) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	game.Mutex.Unlock()
	//log.Println("Mutex UNLocked")

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

	if game.CurrentPhase == "day" && msg.Action == "vote" {
		if player.VotedFor != "" {
			// Удаляем предыдущий голос
			game.Votes[player.VotedFor]--
		}
		if _, exists := game.Players[msg.Target]; exists {
			player.VotedFor = msg.Target
			log.Printf("Player %s voted for %s", playerID, msg.Target)
			game.Votes[msg.Target]++
		}
	} else if game.CurrentPhase == "night" && (player.Aura == "bad" || player.Role == "Провидец" || player.Role == "Провидец ауры" || player.Role == "Доктор") && msg.Action != "cancel_vote" {
		player.Action = msg.Target
		log.Printf("Player %s (%s) targets %s", playerID, player.Role, msg.Target)
	} else if msg.Action == "start_game" {
		log.Printf("Player %s requested to start the game", playerID)
		startGame(nil, nil) // Запуск игры
	} else if msg.Action == "chat" {
		broadcastChatMessage(playerID, msg.Message)
	} else if game.CurrentPhase == "day" && msg.Action == "cancel_vote" {
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

func processVotes() {
	flowerTarget := ""
	// Подсчет количества живых игроков
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
	}

	// Порог голосов для исключения
	voteThreshold := calculateVoteThreshold(alivePlayers)

	// Подсчет голосов
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

	// Проверка, есть ли кандидат с достаточным количеством голосов
	if maxVotes >= voteThreshold && len(candidates) == 1 {
		excludedPlayerID := candidates[0]
		if player, exists := game.Players[excludedPlayerID]; exists {
			flag := true
			if player.ID == flowerTarget {
				broadcastChatMessage("[SERVER]", fmt.Sprintf("Этого игрока нельзя казнить сегодня."))
				flag = false
			} else if player.Role == "Шут" {
				broadcastWinner("Шут победил!")
				game.GameStarted = false // Останавливаем игру
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

	// Очистка голосов
	game.Votes = make(map[string]int)

	// Обновление статуса игры для всех игроков
	broadcastGameStatus()
}

func calculateVoteThreshold(alivePlayers int) int {
	if alivePlayers%2 == 0 {
		return alivePlayers / 2
	}
	return alivePlayers/2 + 1
}

func gameStatus(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
	//log.Println("Mutex Locked")
	//game.Mutex.Unlock()
	log.Println("Mutex UNLocked")

	status, _ := json.Marshal(struct {
		Phase   string          `json:"phase"`
		Players map[string]bool `json:"players"`
		Day     int             `json:"day"`
	}{
		Phase: game.CurrentPhase,
		Players: func() map[string]bool {
			players := make(map[string]bool)
			for id, player := range game.Players {
				players[id] = player.IsAlive
			}
			return players
		}(),
		Day: game.DayNumber,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write(status)
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

		// Когда таймер истекает, вызываем переданную функцию (завершение фазы)
		endPhaseFunc()
	}()
}
