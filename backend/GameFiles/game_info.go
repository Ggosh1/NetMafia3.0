package GameFiles

import "fmt"

type Phase string

const (
	day   Phase = "day"
	night Phase = "night"
)

func (g *Game) GetVotesMap() map[string]int {
	var voteCnt map[string]int = make(map[string]int)
	for _, p := range g.Players {
		if p.IsAlive && p.InRoom && p.VotedFor != "" {
			voteCnt[p.VotedFor]++
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
		if p.IsAlive && p.Team == mafia {
			aliveMafia++
		}
	}
	return aliveMafia
}

func (g *Game) GetVillagerCnt() int {
	aliveVillager := 0
	for _, p := range g.Players {
		if p.IsAlive && p.Team == villager {
			aliveVillager++
		}
	}
	return aliveVillager
}

func (g *Game) CheckGameOver() (bool, string) {
	aliveMafia := g.GetMafiaCnt()
	aliveVillagers := g.GetVillagerCnt()

	hackerAlive := false
	/*for _, player := range g.Players {
		if player.IsAlive {
			if player.Role == "Хакер" {
				hackerAlive = true
			}
		}
	}*/

	if aliveMafia == 0 && !hackerAlive {
		return true, "Villagers win!"
	}

	if aliveMafia >= aliveVillagers && !hackerAlive {
		return true, "Mafia wins!"
	}
	if hackerAlive && aliveVillagers == 0 && aliveMafia == 0 {
		return true, "Hacker win"
	}

	return false, ""
}
