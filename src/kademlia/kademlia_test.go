package kademlia

import (
	"testing"
	"net"
)

func Test_kBuckets(t *testing.T) {
	id := NewRandomID()

	AddrBook := BuildKBuckets(id)
	AddrBook.Find(id)

	// add remote Contact to the address book
	var conCh = make (chan *Contact)
	remoteID := NewRandomID()
	conCh <- &Contact{remoteID, net.IPv4(127,0,0,1), 7809}

	AddrBook.HandleContact(conCh)
	
	C, _ := AddrBook.Find(remoteID)

	if (C.NodeID != remoteID) {
		t.Error("Error!")
	}
}