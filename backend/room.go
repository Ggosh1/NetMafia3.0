package backend

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	rooms    = make(map[string]*Room) // все созданные комнаты
	roomLock sync.Mutex               // для синхронизации доступа к rooms
)

func generateRoomID() string {
	return fmt.Sprintf("room-%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
}

// joinRoom ищет свободную комнату (где количество игроков < 16) и выбирает ту, где уже больше всего игроков.
// Если ни одной такой комнаты нет, создаётся новая.
func joinRoom(p *Player) *Room {
	roomLock.Lock()
	defer roomLock.Unlock()

	var bestRoom *Room
	for _, room := range rooms {
		if len(room.Players) < 16 {
			if bestRoom == nil || len(room.Players) > len(bestRoom.Players) {
				bestRoom = room
			}
		}
	}
	if bestRoom == nil {
		bestRoom = &Room{
			ID:      generateRoomID(),
			Players: make(map[string]*Player),
		}
		rooms[bestRoom.ID] = bestRoom
		log.Printf("Создана новая комната: %s", bestRoom.ID)
	}
	bestRoom.Players[p.ID] = p
	log.Printf("Игрок %s добавлен в комнату %s (игроков: %d)", p.ID, bestRoom.ID, len(bestRoom.Players))
	return bestRoom
}
