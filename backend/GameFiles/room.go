package GameFiles

import "fmt"

// Room представляет игровую комнату с идентификатором и игрой
type Room struct {
	ID      string
	Game    *Game
	Players map[string]*Player
}

// NewRoom создаёт новую комнату с заданным идентификатором и инициализирует карту игроков
func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		Players: make(map[string]*Player),
		Game:    NewGame(),
	}
}

// AddPlayer добавляет игрока в комнату.
// Если игрок с таким ID уже присутствует, выводится сообщение.
func (r *Room) AddPlayer(player *Player) {
	if r.Game.Players == nil {
		r.Game.Players = make(map[string]*Player)
	}
	if _, exists := r.Game.Players[player.ID]; exists {
		fmt.Printf("Игрок %s уже присутствует в комнате %s\n", player.ID, r.ID)
		return
	}
	r.Game.Players[player.ID] = player
	fmt.Printf("Игрок %s добавлен в комнату %s\n", player.ID, r.ID)
}

// RemovePlayer удаляет игрока из комнаты по его идентификатору.
// Если игрок не найден, выводится сообщение.
func (r *Room) RemovePlayer(playerID string) {
	if _, exists := r.Game.Players[playerID]; !exists {
		fmt.Printf("Игрок %s не найден в комнате %s\n", playerID, r.ID)
		return
	}
	r.Game.RemovePlayer(playerID)
	fmt.Printf("Игрок %s удалён из комнаты %s\n", playerID, r.ID)
}

func (r *Room) GetPlayerNames() []string {
	var playerNames []string
	for _, player := range r.Game.Players {
		playerNames = append(playerNames, player.ID)
	}
	return playerNames
}
