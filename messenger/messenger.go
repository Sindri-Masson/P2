package main

import (
	"fmt"
	"net"
	"os"
	"P2/minichord"
	"google.golang.org/protobuf/proto"
	"strconv"
	"math/rand"
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


func readMessage(node MessagingNode) {
	// read message from node using minichord
	fmt.Println("reading message from node")
	data := make([]byte, 1024)
	n, err := node.conn.Read(data)
	if err != nil {
		fmt.Println("Error reading message:", err.Error())
	}
	fmt.Println("received message: ", data[:n])
	envelope := &minichord.MiniChord{}
	err = proto.Unmarshal(data[:n], envelope)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err.Error())
	}
	fmt.Println("unmarshalled message: ", envelope)
	switch envelope.Message.(type) {
	case *minichord.MiniChord_RegistrationResponse:
		fmt.Println("Registration response received")
		fmt.Println("Node ID: ", envelope.GetRegistrationResponse().Result)
		fmt.Println("Info: ", envelope.GetRegistrationResponse().Info)
	case *minichord.MiniChord_DeregistrationResponse:
		fmt.Println("Deregistration response received")
		fmt.Println("Node ID: ", envelope.GetDeregistrationResponse().Result)
		fmt.Println("Info: ", envelope.GetDeregistrationResponse().Info)
	case *minichord.MiniChord_NodeRegistry:
		fmt.Println("Node registry received")
		fmt.Println("Node ID: ", envelope.GetNodeRegistry().Ids)
		fmt.Println("Info: ", envelope.GetNodeRegistry().Peers)
	case *minichord.MiniChord_NodeRegistryResponse:
		fmt.Println("Node registry response received")
		fmt.Println("Node ID: ", envelope.GetNodeRegistryResponse().Result)
		fmt.Println("Info: ", envelope.GetNodeRegistryResponse().Info)
	case *minichord.MiniChord_InitiateTask:
		fmt.Println("Initiate task received")
		fmt.Println("Packets: ", envelope.GetInitiateTask().Packets)
	case *minichord.MiniChord_NodeData:
		fmt.Println("Node data received")
		fmt.Println("Source: ", envelope.GetNodeData().Source)
		fmt.Println("Info: ", envelope.GetNodeData().Payload)
	case *minichord.MiniChord_TaskFinished:
		fmt.Println("Task finished received")
		fmt.Println("Node ID: ", envelope.GetTaskFinished().Id)
		fmt.Println("Info: ", envelope.GetTaskFinished().Address)
	case *minichord.MiniChord_RequestTrafficSummary:
		fmt.Println("Request traffic summary received")
	case *minichord.MiniChord_ReportTrafficSummary:
		fmt.Println("Report traffic summary received")
		fmt.Println("Node ID: ", envelope.GetReportTrafficSummary().Id)
		fmt.Println("Sent: ", envelope.GetReportTrafficSummary().Sent)
		fmt.Println("Relayed: ", envelope.GetReportTrafficSummary().Relayed)
		fmt.Println("Received: ", envelope.GetReportTrafficSummary().Received)
		fmt.Println("Total Sent: ", envelope.GetReportTrafficSummary().TotalSent)
		fmt.Println("Total Received: ", envelope.GetReportTrafficSummary().TotalReceived)
	default:
		fmt.Println("Unknown message received")
	}
}



func sendMessage(node MessagingNode, message string) {
	// send message to node using minichord
	fmt.Println("sending message to node: ", message)
	envelope := constructMessage(message, "registration")
	data, err := proto.Marshal(envelope)
	fmt.Println("marshalled data: ", data)
	new_data, err := node.conn.Write(data)
	if err != nil {
		fmt.Println("Error sending message:", err.Error())
	}
	fmt.Println("sending message: ", new_data)
	
}


func handleConnection(conn net.Conn, connections *[]MessagingNode, otherPort string) {
	// add connection to list of connections
	// create a new MessagingNode
	fmt.Printf("handling connection")
	node := MessagingNode{len(*connections)+1, conn.RemoteAddr().String(), conn}
	*connections = append(*connections, node)
}

func main() {
	// get command line arguments go tun messenger.go <port> <otherPort>
	address := os.Args[1]
	var connections []MessagingNode
	port := rand.Intn(65536-1024) + 1024
	port_str := strconv.Itoa(port)
	// bind to port and start listening
	ln, err := net.Listen("tcp", "localhost:" + port_str)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error dialing:", err.Error())
		os.Exit(1)
	}
	node := MessagingNode{1, address, conn}
	sendMessage(node, port_str)
	// close the listener when the application closes
	go func () {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			fmt.Printf("ding")
			go handleConnection(conn, &connections, address)
		}
	}()
	go readMessage(node)
	select{}
}
