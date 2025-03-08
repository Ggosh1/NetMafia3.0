package GameFiles

import "fmt"

type Phase string

const (
	day      Phase = "day"
	night    Phase = "night"
	gameover Phase = "gameover"
)

func (g *Game) GetVotesMap() map[string]int {
	var voteCnt map[string]int = make(map[string]int)
	for _, p := range g.Players {
		if p.IsAlive && p.InRoom && p.VotedFor != "" {
			voteCnt[p.VotedFor] += p.VoteValue(g.CurrentPhase)
		}
	}
	return voteCnt
}

func (g *Game) GetPlayer(id string) (*Player, error) {
	if player, exists := g.Players[id]; exists {
		return player, nil
	} else {
		return nil, fmt.Errorf("Игрок не найден")
	}
}

func (g *Game) GetMafiaCnt() int {
	aliveMafia := 0
	for _, p := range g.Players {
		if p.IsAlive && p.GetTeam() == mafia {
			aliveMafia++
		}
	}
	return aliveMafia
}

func (g *Game) GetVillagerCnt() int {
	aliveVillager := 0
	for _, p := range g.Players {
		if p.IsAlive && p.GetTeam() == villager {
			aliveVillager++
		}
	}
	return aliveVillager
}

func (g *Game) GetSoloKillersCnt() int {
	alive := 0
	for _, p := range g.Players {
		if p.IsAlive && p.IsSoloKiller() {
			alive++
		}
	}
	return alive
}

func (g *Game) GetAliveCnt() int {
	alive := 0
	for _, p := range g.Players {
		if p.IsAlive {
			alive++
		}
	}
	return alive
}

func (g *Game) CheckGameOver() (bool, string) {

	if g.GameStarted == false {
		return false, ""
	}

	for _, p := range g.Players {
		if _, ok := p.Role.(*JesterRole); ok && p.diedBy == voting {
			g.CurrentPhase = gameover
			return true, "Jester wins!"
		}
	}

	if g.GetSoloKillersCnt() == 1 && g.GetAliveCnt() <= 1 {
		g.CurrentPhase = gameover
		return true, "Solo killer wins!"
	}

	if g.GetMafiaCnt()+g.GetSoloKillersCnt() == 0 {
		g.CurrentPhase = gameover
		return true, "Villagers win!"
	}

	if g.GetMafiaCnt() >= (g.GetAliveCnt()+1)/2 {
		g.CurrentPhase = gameover
		return true, "Mafia wins!"
	}

	return false, ""
}

func (g *Game) PlayerCanVote(id string) bool {
	player, err := g.GetPlayer(id)
	if err != nil {
		return false
	}

	if player.IsAlive && player.InRoom && !player.IsHacked && (g.CurrentPhase == day || player.GetTeam() == mafia) {
		return true
	}
	return false
}
