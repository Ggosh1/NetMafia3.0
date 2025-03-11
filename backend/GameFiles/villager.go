package GameFiles

import "log"

// Роль мирного жителя – ничего не делает ночью.
type VillagerRole struct{}

func (v *VillagerRole) HaveNightAction() bool {
	return false
}

func (v *VillagerRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("Villager %s does nothing at night\n", owner.ID)
}

func (v *VillagerRole) HaveDayAction() bool {
	return false
}

func (v *VillagerRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("Villager %s does nothing at day\n", owner.ID)
}

func (v *VillagerRole) GetRussianName() string { return "Мирный житель" }

func (v *VillagerRole) GetTeam() Team { return villager }

func (v *VillagerRole) NeedTarget(phase Phase) bool { return false }

func (v *VillagerRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (v *VillagerRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (v *VillagerRole) CanExecuteAction(p *Player) bool {
	return false
}
func (v *VillagerRole) ExecuteAction(p *Player) {
	log.Printf("Villager %s does nothing\n", p.ID)
}

func (s *VillagerRole) IsSoloKiller() bool {
	return false
}

func (v *VillagerRole) GetAura() Aura {
	return good
}
