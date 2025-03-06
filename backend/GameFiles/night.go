package GameFiles

import (
	"fmt"
	"log"
)

func (g *Game) startNightPhase() {
	g.Mutex.Lock()
	g.CurrentPhase = night
	log.Println("night phase started.")
	g.BroadcastGameStatusToAllPlayers() // Рассылаем обновление о фазе всем клиентам
	g.Mutex.Unlock()
	g.startPhaseTimer(15, g.EndNightPhase)
}

func (g *Game) EndNightPhase() {
	log.Println("Ending night phase. Starting new day.")

	for _, player := range g.Players {
		if player.IsAlive {
			player.NightAction(g.Players)
		}
	}
	/*if gameOver, winner := checkGameOver(); gameOver {
		log.Println(winner)
		broadcastWinner(winner)
		game.GameStarted = false
		return
	}
	game.DayNumber++
	StartDayPhase()*/
}

func (g *Game) ExecuteNightActions() {
	fmt.Println("night phase begins...")
	for _, player := range g.Players {
		player.NightAction(g.Players)
	}
}
