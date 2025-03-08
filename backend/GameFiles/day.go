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
	g.BroadcastGameStatusToAllPlayers()
	if isGameOver {
		return
	}

	g.startNightPhase()
}
