package kademlia

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

// ======================= Set UP===========================
var instance = SetUpNetwork()

// This is total # of Kademlia created for test, change this # if you want.
var TOTAL = 50

func SetUpNetwork() []*Kademlia {
	rand.Seed(time.Now().UnixNano())
	ports := make([]uint16, 0, 50)
	for i := 7890; i < 7890+TOTAL; i++ {
		ports = append(ports, uint16(i))
	}
	var instance []*Kademlia
	for _, p := range ports {
		instance = append(instance, NewKademlia("localhost:"+strconv.Itoa(int(p))))
	}
	fmt.Printf("Testing with %d nodes:\n", len(ports))
	fmt.Println("Node 0: " + instance[0].NodeID.AsString())
	for i := 1; i < len(ports); i++ {
		instance[i].DoPing(net.IPv4(127, 0, 0, 1), ports[0])
		instance[i].DoIterativeFindNode(instance[i].NodeID)
		fmt.Printf("Node %d: "+instance[i].NodeID.AsString()+"\n", i)
	}
	fmt.Println("Done creating network")
	return instance
}

// ======================= Primitive operations ===========================
func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	instance1 := instance[0]
	instance2 := instance[1]
	instance1.DoPing(instance2.SelfContact.Host, instance2.SelfContact.Port)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	assertStringEqual(
		instance1.NodeID.AsString(),
		contact1.NodeID.AsString(),
		"Instance 1 ID incorrectly stored in Instance 2's contact list",
		t)
	assertStringEqual(
		instance2.NodeID.AsString(),
		contact2.NodeID.AsString(),
		"Instance 2 ID incorrectly stored in Instance 1's contact list",
		t)
}

func Test_Store(t *testing.T) {
	instance1 := instance[0]
	instance2 := instance[1]
	key := NewRandomID()
	value := []byte(key.AsString())
	instance1.DoStore(&instance2.SelfContact, key, value)
	// Found value with given key
	assertStringEqual(
		"OK: Found value: "+string(value),
		instance2.LocalFindValue(key),
		"Instance 2 store wrong value!",
		t)
	// Not found value and got nodes instead
	assertContains(
		instance1.DoFindValue(&instance2.SelfContact, NewRandomID()),
		"OK: Found nodes:",
		"DoFindValue from Instance 1 retrieve wrong value!",
		t)
}

func Test_FindValue(t *testing.T) {
	instance1 := instance[0]
	instance2 := instance[1]
	key := NewRandomID()
	value := []byte(key.AsString())
	instance1.DoStore(&instance2.SelfContact, key, value)
	assertStringEqual(
		"OK: Found value: "+string(value),
		instance1.DoFindValue(&instance2.SelfContact, key),
		"DoFindValue from Instance 1 retrieve wrong value!",
		t)
}

// =============================== Iterative ==============================

/*
 * Try find existing node
 * - for each node, try to find other node id and should get it in return list.
 */
func Test_DoIterativeFindNodeSucc(t *testing.T) {
	N := len(instance)
	for i := 0; i < N/k+1; i++ {
		from := (rand.Int() % (N - 1)) + 1
		to := (rand.Int() % (N - 1)) + 1
		for to == from {
			to = (rand.Int() % (N - 1)) + 1
		}
		fmt.Printf(
			"Looking for: %s\nFrom: %s\n",
			instance[from].NodeID.AsString(),
			instance[to].NodeID.AsString())
		result := instance[from].DoIterativeFindNode(instance[to].NodeID)
		assertContains(
			result,
			instance[to].NodeID.AsString(),
			fmt.Sprintf("Cannot find node %d from node %d", to, from),
			t)
	}
}

/*
 * Try find not existing node and test number of return contacts
 */
func Test_DoIterativeFindNodeFail(t *testing.T) {
	notexistid := NewRandomID()
	list := instance[0].DoIterativeFindNode(notexistid)
	assertNotContains(
		list,
		notexistid.AsString(),
		"Found not existing id",
		t)
}

/*
 * Try iterative store and iterative find successfully
 * - for each node, iterative store a key-value pair
 * - if N <= k call LocalFindValue on each other nodes and should get it
 * - then call iterativeFindNode on this node and should find it
 */
func Test_DoIterativeStoreFindSucc(t *testing.T) {
	N := len(instance)
	for i := 0; i < N; i++ {
		key := NewRandomID()
		value := []byte(key.AsString())
		fmt.Println("Testing: " + instance[i].NodeID.AsString())
		instance[i].DoIterativeStore(key, value)
		if N <= k {
			for j := 0; j < N; j++ {
				if j == i {
					continue
				}
				res := instance[j].LocalFindValue(key)
				assertStringEqual(
					"OK: Found value: "+string(value),
					res,
					fmt.Sprintf("Cannot find value in %d stored by %d", j, i),
					t)
			}
		}
		result := instance[i].DoIterativeFindValue(key)
		assertContains(
			result,
			string(value),
			fmt.Sprintf("Cannot iterative find value from node %d", i),
			t)
	}
}

/*
 * Try iterative find not existing key and test number of return contacts
 */
func Test_DoIterativeStoreFindFail(t *testing.T) {
	notexistid := NewRandomID()
	result := instance[0].DoIterativeFindValue(notexistid)
	assertNotContains(
		result,
		notexistid.AsString(),
		"Key-Value not found but return number of contacts not match",
		t)
}
