package kademlia

import (
	"net"
	"testing"
)

func Test_kBuckets(t *testing.T) {
	id := NewRandomID()

	AddrBook := BuildKBuckets(id)

	// add remote Contact to the address book
	var conCh = make(chan *Contact)
	go AddrBook.HandleContact(conCh)
	remoteID := NewRandomID()
	conCh <- &Contact{remoteID, net.IPv4(127, 0, 0, 1), 7809}
	C, _ := AddrBook.Find_helper(remoteID)
	if !C.NodeID.Equals(remoteID) {
		t.Error("Error!")
	}
}
