package socket

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type socketUDP struct {
	socket *net.UDPConn
}

type Message struct {
	Msg  []byte
	Addr *net.UDPAddr
}

var (
	singl    *socketUDP
	chOutput chan Message
	chInput  chan Message
)

func Init() {
	soc, err := net.ListenUDP("udp", nil)
	if nil != err {
		panic(err.Error())
	} else {
		singl = &socketUDP{socket: soc}
	}
	chOutput = make(chan Message)
	chInput = make(chan Message)

	go singl.readUnicast()
	go singl.sendUnicast()
}

func GetLAddr() *net.UDPAddr {
	addr := singl.socket.LocalAddr()

	//tmp :=

	port := strings.Split(addr.String(), ":")

	localhost, _ := net.LookupIP("localhost")

	pport, _ := strconv.Atoi(port[3])
	t := &net.UDPAddr{
		IP:   net.IP(localhost[0]),
		Port: pport,
		Zone: "",
	}
	fmt.Println(t)
	return t
}
func GetSocket() *socketUDP {
	return singl
}
func GetChannel() chan Message {
	return chOutput
}

func (s socketUDP) Close() {
	close(chOutput)
	close(chInput)
	s.socket.Close()
	//горутина закроется при закрытом сокете
}

func (s socketUDP) readUnicast() {
	for {
		msg := make([]byte, 65507)
		n, addr, err := s.socket.ReadFromUDP(msg)
		if nil != err {
			fmt.Println(err)
		}

		chOutput <- Message{
			Msg:  msg[:n],
			Addr: addr,
		}
	}
}

func (s socketUDP) SendUnicast(msg []byte, addr *net.UDPAddr) {
	chInput <- Message{
		Msg:  msg,
		Addr: addr,
	}
}

func (s socketUDP) sendUnicast() {
	for {
		msg := <-chInput
		//fmt.Println("send", ok)
		bmsg, addr := msg.Msg, msg.Addr
		_, err := s.socket.WriteToUDP(bmsg, addr)
		if nil != err {
			fmt.Println(err.Error())
			return
		}
	}
}
