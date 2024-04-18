package nodes

import (
	"net"
	"snake/model/proto"
	"time"
)

type NodeInterface interface {
	GetMyID() *int32
	GetPeriod() time.Duration
	GetMasterAddr() *net.UDPAddr
	GetDeputyAddr() *net.UDPAddr

	SetMasterAddr(*net.UDPAddr)
	SetDeputyAddr(*net.UDPAddr)

	StateMessage(*proto.GameMessage_StateMsg)
	AckMessage(*proto.GameMessage)
	ErrorMessage(*proto.GameMessage_ErrorMsg)
	JoinMessage(*proto.GameMessage_JoinMsg, *net.UDPAddr)
	PingMessage(*proto.GameMessage_PingMsg)
	RoleChangeMessage(*proto.GameMessage)
	SteerMessage(*proto.GameMessage_SteerMsg, *net.UDPAddr)
	AnnouncementMessage(*proto.GameMessage_AnnouncementMsg)

	Close()
}
