package GameFiles

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Game struct {
	Players       map[string]*Player
	CurrentPhase  Phase
	TimeRemaining int
	GameStarted   bool
	Mutex         sync.Mutex
}

func NewGame() *Game {
	return &Game{
		Players:      make(map[string]*Player),
		CurrentPhase: "",
	}
}

func (g *Game) StartGame(playerID string) {
	log.Println("Обработка запроса на запуск игры")

	log.Println("0тест")

	g.Mutex.Lock()
	g.Mutex.Unlock()

	log.Println("1тест")

	if g.GameStarted {
		errorMessage, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: "Game already started",
		})
		g.Players[playerID].Conn.WriteMessage(websocket.TextMessage, errorMessage)
		//http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	log.Println("2тест")

	if len(g.Players) < 4 {
		//log.Println("зашел сюда")
		errorMessage, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{
			Error: "Not enough players to start the GameFiles",
		})
		g.Players[playerID].Conn.WriteMessage(websocket.TextMessage, errorMessage)
		//http.Error(w, "Not enough players to start the GameFiles", http.StatusBadRequest)
		return
	}

	log.Println("3тест")

	//game.Roles = generateRoles(len(game.Players))
	log.Println("Starting GameFiles...")
	//assignRoles()

	for _, player := range g.Players {
		player.IsAlive = true
	}
	g.assignRoles()

	g.GameStarted = true
	g.StartDayPhase()
}

func (g *Game) startPhaseTimer(duration int, endPhaseFunc func()) {
	g.TimeRemaining = duration
	g.BroadcastGameStatusToAllPlayers() // Отправляем начальное значение таймера
	go func() {
		for g.TimeRemaining > 0 {
			time.Sleep(1 * time.Second)
			g.Mutex.Lock()
			g.TimeRemaining--
			g.Mutex.Unlock()
			g.BroadcastGameStatusToAllPlayers() // Обновляем таймер у всех клиентов
		}
		endPhaseFunc()
	}()
}

func (g *Game) ResetVotes() {
	for _, p := range g.Players {
		p.ResetVotedPlayer()
	}
}

func (g *Game) assignRoles() {
	roles := g.ShuffleRoles(g.GenerateRoles(len(g.Players)))
	index := 0
	for _, player := range g.Players {
		player.Role = roles[index]
		log.Printf("Assigned role %s to player %s", player.Role, player.ID)
		index++
	}
	g.BroadcastGameStatusToAllPlayers()
}

func (g *Game) GenerateRoles(playerCount int) []Role {
	var roles []Role
	switch playerCount {
	case 4:
		roles = []Role{&AlphaWolfRole{}, &SeerRole{}, &JesterRole{}, &ScreamerRole{}}
	}
	return roles
}

func (g *Game) ShuffleRoles(roles []Role) []Role {
	shuffled := make([]Role, len(roles))
	copy(shuffled, roles)
	for i := range shuffled {
		j := i + rand.Intn(len(shuffled)-i)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func (g *Game) RemovePlayer(playerID string) {
	player, exists := g.Players[playerID]
	if !exists {
		log.Printf("Игрок %s не найден в игре\n", playerID)
		return
	}

	// Закрываем соединение, если оно открыто
	if player.Conn != nil {
		player.Conn.Close()
	}
	player.InRoom = false
	player.RoomID = ""

	delete(g.Players, playerID)
	log.Printf("Игрок %s удалён из игры\n", playerID)
	// Отправляем обновлённый статус всем игрокам
	g.BroadcastGameStatusToAllPlayers()
}

func (g *Game) KillPlayer(playerID string, dieType DieType) {
	player, exists := g.Players[playerID]
	if !exists {
		log.Printf("Игрок %s не найден в игре\n", playerID)
		return
	}

	if _, ok := player.Role.(*ScreamerRole); ok {
		t, err := g.GetPlayer(player.Target)
		if err == nil {
			g.broadcastChatMessage(SERVER, fmt.Sprintf("Крикун %s погибает и раскрывает роль игрока %s - %s", player.ID, t.ID, t.Role.GetRussianName()))
		}
	}
	player.Die(dieType)
}
