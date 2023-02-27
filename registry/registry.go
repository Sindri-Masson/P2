package registry

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
)

// MessagingNode represents a registered node in the system
type RMessagingNode struct {
	ID         int // randomly assigned identifier
	Connection net.Conn // connection to the messaging node
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

// hasNodeWithID returns true if a messaging node with the specified ID already exists
func (r *Registry) hasNodeWithID(id int) bool {
	for _, node := range r.nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}

func main() {
	registry := NewRegistry()

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
