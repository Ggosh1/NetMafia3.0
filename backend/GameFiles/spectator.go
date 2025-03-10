package GameFiles

import "log"

type SpectatorRole struct{}

func (s *SpectatorRole) HaveNightAction() bool {
	return false
}
func (s *SpectatorRole) NightAction(owner, target *Player, game *Game) {
	log.Printf("Spectator %s does nothing at night\n", owner.ID)
}

func (s *SpectatorRole) HaveDayAction() bool {
	return false
}
func (s *SpectatorRole) DayAction(owner, target *Player, game *Game) {
	log.Printf("Spectator %s does nothing at day\n", owner.ID)
}

func (s *SpectatorRole) GetRussianName() string {
	return "Зритель"
}

func (s *SpectatorRole) GetTeam() Team {
	return solo
}

func (s *SpectatorRole) NeedTarget(phase Phase) bool {
	return false
}

func (s *SpectatorRole) VoteValue(phase Phase) int {
	return 0
}

func (s *SpectatorRole) GetSpeakArea(p *Player, phase Phase) SpeakArea {
	return nobody
}

func (s *SpectatorRole) CanExecuteAction(p *Player) bool {
	return false
}
func (s *SpectatorRole) ExecuteAction(p *Player) {
	log.Printf("Spectator %s does nothing\n", p.ID)
}

func (s *SpectatorRole) IsSoloKiller() bool {
	return false
}

func (s *SpectatorRole) GetAura() Aura {
	return unknown
}
