package GameFiles

import "fmt"

// Волк – убийца ночью
type Wolf struct {
	*PlayerInfo
}

// Реализация ночного действия – убийство игрока
func (w *Wolf) NightAction(players map[string]*Player) {
	if !w.IsAlive || w.VotedFor == "" {
		return
	}

	target, exists := players[w.VotedFor]
	if exists && target.IsAlive {
		target.Die()
		fmt.Printf("Wolf %s killed Player %s\n", w.ID, w.VotedFor)
	} else {
		fmt.Println("Invalid target for Wolf action")
	}
}

// Мирный житель – без спецспособностей
type Villager struct {
	*Player
}

// Ночное действие у мирного отсутствует
func (v *Villager) NightAction(players map[string]*Player) {
	fmt.Printf("Villager %s does nothing at night\n", v.ID)
}
