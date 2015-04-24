package kademlia

import (
	"container/list"
	"fmt"
	"net"
	"testing"
)

func Test_kBuckets(t *testing.T) {
	id := NewRandomID()

	AddrBook := BuildKBuckets(id)

	// add remote Contact to the address book
	remoteID := NewRandomID()
	AddrBook.Update(Contact{remoteID, net.IPv4(127, 0, 0, 1), 7809})
	C := AddrBook.Find_for_testing(remoteID)
	if con, ok := C.(*Contact); !ok {
		t.Error("Type is not *Contact")
	} else {
		if !con.NodeID.Equals(remoteID) {
			t.Error("Error!")
		}
	}

	notexistID := NewRandomID()
	notexist := AddrBook.Find_for_testing(notexistID)
	if l, ok := notexist.(*list.List); !ok {
		t.Error("Return type is not *list.List")
	} else {
		fmt.Printf("Return list length is %d\n", l.Len())
	}
}
