package backend

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID                      string          `json:"id"`
	Conn                    *websocket.Conn `json:"-"`
	Role                    string          `json:"role"`
	IsAlive                 bool            `json:"is_alive"`
	VotedFor                string          `json:"voted_for"`
	Action                  string          `json:"action"` // Используется для ночных действий
	Aura                    string          `json:"aura"`
	TargetedScreamerPlayer  string          `json:"targeted_screamer_player"`
	TargetedSunFlowerPlayer string          `json:"targeted_sun_flower_player"`
	Hacked                  bool            `json:"hacked"`
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

type Room struct {
	ID      string             `json:"id"`
	Players map[string]*Player `json:"players"` // Ключ – ID игрока
}

// Глобальное состояние игры
var game = Game{
	Players: make(map[string]*Player),
	Votes:   make(map[string]int),
	// Изначальные роли (пример – впоследствии перезаписываются функцией generateRoles)
	Roles: []string{"mafia", "detective", "villager", "villager"},
}
