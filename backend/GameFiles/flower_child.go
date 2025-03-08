package GameFiles

import "log"

// Роль доктора – защищает цель, устанавливая isProtected в true.
type FlowerRole struct{}

func (f *FlowerRole) HaveNightAction() bool {
	return true
}

func (f *FlowerRole) NightAction(owner, target *Player, game *Game) {
	// Используем мьютекс целевого игрока для безопасного изменения состояния.
	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	target.IsProtected = true
	log.Printf("Doctor %s protects %s\n", owner.ID, target.ID)
}

func (f *FlowerRole) GetRussianName() string { return "Дитя цветов" }

func (f *FlowerRole) GetTeam() Team { return villager }

func (f *FlowerRole) NeedTarget(phase Phase) bool { return phase == night }

func (f *FlowerRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (f *FlowerRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (f *FlowerRole) CanExecuteAction(p *Player) bool {
	return false
}
func (f *FlowerRole) ExecuteAction(p *Player) {
	log.Printf("FlowerChild %s does nothing\n", p.ID)
}

func (f *FlowerRole) IsSoloKiller() bool {
	return false
}

func (f *FlowerRole) GetAura() Aura {
	return good
}
