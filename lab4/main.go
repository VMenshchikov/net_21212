package main

import (
	model "snake/controller/modelC"
	multicast "snake/controller/net/multicast"
	socket "snake/controller/net/socket"
	stats "snake/controller/stats"
	viewC "snake/controller/viewC"
	"snake/view"
)

func main() {
	socket.Init()
	defer socket.GetSocket().Close()

	go multicast.GetMultSocket().Reader()
	defer multicast.GetMultSocket().Close()

	go viewC.StartUpdateAnounces()
	defer viewC.StopUpdateAnounces()

	table := make(chan [][]byte)
	statsCh := make(chan []stats.Stats)
	defer close(table)
	defer close(statsCh)

	go viewC.PaintTable(table)
	go viewC.UpdateStats(statsCh)

	model.SetChannel(table, statsCh)

	view.CreateApp(50, 50)
}
