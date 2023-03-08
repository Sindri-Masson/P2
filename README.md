# P2
REGISTRY GOALS
 Describe your method of allocating Ids to nodes.
 Describe the method you have implemented to avoid partitions.
 Explain why your solution does not have deadlocks.


MESSENGER GOALS
 Explain why your methods delivers data packets. How do you
 avoid that packets travel forever?
 Explain how you avoid duplications in routing the packets.
 Explain why your mechanism for task completion and retrieval
 are working correctly.
 
 
 
 Running the project
  registry: go run registry.go <port>
  
  messenger: go run messenger.go <address>:<port>
  
  To run multiple messengers at the same time, 
  we have included a bash script, which runs 10 messengers at the same time.
  
  bash script: ./build.sh
  
 
 
 REGISTRY.
  The registry lets messengers(nodes) connect to them, and send messages. 
  It is a server that keeps track of what messengers are connected and sends appropriate data packets to them.
  
  
  
  When allocating ID's to nodes, we choose a random number from 1 to 128, 
  if the node is already registered, we respond with a message that indicates that the node is already registered.
  When allocating an ID to a node, we check if the node is already being used and if so we, 
  run through a for loop until we find an id that is not being used.
  
  It is not possible to have partitions, because each node knows what node is in front of it,
  this leads to system that is free from partitions because it is always possible to jump through all nodes in the system.
  
 
  
 
 MESSENGER.
  The messenger connects to the registry, and in turn is registered into the network of nodes. 
  The messenger receives data packets from the registry and sends data packets to the registry.
  
  We deliver data packets through our Sender function, which dials to the receiver of supposed message.
  In Sender, we then construct the message with the minichord template, and then marshal the message,
  if these actions succeed we write the data to the receiver. The data packets do not travel forever,
  this is because we error check the data, and if the data is not fit for sending, we send back an appropriate response.
  
  Each node gets one packet and only sends one packet, there is no broadcasting. Every time a node gets a packet that is not intented for itself,
  it sends to the closest node.
  
  The mechanism for task completion and retrieval are not working correctly lmao.
  