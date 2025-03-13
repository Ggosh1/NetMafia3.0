package GameFiles

import (
	"log"
)

func (g *Game) startNightPhase() {
	g.CurrentPhase = night
	log.Println("night phase started.")
	g.BroadcastGameStatusToAllPlayers() // Рассылаем обновление о фазе всем клиентам
	g.startPhaseTimer(30, g.EndNightPhase)
}

func (g *Game) EndNightPhase() {
	log.Println("Ending night phase. Starting new day.")

	g.ExecuteNightActions()

	var maxVotesCnt int = 0
	var playerWithMaxVotes string = ""

	for player, cnt := range g.GetVotesMap() {
		if cnt > maxVotesCnt {
			maxVotesCnt = cnt
			playerWithMaxVotes = player
		} else if cnt == maxVotesCnt {
			playerWithMaxVotes = ""
		}
	}

	if maxVotesCnt >= (g.GetMafiaCnt()+1)/2 && playerWithMaxVotes != "" {
		player, err := g.GetPlayer(playerWithMaxVotes)
		if err != nil {
			return
		}
		g.KillPlayer(player.ID, mafiaVoting)
	}

	isGameOver, _ := g.CheckGameOver()
	g.ResetVotes()
	g.ResetProtect()
	//g.ResetTarget()
	g.BroadcastGameStatusToAllPlayers()
	if isGameOver {
		return
	}

	g.StartDayPhase()
}

func (g *Game) ExecuteNightActions() {
	log.Printf("Executing night actions for %d players", len(g.Players))

	for _, player := range g.Players {
		target, err := g.GetPlayer(player.Target)
		if err != nil {
			//log.Printf("Не найдена цель: " + err.Error())
			continue
		}
		player.NightAction(player, target, g)
	}
}
