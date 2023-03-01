package main

import (
	"P2/minichord"
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"google.golang.org/protobuf/proto"
)

var registry string
var registryConn net.Conn
var address string
var id int32
var routingTable map[int32]string
var AllNodeIds []int32


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
		id, _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_Deregistration{
				Deregistration: &minichord.Deregistration{
					Node: &minichord.Node{
						Id: int32(id),
						Address: address,
					},
				},
			},
		}
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


func readMessage(conn net.Conn) {
	// read message from node using minichord
	fmt.Println("reading message from someone")
	data := make([]byte, 65535)

	n, _ := conn.Read(data)
	
	envelope := &minichord.MiniChord{}
	err := proto.Unmarshal(data[:n], envelope)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err.Error())
	}
	fmt.Println("unmarshalled message: ", envelope)
	switch envelope.Message.(type) {
	case *minichord.MiniChord_Registration:
		fmt.Println("Registration received, should not happen")
	case *minichord.MiniChord_RegistrationResponse:
		fmt.Println("Registration response received")
		fmt.Println("Node ID: ", envelope.GetRegistrationResponse().Result)
		fmt.Println("Info: ", envelope.GetRegistrationResponse().Info)
		if envelope.GetRegistrationResponse().Result < 1 {
			fmt.Println("Registration failed")
			os.Exit(1)
		}
		id = envelope.GetRegistrationResponse().Result

	case *minichord.MiniChord_DeregistrationResponse:
		fmt.Println("Deregistration response received, should not happen")

	case *minichord.MiniChord_NodeRegistry:
		fmt.Println("Node registry received")
		fmt.Println("Size of table: ", envelope.GetNodeRegistry().NR)
		fmt.Println("Peers: ", envelope.GetNodeRegistry().Peers)
		fmt.Println("Number of IDs in network: ", envelope.GetNodeRegistry().NoIds)
		fmt.Println("All Ids: ", envelope.GetNodeRegistry().Ids)
		AllNodeIds = envelope.GetNodeRegistry().Ids
		routingTable = make(map[int32]string)
		for i := 0; i < len(envelope.GetNodeRegistry().Peers); i++ {
			routingTable[envelope.GetNodeRegistry().Peers[i].Id] = envelope.GetNodeRegistry().Peers[i].Address
		}

	case *minichord.MiniChord_NodeRegistryResponse:
		fmt.Println("Node registry response received, should not happen")

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

func Sender(receiver string, senderMessage string, typ string) {
	conn, err := net.Dial("tcp", receiver)
	if err != nil {fmt.Println("Error Conn", err)}

	message := constructMessage(senderMessage, typ)
	data, err := proto.Marshal(message)
	if err != nil {fmt.Println("Error Marshal", err)}

	// send message
	_, err = conn.Write(data)
	if err != nil {fmt.Println("Error Write", err)}
}


func sendMessage(node MessagingNode, message string) {
	// send message to node using minichord
	fmt.Println("sending message to node: ", message)
	envelope := constructMessage(message, "registration")
	fmt.Println("envelope: ", envelope)
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

func readUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		cmd = strings.Trim(cmd, "\n")
		var cmdSlice = strings.Split(cmd, " ")
		switch (cmdSlice[0]) {
		case "print":
			fmt.Println("Asked to print")
		case "exit":
			fmt.Println("Deregistering from the network")
			// change id to string
			id_string := strconv.Itoa(int(id))

			Sender(registry, id_string, "deregistration")
		default:
			fmt.Printf("command not understood: %s\n", cmd)
		}
	}
}

func main() {
	go readUserInput()
	// get command line arguments go tun messenger.go <port> <otherPort>
	registry = os.Args[1]
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(1000) + 1024
	port_str := strconv.Itoa(port)
	fmt.Println("port: ", port_str)
	address = "127.0.0.1:" + port_str

	// bind to port and start listening
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("connecting to ", registry)
	conn, err := net.Dial("tcp", registry)
	if err != nil {
		fmt.Println("Error connecting to registry:", err.Error())
		os.Exit(1)
	}
	fmt.Println("connected to ", conn.RemoteAddr().String())
	if err != nil {
		fmt.Println("Error dialing:", err.Error())
		os.Exit(1)
	}
	
	// send registration message to registry
	Sender(registry, address, "registration")
	
	// continuously accept connections
	go func () {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			fmt.Printf("ding")
			go readMessage(conn)
		}
		
	}()
	select{}
}
