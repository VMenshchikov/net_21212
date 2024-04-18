package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

type speedStruct struct {
	connTime   time.Time
	lastUse    time.Time
	bTotal     uint64
	bLastPrint uint64
	channle    chan int
	alive      chan bool
}

var mapUsers map[string]*speedStruct = nil

func updateSpeed(st *speedStruct) {
	for {
		bytes, status := <-st.channle
		//fmt.Println(bytes, status)
		if !status {
			close(st.alive)
			return
		}
		st.bTotal += uint64(bytes)
	}
}

func printSpeed(st *speedStruct, addr string) {
	execute := true
	for execute {
		select {
		case <-time.After(3 * time.Second):
		case <-st.alive:
			execute = false
			fmt.Println("REEEEEEEAD")
		}

		cTime := time.Now()
		alltime := cTime.Sub(st.connTime)
		totalSpeed := float64(st.bTotal) / alltime.Seconds() / 1024
		lastSpeed := float64(st.bTotal-st.bLastPrint) / time.Now().Sub(st.lastUse).Seconds() / 1024
		st.bLastPrint = st.bTotal
		st.lastUse = time.Now()

		fmt.Printf("Соединение %s: \n\t Средняя скорость: %f \n\t Мгновенная скорость: %f\n", addr, totalSpeed, lastSpeed)
	}
}

func slut(conn net.Conn) {
	defer conn.Close()
	defer close(mapUsers[conn.RemoteAddr().String()].channle)
	var (
		nameSise uint8
		nameFile string
		sizeFile uint64
	)
	nameSizeBuffer := make([]byte, 1)
	_, err := io.ReadFull(conn, nameSizeBuffer)
	if nil != err {
		fmt.Println("Ошибка при чтении:", err.Error())
		return
	}
	nameSise = uint8(nameSizeBuffer[0])

	nameFileBuffer := make([]byte, nameSise)
	_, err = io.ReadFull(conn, nameFileBuffer)
	if nil != err {
		fmt.Println("Ошибка при чтении:", err.Error())
		return
	}

	sizeFileBuffer := make([]byte, 8)
	_, err = io.ReadFull(conn, sizeFileBuffer)
	if nil != err {
		fmt.Println("Ошибка при чтении:", err.Error())
		return
	}
	sizeFile = binary.BigEndian.Uint64(sizeFileBuffer)

	os.Mkdir("uploads", 0777)
	nameFile = string(nameFileBuffer)
	file, err := os.OpenFile(filepath.Join("uploads", nameFile), os.O_WRONLY|os.O_CREATE, 0777)
	if nil != err {
		fmt.Println("Ошибка при открытии файла:", err.Error())
		return
	}
	defer file.Close()

	step := uint64(512)
	for i := uint64(0); i < sizeFile; {
		if sizeFile-i < step {
			step = sizeFile - i
		}
		buffer := make([]byte, step)
		n, err := io.ReadFull(conn, buffer)
		if nil != err {
			fmt.Println("Ошибка при чтении:", err.Error())
			return
		}
		i += uint64(n)
		file.Write(buffer)
		//fmt.Println("Я ПРОЧИТАЛ")
		//fmt.Println(mapUsers[conn.RemoteAddr().String()], "\n", mapUsers[conn.RemoteAddr().String()].channle)
		mapUsers[conn.RemoteAddr().String()].channle <- int(step)
		//fmt.Println("СООООКЕТ")

	}
	conn.Write([]byte{1})

	fmt.Printf("Файл %s от %s принят\n", nameFile, conn.RemoteAddr().String())
	return
}

func Init(port uint16) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	//defer close?
	if nil != err {
		return err
	}

	fmt.Printf("Сервер запущен на %s\n", listener.Addr().String())

	mapUsers = make(map[string]*speedStruct)

	for {
		conn, err := listener.Accept()
		flagContinue := true
		if nil != err {
			fmt.Println("Ошибка при принятии соеденения:", err)
			continue
		}
		for {
			protocol := make([]byte, 10)
			count, err := io.ReadFull(conn, protocol)
			if nil != err {
				fmt.Println("Не удалось прочитать")
				flagContinue = false
			}
			if 10 == count {
				if string(protocol) != "myProtocol" {
					fmt.Println("Ошибочный протокол")
					flagContinue = false
				}
				break
			}
		}
		err = conn.SetDeadline(time.Time{})
		if !flagContinue {
			conn.Close()
			continue
		}
		m := new(speedStruct)
		m.lastUse = time.Now()
		m.connTime = time.Now()
		m.alive = make(chan bool)
		m.channle = make(chan int)

		mapUsers[conn.RemoteAddr().String()] = m
		go slut(conn)
		go updateSpeed(m)
		go printSpeed(m, conn.RemoteAddr().String())
	}

}
