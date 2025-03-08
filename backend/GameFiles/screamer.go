package GameFiles

import (
	"log"
)

// Роль доктора – защищает цель, устанавливая isProtected в true.
type ScreamerRole struct{}

func (s *ScreamerRole) HaveNightAction() bool {
	return false
}

func (s *ScreamerRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("Screamer %s does nothing at night\n", owner.ID)
}

func (s *ScreamerRole) GetRussianName() string { return "Крикун" }

func (s *ScreamerRole) GetTeam() Team { return villager }

func (s *ScreamerRole) NeedTarget(phase Phase) bool { return phase == day }

func (s *ScreamerRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (s *ScreamerRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (s *ScreamerRole) CanExecuteAction(p *Player) bool {
	return false
}
func (s *ScreamerRole) ExecuteAction(p *Player) {
	log.Printf("Seer %s does nothing\n", p.ID)
}

func (s *ScreamerRole) IsSoloKiller() bool {
	return false
}

func (s *ScreamerRole) GetAura() Aura {
	return good
}
