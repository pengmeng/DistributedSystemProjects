package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"strings"
	"time"
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
	srv := rpc.NewServer()
	srv.Register(&KademliaCore{k})
	_, port, _ := net.SplitHostPort(laddr)
	srv.HandleHTTP(rpc.DefaultRPCPath+port, rpc.DefaultDebugPath+port)
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
	return k.AddrBook.FindOne(nodeId)
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
	port_str := strconv.Itoa(int(port))
	client, err := rpc.DialHTTPPath(
		"tcp",
		fmt.Sprintf("%s:%d", host.String(), port),
		rpc.DefaultRPCPath+port_str)
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
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}
	k.AddrBook.Update(pong.Sender)
	return "OK: Ping " + pong.Sender.NodeID.AsString()
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	port_str := strconv.Itoa(int((*contact).Port))
	client, err := rpc.DialHTTPPath(
		"tcp",
		fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port),
		rpc.DefaultRPCPath+port_str,
	)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}

	req := StoreRequest{k.SelfContact, NewRandomID(), key, value}
	var res StoreResult

	err = client.Call("KademliaCore.Store", req, &res)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}

	return "OK:"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	port_str := strconv.Itoa(int((*contact).Port))
	client, err := rpc.DialHTTPPath(
		"tcp",
		fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port),
		rpc.DefaultRPCPath+port_str,
	)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}
	req := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindNodeResult
	err = client.Call("KademliaCore.FindNode", req, &res)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}
	for _, each := range res.Nodes {
		k.AddrBook.Update(each)
	}
	return fmt.Sprintf("OK: Found %d Nodes", len(res.Nodes))
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	port_str := strconv.Itoa(int((*contact).Port))
	client, err := rpc.DialHTTPPath(
		"tcp",
		fmt.Sprintf("%s:%d", (*contact).Host.String(), (*contact).Port),
		rpc.DefaultRPCPath+port_str,
	)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}
	req := FindValueRequest{k.SelfContact, NewRandomID(), searchKey}
	var res FindValueResult

	err = client.Call("KademliaCore.FindValue", req, &res)
	if err != nil {
		//log.Fatal("ERR: ", err)
		return "ERR: " + err.Error()
	}
	if res.Value != nil {
		return "OK: Found value: " + string(res.Value)
	} else if res.Nodes != nil {
		return "OK: Found nodes: " + string(len(res.Nodes))
	} else {
		return "ERR: Not Found"
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
	var buffer bytes.Buffer
	result := k.iterativeFindNode(id)
	for _, each := range result {
		buffer.WriteString(each.NodeID.AsString() + "\n")
	}
	return buffer.String()
}

func (k *Kademlia) iterativeFindNode(id ID) []Contact {
	findCh := make(chan *Contact, alpha)
	resCh := make(chan string, alpha)
	go k.callFindNode(id, findCh, resCh)
	shortlist := k.AddrBook.Find(id)
	statusMap := make(map[ID]int)
	for !validate(&statusMap, &shortlist) {
		todo := make([]Contact, 0, 3)
		j := 0
		for i := 0; j < alpha && i < len(shortlist); i++ {
			each := shortlist[i]
			if _, ok := statusMap[each.NodeID]; !ok {
				findCh <- &each
				statusMap[each.NodeID] = 0
				todo = append(todo, each)
				j++
			}
		}
		for i := 0; i < len(todo); i++ {
			select {
			case s := <-resCh:
				if strings.Contains(s, "OK:") {
					statusMap[todo[i].NodeID] = 1
				} else {
					statusMap[todo[i].NodeID] = 2
					k.AddrBook.Remove(todo[i].NodeID)
				}
			case <-time.After(time.Millisecond * 300):
				statusMap[todo[i].NodeID] = 2
			}
		}
		shortlist = k.AddrBook.Find(id)
	}
	return shortlist
}

func validate(statusMap *map[ID]int, shortlist *[]Contact) bool {
	for _, each := range *shortlist {
		if v, ok := (*statusMap)[each.NodeID]; !ok {
			return false
		} else if v != 1 {
			return false
		}
	}
	return true
}

func (k *Kademlia) callFindNode(id ID, findCh chan *Contact, resCh chan string) {
	for {
		select {
		case con := <-findCh:
			resCh <- k.DoFindNode(con, id)
		}
	}
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeFindValue(key ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
