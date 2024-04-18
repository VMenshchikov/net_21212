package deputy

import (
	"fmt"
	"net"
	"snake/controller/net/socket"
	stats "snake/controller/stats"
	"snake/model/proto"
	"snake/model/snakeTable"
	"time"

	pb "google.golang.org/protobuf/proto"
)

var ()

type Deputy struct {
	updTable chan [][]byte
	statCh   chan []stats.Stats

	myID       int32
	seq        int64
	masterAddr *net.UDPAddr
	deputyAddr *net.UDPAddr
	period     time.Duration

	masterTimer *time.Timer
	closeCh     chan struct{}
}

func (n *Deputy) GetPeriod() time.Duration {
	return n.period
}

func Create(ch chan [][]byte, statCh chan []stats.Stats, id int32, addr *net.UDPAddr, duration time.Duration) *Deputy {

	fmt.Println(duration, "!!!!!!!!!")
	n := &Deputy{
		updTable:    ch,
		statCh:      statCh,
		myID:        id,
		seq:         0,
		masterAddr:  addr,
		deputyAddr:  addr,
		period:      duration,
		masterTimer: time.NewTimer(duration * 3),
		closeCh:     make(chan struct{}),
	}
	go n.CheckMaster()
	return n
}

func (n *Deputy) CheckMaster() {
	for {
		select {
		/* case <-n.closeCh:
		{
			n.masterTimer.Stop()
		} */
		case a, ok := <-n.masterTimer.C:
			{
				fmt.Println(a, ok)
				if !ok {
					return
				}
				if n.deputyAddr == nil {
					return
				}
				n.masterAddr = n.deputyAddr
				n.deputyAddr = nil

				//socket.GetChannel()

				m := proto.GameMessage{
					MsgSeq:     new(int64),
					SenderId:   new(int32),
					ReceiverId: new(int32),
					Type: &proto.GameMessage_RoleChange{
						RoleChange: &proto.GameMessage_RoleChangeMsg{
							SenderRole:   proto.NodeRole_DEPUTY.Enum(),
							ReceiverRole: proto.NodeRole_MASTER.Enum(),
						},
					},
				}

				bmsg, _ := pb.Marshal(&m)

				msg := socket.Message{
					Msg:  bmsg,
					Addr: socket.GetLAddr(),
				}

				socket.GetChannel() <- msg
				fmt.Println("Мастер отпал")
				return

			}
		}
	}
}

func (n *Deputy) GetMyID() *int32 {
	return &n.myID
}

func (n *Deputy) GetMasterAddr() *net.UDPAddr {
	return n.masterAddr
}

func (n *Deputy) GetDeputyAddr() *net.UDPAddr {
	return n.deputyAddr
}

func (n *Deputy) SetMasterAddr(addr *net.UDPAddr) {
	n.masterAddr = addr
}

func (n *Deputy) SetDeputyAddr(addr *net.UDPAddr) {
	n.deputyAddr = addr
}

func (n *Deputy) StateMessage(msg *proto.GameMessage_StateMsg) {
	n.masterTimer.Reset(n.period * 3)

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

}

func (n *Deputy) AckMessage(*proto.GameMessage) {
	n.masterTimer.Reset(n.period * 3)
}
func (n *Deputy) ErrorMessage(*proto.GameMessage_ErrorMsg)             {}
func (n *Deputy) JoinMessage(*proto.GameMessage_JoinMsg, *net.UDPAddr) {}
func (n *Deputy) PingMessage(*proto.GameMessage_PingMsg) {

	n.masterTimer.Reset(n.period * 3)
}
func (n *Deputy) RoleChangeMessage(*proto.GameMessage) {
	//
}
func (n *Deputy) SteerMessage(*proto.GameMessage_SteerMsg, *net.UDPAddr) {}
func (n *Deputy) AnnouncementMessage(*proto.GameMessage_AnnouncementMsg) {}
func (n *Deputy) Close() {
	n.masterTimer.Stop()
}
