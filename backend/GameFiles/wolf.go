package GameFiles

import "log"

// Роль волка – также не выполняет действий ночью.
type WolfRole struct{}

func (w *WolfRole) HaveNightAction() bool {
	return false
}

func (w *WolfRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("Wolf %s does nothing at night\n", owner.ID)
}

func (w *WolfRole) GetRussianName() string { return "Волк" }

func (w *WolfRole) GetTeam() Team { return mafia }

func (w *WolfRole) NeedTarget(phase Phase) bool { return false }

func (w *WolfRole) VoteValue(phase Phase) int {
	return 1
}

func (w *WolfRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
	if phase == day {
		if p.IsHacked {
			return nobody
		} else {
			return all
		}
	} else {
		if p.IsJailed {
			return prison
		} else {
			return wolfs
		}
	}
}

func (w *WolfRole) CanExecuteAction(p *Player) bool {
	return false
}
func (w *WolfRole) ExecuteAction(p *Player) {
	log.Printf("Wolf %s does nothing\n", p.ID)
}

func (w *WolfRole) IsSoloKiller() bool {
	return false
}

func (w *WolfRole) GetAura() Aura {
	return evil
}
