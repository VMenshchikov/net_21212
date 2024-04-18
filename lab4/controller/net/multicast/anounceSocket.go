package multicast

import (
	"fmt"
	"net"
	"snake/model/anouncement"
)

type multSocket struct {
	multicastSocket *net.UDPConn
}

var (
	singl *multSocket
)

func (s multSocket) Close() {
	s.multicastSocket.Close()
}
func GetMultSocket() *multSocket {
	if singl == nil {
		ifi, _ := net.Interfaces()
		multAddr := net.UDPAddr{
			IP:/* net.IPv4(224, 0, 0, 1), */ net.IPv4(239, 192, 0, 4),
			Port:/* 8888,  */ 9192,
			Zone: "",
		}
		soc, err := net.ListenMulticastUDP("udp", &ifi[1], &multAddr)
		if nil != err {
			fmt.Printf("%s", err.Error())
		} else {
			singl = &multSocket{multicastSocket: soc}
		}
	}
	return singl
}

func (s multSocket) Reader() {
	for {
		msg := make([]byte, 65507)
		n, addr, err := s.multicastSocket.ReadFromUDP(msg)
		if err != nil {
			fmt.Printf("%s", err.Error())
			break
		}
		if n == 0 {
			continue
		}
		anouncement.NewAnounce(addr, msg)

		/*
		 */

	}

}
