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
	UserID    int
	Username  string
	Handshake string
	Message   string
}

//!+broadcaster
type client chan<- []byte // an outgoing message channel

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan []byte) // all incoming client messages
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
	ch := make(chan []byte) // outgoing client messages
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
		if userinfo.Handshake == "Hello" {
			var userinfoOut UserInfo
			userinfoOut.Message = "<<<<Broadcast>>>>  You are " + who
			userinfoOut.Username = ""
			buffer, _ := json.Marshal(userinfoOut)
			ch <- protocol.Packet(buffer)

			userinfoOut.Message = "<<<<Broadcast>>>>  " + who + " has arrived"
			userinfoOut.Username = who
			buffer, _ = json.Marshal(userinfoOut)
			messages <- protocol.Packet(buffer)

			entering <- ch
		} else {
			if userinfo.Handshake == "quit" {
				break
			} else {
				var userinfoOut UserInfo
				userinfoOut.Message = who + " : " + userinfo.Message
				userinfoOut.Username = who
				buffer, _ := json.Marshal(userinfoOut)
				messages <- protocol.Packet(buffer)
			}
		}
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- ch
	var userinfoOut UserInfo
	userinfoOut.Message = who + " has left"
	userinfoOut.Username = who
	buffer, _ = json.Marshal(userinfoOut)
	messages <- protocol.Packet(buffer)
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan []byte) {
	for msg := range ch {
		conn.Write(msg) // NOTE: ignoring network errors
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
