package GameFiles

import "log"

func (g *Game) StartDayPhase() {
	g.Mutex.Lock()
	g.CurrentPhase = day
	log.Println("day phase started.")
	g.BroadcastGameStatusToAllPlayers() // Рассылаем обновление о фазе всем клиентам
	g.Mutex.Unlock()
	g.startPhaseTimer(15, g.EndDayPhase)
}

func (g *Game) EndDayPhase() {
	log.Println("Ending day phase. Processing votes...")

	g.ExecuteDayActions()

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

	if maxVotesCnt >= (g.GetAliveCnt()+1)/2 && playerWithMaxVotes != "" {
		player, err := g.GetPlayer(playerWithMaxVotes)
		if err != nil {
			return
		}
		g.KillPlayer(player.ID, voting)
	}

	isGameOver, _ := g.CheckGameOver()
	g.ResetVotes()
	g.ResetProtect()
	//g.ResetTarget()
	g.BroadcastGameStatusToAllPlayers()
	if isGameOver {
		return
	}

	g.startNightPhase()
}

func (g *Game) ExecuteDayActions() {
	log.Printf("Executing day actions for %d players", len(g.Players))

	for _, player := range g.Players {
		target, err := g.GetPlayer(player.Target)
		if err != nil {
			//log.Printf("Не найдена цель: " + err.Error())
			continue
		}
		player.DayAction(player, target, g)
	}
}
