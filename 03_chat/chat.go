package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Client struct {
	conn net.Conn
	ch   chan<- string
}

func main() {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgchan := make(chan string)
	addchan := make(chan Client)
	rmchan := make(chan net.Conn)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func handleConnection(c net.Conn, msgchan chan<- string, addchan chan<- Client, rmchan chan<- net.Conn) {
	ch := make(chan string)

	msgs := make(chan string)

	addchan <- Client{c, ch}

	go func() {
		defer close(msgs)

		bufc := bufio.NewReader(c)

		c.Write([]byte("\033[1;30;41mWelcome to the fancy demo chat!\033[0m\r\nWhat is your nick? "))
		nick, _, err := bufc.ReadLine()
		if err != nil {
			return
		}

		nickname := string(nick)

		c.Write([]byte("Welcome, " + nickname + "!\r\n\r\n"))

		msgs <- "New user " + nickname + " has joined the chat room."

		for {
			line, _, err := bufc.ReadLine()
			if err != nil {
				break
			}
			msgs <- nickname + ": " + string(line)
		}

		msgs <- "User " + nickname + " left the chat room."
	}()

LOOP:
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				break LOOP
			}
			msgchan <- msg
		case msg := <-ch:
			_, err := c.Write([]byte(msg))
			if err != nil {
				break LOOP
			}
		}
	}

	c.Close()
	fmt.Printf("Connection from %v closed.\n", c.RemoteAddr())
	rmchan <- c
}

func handleMessages(msgchan <-chan string, addchan <-chan Client, rmchan <-chan net.Conn) {
	clients := make(map[net.Conn]chan<- string)

	for {
		select {
		case msg := <-msgchan:
			fmt.Printf("new message: %s\n", msg)
			for _, ch := range clients {
				go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m\r\n" }(ch)
			}
		case client := <-addchan:
			fmt.Printf("New client: %v\n", client.conn)
			clients[client.conn] = client.ch
		case conn := <-rmchan:
			fmt.Printf("Client disconnects: %v\n", conn)
			delete(clients, conn)
		}
	}
}
