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

	addDataChan chan Pair
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.LocalData = make(map[ID][]byte)
	k.AddrBook = BuildKBuckets(k.NodeID)
	k.addDataChan = make(chan Pair)

	go k.addDataWorker()

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
	return nil, nil
	//return k.AddrBook.Find(nodeId), nil
}

func (k *Kademlia) addDataWorker() {
	// loop forever, wait for input from addDataChan
	// add pair of key, value to map
	for {
		pair := <-k.addDataChan
		k.LocalData[pair.key] = pair.value
	}
}

func (k *Kademlia) addData(p Pair) {
	k.addDataChan <- p
}

// ========================== RPC client code =========================
// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	// client Ping
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", host.String(), port))

	if err != nil {
		log.Fatal("ERR: ", err)
	}

	ping := PingMessage{k.SelfContact, NewRandomID()}
	var pong PongMessage

	err = client.Call("KademliaCore.Ping", ping, &pong)

	if err != nil {
		log.Fatal("ERR: ", err)
	}

	k.AddrBook.Update(pong.Sender)

	return "OK: Ping successful"
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

	return "OK: store successfully"
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

	return "OK: Find nodes"
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
	return "ERR: Not implemented"
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	if value, ok := k.LocalData[searchKey]; ok {
		return "OK: " + string(value)
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
