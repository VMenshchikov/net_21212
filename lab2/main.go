package main

import (
	"fmt"
	"myftp/client"
	"myftp/server"
)

func main() {
	var mode uint8
	for true {
		fmt.Printf("Select mode:\n\t1) Server\n\t2) Client\nMode: ")
		fmt.Scan(&mode)
		if 1 == mode || 2 == mode {
			break
		}
	}

	if 1 == mode {
		var port uint16
		fmt.Printf("Enter port:")
		fmt.Scan(&port)
		err := server.Init(port)
		if nil != err {
			fmt.Println(err.Error())
		}
	}
	if 2 == mode {
		client.Client()
	}

}
