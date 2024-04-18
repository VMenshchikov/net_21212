package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func getIP(s string) net.IP {
	ip := net.ParseIP(s)

	if nil == ip {
		arrIP, _ := net.LookupIP(s)
		if nil == arrIP {
			return nil
		}
		return arrIP[0]
	}
	return ip
}

func Client() {
	var path string

	interfaces, err := net.Interfaces()

	if err != nil {
		println("Ошибка при получении интерфейсов:" + err.Error())
		return
	}

	fmt.Println("Введите путь к файлу:")
	fmt.Scanf("%s", &path)

	file, err := os.OpenFile(path, os.O_RDONLY, os.FileMode(0755))
	if nil != err {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()

	filename := filepath.Base(path)

	fmt.Println("Введите DNS или IP:")
	var tmp string
	fmt.Scanf("%s", &tmp)

	ip := getIP(tmp)

	if nil == ip {
		println("Введен не DNS или IP")
		return
	}

	fmt.Println("Введите номер порта:")
	var port int
	fmt.Scan(&port)

	for i, v := range interfaces {
		fmt.Println(i+1, v)
	}

	fmt.Println("Выберете интерфейс для отправки:")
	var interf int
	fmt.Scan(&interf)

	ips, _ := interfaces[interf-1].Addrs()

	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(strings.Split(ips[0].String(), "/")[0]),
		Port: 0,
	}

	soc, err := net.DialTCP("tcp", (*net.TCPAddr)(localAddr), &net.TCPAddr{
		IP:   ip,
		Port: port,
	})

	if nil != err {
		fmt.Println(err.Error())
		return
	}

	fStat, err := file.Stat()
	if nil != err {
		fmt.Println(err.Error())
		return
	}

	soc.Write([]byte("myProtocol"))
	soc.Write([]byte{byte(len(filename))})
	soc.Write([]byte(filename))
	sizeFile := [8]byte{}
	binary.BigEndian.PutUint64(sizeFile[:], uint64(fStat.Size()))
	soc.Write(sizeFile[:])

	buffer := [512]byte{}
	for {
		//fmt.Println(n)
		n, err := file.Read(buffer[:])
		if nil != err {
			fmt.Println(err.Error())
			break
		}
		m, err := soc.Write(buffer[:n])
		if nil != err {
			fmt.Println(err.Error(), m)
			break
		}
	}

	result := [1]byte{}
	soc.Read(result[:])

	if result[0] == byte(1) {
		fmt.Printf("Файл %s для %s:%d успешно передан\n", filename, ip.String(), port)
	} else {
		fmt.Printf("Файл %s для %s:%d не был передан\n", filename, ip.String(), port)

	}
	return
}
