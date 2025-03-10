package GameFiles

import "log"

// Роль волка – также не выполняет действий ночью.
type AlphaWolfRole struct{}

func (w *AlphaWolfRole) HaveNightAction() bool {
	return false
}

func (w *AlphaWolfRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("AlphaWolf %s does nothing at night\n", owner.ID)
}

func (w *AlphaWolfRole) HaveDayAction() bool {
	return false
}

func (w *AlphaWolfRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("AlphaWolf %s does nothing at day\n", owner.ID)
}

func (w *AlphaWolfRole) GetRussianName() string { return "Альфа-Волк" }

func (w *AlphaWolfRole) GetTeam() Team { return mafia }

func (w *AlphaWolfRole) NeedTarget(phase Phase) bool { return false }

func (w *AlphaWolfRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 2
	}
}

func (w *AlphaWolfRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (w *AlphaWolfRole) CanExecuteAction(p *Player) bool {
	return false
}
func (w *AlphaWolfRole) ExecuteAction(p *Player) {
	log.Printf("AlphaWolf %s does nothing\n", p.ID)
}

func (w *AlphaWolfRole) IsSoloKiller() bool {
	return false
}

func (w *AlphaWolfRole) GetAura() Aura {
	return evil
}
