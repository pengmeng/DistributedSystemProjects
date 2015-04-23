package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
)

type KademliaCore struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (kc *KademliaCore) Ping(ping PingMessage, pong *PongMessage) error {
	// TODO: Finish implementation
	// server Ping
	pong.MsgID = CopyID(ping.MsgID)
	go kc.kademlia.AddrBook.Update(ping.Sender)
	
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (kc *KademliaCore) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.
	go kc.kademlia.AddrBook.Update(req.Sender)
	go kc.kademlia.addData(Pair{req.Key, req.Value})
	res.MsgID = CopyID(req.MsgID)	
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	// find closest nodes to the key
	go kc.kademlia.AddrBook.Update(req.Sender)

	contacts := kc.kademlia.AddrBook.Find(req.NodeID)

	res.MsgID = CopyID(req.MsgID)
	res.Nodes = contacts

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	go kc.kademlia.AddrBook.Update(req.Sender)
	
	key := req.Key

	// check if key is in LocalData
	if value, ok := kc.kademlia.LocalData[key]; ok {
		res.Value = value

	} else {
		res.MsgID = CopyID(req.MsgID)
		res.Nodes = kc.kademlia.AddrBook.Find(req.Key)		
	}

	return nil
}
