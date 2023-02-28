package main

import (
	"fmt"
	"net"
	"os"
	"P2/minichord"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type MessagingNode struct {
	id      int
	address string
	conn    net.Conn
}

func constructMessage(message string, typ string) *minichord.MiniChord {
	switch typ {
	case "registration":
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_Registration{
				Registration: &minichord.Registration{
					Address: message,
				},
			},
		}
	case "registrationResponse":
		id, _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_RegistrationResponse{
				RegistrationResponse: &minichord.RegistrationResponse{
					Result: int32(id),
					Info: "Registered successfully",
				},
			},
		}
	case "deregistration":
		fmt.Println("Message: ", message)
		/* return &minichord.MiniChord{
			Message: &minichord.MiniChord_Deregistration{
				Deregistration: &minichord.Deregistration{
					Node: message,
				},
			},
		} */
		return nil
	case "deregistrationResponse":
		id, _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
					Result: int32(id),
					Info: "Deregistered successfully",
				},
			},
		}
	case "nodeRegistry":
		// TODO: Implement
		return nil
	case "nodeRegistryResponse":
		// TODO: Implement
		return nil
	case "initiateTask":
		// TODO: Implement
		return nil
	case "nodeData":
		// TODO: Implement
		return nil
	case "taskFinished":
		// TODO: Implement
		return nil
	case "requestTrafficSummary":
		// TODO: Implement
		return nil
	case "reportTrafficSummary":
		// TODO: Implement
		return nil
	default:
		return nil
	}
}


func sendMessage(node MessagingNode, message string) {
	// send message to node using minichord
	envelope := constructMessage(message, "registration")
	data, err := proto.Marshal(envelope)
	fmt.Println(data)
	new_data, err := node.conn.Write([]byte(data))
	if err != nil {
		fmt.Println("Error sending message:", err.Error())
	}
	fmt.Println("sending message: ", new_data)
}


func handleConnection(conn net.Conn, connections *[]MessagingNode, otherPort string) {
	// add connection to list of connections
	// create a new MessagingNode
	node := MessagingNode{len(*connections)+1, conn.RemoteAddr().String(), conn}
	*connections = append(*connections, node)
}

func main() {
	// get command line arguments go tun messenger.go <port> <otherPort>
	address := os.Args[1]
	var connections []MessagingNode
	// bind to port and start listening
	ln, err := net.Listen("tcp", address)
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
			go handleConnection(conn, &connections, address)
		}
	}()	
}
