package GameFiles

import "log"

// Роль доктора – защищает цель, устанавливая isProtected в true.
type FlowerChildRole struct{}

func (f *FlowerChildRole) HaveNightAction() bool {
	return false
}

func (f *FlowerChildRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("FlowerChild does nothing at night \n")
}

func (f *FlowerChildRole) HaveDayAction() bool {
	return true
}

func (f *FlowerChildRole) DayAction(owner, target *Player, game *Game) {
	// Используем мьютекс целевого игрока для безопасного изменения состояния.
	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	target.IsProtected = true
	log.Printf("FlowerChild %s protects %s\n", owner.ID, target.ID)
}

func (f *FlowerChildRole) GetRussianName() string { return "Дитя цветов" }

func (f *FlowerChildRole) GetTeam() Team { return villager }

func (f *FlowerChildRole) NeedTarget(phase Phase) bool { return phase == day }

func (f *FlowerChildRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (f *FlowerChildRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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
			return nobody
		}
	}
}

func (f *FlowerChildRole) CanExecuteAction(p *Player) bool {
	return false
}
func (f *FlowerChildRole) ExecuteAction(p *Player) {
	log.Printf("FlowerChild %s does nothing\n", p.ID)
}

func (f *FlowerChildRole) IsSoloKiller() bool {
	return false
}

func (f *FlowerChildRole) GetAura() Aura {
	return good
}
