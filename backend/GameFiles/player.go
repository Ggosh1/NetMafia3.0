package GameFiles

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Team string

const (
	mafia    Team = "mafia"
	villager Team = "villager"
	solo     Team = "solo"
)

type Aura string

const (
	good    Aura = "good"
	evil    Aura = "evil"
	unknown Aura = "unknown"
)

type SpeakArea string

const (
	all    SpeakArea = "all"
	nobody SpeakArea = "nobody"
	wolfs  SpeakArea = "wolfs"
	prison SpeakArea = "prison"
)

type DieType string

const (
	voting      DieType = "voting"
	mafiaVoting DieType = "mafiaVoting"
	hack        DieType = "hack"
	jail        DieType = "jail"
)

// Базовая структура игрока
type PlayerInfo struct {
	ID           string
	InRoom       bool
	RoomID       string
	Conn         *websocket.Conn
	IsAlive      bool
	Target       string
	VotedFor     string // ID выбранного игрока
	Mutex        sync.Mutex
	ChatHistory  []ChatMessage
	CanStartGame bool
	IsProtected  bool
	IsHacked     bool
	IsJailed     bool
	NeedExecute  bool
	diedBy       DieType
}

func NewPlayerInfo(playerID string) *PlayerInfo {
	return &PlayerInfo{
		ID:           playerID,
		InRoom:       false,
		RoomID:       "",
		IsAlive:      false,
		VotedFor:     "",
		ChatHistory:  []ChatMessage{},
		Target:       "",
		CanStartGame: false,
		IsProtected:  false,
		IsHacked:     false,
		IsJailed:     false,
		NeedExecute:  false,
		diedBy:       "",
	}
}

func (p *Player) ResetPlayer() {
	p.InRoom = false
	p.RoomID = ""
	p.IsAlive = false
	p.Target = ""
	p.VotedFor = ""
	p.ChatHistory = []ChatMessage{}
	p.CanStartGame = false
	p.IsProtected = false
	p.IsHacked = false
	p.IsJailed = false
	p.NeedExecute = false
	p.diedBy = ""
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
func (p *Player) Die(dieType DieType) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	if p.IsProtected && (dieType == mafiaVoting || dieType == voting) {
		log.Printf("Player %s has protected\n", p.ID)
		return
	}

	p.IsAlive = false
	p.diedBy = dieType
	log.Printf("Player %s has died\n", p.ID)
}

type Player struct {
	*PlayerInfo
	Role
}

func NewPlayer(playerID string) *Player {
	return &Player{
		PlayerInfo: NewPlayerInfo(playerID),
		Role:       &SpectatorRole{},
	}
}

func (p *Player) LeaveRoom(role Role) {
	p.ResetPlayer()
}
func (p *Player) JoinRoom(roomID string) error {
	if roomID == "" {
		return errors.New("roomID не может быть пустым")
	}

	if p.InRoom && p.RoomID != roomID {
		return errors.New("игрок уже находится в комнате")
	}

	if p.RoomID == "" {
		// Если ResetPlayer изменяет поля, связанные с состоянием игрока,
		// его следует вызывать под мьютексом.
		p.ResetPlayer()

		p.RoomID = roomID
		p.InRoom = true
	}

	return nil
}

func (p *Player) AddChatMessage(message ChatMessage) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.ChatHistory = append(p.ChatHistory, message)
}

func (p *Player) ChooseTarget(id string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	if p.Target == id {
		p.Target = ""
	} else {
		p.Target = id

	}
}
