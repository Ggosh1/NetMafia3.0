package GameFiles

import "log"

// Роль мирного жителя – ничего не делает ночью.
type JesterRole struct{}

func (j *JesterRole) HaveNightAction() bool {
	return false
}

func (j *JesterRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("Jester %s does nothing at night\n", owner.ID)
}

func (j *JesterRole) HaveDayAction() bool {
	return false
}

func (j *JesterRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("Jester %s does nothing at night\n", owner.ID)
}

func (j *JesterRole) GetRussianName() string { return "Шут" }

func (j *JesterRole) GetTeam() Team { return solo }

func (j *JesterRole) NeedTarget(phase Phase) bool { return false }

func (j *JesterRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (j *JesterRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (j *JesterRole) CanExecuteAction(p *Player) bool {
	return false
}
func (j *JesterRole) ExecuteAction(p *Player) {
	log.Printf("Jester %s does nothing\n", p.ID)
}

func (j *JesterRole) IsSoloKiller() bool {
	return false
}

func (j *JesterRole) GetAura() Aura {
	return unknown
}
