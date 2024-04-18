package master

import (
	"fmt"
	"net"
	"snake/controller/net/socket"
	stats "snake/controller/stats"
	gamemodel "snake/model/gameModel"

	"snake/model/proto"
	"snake/model/snakeTable"
	"sync"
	"sync/atomic"
	"time"

	pb "google.golang.org/protobuf/proto"

	mt "snake/controller"
)

var (
	update = make(chan struct{})
)

type idConnS struct {
	lock sync.Mutex
	role proto.NodeRole
	life *time.Timer
}

func (m *Master) selectDeputy() {
	fmt.Println("Выбор депути", len(m.idConn), m.deputyAddr)
	if len(m.idConn) == 0 {
		return
	}
	if m.deputyAddr != nil {
		return
	}
	for k := range m.idConn {
		m.idConn[k].role = *proto.NodeRole_DEPUTY.Enum()
		m.updTImerIdConn(k)
		m.deputyAddr = m.iDaddr[k]

		fmt.Println("установил депути", m.iDaddr[k])

		tmpID := int32(k)
		msg := mt.PMessage{
			Msg: &proto.GameMessage{
				MsgSeq:     new(int64),
				SenderId:   &m.myID,
				ReceiverId: &tmpID,
				Type: &proto.GameMessage_RoleChange{
					RoleChange: &proto.GameMessage_RoleChangeMsg{
						SenderRole:   proto.NodeRole_MASTER.Enum(),
						ReceiverRole: proto.NodeRole_DEPUTY.Enum(),
					},
				},
			},
			Addr: m.iDaddr[k],
		}
		m.sendChan <- msg
		break
	}
}

func (m *Master) checkLife(id int) {
	tmp, ok := m.idConn[id]
	if !ok {
		return
	}

	select {
	case <-tmp.life.C:
	}

	tmp.lock.Lock()
	fmt.Println("сдох!!!!!!!!!!!!!!1")
	delete(m.idConn, id)
	delete(m.addrID, m.iDaddr[id].String())
	delete(m.iDaddr, id)
	m.deputyAddr = nil
	//tmp.lock.Unlock()

	if tmp.role == *proto.NodeRole_DEPUTY.Enum() {
		m.selectDeputy()
	}
}

func (m *Master) updTImerIdConn(id int) {
	for {
		tmp, ok := m.idConn[id]
		if !ok {
			return
		}

		if tmp.lock.TryLock() == false {
			continue
		}
		tmp.life.Reset(m.period * 3)
		tmp.lock.Unlock()
		break
	}
}

func (m *Master) addIdConn(id int, role *proto.NodeRole) {
	tmp := idConnS{
		lock: sync.Mutex{},
		role: *role,
		life: time.NewTimer(m.period * 10),
	}

	if role == proto.NodeRole_DEPUTY.Enum() {
		m.deputyAddr = m.iDaddr[id]
	}

	m.idConn[id] = &tmp

	go m.checkLife(id)

}

type Master struct {
	updTable   chan [][]byte
	statCh     chan []stats.Stats
	myID       int32
	masterAddr *net.UDPAddr
	deputyAddr *net.UDPAddr
	period     time.Duration

	addrID map[string]int
	iDaddr map[int]*net.UDPAddr

	idConn      map[int]*idConnS
	deputyIsSet atomic.Bool

	game, prevGame *gamemodel.GameStruct

	sendChan chan mt.PMessage
}

func (m *Master) GetPeriod() time.Duration {
	return m.period
}

func (m *Master) sendAnonounces() {
	for {
		select {
		case <-update:
			{
				return
			}
		case <-time.NewTimer(time.Second).C:
			{
				tmpTrue := true
				tmpName := "Vlad"
				tmpDur := int32(m.period)
				announce := proto.GameAnnouncement{
					Players: m.game.GetPlayers(),
					Config: &proto.GameConfig{
						Width:        m.game.GetWidth(),
						Height:       m.game.GetHeight(),
						FoodStatic:   m.game.GetFood(),
						StateDelayMs: &tmpDur,
					},
					CanJoin:  &tmpTrue,
					GameName: &tmpName,
				}
				msg := proto.GameMessage{
					MsgSeq:     new(int64),
					SenderId:   new(int32),
					ReceiverId: new(int32),
					Type: &proto.GameMessage_Announcement{
						Announcement: &proto.GameMessage_AnnouncementMsg{
							Games: []*proto.GameAnnouncement{&announce},
						},
					},
				}

				bmsg, _ := pb.Marshal(&msg)

				socket.GetSocket().SendUnicast(bmsg, &net.UDPAddr{
					IP:   []byte{239, 192, 0, 4},
					Port: 9192,
					Zone: "",
				})
				fmt.Println("send announce")
			}
		}
	}
}

func Create(ch chan [][]byte, statCh chan []stats.Stats, id int32, addr *net.UDPAddr, w, h, food int, duration time.Duration, sendCh chan mt.PMessage) *Master {

	g := gamemodel.CreateNew(w, h, food, addr)
	g.AddPlayer("Vlad", proto.NodeRole_MASTER.Enum(), socket.GetLAddr(), &id)

	g.AddSnake(id)

	node := &Master{
		updTable:    ch,
		statCh:      statCh,
		myID:        id,
		masterAddr:  addr,
		deputyAddr:  nil,
		period:      duration,
		addrID:      map[string]int{},
		iDaddr:      map[int]*net.UDPAddr{},
		idConn:      map[int]*idConnS{},
		deputyIsSet: atomic.Bool{},
		game:        g,
		prevGame:    g.CreateCopy(),
		sendChan:    sendCh,
	}

	node.addrID[addr.String()] = int(id)
	node.iDaddr[int(id)] = addr

	go node.startUpdate()
	go node.sendAnonounces()

	//Создали мастера
	return node
}

func CreateFromState(ch chan [][]byte, statCh chan []stats.Stats, id int32, w, h, food int, addr *net.UDPAddr, duration time.Duration, sendCh chan mt.PMessage, state *proto.GameMessage_StateMsg) *Master {
	g := gamemodel.CreateWithGamestate(w, h, food, addr, state)
	node := &Master{
		updTable:    ch,
		statCh:      statCh,
		myID:        id,
		masterAddr:  addr,
		deputyAddr:  nil,
		period:      duration,
		addrID:      map[string]int{},
		iDaddr:      map[int]*net.UDPAddr{},
		idConn:      map[int]*idConnS{},
		deputyIsSet: atomic.Bool{},
		game:        g,
		prevGame:    g.CreateCopy(),
		sendChan:    sendCh,
	}
	fmt.Println(11)
	for _, v := range state.GetState().GetPlayers().GetPlayers() {
		id := v.GetId()
		addr := &net.UDPAddr{
			IP:   net.ParseIP(v.GetIpAddress()),
			Port: int(v.GetPort()),
			Zone: "",
		}
		fmt.Println(" игрок !!", addr, v.GetRole(), v.GetId())

		msg := proto.GameMessage{
			MsgSeq:     new(int64),
			SenderId:   &node.myID,
			ReceiverId: &id,
			Type: &proto.GameMessage_RoleChange{
				RoleChange: &proto.GameMessage_RoleChangeMsg{
					SenderRole:   proto.NodeRole_MASTER.Enum(),
					ReceiverRole: proto.NodeRole_NORMAL.Enum(),
				},
			},
		}
		bmsg, _ := pb.Marshal(&msg)

		socket.GetSocket().SendUnicast(bmsg, addr)

		node.addrID[addr.String()] = int(id)
		node.iDaddr[int(id)] = addr
		if node.myID != id {
			node.addIdConn(int(id), proto.NodeRole_NORMAL.Enum())
		}

	}
	fmt.Println(22)
	node.selectDeputy()

	go node.startUpdate()
	go node.sendAnonounces()
	fmt.Println(33)

	return node
}

func (m *Master) startUpdate() {
	for {
		fmt.Println("step", m.period)
		select {
		case <-update:
			return
		case <-time.NewTimer(m.period).C:
			{
				m.game.NextStep()
				m.prevGame = m.game.CreateCopy()
				fmt.Println("go sState", len(m.addrID))
				go m.sendStates()
			}
		}
	}
}

func (m *Master) sendStates() {
	state := m.createState()

	for id, addr := range m.iDaddr {
		fmt.Println("Отправляем стейт на", addr)
		id32 := int32(id)
		m.sendChan <- mt.PMessage{
			Msg: &proto.GameMessage{
				MsgSeq:     nil,
				SenderId:   &m.myID,
				ReceiverId: &id32,
				Type: &proto.GameMessage_State{
					State: &proto.GameMessage_StateMsg{
						State: state,
					},
				},
			},
			Addr: addr,
		}
	}

}

func (m *Master) createState() *proto.GameState {
	a := m.prevGame.GetOrder()
	b := m.prevGame.GetSnakes()
	c := m.prevGame.GetFoods()
	d := m.prevGame.GetPlayers()

	state := &proto.GameState{
		StateOrder: a,
		Snakes:     b, //m.prevGame.GetSnakes(),
		Foods:      c, // m.prevGame.GetFoods(),
		Players:    d, //m.prevGame.GetPlayers(),
	}
	return state
}

func (m *Master) stopUpdate() {
	close(update)
}

func (m *Master) GetMyID() *int32 {
	return &m.myID
}

func (m *Master) GetMasterAddr() *net.UDPAddr {
	return m.masterAddr
}

func (m *Master) GetDeputyAddr() *net.UDPAddr {
	return m.deputyAddr
}

func (m *Master) SetMasterAddr(addr *net.UDPAddr) {
	m.masterAddr = addr
}

func (m *Master) SetDeputyAddr(addr *net.UDPAddr) {
	m.deputyAddr = addr
}

func (m *Master) StateMessage(msg *proto.GameMessage_StateMsg) {
	state := msg.GetState()
	go func() {
		snakeTable.GetTable().UpdateTable(state, int(m.myID))
		m.updTable <- snakeTable.GetTable().GetGameZone()
	}()

	pStats := make([]stats.Stats, 0)
	for _, v := range state.Players.Players {
		pStats = append(pStats, stats.Stats{
			Name:  v.GetName(),
			Score: int(v.GetScore()),
		})

	}
	m.statCh <- pStats

}

func (m *Master) AckMessage(msg *proto.GameMessage) {
	m.updTImerIdConn(int(msg.GetSenderId()))
	//fmt.Println("обновил таймер, ак")
}
func (m *Master) ErrorMessage(*proto.GameMessage_ErrorMsg) {}
func (m *Master) JoinMessage(msg *proto.GameMessage_JoinMsg, addr *net.UDPAddr) {
	id := m.game.AddPlayer(msg.GetPlayerName(), proto.NodeRole_NORMAL.Enum(), addr, nil)
	err := m.game.AddSnake(id)
	fmt.Println("join")
	if err != nil {
		errorMsg := "Не получилось найти место"
		fmt.Println(errorMsg)

		castid := int32(id)

		pbmsg := proto.GameMessage{
			MsgSeq:     nil,
			SenderId:   &m.myID,
			ReceiverId: &castid,
			Type: &proto.GameMessage_Error{
				Error: &proto.GameMessage_ErrorMsg{
					ErrorMessage: &errorMsg,
				},
			},
		}

		bmsg, _ := pb.Marshal(&pbmsg)
		socket.GetSocket().SendUnicast(bmsg, addr)

	} else {
		m.addrID[addr.String()] = int(id)
		m.iDaddr[int(id)] = addr
		m.addIdConn(int(id), msg.RequestedRole)

		castid := int32(id)

		pbmsg := proto.GameMessage{
			MsgSeq:     nil,
			SenderId:   &m.myID,
			ReceiverId: &castid,
			Type:       &proto.GameMessage_Ack{},
		}

		bmsg, _ := pb.Marshal(&pbmsg)
		socket.GetSocket().SendUnicast(bmsg, addr)

		m.selectDeputy()
	}

}
func (m *Master) PingMessage(msg *proto.GameMessage_PingMsg) {
	//m.updTImerIdConn(int(msg.GetSenderId()))

}
func (m *Master) RoleChangeMessage(*proto.GameMessage) {
	//
}
func (m *Master) SteerMessage(msg *proto.GameMessage_SteerMsg, addr *net.UDPAddr) {
	m.game.UpdateDir(m.prevGame, m.addrID[addr.String()], msg.Direction)
	m.updTImerIdConn(m.addrID[addr.String()])
	//fmt.Println("Обновил таймер стир")

}
func (m *Master) AnnouncementMessage(*proto.GameMessage_AnnouncementMsg) {}

func (m *Master) Close() {
	m.stopUpdate()
}
