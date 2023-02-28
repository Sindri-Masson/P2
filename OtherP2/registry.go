package registry

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

	_, err = conn.Write(data)
	if err != nil {fmt.Println("Error Write", err)}
}

/*
	This function constructs the message into a minichord message format
*/
func constructMessage(message string, typ string) *minichord.MiniChord{ 
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
					Info: "Registration successful",
				},
			},
		}

	case "deregistration":
		fmt.Println("Messge: ",message)
		newMessage := strings.Split(message, ",")
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
		result , _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
					Result: int32(result),
					Info: "Successfully deregistered",
				},
			},
		}
	case "nodeRegistry":
		fmt.Println("Routing table initialized")
		newMessage := strings.Split(message, ",")
		newId, _ := strconv.Atoi(newMessage[0])
		
		routingTable := PopulatingRoutingTable(NodesInOverlay)
		
		thisNode := routingTable[newId]
	
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_NodeRegistry{
				NodeRegistry: &minichord.NodeRegistry{
					NR: uint32(NrOfConnectionPerNode),
					Peers: CreatePeers(thisNode),
					NoIds: uint32(len(NodesInOverlay)),
					Ids: NodeListInt32,
				},
			},
		}
	case "initiateTask":
		numberOfPackets, _ := strconv.Atoi(message)
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_InitiateTask{
				InitiateTask: &minichord.InitiateTask{
					Packets: uint32(numberOfPackets),
				},
			},
		}

	case "requestTrafficSummary":
		return &minichord.MiniChord{
			Message: &minichord.MiniChord_RequestTrafficSummary {
				RequestTrafficSummary: &minichord.RequestTrafficSummary{},
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




package registry

import (
	"PA2/minichord"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

var NrConn int = 0
var FinishedCounter = 0

// Registry instance variables
var NodeID int32;
var Sent uint32 = 0;
var Received uint32 = 0;
var Relayed uint32 = 0;
var TotalSent int64 = 0;
var TotalReceived int64 = 0;

/*
	Creates a listen socket for the given address
*/
func Listener(address string) {

	var l = createListener(address)
	listening(l)
}


/*
	Helping function for the Listener function
*/
func createListener(address string) net.Listener {
	l, err := net.Listen("tcp", address)
	if err != nil {fmt.Println("Error Listen: ", err);return nil}
	return l
}

/*
	Listens for new connections and creates new goroutines each time a new connection is established
*/
func listening(l net.Listener){
	for {
		conn, err := l.Accept()
		NrConn = NrConn + 1
		if err != nil {fmt.Println("Error Conn: ", err)}
		go readMessage(conn)
	}
}

/*
	Reads a message each minichord message type that 
	is sent to the registry listen socket
*/
func readMessage(conn net.Conn){
	var id int32;
	for {
		buffer := make([]byte, 65535)
		length, _:= conn.Read(buffer)
		message := &minichord.MiniChord{}
		err := proto.Unmarshal(buffer[:length], message)
		if err != nil {fmt.Println("Error Unmarshal: ", err)}
		if message.Message == nil {return}
		
		switch message.Message.(type){
		case *minichord.MiniChord_Registration:
			// fmt.Println("Reading registration message: ", message.Message)
			id := AssignIDs()
			Sender(message.GetRegistration().Address, strconv.Itoa(int(id)), "registrationResponse")
			NrOfNodesInNetwork = NrOfNodesInNetwork + 1
			IdAddr[id] = message.GetRegistration().Address
			
		case *minichord.MiniChord_RegistrationResponse:
			fmt.Println("Reading registration response message: ", message.Message)

		case *minichord.MiniChord_Deregistration:
			id = message.GetDeregistration().Id
			fmt.Println("Reading Deregistration message: ", message.Message)
			fmt.Println("Reading DEREGISTERED node ID: ", id)
			Sender(message.GetDeregistration().Address, strconv.Itoa(int(id)), "deregistrationResponse")
			NrOfNodesInNetwork = NrOfNodesInNetwork - 1
			delete(IdAddr, id)
			NodesInOverlay = RemoveID(int(id))
		
		case *minichord.MiniChord_NodeRegistryResponse:
			fmt.Println("Node registry response message: ", message.Message)

		case *minichord.MiniChord_TaskFinished:
			fmt.Println("Task finished by: ", message.GetTaskFinished().Id)
			finishedAddress := message.GetTaskFinished().Address
			time.Sleep(time.Second*2)
			Sender(finishedAddress, "", "requestTrafficSummary")
			
			
		case *minichord.MiniChord_ReportTrafficSummary:
			FinishedCounter += 1
			NodeID = message.GetReportTrafficSummary().Id
			Sent = message.GetReportTrafficSummary().Sent
			Relayed = message.GetReportTrafficSummary().Relayed
			Received = message.GetReportTrafficSummary().Received
			TotalSent = message.GetReportTrafficSummary().TotalSent
			TotalReceived = message.GetReportTrafficSummary().TotalReceived
			
			tempList := []string {strconv.Itoa(int(Sent)), strconv.Itoa(int(Received)),strconv.Itoa(int(Relayed)), strconv.Itoa(int(TotalSent)),strconv.Itoa(int(TotalReceived))}
			SummaryMap[NodeID] = tempList
			if FinishedCounter == len(IdAddr) {
				PrintSummary()
			}
		}

	}
}




package registry

import (
	"PA2/minichord"
	"fmt"
	"math"
	"strconv"
)

// const nodesInOverlay = []

/*
Returns true if the element elem exists in the slice idsSlice
otherwise returns false
*/
func contains(idsSlice []int, elem int) bool {
	for _, item := range idsSlice {
		if item == elem {
			return true
		}
	}
	return false
}

// Registry instance variables
var NrOfNodesInNetwork int = 0
var NrOfConnectionPerNode int = 3 // Default value is 3
var NodesInOverlay = make([]int, 0, 127);
var NodeListInt32 = make([]int32, 0, 127)
var IdAddr = make(map[int32]string)
var SummaryMap = make(map[int32][]string)

/*
	Removes a node ID from the nodesInOverlay list
*/
func RemoveID(id int) []int{
    for index, item := range NodesInOverlay{
        if item == id {
            return append(NodesInOverlay[:index], NodesInOverlay[index+1:]...)
        }
	}
	return NodesInOverlay
}

/*
	Iterates through numbers from 0 to 127 and appends the id to a list
	called nodesInOverlay which contains all the node IDs that are in the overlay
	the nodesInOverlay list will then contain node Ids sorted in a ascending order
*/
func AssignIDs() int32 {
	for i := 1; i <= 127; i++ {
		if !contains(NodesInOverlay, i) {
			NodesInOverlay = append(NodesInOverlay, i)
			NodeListInt32 = append(NodeListInt32, int32(i))
			return int32(i)
		}
	}
	return -1
}

/*
	Logic for calculating the peers in the overlay needed for the routing table
	further explained in the report
*/
func PopulatingRoutingTable(IDs []int) map[int][]int {
    routingTableMap := make(map[int][]int)
    for index, id := range IDs {
        connectionNodeSlice := make([]int, NrOfConnectionPerNode)
        for i := 1; i <= NrOfConnectionPerNode; i++ {
            hops := int(math.Pow(float64(2), float64((i - 1))))
			node := IDs[(index+hops)%(len(IDs))]
            if !contains(connectionNodeSlice, node) && node != id {	
				connectionNodeSlice[i-1] = node
			} else {
				for {
					hops = hops + 1
					node := IDs[(index+hops)%(len(IDs))]
					if !contains(connectionNodeSlice, node) && node != id {
						connectionNodeSlice[i-1] = node
						break
					}
				}
			}
            
        }
        routingTableMap[id] = connectionNodeSlice
    }
    return routingTableMap
}


/*
	Iterates through nodes in the overlay and sends nodeRegistry to all of them
*/
func Broadcast() { 
	for id, address := range IdAddr {
		Sender(address, strconv.Itoa(int(id)) +  "," + address,"nodeRegistry")
	}
}

/*
	Prints formatted the node id and the node address
*/
func PrintMessenger() { 
	for id, address := range IdAddr {
		fmt.Println("Node " + strconv.Itoa(int(id)) + ": Address: " + address)
	}
}

/*
	returns the address of a given node ID
*/
func MapIdWithAddress(inputID int) string {
	for id, address := range IdAddr {
		if id == int32(inputID) {
			return address
		}
	}
	return "Address not found" //laga
}

/*
	Creates the peer list 
*/
func CreatePeers(peerIdList []int) []*minichord.Deregistration{
	var Peers []*minichord.Deregistration

	for i, item := range peerIdList {
		Peers = append(Peers, &minichord.Deregistration{
		Id: int32(item),
		Address: MapIdWithAddress(peerIdList[i])})
	}
	return Peers
}

/*
	Sends InitiateTask minichord message to each node
*/
func SendPackets(numberOfPackets int) {
	for _, address := range IdAddr {
		Sender(address, strconv.Itoa(numberOfPackets), "initiateTask")
	}
}

/*
	Prints the summary for each node in the overlay for statistics purposes
*/
func PrintSummary() {
	sentPacketsSum := 0
	receivedPacketsSum := 0
	relayedPacketsSum := 0
	sentSum := 0
	receivedSum := 0

	fmt.Println("			        Packets						Sum Values			")
	fmt.Println("		 Sent		Received	Relayed			  Sent		  Received	")
	fmt.Println("______________________________________________________________________________________________________")
	for i, item := range SummaryMap {
		fmt.Println(" Node " + strconv.Itoa(int(i)) + ":	" + item[0] + "		" + item[1] + "		" + item[2] + "		  " + item[3] + "		" + item[4]) 
		sentPacketsSum2, _ := strconv.Atoi(item[0])
		sentPacketsSum += sentPacketsSum2
		receivedPacketsSum2, _ := strconv.Atoi(item[1])
		receivedPacketsSum += receivedPacketsSum2
		relayedPacketsSum2, _ := strconv.Atoi(item[2])
		relayedPacketsSum += relayedPacketsSum2
		sentSum2, _ := strconv.Atoi(item[3])
		sentSum += sentSum2
		receivedSum2, _ := strconv.Atoi(item[4])
		receivedSum += receivedSum2

	}
	fmt.Println("_______________________________________________________________________________________________________")
	fmt.Println(" Sum 	 	" + strconv.Itoa(int(sentPacketsSum)) + "		" + strconv.Itoa(int(receivedPacketsSum)) + "		" + strconv.Itoa(int(relayedPacketsSum)) + "		 " + strconv.Itoa(int(sentSum)) + "  	" + strconv.Itoa(int(receivedSum))) 
	fmt.Println("_______________________________________________________________________________________________________")
}
