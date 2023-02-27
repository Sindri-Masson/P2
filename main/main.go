package main

import (
	"fmt"
	"net"

	"github.com/Sindri-Masson/P2/messenger"
	"github.com/Sindri-Masson/P2/registry"
)

func main() {
	println(messenger.NewMessagingNode("localhost:8080"))
	registry := registry.NewRegistry()

	// start listening for incoming connections
	ln, err := net.Listen("tcp", ":8080")
	//print
	fmt.Println("Registry is running on port 8080")
	if err != nil {
		// handle listen error
	}
	defer ln.Close()

	// accept incoming connections and register messaging nodes
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle accept error
		}

		// handle registration in a separate goroutine to avoid blocking
		go func(conn net.Conn) {
			err := registry.RegisterNode(conn)
			if err != nil {
				// handle registration error
			}
		}(conn)

		// handle incoming requests from registered messaging nodes
		go func(conn net.Conn) {
			defer conn.Close()
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					// handle read error
					break
				}

				// handle message type
				msgType := string(buf[:n])
				switch msgType {
				case "deregister":
					err := registry.DeregisterNode(conn)
					if err != nil {
						// handle deregistration error
					}
					return
				default:
					// handle unknown message type
				}
			}
		}(conn)
	}
}