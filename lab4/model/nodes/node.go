package nodes

import (
	"fmt"
	"net"
	mt "snake/controller"
	stats "snake/controller/stats"
	"snake/model/nodes/deputy"
	"snake/model/nodes/master"
	"snake/model/nodes/normal"

	"snake/model/proto"
	"time"
)

/* var (
	node NodeInterface
) */

func CreateNormal(ch chan [][]byte, sCh chan []stats.Stats, id int32, addr *net.UDPAddr, dur time.Duration) *normal.Normal {
	return normal.Create(ch, sCh, id, addr, dur)
}

func CreateClearMaster(ch chan [][]byte, sCh chan []stats.Stats, id int32, addr *net.UDPAddr, w, h, food int, dur time.Duration, sendCh chan mt.PMessage) *master.Master {
	m := master.Create(ch, sCh, id, addr, w, h, food, dur, sendCh)
	return m
}

func CreateMasterFromGameState(ch chan [][]byte, sCh chan []stats.Stats, id int32, addr *net.UDPAddr, w, h, food int, dur time.Duration, sendCh chan mt.PMessage, state *proto.GameMessage_StateMsg) *master.Master {
	m := master.CreateFromState(ch, sCh, id, w, h, food, addr, dur, sendCh, state)
	fmt.Println("ВОЗВРАЩАЮ МАСТЕРА")
	return m
}

func CreateDeputy(ch chan [][]byte, sCh chan []stats.Stats, id int32, addr *net.UDPAddr, dur time.Duration) *deputy.Deputy {
	return deputy.Create(ch, sCh, id, addr, dur)
	// chan [][]byte, id int32, addr *net.UDPAddr, duration time.Duration

}
