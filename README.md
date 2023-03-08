# P2

## Running the project

registry:
go run registry/registry.go [port]

messenger:
go run messenger/messenger.go [address]:[port]

To run multiple messengers at the same time,
we have included a bash script, which runs 10 messengers at the same time.

bash script: ./build.sh [address]:[port]

## Discussion

### REGISTRY.

The registry lets messengers(nodes) connect to them, and send messages.
It is a server that keeps track of what messengers are connected and sends appropriate data packets to them.

When allocating ID's to nodes, we choose a random number from 1 to 128, if that number is already taken, we choose another random number, until we get one that is available..
If the node is already registered, we respond with a message that indicates that the node is already registered.

With this routing algorithm it is not possible to have partitions, because each node knows the node in front of it and the last node knows the first node, therefore all nodes can be reached from any given node by just traversing to the next node until you reach your destination.

Our solution does not have deadlocks because we dont have a single mutex lock.

### MESSENGER.

The messenger connects to the registry, and in turn is registered into the network of nodes.
The messenger receives data packets from the registry and sends data packets to the registry.

We deliver data packets through our Sender function, which dials to the receiver of supposed message.
In Sender, we then construct the message with the minichord template, and then marshal the message,
if these actions succeed we write the data to the receiver. The data packets do not travel forever,
this is because every time a node receives a data packet that is not intended for it, the node searches for the single best destination to send that packet to, the best destination being either the destination node itself or the closest node in the routing table.

In conclusion, each node gets one packet and only sends one packet, there is no broadcasting. Every time a node gets a packet that is not intented for itself, it sends to the closest node.

The mechanism for task completion and retrieval are working correctly because when a node sends a packet, it increments the sendTracker global variable by one and adds the payload to the sendSummation global variable. It does the same when receiving, however uses the global variables receiveTracker and receiveSummation. When the node receives a packet not intended for itself, it relays the packet and increments the relayTracker global variable. The node then uses those variables for the reportTrafficSummary message to the registry. The registry then calculates the totals whan all summaries have been received.
