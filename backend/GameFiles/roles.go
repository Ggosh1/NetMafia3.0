package GameFiles

// Интерфейс для ролевых действий
type Role interface {
	HaveNightAction() bool
	NightAction(owner, target *Player, game *Game) // Ночное действие
	GetRussianName() string
	GetTeam() Team
	NeedTarget(phase Phase) bool
	VoteValue(phase Phase) int
	GetSpeakArea(p *Player, phase Phase) SpeakArea
	CanExecuteAction(p *Player) bool
	ExecuteAction(p *Player)
	IsSoloKiller() bool
	GetAura() Aura
}
