package main

import (
	"P2/minichord"
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"google.golang.org/protobuf/proto"
)

var registry string
var address string
var id int32
var routingTable map[int32]ConnectedNode
var ConnectedIds []int
var AllNodeIds []int32

type ConnectedNode struct {
	Address string
	Conn   net.Conn
}

func findClosest(dest int32) int32 {
	// Find closest node
	var _, ok = routingTable[dest]
	if ok {
		return dest
	}
	i := 0
	for i < len(AllNodeIds) {
		if AllNodeIds[i] == dest {
			break
		}
		i++
	}
	index := i
	for i := index; i > 0; i-- {
		_, ok = routingTable[AllNodeIds[i]]
		if ok {
        	return AllNodeIds[i]
		}
    }
	for i := len(AllNodeIds) - 1; i >= index; i-- {
		_, ok = routingTable[AllNodeIds[i]]
		if ok {
        	return AllNodeIds[i]
		}
    }

		

	return -1
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
	case "nodeRegistryResponse":
		if message == "1" {
			return &minichord.MiniChord{
				Message: &minichord.MiniChord_NodeRegistryResponse{
					NodeRegistryResponse: &minichord.NodeRegistryResponse{
						Result: 1,
						Info: "Node registry received",
					},
				},
			}
		} else {
			return &minichord.MiniChord{
				Message: &minichord.MiniChord_NodeRegistryResponse{
					NodeRegistryResponse: &minichord.NodeRegistryResponse{
						Result: 0,
						Info: "Node registry failed",
					},
				},
			}
		}
	case "nodeData":
		data := strings.Split(message, ",")
		dest, _ := strconv.Atoi(data[0])
		source, _ := strconv.Atoi(data[1])
		payload, _ := strconv.Atoi(data[2])
		hopCount, _ := strconv.Atoi(data[3])

		trace := StoI(data[4:])

		return &minichord.MiniChord{
			Message: &minichord.MiniChord_NodeData{
				NodeData: &minichord.NodeData{
					Destination: int32(dest),
					Source: int32(source),
					Payload: int32(payload),
					Hops: uint32(hopCount),
					Trace: trace,
				},
			},
		}
	case "taskFinished":
		// TODO: Implement
		new_message := strings.Split(message, ",")
		id := new_message[0]
		int_id, _ := strconv.Atoi(id)
		address := new_message[1]
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_TaskFinished{
				TaskFinished: &minichord.TaskFinished{
					Id: int32(int_id),
					Address: address,
				},
			},
		}
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
func ConnectToNode (address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to node:", err.Error())
	}
	return conn
}

func sendPackets (number int) {
	for i := 0; i < number; i++ {
		// Get random node
		rand.Seed(time.Now().UnixNano())
		receiver := AllNodeIds[rand.Intn(len(AllNodeIds))]
		// Get random payload
		randomPayload := rand.Intn(4294967295) - 2147483648
		// Send message
		message := strconv.Itoa(int(receiver)) + "," + strconv.Itoa(int(id)) + "," + strconv.Itoa(randomPayload) + "," + strconv.Itoa(0) + "," + strconv.Itoa(int(id))
		// find closest node in routing table
		closestNode := findClosest(int32(receiver))
		fmt.Println("My table is ", ConnectedIds, "Closest node to ", receiver, " is ", closestNode)
		
		Sender(routingTable[closestNode].Address, message, "nodeData")
	}
}

func remove(slice []int32, s int32) []int32 {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func readMessage(data []byte) {
	// read message from node using minichord
	envelope := &minichord.MiniChord{}
	err := proto.Unmarshal(data, envelope)
	if err != nil {
		fmt.Println("Error unmarshalling message:", err.Error())
	}
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
		AllNodeIds = envelope.GetNodeRegistry().Ids
		AllNodeIds = remove(AllNodeIds, id)
		sort.Slice(AllNodeIds, func(i, j int) bool { return AllNodeIds[i] < AllNodeIds[j] })
		routingTable = make(map[int32]ConnectedNode)
		failed := false
		for i := 0; i < len(envelope.GetNodeRegistry().Peers); i++ {
			conn, err := net.Dial("tcp", envelope.GetNodeRegistry().Peers[i].Address)
			if err != nil {
				fmt.Println("Error connecting to node:", err.Error())
				failed = true
				break
			}
			routingTable[envelope.GetNodeRegistry().Peers[i].Id] = ConnectedNode{Address: envelope.GetNodeRegistry().Peers[i].Address, Conn: conn}
			ConnectedIds = append(ConnectedIds, int(envelope.GetNodeRegistry().Peers[i].Id))
		}
		if failed {
			fmt.Println("Error connecting to nodes, exiting")
			go Sender(registry, "2", "nodeRegistryResponse")
			ConnectedIds = nil
		} else {
			// Sort the connected ids
			sort.Ints(ConnectedIds)
			go Sender(registry, "1", "nodeRegistryResponse")
		}
	case *minichord.MiniChord_NodeRegistryResponse:
		fmt.Println("Node registry response received, should not happen")
	case *minichord.MiniChord_InitiateTask:
		go sendPackets(int(envelope.GetInitiateTask().Packets))
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
		Sender(registry,"", "reportTrafficSummary")
		// TODO: Implement
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


// func sendMessage(node MessagingNode, message string) {
// 	// send message to node using minichord
// 	fmt.Println("sending message to node: ", message)
// 	envelope := constructMessage(message, "registration")
// 	fmt.Println("envelope: ", envelope)
// 	data, err := proto.Marshal(envelope)
// 	fmt.Println("marshalled data: ", data)
// 	new_data, err := node.conn.Write(data)
// 	if err != nil {
// 		fmt.Println("Error sending message:", err.Error())
// 	}
// 	fmt.Println("sending message: ", new_data)
	
// }


func handleConnection(conn net.Conn) {
	defer conn.Close()

    for {
        // Create a buffer to read into
        buf := make([]byte, 1024)

        // Read from the connection into the buffer
        n, err := conn.Read(buf)
        if err != nil {
            if err == io.EOF {
                // Connection closed by remote host
                return
            }
            continue
        }

        // Process the data that was read
        data := buf[:n]
		readMessage(data)
	}
	
}

func StoI(s []string) []int32 {
	var t2 = []int32{}
	for _, v := range s {
		t, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		t2 = append(t2, int32(t))
	}
	return t2
}

func readUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		// change id to string
		id_string := strconv.Itoa(int(id))
		cmd = strings.Trim(cmd, "\n")
		var cmdSlice = strings.Split(cmd, " ")
		switch (cmdSlice[0]) {
		case "print":
			Sender(registry, id_string, "requestTrafficSummary")
		case "exit":
			fmt.Println("Deregistering from the network")
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
			go handleConnection(conn)
		}
		
	}()
	select{}
}
