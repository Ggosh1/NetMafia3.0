package GameFiles

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
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
