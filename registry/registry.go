package main

import (
	"P2/minichord"
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	//"time"

	"google.golang.org/protobuf/proto"
)

const TYPE = "tcp"

var registry Registry;
var routingTables map[int][]int = make(map[int][]int)
var readyCounter = 0
var finishedCounter = 0
var summaryMap = make(map[int32]Summary)
var counter = 0

type Summary struct {
	Sent uint32
	Received uint32
	Relayed uint32
	SentSummary int64
	ReceivedSummary int64
}

// Registry represents the system registry
type Registry struct {
	nodes map[int]string // map of registered messaging nodes
	mu    sync.Mutex // mutex to protect concurrent access to the nodes map
}

// NewRegistry creates a new registry instance
func NewRegistry() Registry {
	return Registry{
		nodes: make(map[int]string),
	}
}

// RegisterNode registers a messaging node in the system and assigns an ID
func RegisterNode(conn net.Conn, message minichord.MiniChord) error {
	addr := message.GetRegistration().GetAddress()
	fmt.Println("Registering node: ", addr)
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// check if registry is full
	if len(registry.nodes) == 128 {
		Sender(addr, "-1", "registrationResponse")
		return fmt.Errorf("registry is full")
	}

	// check if address is the same as message sender
	/* if addr.String() != message.GetRegistration().Address {
		Sender(addr.String(), "-3", "registrationResponse")
		return fmt.Errorf("address mismatch: %s != %s", addr, message.GetRegistration().Address)
	} */

	// check if node is already registered
	for _, x := range registry.nodes {
		if x == addr {
			Sender(addr, "-2", "registrationResponse")
			return fmt.Errorf("node already registered: %s", addr)
		}
	}
	

	id := rand.Intn(128)
	_, ok := registry.nodes[id]
	for ok { // ensure no ID collisions
		id = rand.Intn(128)
		_, ok = registry.nodes[id]
	}
	registry.nodes[id] = addr
	fmt.Println("Registered node: ", id, addr)

	// send response to the messaging node

	Sender(addr, strconv.Itoa(int(id)), "registrationResponse")

	return nil
}

// DeregisterNode deregisters a messaging node from the system
func DeregisterNode(add string, id int) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	fmt.Println("Deregistering node:", id, "At address: ", add, "Node address: ", registry.nodes[id])

	// check if node is registered
	var _, ok = registry.nodes[id]
	if !ok {
		Sender(add, "-1", "deregistrationResponse")
		return fmt.Errorf("node not registered: %d", id)
	}

	// check if node is deregistering itself
	if registry.nodes[id] != add {
		Sender(add, "-2", "deregistrationResponse")
		return fmt.Errorf("node not deregistering itself: %d", id)
	}

	// deregister node
	delete(registry.nodes, id)

	// send response to the messaging node
	Sender(add, strconv.Itoa(id), "deregistrationResponse")

	return nil
}

func distance(n1, n2, maxID int) int {
	d := n2 - n1
	if d < 0 {
		d += maxID
	}
	return d
}


func GetRoutingTable(nodePlacement, size int, nodes []int) []int {
	// Compute the size of the ID space
	numIds := len(nodes)

	// Initialize the routing table
	routingTable := []int{}

	// Compute the routing table for the given node
	for i := 0; i < size; i++ {
		hops := math.Pow(2, float64(i))
		target := (nodePlacement + int(hops)) % numIds

		if target == nodePlacement {
			target = (target + 1) % numIds
		}

		routingTable = append(routingTable, nodes[target])
	}

	return routingTable
	
}

//Send TaskInitiate message in start function
func (r *Registry) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()


	return nil
}

func readMessage(conn net.Conn) {
	var id int32;
	for {
		buffer := make([]byte, 65535)
		//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		length, _ := conn.Read(buffer)
		message := &minichord.MiniChord{}
		err := proto.Unmarshal(buffer[:length], message)
		if err != nil {fmt.Println("Error in unmarshalling")}
		if message.Message == nil{return}
		switch message.Message.(type) {
		case *minichord.MiniChord_Registration:
			RegisterNode(conn, *message)
		case *minichord.MiniChord_RegistrationResponse:
			fmt.Println("Registration response received: ", message.Message)
		case *minichord.MiniChord_Deregistration:
			id = message.GetDeregistration().Node.Id
			DeregisterNode(message.GetDeregistration().Node.Address, int(id))
		
		case *minichord.MiniChord_DeregistrationResponse:
			fmt.Println("Deregistration response received: ", message.Message)
		
		case *minichord.MiniChord_NodeRegistry:
			//TODO: Implement
			fmt.Println("Node registry received: ", message.Message)
		
		case *minichord.MiniChord_NodeRegistryResponse:
			if message.GetNodeRegistryResponse().Result == 1 {
				fmt.Println("Node registry response received: ", message.Message)
				readyCounter++
			} else {
				fmt.Println("Node registry response received: ", message.Message)
				readyCounter = -100
			}
		case *minichord.MiniChord_InitiateTask:
			//TODO: Implement
			fmt.Println("Task initiate received: ", message.Message)
		
		case *minichord.MiniChord_TaskFinished:
			//TODO: Implement
			fmt.Println("Task finished received: ", message.Message)
			finishedCounter++

		
		case *minichord.MiniChord_NodeData:
			//TODO: Implement
			fmt.Println("Node data received: ", message.Message)
		
		case *minichord.MiniChord_RequestTrafficSummary:
			//TODO: Implement
			fmt.Println("Request traffic summary received: ", message.Message)
		
		case *minichord.MiniChord_ReportTrafficSummary:
			//TODO: Implement
			counter++
			fmt.Println("Traffic summary received: ", message.Message)
			NodeID := message.GetReportTrafficSummary().Id
			Sent := message.GetReportTrafficSummary().Sent
			Received := message.GetReportTrafficSummary().Received
			Relayed := message.GetReportTrafficSummary().Relayed
			TotalSent := message.GetReportTrafficSummary().TotalSent
			TotalReceived := message.GetReportTrafficSummary().TotalReceived

			summaryMap[NodeID] = Summary{Sent, Received, Relayed, TotalSent, TotalReceived}
			if counter == len(registry.nodes) {
				printSummary()
			}

		}
	}
}

func Sender(receiver string, message string, typ string) {
	conn, err := net.Dial("tcp", receiver)
	if err != nil {fmt.Println("Error in dialing"); return}
	
	conMess := constructMessage(message, typ)

	data, err := proto.Marshal(conMess)
	if err != nil {fmt.Println("Error in marshalling"); return}
	_, err = conn.Write(data)
	if err != nil {fmt.Println("Error in writing")}
}

func printSummary() {
	var sentPacketsSum int32
	var receivedPacketsSum int32
	var relayedPacketsSum int32
	var totalSentPacketsSum int64
	var totalReceivedPacketsSum int64
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Println(" 			        Packets 	           Sum Values			")
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Println("       		Sent   Received   	Relayed   	Sent   	Received ")
	fmt.Println("----------------------------------------------------------------------------------")
	for id, item := range summaryMap {
		fmt.Println("Node", id, "    	", item.Sent, "    	", item.Received, "   		", item.Relayed, "   	", item.SentSummary, "   	", item.ReceivedSummary)
		sentPacketsSum += int32(item.Sent)
		receivedPacketsSum += int32(item.Received)
		relayedPacketsSum += int32(item.Relayed)
		totalSentPacketsSum += item.SentSummary
		totalReceivedPacketsSum += item.SentSummary
	}
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Println("Sum:		", sentPacketsSum, "	", receivedPacketsSum, "		", relayedPacketsSum, "	", totalSentPacketsSum, "		", totalReceivedPacketsSum)
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
		info := "Registered successfully"
		if id == -1 {
			info = "Registration failed, registry is full"
		}else if id == -2{
			info = "Registration failed, node already registered"
		} else if id == -3 {
			info = "Registration failed, node address does not match sender"
		}
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_RegistrationResponse{
				RegistrationResponse: &minichord.RegistrationResponse{
					Result: int32(id),
					Info: info,
				},
			},
		}
	case "deregistration":
		fmt.Println("Should not be happening")
		return nil
	case "deregistrationResponse":
		id, _ := strconv.Atoi(message)
		info := "Deregistered successfully"
		if id == -1 {
			info = "Deregistration failed, node not registered"
		} else if id == -2 {
			info = "Deregistration failed, node must deregister itself"
		}

		return &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
					Result: int32(id),
					Info: info,
				},
			},
		}
	case "nodeRegistry":
		who, _ := strconv.Atoi(message)
		table := routingTables[who]


		num := len(table)
		peers := []*minichord.Node{}
		for i := 0; i < num; i++ {
			peer := &minichord.Node{
				Id: int32(table[i]),
				Address: registry.nodes[table[i]],
			}
			peers = append(peers, peer)
		}
		ids := make([]int32, len(registry.nodes))
		i := 0
  		for k := range registry.nodes {
    		ids[i] = int32(k)
    		i++
  		}
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_NodeRegistry{
				NodeRegistry: &minichord.NodeRegistry{
					NR: uint32(num),
					Peers: peers,
					NoIds: uint32(len(registry.nodes)),
					Ids: ids,
				},
			},
		}
	case "nodeRegistryResponse":
		// TODO: Implement
		return nil
	case "initiateTask":
		// TODO: Implement
		num, _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_InitiateTask{
				InitiateTask: &minichord.InitiateTask{
					Packets: uint32(num),
				},
			},
		}
	case "nodeData":
		// TODO: Implement
		return nil
	case "taskFinished":
		// TODO: Implement
		return nil
	case "requestTrafficSummary":
		// TODO: Implement
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_RequestTrafficSummary{
				RequestTrafficSummary: &minichord.RequestTrafficSummary{},
			},
		}
	case "reportTrafficSummary":
		// TODO: Implement
		return nil
	default:
		return nil
	}
}



func setupOverlay(num string) {
	fmt.Println("Setup overlay with ", num, " nodes")
	var numInt, err = strconv.Atoi(num)
	if err != nil || numInt < 1 || numInt >= len(registry.nodes) {
		fmt.Println("Invalid number of nodes")
		return
	}
	ids := []int{}
	for id := range registry.nodes {
		ids = append(ids, id)
	}
	// sort the ids
	sort.Ints(ids)
	for id := range ids {
		routingTables[ids[id]] = GetRoutingTable(id, numInt, ids)
		go Sender(registry.nodes[ids[id]], strconv.Itoa(ids[id]), "nodeRegistry")
	}
	for readyCounter < len(registry.nodes) {
		if readyCounter < 0 {
			fmt.Println("Setup failed")
			return
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Registry is now ready to initiate tasks")
	readyCounter = 0
}

func listnodes() {
	fmt.Println("List of nodes,", len(registry.nodes), "nodes in total")
	for id, addr := range registry.nodes {
		fmt.Println(id, addr)
	}
}

func startTask(num string) {
	for id, addr := range registry.nodes {
		fmt.Println("Sending task to ", id, addr)
		go Sender(addr, num, "initiateTask")
	}
}

func GetTrafficSummary() {
	for id, addr := range registry.nodes {
		fmt.Println("Requesting traffic summary from ", id, addr)
		go Sender(addr, "", "requestTrafficSummary")
	}
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
		case "list":
			go listnodes()
		case "setup":
			if len(cmdSlice) != 2 {
				cmdSlice = append(cmdSlice, "3")
			}
			go setupOverlay(cmdSlice[1])
		case "route":
			for id, table := range routingTables {
				fmt.Println("Routing table for node ", id, ":", table)	
			}
		case "start":
			if len(cmdSlice) != 2 {
				cmdSlice = append(cmdSlice, "100")
			}
			go startTask(cmdSlice[1])
			for finishedCounter < len(registry.nodes) {
				time.Sleep(1 * time.Second)
				continue
			}
			fmt.Println("All nodes are finished")
			finishedCounter = 0
			go GetTrafficSummary()
		default:
			fmt.Printf("command not understood: %s\n", cmd)
		}
	}
}


func main() {
	go readUserInput()

	registry = NewRegistry()

	// start listening for incoming connections
	port := os.Args[1]
	ln, err := net.Listen("tcp", "localhost:" + port)
	//print
	if err != nil {
		// handle listen error
		fmt.Println("Error listening", err)
		os.Exit(1)
	}
	fmt.Println("Registry is running on port: ", port)
	defer ln.Close()

	// accept incoming connections and register messaging nodes
	for {
		fmt.Println("Waiting for new connection")
		conn, err := ln.Accept()
		fmt.Println(conn.RemoteAddr().String())
		if err != nil {
			// handle accept error
			fmt.Println("Error in accepting")
			continue
		}

		// handle messages in a separate goroutine to avoid blocking
		fmt.Println("New connection")
		go readMessage(conn)
	}
}
