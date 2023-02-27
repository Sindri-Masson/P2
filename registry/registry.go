package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"os"
)

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
	nodes map[net.Addr]*RMessagingNode // map of registered messaging nodes
	mu    sync.Mutex // mutex to protect concurrent access to the nodes map
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[net.Addr]*RMessagingNode),
	}
}

// RegisterNode registers a messaging node in the system and assigns a random ID
func (r *Registry) RegisterNode(conn net.Conn) error {
	addr := conn.RemoteAddr()
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.nodes[addr]; ok {
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
	r.nodes[addr] = node

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

	if _, ok := r.nodes[addr]; !ok {
		return fmt.Errorf("node not registered: %s", addr)
	}

	delete(r.nodes, addr)

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


//
// main
//


func main() {
	registry := NewRegistry()

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
