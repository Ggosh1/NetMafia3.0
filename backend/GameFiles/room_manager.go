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
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	room, exists := rm.Rooms[id]
	if !exists {
		return nil, fmt.Errorf("комната с id %s не найдена", id)
	}
	return room, nil
}

// DeleteRoom удаляет комнату из менеджера
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

	// 3. Попытка присоединить игрока к комнате (на уровне структуры Player)
	if err := player.JoinRoom(roomID); err != nil {
		return fmt.Errorf("не удалось присоединить игрока %s к комнате %s: %v", playerID, roomID, err)
	}

	// 4. Добавляем игрока в конкретную комнату (Room.AddPlayer)
	room.AddPlayer(player)

	return nil
}
