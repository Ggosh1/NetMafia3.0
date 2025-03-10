package GameFiles

import (
	"fmt"
	"log"
)

// Роль доктора – защищает цель, устанавливая isProtected в true.
type SeerRole struct{}

func (s *SeerRole) HaveNightAction() bool {
	return true
}

func (s *SeerRole) NightAction(owner, target *Player, game *Game) {
	// Используем мьютекс целевого игрока для безопасного изменения состояния.
	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	game.broadcastChatMessageToPlayer(SERVER, owner.ID, fmt.Sprintf("Роль игрока %s - %s", target.ID, target.Role.GetRussianName()))
	log.Printf("Seer %s asks %s\n role", owner.ID, target.ID)
}

func (s *SeerRole) HaveDayAction() bool {
	return false
}

func (s *SeerRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("Seer %s does nothing at day\n", owner.ID)
}

func (s *SeerRole) GetRussianName() string { return "Провидец" }

func (s *SeerRole) GetTeam() Team { return villager }

func (s *SeerRole) NeedTarget(phase Phase) bool { return phase == night }

func (s *SeerRole) VoteValue(phase Phase) int {
	if phase == day {
		return 1
	} else {
		return 0
	}
}

func (s *SeerRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
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

func (s *SeerRole) CanExecuteAction(p *Player) bool {
	return false
}
func (s *SeerRole) ExecuteAction(p *Player) {
	log.Printf("Seer %s does nothing\n", p.ID)
}

func (s *SeerRole) IsSoloKiller() bool {
	return false
}

func (s *SeerRole) GetAura() Aura {
	return good
}
