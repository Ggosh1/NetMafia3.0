package GameFiles

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Team string

const (
	mafia    Team = "mafia"
	villager Team = "villager"
	neutral  Team = "neutral"
)

// Интерфейс для ролевых действий
type Role interface {
	NightAction(players map[string]*Player) // Ночное действие
}

// Базовая структура игрока
type PlayerInfo struct {
	ID              string
	InRoom          bool
	RoomID          string
	Conn            *websocket.Conn
	IsAlive         bool
	HaveNightAction bool
	ChosenPlayer    string
	VotedFor        string // ID выбранного игрока
	Mutex           sync.Mutex
	Team            Team
	ChatHistory     []ChatMessage
	CanStartGame    bool
	IsProtected     bool
	IsHacked        bool
}

func NewPlayerInfo(playerID string) *PlayerInfo {
	return &PlayerInfo{
		ID:              playerID,
		InRoom:          false,
		RoomID:          "",
		IsAlive:         false,
		HaveNightAction: false,
		VotedFor:        "",
		Team:            neutral,
		ChatHistory:     []ChatMessage{},
		ChosenPlayer:    "",
		CanStartGame:    false,
		IsProtected:     false,
		IsHacked:        false,
	}
}

func (p *Player) ResetPlayer() {
	p.InRoom = false
	p.RoomID = ""
	p.IsAlive = false
	p.HaveNightAction = false
	p.VotedFor = ""
	p.Team = neutral
	p.ChatHistory = []ChatMessage{}
	p.ChosenPlayer = ""
	p.CanStartGame = false
	p.IsProtected = false
	p.IsHacked = false
}

// Метод выбора игрока
func (p *Player) VoteForPlayer(targetID string) {
	if p.ID == targetID {
		return
	}

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	if p.VotedFor == targetID {
		p.VotedFor = ""
	} else {
		p.VotedFor = targetID
	}

	fmt.Printf("Player %s chose %s\n", p.ID, targetID)
}

func (p *Player) ResetVotedPlayer() {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.VotedFor = ""
}

// Метод смерти игрока
func (p *Player) Die() {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.IsAlive = false
	fmt.Printf("Player %s has died\n", p.ID)
}

type Player struct {
	*PlayerInfo
	Role
}

func NewPlayer(playerID string) *Player {
	return &Player{
		PlayerInfo: NewPlayerInfo(playerID),
		Role:       nil,
	}
}

func (p *Player) LeaveRoom(role Role) {
	p.ResetPlayer()
}
func (p *Player) JoinRoom(roomID string) error {
	if roomID == "" {
		return errors.New("roomID не может быть пустым")
	}

	if p.InRoom {
		return errors.New("игрок уже находится в комнате")
	}

	// Если ResetPlayer изменяет поля, связанные с состоянием игрока,
	// его следует вызывать под мьютексом.
	p.ResetPlayer()

	p.RoomID = roomID
	p.InRoom = true

	return nil
}

func (p *Player) AddChatMessage(message ChatMessage) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.ChatHistory = append(p.ChatHistory, message)
}
