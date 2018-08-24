// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 227.

// Netcat is a simple read/write client for TCP servers.
package main

import (
	"GroupChat/protocol"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type UserInfo struct {
	UserID    int
	Username  string
	Handshake string
	Message   string
}

//!+
func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Please enter your username:")
	input := bufio.NewScanner(os.Stdin)

	done := make(chan struct{})
	if input.Scan() {
		go func(uname string) {
			for true {
				tmpBuffer := make([]byte, 0)
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					break
				}
				tmpBuffer = protocol.Unpack(buffer, n)
				var userinfo UserInfo
				json.Unmarshal(tmpBuffer, &userinfo)
				if userinfo.Username != uname {
					fmt.Println(userinfo.Message)
				}
			}
			//io.Copy(os.Stdout, conn) // NOTE: ignoring errors
			log.Println("Disconnted")
			done <- struct{}{} // signal the main goroutine
		}(input.Text())
		mustCopy(conn, input.Text())
	}

	conn.Close()
	<-done // wait for background goroutine to finish
}

//!-
func mustCopy(dst net.Conn, uname string) {

	userInfs := UserInfo{
		Username:  uname,
		Handshake: "Hello",
		Message:   "",
	}
	buffer, _ := json.Marshal(userInfs)
	dst.Write(protocol.Packet(buffer))
	//Log(string(protocol.Packet(buffer)))
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		userInfs := UserInfo{
			Handshake: "",
			Username:  uname,
			Message:   input.Text(),
		}
		if userInfs.Message == "/quit" {
			userInfs.Handshake = "quit"
			userInfs.Message = ""
		}

		buffer, _ := json.Marshal(userInfs)
		dst.Write(protocol.Packet(buffer))
		//Log(string(protocol.Packet(buffer)))
	}
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}
