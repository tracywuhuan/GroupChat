// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 254.
//!+

// Chat is a server that lets clients chat with each other.
package main

import (
	"GroupChat/protocol"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type UserInfo struct {
	Username string
	Message  string
}

//!+broadcaster
type client chan<- string // an outgoing message channel

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client messages
)

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

//!-broadcaster

//!+handleConn
func handleConn(conn net.Conn) {
	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()

	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	for true {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		tmpBuffer = protocol.Unpack(buffer, n)
		var userinfo UserInfo
		json.Unmarshal(tmpBuffer, &userinfo)
		who = userinfo.Username
		if userinfo.Message == "" {
			ch <- "You are " + who
			messages <- who + " has arrived"
			entering <- ch
		} else {
			if userinfo.Message == "quit" {
				break
			} else {
				messages <- who + ": " + userinfo.Message
			}
		}
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}

//!-handleConn

//!+main
func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		Log(conn.RemoteAddr().String(), " tcp connect success")
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
func Log(v ...interface{}) {
	fmt.Println(v...)
}

//!-main
