package kademlia

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func Test_Find(t *testing.T) {
	id := NewRandomID()
	AddrBook := BuildKBuckets(Contact{id, net.IPv4(127, 0, 0, 1), 6999})

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
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 6999})
	start := 7000
	end := start + 100
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

func Test_Kademlia(t *testing.T) {
	k := NewKademlia("127.0.0.1:7809")
	key := NewRandomID()
	value := []byte("hello world")

	k.addData(Pair{key, value})

	res, err := k.getData(key)
	if !bytes.Equal(res, value) || err != nil {
		t.Error("Retrieve wrong value!")
	}

	res, err = k.getData(NewRandomID())
	if res != nil || err == nil {
		t.Error("Error dealing with non existent key")
	}

}

func Test_PING(t *testing.T) {

}
