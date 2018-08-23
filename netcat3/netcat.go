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
	"io"
	"log"
	"net"
	"os"
)

type UserInfo struct {
	Username string
	Message  string
}

//!+
func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn) // NOTE: ignoring errors
		log.Println("done")
		done <- struct{}{} // signal the main goroutine
	}()

	mustCopy(conn, os.Args[1])
	conn.Close()
	<-done // wait for background goroutine to finish
}

//!-
func mustCopy(dst net.Conn, uname string) {

	userInfs := UserInfo{
		Username: uname,
		Message:  "",
	}
	buffer, _ := json.Marshal(userInfs)
	dst.Write(protocol.Packet(buffer))
	//Log(string(protocol.Packet(buffer)))
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		userInfs := UserInfo{
			Username: uname,
			Message:  input.Text(),
		}
		buffer, _ := json.Marshal(userInfs)
		dst.Write(protocol.Packet(buffer))
		//Log(string(protocol.Packet(buffer)))
	}
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}
