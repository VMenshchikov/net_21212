package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func read(socket *net.UDPConn, addr chan *net.UDPAddr) {
	_, ad, err := socket.ReadFromUDP([]byte{})
	if nil != err {
		fmt.Printf("Ошибка получения: %v\n", err)
	}
	addr <- ad
}

func createSendUDPSocket(addr *net.UDPAddr) (*net.UDPConn, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, errors.New("Ошибка при получении интерфейсов:" + err.Error())
	}

	fmt.Println("Список интерфейсов:")
	for ind, val := range interfaces {
		fmt.Print("\t", ind+1, ". ", val, "\n")
	}

	fmt.Println("Выберете интерфейс для отправки:")
	var interf int
	fmt.Scan(&interf)

	ips, _ := interfaces[interf-1].Addrs()

	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(strings.Split(ips[0].String(), "/")[0]),
		Port: 0,
	}

	soc, err := net.DialUDP("udp", localAddr, addr)

	return soc, nil
}

func createMulticastSocket(addr *net.UDPAddr) (*net.UDPConn, error) {

	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, errors.New("Ошибка при получении интерфейсов:" + err.Error())
	}

	fmt.Println("Список интерфейсов:")
	for ind, val := range interfaces {
		fmt.Print("\t", ind+1, ". ", val, "\n")
	}
	fmt.Println("Выберете интерфейс для чтения:")
	var interf int
	fmt.Scan(&interf)

	listenSocket, err := net.ListenMulticastUDP("udp", &interfaces[interf-1], addr)

	if err != nil {
		return nil, errors.New("Ошибка при бинде к мультикаст группе:" + err.Error())
	}

	return listenSocket, nil
}

func selectMulticastAddres() (*net.UDPAddr, error) {
	fmt.Print("Введите адрес мультикаст группы: ")
	var addrs string
	fmt.Scan(&addrs)

	multicastAddres := net.ParseIP(addrs)
	if multicastAddres == nil {
		return nil, errors.New("Введенная строка не является ip адресом.")
	}

	if !multicastAddres.IsMulticast() {
		return nil, errors.New("Передан не multicast адрес.")
	}

	const port = 1234
	ret := new(net.UDPAddr)
	ret.IP = multicastAddres
	ret.Port = port
	return ret, nil
}

func printMap(m map[string]uint8) {
	fmt.Printf("\n Актуальный список (%d пользователей подключено):\n", len(m))

	for key, val := range m {
		fmt.Printf("\t Пользователь %s побъявлялся %d циклов назад\n", key, val)
	}
}

func main() {
	multicastAddr, err := selectMulticastAddres()
	if err != nil {
		fmt.Println("Ошибка мультикаст-адреса:", err)
		os.Exit(0)
	}

	listenSocket, err := createMulticastSocket(multicastAddr)
	if err != nil {
		fmt.Println("Ошибка при создании мультикаст-сокета", err)
		os.Exit(0)
	}
	defer listenSocket.Close()

	sendSocket, err := createSendUDPSocket(multicastAddr)
	if err != nil {
		fmt.Println("Ошибка при создании сокета отправки", err)
		os.Exit(0)
	}
	defer sendSocket.Close()

	if err != nil {
		fmt.Println("Ошибка при создании UDP-соединения:", err)
		os.Exit(0)
	}
	defer sendSocket.Close()

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	m := make(map[string]uint8)
	addr := make(chan *net.UDPAddr)
	go read(listenSocket, addr)

	for true {
		for key, val := range m {
			m[key] = val + 1
			if m[key] == 3 {
				delete(m, key)
				fmt.Println("Пользователь отключился:", key)
				printMap(m)
			}
		}

		sendSocket.Write([]byte{})

		//fmt.Printf("\"before for\": %v\n", "before for")
		channel := make(chan bool)
		go func() {
			select {
			case <-time.After(3 * time.Second):
				channel <- false
			}
		}()
		isRun := true
		for isRun {

			//fmt.Printf("\"for\": %v\n", "for")
			select {
			case isRun = <-channel:
			case data := <-addr:
				_, exist := m[data.String()]
				m[data.String()] = 0
				if exist == false {
					fmt.Println("Подключился:", data.String())
					printMap(m)
				}
				go read(listenSocket, addr)
			}
		}

	}
}
