package normal

import (
	"fmt"
	"net"
	stats "snake/controller/stats"
	"snake/model/proto"
	"snake/model/snakeTable"
	"time"
)

var ()

type Normal struct {
	updTable   chan [][]byte
	statCh     chan []stats.Stats
	myID       int32
	seq        int64
	masterAddr *net.UDPAddr
	deputyAddr *net.UDPAddr
	period     time.Duration

	masterTimer time.Timer
	closeCh     chan struct{}
}

func (n *Normal) GetPeriod() time.Duration {
	fmt.Println(n.period, "get normal")
	return n.period
}

func Create(ch chan [][]byte, statCh chan []stats.Stats, id int32, addr *net.UDPAddr, duration time.Duration) *Normal {
	fmt.Println(duration, "create normal")
	n := &Normal{
		updTable:    ch,
		statCh:      statCh,
		myID:        id,
		seq:         0,
		masterAddr:  addr,
		deputyAddr:  nil,
		period:      duration,
		masterTimer: *time.NewTimer(duration * 3),
		closeCh:     make(chan struct{}),
	}
	go n.CheckMaster()
	return n
}

func (n *Normal) CheckMaster() {
	for {
		select {
		/* case <-n.closeCh:
		{
			n.masterTimer.Stop()
		} */
		case _, ok := <-n.masterTimer.C:
			{
				if !ok {
					return
				}
				if n.deputyAddr == nil {
					return
				}
				n.masterAddr = n.deputyAddr
				n.deputyAddr = nil
			}
		}
	}
}

func (n *Normal) GetMyID() *int32 {
	return &n.myID
}

func (n *Normal) GetMasterAddr() *net.UDPAddr {
	return n.masterAddr
}

func (n *Normal) GetDeputyAddr() *net.UDPAddr {
	return n.deputyAddr
}

func (n *Normal) SetMasterAddr(addr *net.UDPAddr) {
	n.masterAddr = addr
}

func (n *Normal) SetDeputyAddr(addr *net.UDPAddr) {
	n.deputyAddr = addr
}

func (n *Normal) StateMessage(msg *proto.GameMessage_StateMsg) {
	state := msg.GetState()
	snakeTable.GetTable().UpdateTable(state, int(n.myID))
	n.updTable <- snakeTable.GetTable().GetGameZone()

	pStats := make([]stats.Stats, 0)
	for _, v := range state.Players.Players {
		pStats = append(pStats, stats.Stats{
			Name:  v.GetName(),
			Score: int(v.GetScore()),
		})

	}
	n.statCh <- pStats

	n.masterTimer.Reset(n.period * 3)
}

func (n *Normal) AckMessage(*proto.GameMessage)                        {}
func (n *Normal) ErrorMessage(*proto.GameMessage_ErrorMsg)             {}
func (n *Normal) JoinMessage(*proto.GameMessage_JoinMsg, *net.UDPAddr) {}
func (n *Normal) PingMessage(*proto.GameMessage_PingMsg)               {}
func (n *Normal) RoleChangeMessage(*proto.GameMessage) {
	//
}
func (n *Normal) SteerMessage(*proto.GameMessage_SteerMsg, *net.UDPAddr) {}
func (n *Normal) AnnouncementMessage(*proto.GameMessage_AnnouncementMsg) {}
func (n *Normal) Close() {
	n.masterTimer.Stop()
}
