package kademlia

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func Test_Remove(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 7000})

	// add remote Contact to the address book
	remoteID := NewRandomID()
	AddrBook.Update(Contact{remoteID, net.IPv4(127, 0, 0, 1), 7809})
	if C, err := AddrBook.FindOne(remoteID); err == nil {
		if !C.NodeID.Equals(remoteID) {
			t.Error("Return ID not match.")
		}
	} else {
		t.Error("Exist ID not found.")
	}
	AddrBook.Remove(remoteID)
	if _, err := AddrBook.FindOne(remoteID); err == nil {
		t.Error("Delete ID found.")
	}
}

func Test_FindOne(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 7000})

	// add remote Contact to the address book
	remoteID := NewRandomID()
	AddrBook.Update(Contact{remoteID, net.IPv4(127, 0, 0, 1), 7809})
	if C, err := AddrBook.FindOne(remoteID); err == nil {
		if !C.NodeID.Equals(remoteID) {
			t.Error("Return ID not match.")
		}
	} else {
		t.Error("Exist ID not found.")
	}

	notexistID := NewRandomID()
	if _, err := AddrBook.FindOne(notexistID); err == nil {
		t.Error("Not exist ID found.")
	}
}

func Test_FindWithMoreThanK(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 7000})

	start := 7002
	end := start + 50
	for i := start; i <= end; i++ {
		id := NewRandomID()
		AddrBook.Update(Contact{id, net.IPv4(127, 0, 0, 1), uint16(i)})
	}

	idForTest := NewRandomID()
	AddrBook.Update(Contact{idForTest, net.IPv4(127, 0, 0, 1), uint16(7001)})

	if result := AddrBook.Find(idForTest); result == nil {
		t.Error("Return nothing.")
	} else {
		if len(result) != k {
			t.Errorf("Return %d contact but expect %d\n", len(result), k)
		} else {
			for _, each := range result {
				dis := each.NodeID.Xor(idForTest).PrefixLen()
				fmt.Printf("%d ", dis)
			}
			fmt.Println()
		}
	}
}

func Test_FindWithLessThanK(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 7000})

	idForTest := NewRandomID()
	AddrBook.Update(Contact{idForTest, net.IPv4(127, 0, 0, 1), uint16(7001)})

	count := 10
	start := 7002
	end := start + count
	for i := start; i < end; i++ {
		id := NewRandomID()
		AddrBook.Update(Contact{id, net.IPv4(127, 0, 0, 1), uint16(i)})
	}

	if result := AddrBook.Find(idForTest); result == nil {
		t.Error("Return nothing.")
	} else {
		if len(result) != count+1 {
			t.Errorf("Return %d contact but expect %d\n", len(result), count+1)
		} else {
			for _, each := range result {
				dis := each.NodeID.Xor(idForTest).PrefixLen()
				fmt.Printf("%d ", dis)
			}
			fmt.Println()
		}
	}

	newID := NewRandomID()
	if result := AddrBook.Find(newID); result == nil {
		t.Error("Return nothing.")
	} else {
		if len(result) != count+1 {
			t.Errorf("Return %d contact but expect %d\n", len(result), count+1)
		}
	}
}

func Test_FindThree(t *testing.T) {
	self := NewRandomID()
	AddrBook := BuildKBuckets(Contact{self, net.IPv4(127, 0, 0, 1), 7000})
	start := 7002
	end := start + 10
	for i := start; i < end; i++ {
		id := NewRandomID()
		AddrBook.Update(Contact{id, net.IPv4(127, 0, 0, 1), uint16(i)})
	}

	if result := AddrBook.FindThree(NewRandomID()); result == nil {
		t.Error("Return nothing.")
	} else {
		if len(result) != 3 {
			t.Errorf("Return %d contact but expect %d\n", len(result), 3)
		}
	}
}

func Test_LocalData(t *testing.T) {
	k := NewKademlia("127.0.0.1:7000")
	key := NewRandomID()
	value := []byte("hello world")

	k.addData(Pair{key, value})

	res, err := k.getData(key)
	if !bytes.Equal(res, value) || err != nil {
		t.Error("Retrieve wrong value!")
	}

	res, err = k.getData(NewRandomID())
	if res != nil || err == nil {
		t.Error("Not exist key found")
	}
}
