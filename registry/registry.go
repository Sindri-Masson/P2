package main

import (
	"P2/minichord"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"

	"google.golang.org/protobuf/proto"
)

var NrOfNodes int32 = 0
var registry *Registry;
var routingTable map[int32]string;

// MessagingNode represents a registered node in the system
type RMessagingNode struct {
	ID         int // randomly assigned identifier
	Connection net.Conn // connection to the messaging node
}
type RoutingTable struct {
    nodeID       int
    size         int
    routingTable [][]int
}

// Registry represents the system registry
type Registry struct {
	nodes map[string]*RMessagingNode // map of registered messaging nodes
	mu    sync.Mutex // mutex to protect concurrent access to the nodes map
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]*RMessagingNode),
	}
}

// RegisterNode registers a messaging node in the system and assigns a random ID
func (r *Registry) RegisterNode(conn net.Conn) error {
	addr := conn.RemoteAddr()
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[addr.String()]; ok {
		return fmt.Errorf("node already registered: %s", addr)
	}

	id := rand.Intn(128)
	for r.hasNodeWithID(id) { // ensure no ID collisions
		id = rand.Intn(128)
	}

	node := &RMessagingNode{
		ID:         id,
		Connection: conn,
	}
	r.nodes[addr.String()] = node

	// send response to the messaging node
	response := fmt.Sprintf("Registered with ID %d", id)
	_, err := conn.Write([]byte(response))
	if err != nil {
		// handle write error
	}

	return nil
}

// DeregisterNode deregisters a messaging node from the system
func (r *Registry) DeregisterNode(conn net.Conn) error {
	addr := conn.RemoteAddr()
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[addr.String()]; !ok {
		return fmt.Errorf("node not registered: %s", addr)
	}

	delete(r.nodes, addr.String())

	// send response to the messaging node
	response := "Deregistered successfully"
	_, err := conn.Write([]byte(response))
	if err != nil {
		// handle write error
	}

	return nil
}

// GetNodeByID returns the messaging node with the specified ID
func (r *Registry) GetNodeByID(id int) (*RMessagingNode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, node := range r.nodes {
		if node.ID == id {
			return node, nil
		}
	}

	return nil, fmt.Errorf("node not found with ID %d", id)
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

// hasNodeWithID returns true if a messaging node with the specified ID already exists
func (r *Registry) hasNodeWithID(id int) bool {
	for _, node := range r.nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}


// GetRoutingTable returns the routing table for the system
// func (r *Registry) GetRoutingTable() (*RoutingTable, error) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	return RoutingTable, nil
// }

// SetupOverlay sets up the overlay network with the specified number of entries
func (r *Registry) SetupOverlay(n int) error {
	r.mu.Lock()
	defer r.mu.Unlock()



	return nil
}

//Send TaskInitiate message in start function
func (r *Registry) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()


	return nil
}

func AssignIDs() int32 {
	return NrOfNodes+1
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
			id := AssignIDs()
			Sender(message.GetRegistration().Address, strconv.Itoa(int(id)), "registrationResponse")
			NrOfNodes++
			routingTable[id] = message.GetRegistration().Address
		
		case *minichord.MiniChord_RegistrationResponse:
			fmt.Println("Registration response received: ", message.Message)

		case *minichord.MiniChord_Deregistration:
			id = message.GetDeregistration().Node.Id
			fmt.Println("Deregistration message received: ", message.Message)
			fmt.Println("Deregistration Node ID received: ", message.GetDeregistration().Node.Id)
			Sender(message.GetDeregistration().Node.Address, strconv.Itoa(int(id)), "deregistrationResponse")
			NrOfNodes--
			delete(routingTable, id)
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


func main() {
	registry = NewRegistry()

	// start listening for incoming connections
	port := os.Args[1]
	ln, err := net.Listen("tcp", ":" + port)
	//print
	fmt.Println("Registry is running on port 8080")
	if err != nil {
		// handle listen error
		fmt.Println("Error in listening")
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
				// case "list":
				// 	routingTable, err := registry.GetRoutingTable()
				// 	if err != nil {
				// 		fmt.Printf("Failed to get routing table: %v", err)
				// 		fmt.Printf(routingTable)
				// 		continue
				// 	}
				// 	// nodes := routingTable.AllNodes()
				// 	// for _, node := range nodes {
				// 	// 	fmt.Printf("Node ID: %d, Hostname: %s, Port: %d", node.ID, node.Hostname, node.Port)
				// 	// }
				// case "setup":
				// 	nStr := string(buf[n:])
				// 	n, err := strconv.Atoi(nStr)
				// 	if err != nil {
				// 		fmt.Printf("Invalid setup command: %s", msgType+nStr)
				// 		continue
				// 	}
				// 	err = registry.SetupOverlay(n)
				// 	if err != nil {
				// 		fmt.Printf("Failed to setup overlay: %v", err)
				// 		continue
				// 	}
				// 	fmt.Printf("Overlay setup with %d entries", n)
				// case "route":
				// 	routingTable, err := registry.GetRoutingTable()
				// 	if err != nil {
				// 		fmt.Printf("Failed to get routing table: %v", err)
				// 		continue
				// 	}

				// case "start":
				// 	nStr := string(buf[n:])
				// 	n, err := strconv.Atoi(nStr)
				// 	if err != nil {
				// 		fmt.Printf("Invalid start command: %s", msgType+nStr)
				// 		continue
				// 	}
				// 	err = registry.Start()
				// 	if err != nil {
				// 		fmt.Printf("Failed to start sending messages: %v", err)
				// 		continue
				// 	}
				default:
					fmt.Printf("Unknown message type: %s", msgType)
				}
			}
		}(conn)
	}
}
