

package messenger

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// Node Instance variables
var RegistryAddress string = "localhost:4000"
var nodesInOverlay = make([]int, 0, 127);
var NodeID int = -1;
var NodeAddress string

var Sent int = 0;
var Received int = 0;
var Relayed int = 0;
var TotalSent int = 0;
var TotalReceived int = 0;

/*
	This will print on to the console the node statistics:
	How many messages the node has sent, received and relayed
	How many packets the node has sent in total and how many packets the node has received 
*/
func PrintCommand(){
	nodePrint(strconv.Itoa(Sent),strconv.Itoa(Received),strconv.Itoa(Relayed), strconv.Itoa(TotalSent),strconv.Itoa(TotalReceived))
}

/*
	This is a helper function for PrintCommand
	It formats the node statistics into a string
*/
func nodePrint(sent string, received string, relayed string, sentSum string, receivedSum string){
	fmt.Println("Sent: "+ sent +" Received: "+received +" Relayed: "+relayed + " SentSum: " + sentSum + " ReceivedSum: " + receivedSum)
}

/*
	This function checks if the id exists in a slice that
	contains ids. Returns true if the id exists, otherwise returns false
*/
func contains(idsSlice []int, elem int) bool {
	for _, item := range idsSlice {
		if item == elem {
			return true
		}
	}
	return false
}

/*
	Iterates through numbers from 0 to 127 and appends the id to a list
	called nodesInOverlay which contains all the node IDs that are in the overlay
	the nodesInOverlay list will then contain node Ids sorted in a ascending order
*/
func AssignIDs() int32 {
	for i := 0; i <= 127; i++ {
		if !contains(nodesInOverlay, i) {
			nodesInOverlay[i] = i
			return int32(i)
		}
	}
	return -1
}

/*
	Removes the element from the slice that is passed into the function
	the index that is passed into the function is the element's index
*/
func removeNode(slice []int, index int) []int {
    return append(slice[:index], slice[index+1:]...)
}

/*
	It tries to establish a connection between nodes with TCP 
	If the connection is not established the function returns 
	false, otherwise it returns true
*/
func TestConnectionToNodes() bool{
	for _, object := range Peers {
		conn, err := net.Dial("tcp", object.Address)
		if err != nil {
			conn.Close()
			return false
		}	
		conn.Close()
	}
	return true
}

/*
	Takes a lower and an upper bound and choses a random integer between that range
	inspired from https://wcsiu.github.io/2018/06/21/go-generate-random-negative-and-positive-numbers-in-range.html
*/
func RandInt(lower, upper int) int {
	rand.Seed(time.Now().UnixNano())
	rng := upper - lower
	return rand.Intn(rng) + lower
}

/*
	Takes a slice that contains elements that are strings and
	converts all those elements to int32
*/
func StoI(sList []string) []int32{
    var t2 = []int32{}

    for _, i := range sList {
        j, err := strconv.Atoi(i)
        if err != nil {
            panic(err)
        }
        t2 = append(t2, int32(j))
    }
    return t2
}

/*
	Takes an int32 slice and joins the list with comma seperated string
*/
func AddTrace(list []int32) string{
	var s string
	for _, i := range list {
		s += "," + strconv.Itoa(int(i))
	}
	return s
}






package messenger

import (
	"PA2/minichord"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

// Node instance variables
var Connections []net.Conn
var PacketsToSend int
var Peers []*minichord.Deregistration
var AllNodeIds []int32


/*
	Creates a listener socket and accepts a connection with TCP
*/
func Listener(address string, exitwg *sync.WaitGroup) {
	
	var l = createListener(address)
	for {
		conn, err := l.Accept()
		if err != nil {fmt.Println("Error Conn: ", err)}
		go readMessage(conn, exitwg)
	}
}


/*
	Creates a listener socket on the given address that is
	passed into the function
*/
func createListener(address string) net.Listener {
	l, err := net.Listen("tcp", address)
	if err != nil {fmt.Println("Error Listen: ", err);return nil}
	return l
}

/*
	Reads the minichord message from the tcp connection variable conn net.Conn
	and prints on to the console descriptive messages corresponding to which type
	of minichord message was received
*/
func readMessage(conn net.Conn, exitwg *sync.WaitGroup){
	var id int32;
	buffer := make([]byte, 65535)
	length, _:= conn.Read(buffer)
	message := &minichord.MiniChord{}
	err := proto.Unmarshal(buffer[:length], message)
	if err != nil {fmt.Println("Error Unmarshal: ", err)}
	if message.Message == nil {return}
	
	switch message.Message.(type){
	case *minichord.MiniChord_Registration:
		fmt.Println("Reading registration message: ", message.Message)
		Sender(message.GetRegistration().Address, conn.LocalAddr().String(), "registrationResponse")
		
	case *minichord.MiniChord_RegistrationResponse:
		id = message.GetRegistrationResponse().Result
		fmt.Println("Reading registration response message: ", message.Message)
		NodeID = int(id)

	case *minichord.MiniChord_Deregistration:
		id = message.GetDeregistration().Id
		fmt.Println("Reading DEregistration message: ", message.Message)
		fmt.Println("Reading DEREGISTERED node ID: ", id)
		Sender(message.GetDeregistration().Address, conn.LocalAddr().String(), "deregistrationResponse")
	case *minichord.MiniChord_DeregistrationResponse:
		id = message.GetDeregistrationResponse().Result
		fmt.Println("Reading DEregistration response message: ", message.Message)
		fmt.Println("Reading DEREG node ID: ", id)	
		exitwg.Done()

	case *minichord.MiniChord_NodeRegistry:
		Peers = message.GetNodeRegistry().Peers
		AllNodeIds = message.GetNodeRegistry().Ids
		if TestConnectionToNodes(){
			Sender(RegistryAddress, "success", "nodeRegistryResponse")
		} else {
			Sender(RegistryAddress, "fail", "nodeRegistryResponse")
		}	
		
	case *minichord.MiniChord_InitiateTask:
		numberOfPackets := message.GetInitiateTask().Packets
		PacketsToSend := numberOfPackets
		payload := RandInt(math.MinInt32, math.MaxInt32)
		payload2 := strconv.Itoa(int(payload))
		rand.Seed(time.Now().UnixNano())


		fmt.Println("Sending ", PacketsToSend, " packets")
		for i := 0; i < int(PacketsToSend); i++{
			randomIndex := rand.Intn(len(AllNodeIds))
			destinationNode := AllNodeIds[randomIndex]
			// message is formatted like: dest_id,src_id,payload,hopcounter,trace
			message := strconv.Itoa(int(destinationNode)) + "," + strconv.Itoa(NodeID) + "," + payload2 + "," + "1" + "," + strconv.Itoa(NodeID)
			TotalSent += int(payload)

			Sender(Peers[0].Address, message, "nodeData")
			Sent += 1
		}
		message2 := strconv.Itoa(NodeID) + "," + NodeAddress
		Sender(RegistryAddress, message2 ,"taskFinished")
		

	case *minichord.MiniChord_NodeData:
		dest := message.GetNodeData().Destination
		payload := message.GetNodeData().Payload
		payload2 := strconv.Itoa(int(payload))
		NID := message.GetNodeData().Source
		hops := message.GetNodeData().Hops
		trace := message.GetNodeData().Trace
		// messageString is formatted like: dest_id,src_id,payload,hopcounter,trace
		messageString := strconv.Itoa(int(dest)) + "," + strconv.Itoa(int(NID)) + "," + payload2 + "," + strconv.Itoa(int(hops)) + AddTrace(trace)

		if dest == int32(NodeID){
			fmt.Println("Node has reached its destination")
			Received += 1
			TotalReceived += int(payload)
		} else {
			Relayed += 1
			var foundDest bool = false
			for index,node := range Peers {
				if node.Id == dest {
					Sender(node.Address, messageString + strconv.Itoa(int(hops) + (index + 1)), "nodeData")
					foundDest = true
				}
			} 
			if !foundDest {
				hops += 1
				Sender(Peers[0].Address, messageString + strconv.Itoa(int(hops)), "nodeData")
			}
		}
	
	case *minichord.MiniChord_RequestTrafficSummary:
		Sender(RegistryAddress, "", "reportTrafficSummary")

	}

}





package messenger

import (
	"PA2/minichord"
	"fmt"
	"net"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

/*
	Sends the senderMessage to the receiver and it construct the message
	with the typ variable that is a string representation of the minichord message
	types. Upon sending the message to the receiver, the message must be marshaled.
*/
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

/*
	This function constructs the message into a minichord message format
*/
func constructMessage(message string, typ string) *minichord.MiniChord{ 
	var id int32;
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
		id = AssignIDs()
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_RegistrationResponse{
				RegistrationResponse: &minichord.RegistrationResponse{
				Result: id,
				Info: "Node registered successfully",
				},
			},
		}

	case "deregistration":
		fmt.Println("Messge: ",message)
		newMessage := strings.Split(message, ",")
		fmt.Println("New message: ", newMessage)
		deregId, _ := strconv.Atoi(newMessage[0])
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_Deregistration{
			  Deregistration: &minichord.Deregistration{
				Id: int32(deregId),
				Address: newMessage[1],
			  },
			},
		  }

	case "deregistrationResponse":
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
				Result: id,
				Info: "",
				},
			},
		}
	
	case "taskFinished":
		newMessage := strings.Split(message, ",")
		id := newMessage[0]
		nid, _ := strconv.Atoi(id)
		address := newMessage[1]
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_TaskFinished{
				TaskFinished: &minichord.TaskFinished{
				Id: int32(nid),
				Address: address,
				},
			},
		}
	case "nodeRegistryResponse":
		if message == "success" {
			return &minichord.MiniChord{
				Message: &minichord.MiniChord_NodeRegistryResponse{
					NodeRegistryResponse: &minichord.NodeRegistryResponse{
					Result: 1,
					Info: "Connections to other nodes successful",
					},
				},
			}
		} else {
			return &minichord.MiniChord{
				Message: &minichord.MiniChord_NodeRegistryResponse{
					NodeRegistryResponse: &minichord.NodeRegistryResponse{
					Result: 0,
					Info: "Connections to other nodes failed",
					},
				},
			}
		}
	case "nodeData":
		trace := make([]int32, 0, 127)

		data := strings.Split(message, ",")
		dest, _ := strconv.Atoi(data[0])
		dest2 := int32(dest)
		source, _:= strconv.Atoi(data[1])
		source2 := int32(source)
		payload,_ := strconv.Atoi(data[2])
		hopCount, _ := strconv.Atoi(data[3])
		hopCounter := uint32(hopCount)
		
		trace = StoI(data[4:])

		return &minichord.MiniChord{
			Message: &minichord.MiniChord_NodeData {
				NodeData: &minichord.NodeData {
					Destination: dest2,
					Source: source2,
					Payload: int32(payload), 
					Hops: hopCounter, 
					Trace: trace,
				},
			},
		}
	
	case "reportTrafficSummary":
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_ReportTrafficSummary {
				ReportTrafficSummary: &minichord.TrafficSummary {
					Id: int32(NodeID),
					Sent: uint32(Sent),
					Relayed: uint32(Relayed),
					Received: uint32(Received),
					TotalSent: int64(TotalSent),
					TotalReceived: int64(TotalReceived),
				},
			},
		}
	
	default:
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_Registration{
				Registration: &minichord.Registration{
				Address: message,
				},
			},
		}
	}	
}