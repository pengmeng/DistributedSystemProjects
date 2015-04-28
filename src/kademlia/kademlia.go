package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

type Pair struct {
	key   ID
	value []byte
}

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	LocalData   map[ID][]byte
	AddrBook    *KBuckets

	addDataChan  chan Pair
	findDataChan chan ID
	resChan      chan []byte
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.LocalData = make(map[ID][]byte)

	k.addDataChan = make(chan Pair)
	k.findDataChan = make(chan ID)
	k.resChan = make(chan []byte)

	go k.MessageWorker()

	// Set up RPC server
	// NOTE: KademliaCore is just a wrapper around Kademlia. This type includes
	// the RPC functions.
	rpc.Register(&KademliaCore{k})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ := net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	k.AddrBook = BuildKBuckets(k.SelfContact)
	return k
}

type NotFoundError struct {
	id  ID
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

// ======================= Kademlia helper functions ===================
func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}
	return k.AddrBook.find_contact(nodeId)
}

func (k Kademlia) MessageWorker() {
	// loop forever
	for {
		select {
		case pair := <-k.addDataChan:
			k.LocalData[pair.key] = pair.value

		case key := <-k.findDataChan:
			// check if key is in LocalData
			if value, ok := k.LocalData[key]; ok {
				k.resChan <- value
			} else {
				k.resChan <- nil
			}
		}
	}
}

func (k Kademlia) addData(p Pair) {
	k.addDataChan <- p
}

func (k Kademlia) getData(key ID) ([]byte, error) {
	k.findDataChan <- key
	result := <-k.resChan

	if result != nil {
		return result, nil
	}
	return nil, &NotFoundError{key, "Key does not exist"}
}

func PingHelper(self Contact, host net.IP, port uint16) (*PongMessage, error) {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", host.String(), port))
	if err != nil {
		return nil, err
	}
	ping := PingMessage{self, NewRandomID()}
	var pong PongMessage

	err = client.Call("KademliaCore.Ping", ping, &pong)
	if err != nil {
		return nil, err
	}
	return &pong, nil
}

// ========================== RPC client code =========================
// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	pong, err := PingHelper(k.SelfContact, host, port)
	if err != nil {
		log.Fatal("ERR: ", err)
	}
	k.AddrBook.Update(pong.Sender)
	return "OK: Ping " + pong.Sender.NodeID.AsString()
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port))
	if err != nil {
		log.Fatal("ERR: ", err)
	}

	req := StoreRequest{k.SelfContact, NewRandomID(), key, value}
	var res StoreResult

	err = client.Call("KademliaCore.Store", req, &res)
	if err != nil {
		log.Fatal("ERR: ", err)
	}

	return "OK:"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port))

	req := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindNodeResult

	err = client.Call("KademliaCore.FindNode", req, &res)

	if err != nil {
		log.Fatal("ERR: ", err)
	}

	return "OK: Found Nodes"
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port))

	req := FindValueRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindValueResult

	err = client.Call("KademliaCore.FindValue", req, &res)
	if err != nil {
		log.Fatal("ERR: ", err)
	}
	if res.Value != nil {
		return "OK: Found value: " + string(res.Value)
	} else if res.Nodes != nil {
		return "OK: Found nodes: " + string(len(res.Nodes))
	} else {
		return "Err: Not Found"
	}
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if value, err := k.getData(searchKey); err == nil {
		return "OK: Found value: " + string(value)
	} else {
		return "ERR: Not Found"
	}
}

func (k *Kademlia) DoIterativeFindNode(id ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeFindValue(key ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
