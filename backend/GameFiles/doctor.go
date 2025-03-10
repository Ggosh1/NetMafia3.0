package GameFiles

import "log"

// Роль доктора – защищает цель, устанавливая isProtected в true.
type DoctorRole struct{}

func (d *DoctorRole) HaveNightAction() bool {
	return true
}

func (d *DoctorRole) NightAction(owner, target *Player, game *Game) {
	// Используем мьютекс целевого игрока для безопасного изменения состояния.
	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	target.IsProtected = true
	log.Printf("Doctor %s protects %s\n", owner.ID, target.ID)
}

func (d *DoctorRole) HaveDayAction() bool {
	return false
}

func (d *DoctorRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("Doctor does nothing ar day \n")
}

func (d *DoctorRole) GetRussianName() string { return "Доктор" }

func (d *DoctorRole) GetTeam() Team { return villager }

func (d *DoctorRole) NeedTarget(phase Phase) bool { return phase == night }

func (d *DoctorRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (d *DoctorRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (d *DoctorRole) CanExecuteAction(p *Player) bool {
	return false
}
func (d *DoctorRole) ExecuteAction(p *Player) {
	log.Printf("Doctor %s does nothing\n", p.ID)
}

func (d *DoctorRole) IsSoloKiller() bool {
	return false
}

func (d *DoctorRole) GetAura() Aura {
	return good
}
