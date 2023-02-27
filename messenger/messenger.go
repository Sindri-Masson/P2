package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/protobuf/proto"
)

type MessagingNode struct {
	id      int
	address string
	conn    net.Conn
}


func sendMessage(node MessagingNode, message string) {
	// send message to node using minichord
	data, err =  proto.Marshal(message)
	_, err := node.conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err.Error())
	}
}


func handleConnection(conn net.Conn, connections *[]MessagingNode, otherPort string) {
	// add connection to list of connections
	// create a new MessagingNode
	node := MessagingNode{len(*connections)+1, conn.RemoteAddr().String(), conn}
	*connections = append(*connections, node)
}

func main() {
	// get command line arguments go tun messenger.go <port> <otherPort>
	port := os.Args[1]
	otherPort := ""
	var connections []MessagingNode
	// check number of arguments
	if len(os.Args) == 3 {
		otherPort = os.Args[2]
	}
	// bind to port and start listening
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	go func () {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			go handleConnection(conn, &connections, otherPort)
		}
	}()
	if otherPort != "" {
		// connect to other node
		// create a connection to the other node
		conn, err := net.Dial("tcp", ":"+otherPort)
		if err != nil {
			fmt.Println("Error dialing:", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, &connections, otherPort)
	}
	
}
