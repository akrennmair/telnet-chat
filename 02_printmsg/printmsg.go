package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgchan := make(chan string)

	go printMessages(msgchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, msgchan)
	}
}

func handleConnection(c net.Conn, msgchan chan<- string) {
	buf := make([]byte, 4096)

	for {
		n, err := c.Read(buf)
		if err != nil || n == 0 {
			c.Close()
			break
		}
		msgchan <- string(buf[0:n])
		n, err = c.Write(buf[0:n])
		if err != nil {
			c.Close()
			break
		}
	}
	fmt.Printf("Connection from %v closed.\n", c.RemoteAddr())
}

func printMessages(msgchan <-chan string) {
	for {
		msg := <-msgchan
		fmt.Printf("new message: %s\n", msg)
	}
}
