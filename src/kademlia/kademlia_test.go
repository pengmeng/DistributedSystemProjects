package kademlia

import (
	"fmt"
	"net"
	"testing"
)

func Test_Find(t *testing.T) {
	id := NewRandomID()

	AddrBook := BuildKBuckets(id)

	// add remote Contact to the address book
	remoteID := NewRandomID()
	AddrBook.Update(Contact{remoteID, net.IPv4(127, 0, 0, 1), 7809})
	C := AddrBook.Find(remoteID)
	if len(C) == 1 {
		if !C[0].NodeID.Equals(remoteID) {
			t.Error("Return Id not match")
		}
	} else {
		t.Errorf("Return array length is %d\n", len(C))
	}

	notexistID := NewRandomID()
	notexist := AddrBook.Find(notexistID)
	if len(notexist) != 0 {
		t.Errorf("Return array length is %d\n", len(notexist))
	}
}

func Test_MultiContacts(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(self)
	start := 7000
	end := start + 160*20 - 1
	rdict := make(map[int]int)
	for i := start; i <= end; i++ {
		id := NewRandomID()
		dis := self.Xor(id).PrefixLen()
		if _, ok := rdict[dis]; ok {
			rdict[dis] += 1
		} else {
			rdict[dis] = 1
		}
		AddrBook.Update(Contact{id, net.IPv4(127, 0, 0, 1), uint16(i)})
	}
	for k, v := range rdict {
		fmt.Printf("Distance %d has contacts %d\n", k, v)
	}
}
