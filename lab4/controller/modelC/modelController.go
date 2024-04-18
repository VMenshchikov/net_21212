package model

import (
	"fmt"
	"net"
	"snake/controller/net/socket"
	stats "snake/controller/stats"
	"snake/model/nodes"
	"snake/model/proto"
	snaketable "snake/model/snakeTable"

	"sync"

	"time"

	mt "snake/controller"

	pb "google.golang.org/protobuf/proto"
)

var (
	lastMsg       *proto.GameMessage_StateMsg
	sw, sh, sfood int
	waitAck       map[string]spectate // key = UDPAddr+int
	waitLock      *sync.Mutex
	node          nodes.NodeInterface
	tableChannel  chan [][]byte
	statsChannel  chan []stats.Stats
	sendCh        chan mt.PMessage
	sendChClose   chan struct{}
	seq           int64
)

type spectate struct {
	doneCh chan struct{}
	msg    *proto.GameMessage
}

func init() {
	waitAck = make(map[string]spectate)
	sendCh = make(chan mt.PMessage)
	//sendCh = make(chan mt.PMessage)
	waitLock = new(sync.Mutex)

}

func SetChannel(tableCh chan [][]byte, statsCh chan []stats.Stats) {
	tableChannel = tableCh
	statsChannel = statsCh
}

func proxy(d time.Duration) {
	sendChClose = make(chan struct{})

	for {
		select {
		case <-sendChClose:
			{
				return
			}
		/* case <-time.NewTimer(d / 2).C:
		{
			seq++
			tmp := seq
			ping := proto.GameMessage{
				MsgSeq:     &tmp,
				SenderId:   new(int32),
				ReceiverId: new(int32),
				Type:       &proto.GameMessage_Ping{Ping: &proto.GameMessage_PingMsg{}},
			}

			bmsg, _ := pb.Marshal(&ping)
			socket.GetSocket().SendUnicast(bmsg, node.GetMasterAddr())
			addSpec(&ping, node.GetMasterAddr(), node.GetPeriod())
			//fmt.Println("пинг на ", node.GetMasterAddr())
		} */

		case stct, ok := <-sendCh:
			{

				if !ok {
					return
				}

				if stct.Msg.GetAck() == nil {
					seq++
					tmp := seq
					stct.Msg.MsgSeq = &tmp

					addSpec(stct.Msg, stct.Addr, node.GetPeriod())
				}

				bmsg, _ := pb.Marshal(stct.Msg)

				socket.GetSocket().SendUnicast(bmsg, stct.Addr)
				//fmt.Println("сообщение", stct.Addr)

			}
		}
	}
}

func NewGame(w, h, food int, dur time.Duration) {
	node = nodes.CreateClearMaster(tableChannel, statsChannel, 0, socket.GetLAddr(), w, h, food, dur, sendCh)
	go manager()
	go proxy(time.Millisecond * time.Duration(dur))
	//send anounces
	//send state

}

func closeAllSpec() {
	waitLock.Lock()
	defer waitLock.Unlock()
	for key := range waitAck {
		close(waitAck[key].doneCh)
		delete(waitAck, key)
	}
}

func closeSpec(addr *net.UDPAddr, seq int64) {
	waitLock.Lock()
	defer waitLock.Unlock()

	key := addr.String() + fmt.Sprintf("#%d", seq)
	spec, ok := waitAck[key]
	if !ok {
	} else {
		close(spec.doneCh)
		delete(waitAck, key)
	}
}

func addSpec(msg *proto.GameMessage, addr *net.UDPAddr, dur time.Duration) {
	waitLock.Lock()
	defer waitLock.Unlock()

	tmp := spectate{
		doneCh: make(chan struct{}),
		msg:    msg,
	}
	waitAck[addr.String()+fmt.Sprintf("#%d", msg.GetMsgSeq())] = tmp

	//функция при отсутствии сообщения(открытом канале) по истечению времени
	//начнет повторно посылать неподтвержденное сообщение
	go func() {
		select {
		case <-tmp.doneCh:
			{
				return
			}
		case <-time.NewTimer(dur).C:
			{
				fmt.Println("Не получили ack от", addr.String(), msg.GetMsgSeq())
				bmsg, err := pb.Marshal(msg)
				if nil != err {
					fmt.Println(err.Error())
				}
				for i := 0; i < 10; i++ {
					select {
					case <-tmp.doneCh:
						{
							fmt.Println("Закрыли сральню", msg.GetMsgSeq())
							return
						}
					case <-time.NewTimer(dur / 10).C:
						{
							//fmt.Println("Сру", msg.GetMsgSeq() /* , msg.GetType() */)
							socket.GetSocket().SendUnicast(bmsg, addr)
						}
					}
				}
				//отрубился!
			}
		}
	}()
}

func UpdateTable(w, h int) {
	snaketable.NewTable(w, h)
}

func SendSteer(dir *proto.Direction) {
	if nil != node /*&& nil != reflect.TypeOf(node) != */ {
		msg := proto.GameMessage{
			MsgSeq:     nil,
			SenderId:   node.GetMyID(),
			ReceiverId: new(int32),
			Type: &proto.GameMessage_Steer{
				Steer: &proto.GameMessage_SteerMsg{
					Direction: dir,
				},
			},
		}

		sendCh <- mt.PMessage{
			Msg:  &msg,
			Addr: node.GetMasterAddr(),
		}
	}
}
func SendAck(seq int64, addr *net.UDPAddr) {
	msg := proto.GameMessage{
		MsgSeq:     &seq,
		SenderId:   node.GetMyID(),
		ReceiverId: new(int32),
		Type:       &proto.GameMessage_Ack{Ack: &proto.GameMessage_AckMsg{}},
	}
	sendCh <- mt.PMessage{
		Msg:  &msg,
		Addr: addr,
	}
}

func manager() {
	for {
		ch := <-socket.GetChannel()
		gmsg := proto.GameMessage{}
		err := pb.Unmarshal(ch.Msg, &gmsg)
		if nil != err {
			fmt.Println(err.Error())
			continue
		}
		//fmt.Println("получил от", ch.Addr, gmsg.GetType())
		switch {
		case nil != gmsg.GetState():
			{
				SendAck(gmsg.GetMsgSeq(), ch.Addr)
				node.StateMessage(gmsg.GetState())
				lastMsg = gmsg.GetState()

				m, d := node.GetMasterAddr(), node.GetDeputyAddr()

				for _, p := range lastMsg.GetState().GetPlayers().GetPlayers() {
					if p.Role.Enum() == proto.NodeRole_MASTER.Enum() {

						curM := (&net.UDPAddr{
							IP:   net.ParseIP(p.GetIpAddress()),
							Port: int(p.GetPort()),
							Zone: "",
						})

						if m.String() != curM.String() {
							node.SetMasterAddr(curM)
						}
					}
					if p.Role.Enum() == proto.NodeRole_DEPUTY.Enum() {
						curD := (&net.UDPAddr{
							IP:   net.ParseIP(p.GetIpAddress()),
							Port: int(p.GetPort()),
							Zone: "",
						})

						if d == nil || d.String() != curD.String() {
							node.SetDeputyAddr(curD)
							continue
						}

					}
				}
			}
		case nil != gmsg.GetAck():
			{
				closeSpec(ch.Addr, gmsg.GetMsgSeq())
				node.AckMessage(&gmsg)
			}
		case nil != gmsg.GetError():
			{
				SendAck(gmsg.GetMsgSeq(), ch.Addr)
				//node.ErrorMessage(gmsg.GetError())
			}
		case nil != gmsg.GetJoin():
			{
				//отдельный ак в мастере
				node.JoinMessage(gmsg.GetJoin(), ch.Addr)
			}
		case nil != gmsg.GetPing():
			{
				SendAck(gmsg.GetMsgSeq(), ch.Addr)
				node.PingMessage(gmsg.GetPing())
			}
		case nil != gmsg.GetRoleChange():
			{
				SendAck(gmsg.GetMsgSeq(), ch.Addr)

				msg := gmsg.GetRoleChange()

				sendRole := msg.GetSenderRole()
				if sendRole == *proto.NodeRole_MASTER.Enum().Enum() {
					fmt.Println("Обновть мастера на", ch.Addr)
					node.SetMasterAddr(ch.Addr)
					if node.GetMasterAddr() == node.GetDeputyAddr() {
						node.SetDeputyAddr(nil)
					}
				}

				recRole := msg.GetReceiverRole()
				switch recRole {
				case *proto.NodeRole_DEPUTY.Enum():
					{
						fmt.Println("Стать депути")
						myId := node.GetMyID()
						dur := node.GetPeriod()
						node.Close()
						node = nodes.CreateDeputy(tableChannel, statsChannel, *myId, ch.Addr, dur)
					}
				case *proto.NodeRole_MASTER.Enum():
					{
						fmt.Println("Стать мастером")

						myId := node.GetMyID()
						dur := node.GetPeriod()
						node.Close()
						closeAllSpec()

						newPlayers := make([]*proto.GamePlayer, 0)

						for i, v := range lastMsg.GetState().GetPlayers().GetPlayers() {
							nPlayer := lastMsg.GetState().GetPlayers().GetPlayers()[i]
							fmt.Println("Пересборка игроков", v.GetRole(), v.GetId())
							/* if nPlayer.GetRole() == proto.NodeRole_MASTER {
								continue
							} */

							if nPlayer.GetId() == *myId {
								//tmpPort := socket.GetLAddr().Port
								//tmpAddr := socket.GetLAddr().IP.String()
								nPlayer.Role = proto.NodeRole_MASTER.Enum()
								//nPlayer.Port
							}
							if nPlayer.GetRole() == proto.NodeRole_MASTER {
								nPlayer.Role = proto.NodeRole_NORMAL.Enum()
							}
							newPlayers = append(newPlayers, nPlayer)
						}

						newState := proto.GameMessage_StateMsg{
							State: &proto.GameState{
								StateOrder: lastMsg.GetState().StateOrder,
								Snakes:     lastMsg.State.Snakes,
								Foods:      lastMsg.State.Foods,
								Players: &proto.GamePlayers{
									Players: newPlayers,
								},
							},
						}

						//close(sendChClose)

						node = nodes.CreateMasterFromGameState(tableChannel, statsChannel, *myId, socket.GetLAddr(), sw, sh, sfood, dur, sendCh, &newState)
						fmt.Println("Стал!!!!!!!!!!!")
						//proxy(dur)
						fmt.Println("ЗАКОНЧИЛ")
					}
				default:

				}

			}
		case nil != gmsg.GetSteer():
			{
				SendAck(gmsg.GetMsgSeq(), ch.Addr)
				node.SteerMessage(gmsg.GetSteer(), ch.Addr)
			}
		case nil != gmsg.GetAnnouncement():
			{
				fmt.Println("ОТ КУДА ТУТ АНОНС?")
				continue
			}
		default:
		}

		//tab := mtable.GetState()
		//snaketable.GetTable().UpdateTable(tab.GetState(), int(*ummsg.ReceiverId))

	}
}

func SendJoin(gameName string, addr *net.UDPAddr, dur time.Duration, w, h, food int) {
	//dur = time.Microsecond * dur
	//fmt.Println("join", w, h, food)
	name := "Vlad"
	protoMsg := proto.GameMessage{
		MsgSeq:     new(int64),
		SenderId:   new(int32),
		ReceiverId: new(int32),
		Type: &proto.GameMessage_Join{
			Join: &proto.GameMessage_JoinMsg{
				PlayerType:    proto.PlayerType_HUMAN.Enum(),
				PlayerName:    &name,
				GameName:      &gameName,
				RequestedRole: proto.NodeRole_NORMAL.Enum(),
			},
		},
	}

	msg, err := pb.Marshal(&protoMsg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	socket.GetSocket().SendUnicast(msg, addr)

	for {
		select {
		case mes := <-socket.GetChannel():
			{
				ummsg := proto.GameMessage{}
				pb.Unmarshal(mes.Msg, &ummsg)
				switch {
				case nil != ummsg.GetAck():
					{

						id := ummsg.GetReceiverId()
						sw = w
						sh = h
						sfood = food
						node = nodes.CreateNormal(tableChannel, statsChannel, id, mes.Addr, dur)
						//normal.Create(tableChannel, id, mes.Addr, dur)
						go manager()
						go proxy(dur)
						return

					}
				case nil != ummsg.GetError():
					{
						m := ummsg.GetError()
						fmt.Println(m.GetErrorMessage())
						return
					}
				default:
					{
						fmt.Println("Не то сообщение")
					}
				}
			}
		case <-time.NewTimer(time.Second * 5).C:
			{
				fmt.Println("Мужик умер. F")
				return
			}
		}
	}
}
