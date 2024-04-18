package controller

import (
	"net"
	"snake/model/proto"
)

type PMessage struct {
	Msg  *proto.GameMessage
	Addr *net.UDPAddr
}
