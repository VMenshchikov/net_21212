package anouncement

import (
	"fmt"
	"net"

	//winCont "snake/controller/viewC"
	"snake/model/proto"
	"sync"
	"time"

	pb "google.golang.org/protobuf/proto"
)

type AnounceAddr struct {
	An   anounce
	Addr net.UDPAddr
}

type anounce struct {
	Can_join                  bool
	Name                      string
	W, H                      int
	Food_static, Food_dynamic int
	Duration                  time.Duration
	//players int,
}

type storageNode struct {
	timerLife    *time.Timer
	remoteAddres *net.UDPAddr
	anounce      anounce
}

type syncMap struct {
	lock           sync.Mutex
	anounceStorage map[string]storageNode
}

const (
	lifeDuration = time.Second * 3
)

var (
	storage = syncMap{
		lock:           sync.Mutex{},
		anounceStorage: make(map[string]storageNode),
	}
)

func GetAllAnounces() []AnounceAddr {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	var ret []AnounceAddr

	for _, sn := range storage.anounceStorage {
		ret = append(ret, AnounceAddr{
			An:   sn.anounce,
			Addr: *sn.remoteAddres,
		})
	}
	return ret

}

func addAnounce(node storageNode) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	val, ok := storage.anounceStorage[node.remoteAddres.String()]

	if false == ok {
		node.timerLife = time.NewTimer(lifeDuration)
		storage.anounceStorage[node.remoteAddres.String()] = node

		go func() {
			select {
			case <-node.timerLife.C:
				{
					storage.lock.Lock()
					defer storage.lock.Unlock()
					delete(storage.anounceStorage, node.remoteAddres.String())
					//winCont.UpdAnounce()
				}
			}
		}()
	} else {
		val.anounce = node.anounce
		val.timerLife.Reset(lifeDuration)
		storage.anounceStorage[node.remoteAddres.String()] = val
	}

	//winCont.UpdAnounce()

}

func NewAnounce(addr *net.UDPAddr, msg []byte) {
	var tmp proto.GameMessage
	pb.Unmarshal(msg, &tmp)

	switch {
	case tmp.GetAnnouncement() != nil:
		{
			anounceMsg := tmp.GetAnnouncement().GetGames()[0]
			an := anounce{
				Can_join:     anounceMsg.GetCanJoin(),
				Name:         anounceMsg.GetGameName(),
				W:            int(anounceMsg.GetConfig().GetWidth()),
				H:            int(anounceMsg.GetConfig().GetHeight()),
				Food_static:  int(anounceMsg.GetConfig().GetFoodStatic()),
				Food_dynamic: 1,
				Duration:     time.Duration(anounceMsg.GetConfig().GetStateDelayMs()), /*  * time.Millisecond */
			}

			storageEl := storageNode{
				timerLife:    nil,
				remoteAddres: addr,
				anounce:      an,
			}
			addAnounce(storageEl)
		}
	default:
		fmt.Println("Unknown message")
	}
}
