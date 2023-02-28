package main

import (
	"P2/minichord"
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

const TYPE = "tcp"

var NrOfNodes int32 = 0
var registry Registry;


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
func RegisterNode(conn net.Conn) error {
	addr := conn.RemoteAddr()

	for !registry.mu.TryLock() {
		continue
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// check if node is already registered
	for _, x := range registry.nodes {
		if x == addr.String() {
			return fmt.Errorf("node already registered: %s", addr)
		}
	}
	

	id := rand.Intn(128)
	_, ok := registry.nodes[id]
	for ok { // ensure no ID collisions
		id = rand.Intn(128)
		_, ok = registry.nodes[id]
	}

	registry.nodes[id] = addr.String()

	// send response to the messaging node

	Sender(addr.String(), strconv.Itoa(int(id)), "registrationResponse")
	NrOfNodes++
	registry.nodes[id] = addr.String()

	return nil
}

// DeregisterNode deregisters a messaging node from the system
func DeregisterNode(conn net.Conn, id int) error {
	for !registry.mu.TryLock() {
		continue
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()

	// check if node is registered
	var _, ok = registry.nodes[id]
	if !ok {
		return fmt.Errorf("node not registered: %d", id)
	}

	delete(registry.nodes, id)

	// send response to the messaging node
	Sender(conn.RemoteAddr().String(), strconv.Itoa(int(id)), "deregistrationResponse")

	return nil
}

func distance(n1, n2, maxID int) int {
	d := n2 - n1
	if d < 0 {
		d += maxID
	}
	return d
}

func GetRoutingTable(nodeID int, nodes []int) map[int][]int {
	// Compute the size of the ID space
	maxID := 1
	for maxID < len(nodes) {
		maxID *= 2
	}

	// Compute the size of the routing table
	k := 0
	for (1 << k) <= maxID {
		k++
	}

	// Initialize the routing table
	routingTable := make(map[int][]int)
	for i := 1; i <= k; i++ {
		routingTable[i] = make([]int, 0)
	}

	// Compute the routing table for the given node
	for i := 0; i < k; i++ {
		hop := 1 << i
		target := (nodeID + hop) % maxID
		for _, n := range nodes {
			if distance(n, target, maxID) < distance(nodeID, target, maxID) {
				routingTable[i+1] = append(routingTable[i+1], n)
			}
		}
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
		length, err := conn.Read(buffer)
		message := minichord.MiniChord{}
		err = proto.Unmarshal(buffer[:length], &message)
		if err != nil {fmt.Println("Error in unmarshalling")}
		if message.Message == nil{return}
		switch message.Message.(type) {
		case *minichord.MiniChord_Registration:
			go RegisterNode(conn)
		
		case *minichord.MiniChord_RegistrationResponse:
			fmt.Println("Registration response received: ", message.Message)

		case *minichord.MiniChord_Deregistration:
			id = message.GetDeregistration().Node.Id
			fmt.Println("Deregistration message received: ", message.Message)
			fmt.Println("Deregistration Node ID received: ", id)
			DeregisterNode(conn, int(id))
			Sender(message.GetDeregistration().Node.Address, strconv.Itoa(int(id)), "deregistrationResponse")
			NrOfNodes--
			//NodesInOverlay = RemoveID(int(id))
		
		case *minichord.MiniChord_DeregistrationResponse:
			fmt.Println("Deregistration response received: ", message.Message)
		
		case *minichord.MiniChord_NodeRegistry:
			//TODO: Implement
			fmt.Println("Node registry received: ", message.Message)
		
		case *minichord.MiniChord_NodeRegistryResponse:
			fmt.Println("Node registry response received: ", message.Message)
		
		case *minichord.MiniChord_InitiateTask:
			//TODO: Implement
			fmt.Println("Task initiate received: ", message.Message)
		
		case *minichord.MiniChord_TaskFinished:
			//TODO: Implement
			fmt.Println("Task finished received: ", message.Message)
		
		case *minichord.MiniChord_NodeData:
			//TODO: Implement
			fmt.Println("Node data received: ", message.Message)
		
		case *minichord.MiniChord_RequestTrafficSummary:
			//TODO: Implement
			fmt.Println("Request traffic summary received: ", message.Message)
		
		case *minichord.MiniChord_ReportTrafficSummary:
			//TODO: Implement
			fmt.Println("Traffic summary received: ", message.Message)
		}
	}
}

func Sender(receiver string, message string, typ string) {
	conn, err := net.Dial("tcp", receiver)
	if err != nil {fmt.Println("Error in dialing")}
	
	conMess := constructMessage(message, typ)

	data, err := proto.Marshal(conMess)
	if err != nil {fmt.Println("Error in marshalling")}
	_, err = conn.Write(data)
	if err != nil {fmt.Println("Error in writing")}
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
			fmt.Println("list")
		case "setup":
			fmt.Println("Setup")
		case "route":
			fmt.Println("route")
		case "start":
			fmt.Println("start")
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
	ln, err := net.Listen("tcp", ":" + port)
	//print
	fmt.Println("Registry is running on port 8080")
	if err != nil {
		// handle listen error
		fmt.Println("Error listening")
		os.Exit(1)
	}
	defer ln.Close()

	// accept incoming connections and register messaging nodes
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle accept error
			fmt.Println("Error in accepting")
			continue
		}

		// handle messages in a separate goroutine to avoid blocking
		go readMessage(conn)
	}
}
