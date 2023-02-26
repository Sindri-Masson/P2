package main

import (
	"fmt"
	"net"
	"time"
	"math/rand"
	"strings"
)

type MessagingNode struct {
	id      int
	address string
	conn    net.Conn
}

func NewMessagingNode(registryAddr string) (*MessagingNode, error) {
	// choose a random port to listen on
	port := rand.Intn(65536-1024) + 1024
	addr := fmt.Sprintf(":%d", port)

	// start listening for incoming connections
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Listening on %s\n", ln)

	// register with the registry
	conn, err := net.Dial("tcp", registryAddr)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(conn, "register %s\n", addr)

	// create and return new messaging node
	id := rand.Intn(128)
	return &MessagingNode{
		id:      id,
		address: addr,
		conn:    conn,
	}, nil
}

func (m *MessagingNode) Run() {
	// start accepting incoming connections
	go func() {
		for {
			conn, err := m.acceptConnection()
			if err != nil {
				// handle accept error
			}
			go m.handleIncomingMessages(conn)
		}
	}()

	// periodically send messages to the sink
	for {
		m.sendRandomMessage()
		time.Sleep(1 * time.Second)
	}
}

// func (m *MessagingNode) acceptConnection() (net.Conn, error) {
// 	conn, err := m.conn.Accept()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

func (m *MessagingNode) handleIncomingMessages(conn net.Conn) {
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
		case "message":
			// handle message
			fmt.Printf("Node %d received message: %s\n", m.id, string(buf[n:]))
		default:
			// handle unknown message type
		}
	}
}

func (m *MessagingNode) sendRandomMessage() {
	// choose a random node to forward message to
	destID := rand.Intn(128)
	destAddr, err := m.getRoutingTableEntry(destID)
	if err != nil {
		// handle routing error
		return
	}

	// send message to destination node
	conn, err := net.Dial("tcp", destAddr)
	if err != nil {
		// handle dial error
		return
	}
	defer conn.Close()

	msg := fmt.Sprintf("message from %d to %d", m.id, destID)
	conn.Write([]byte(msg))
}

func (m *MessagingNode) getRoutingTableEntry(id int) (string, error) {
	// query the registry for routing information
	fmt.Fprintf(m.conn, "get_routing_table %d\n", id)
	buf := make([]byte, 1024)
	n, err := m.conn.Read(buf)
	if err != nil {
		return "", err
	}

	// parse routing table entry from response
	resp := string(buf[:n])
	parts := strings.Split(resp, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid routing table entry: %s", resp)
	}

	return parts[1], nil
}

func mmain() {
	// Initialize the registry
	// registry := NewRegistry()

	// Initialize the messaging nodes
	// var nodes []*MessagingNode
	// for i := 0; i < 10; i++ {
	// 	node, err := NewMessagingNode(registry.Address)
	// 	if err != nil {
	// 		fmt.Printf("Failed to create node: %v\n", err)
	// 		continue
	// 	}
	// 	nodes = append(nodes, node)

	// 	// Start the node's main loop
	// 	go node.Run()
	// }

	// Wait for messages to be sent and received
	select {}
}

