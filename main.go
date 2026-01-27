package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("listening on port 6379")

	// Create a new server
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Listen for connection
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	for {
		buf := make([]byte, 1024)

		// Read message from client
		_, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error Reading from client: ", err.Error())
			os.Exit(1)
		}

		// Ignore requst and send back a PONG
		conn.Write([]byte("+OK\r\n"))
	}
}