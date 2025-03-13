package GameFiles

import (
	"fmt"
	"log"
	"sync"
)

// RoomManager управляет созданными комнатами
type RoomManager struct {
	Rooms   map[string]*Room
	Players map[string]*Player
	Mutex   sync.Mutex
}

// создаёт новый менеджер комнат
func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms:   make(map[string]*Room),
		Players: make(map[string]*Player),
	}
}

// создаёт комнату с заданным ID.
// Если комната с таким ID уже существует, возвращается ошибка.
func (rm *RoomManager) CreateRoom(id string) (*Room, error) {
	log.Printf("Попытка создать комнату %s из room_manager.go", id)
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	if _, exists := rm.Rooms[id]; exists {
		log.Printf("комната с id %s уже существует", id)
		return nil, fmt.Errorf("комната с id %s уже существует", id)
	}
	room := NewRoom(id)
	rm.Rooms[id] = room
	log.Printf("Вроде создал комнату с id %s", id)

	return room, nil
}

// GetRoom возвращает комнату по ID, если она существует.
func (rm *RoomManager) GetRoom(id string) (*Room, error) {
	room, exists := rm.Rooms[id]
	if !exists {
		return nil, fmt.Errorf("комната с id %s не найдена", id)
	}
	return room, nil
}

// удаляет комнату из менеджера
func (rm *RoomManager) DeleteRoom(id string) error {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	if _, exists := rm.Rooms[id]; !exists {
		return fmt.Errorf("комната с id %s не существует", id)
	}
	delete(rm.Rooms, id)
	return nil
}

func (rm *RoomManager) CreatePlayer(id string) (*Player, error) {
	log.Printf("Попытка создать игрока %s из room_manager.go", id)
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	if _, exists := rm.Players[id]; exists {
		log.Printf("Игрок с id %s уже существует", id)
		return nil, fmt.Errorf("Игрок с id %s уже существует", id)
	}
	player := NewPlayer(id)
	rm.Players[id] = player
	log.Printf("Вроде создал игрока с id %s", id)

	return player, nil
}

func (rm *RoomManager) GetPlayer(id string) (*Player, error) {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	player, exists := rm.Players[id]
	if !exists {
		return nil, fmt.Errorf("игрок с id %s не найден", id)
	}
	return player, nil
}

func (rm *RoomManager) AddPlayerToRoom(roomID, playerID string) error {
	// 1. Проверка существования комнаты
	room, exists := rm.Rooms[roomID]
	if !exists {
		return fmt.Errorf("комната с id %s не найдена", roomID)
	}

	// 2. Проверка существования игрока
	player, exists := rm.Players[playerID]
	if !exists {
		return fmt.Errorf("игрок с id %s не найден", playerID)
	}

	if player.InRoom {
		return fmt.Errorf("Игрок уже в комнате")
	}

	if room.Game.GameStarted {

		return fmt.Errorf("Невозможно зайти в комнату, когда игра в ней уже началась")
	}

	player.ResetPlayer()

	// 3. Попытка присоединить игрока к комнате (на уровне структуры Player)
	if err := player.JoinRoom(roomID); err != nil {
		return fmt.Errorf("не удалось присоединить игрока %s к комнате %s: %v", playerID, roomID, err)
	}

	// 4. Добавляем игрока в конкретную комнату (Room.AddPlayer)
	room.AddPlayer(player)

	return nil
}

func (rm *RoomManager) RemovePlayerFromRoom(roomID, playerID string) {
	room, err := rm.GetRoom(roomID)
	if err != nil {
		log.Printf("Комната не найдена %s\n", roomID)
		return
	}

	if _, exists := room.Game.Players[playerID]; !exists {
		log.Printf("Игрок %s не найден в комнате %s\n", playerID, room.ID)
		return
	}
	room.Game.RemovePlayer(playerID)
	log.Printf("Игрок %s удалён из комнаты %s\n", playerID, room.ID)

	if len(room.Game.Players) == 0 {
		rm.DeleteRoom(roomID)
		log.Printf("Комната удалена %s\n", roomID)
	}
}

func (rm *RoomManager) GetRooms() []*Room {
	rooms := make([]*Room, 0, len(rm.Rooms))
	for _, room := range rm.Rooms {
		rooms = append(rooms, room)
	}
	return rooms

}
func (rm *RoomManager) GetRoomsList() []*Room {
	var rooms []*Room
	for _, room := range rm.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (rm *RoomManager) GetBestRoom() *Room {
	var bestRoom *Room = nil
	for _, room := range rm.Rooms {
		if len(room.Game.Players) == 16 {
			continue
		}

		if bestRoom == nil || len(room.Game.Players) > len(bestRoom.Game.Players) {
			bestRoom = room
		}
	}

	return bestRoom
}
