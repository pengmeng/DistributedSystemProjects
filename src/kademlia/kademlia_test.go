package kademlia

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"os"
	"os/exec"
)

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

func testPing(instance1 *Kademlia, instance2 *Kademlia, t *testing.T) {
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

func test_StoreFind(instance1 *Kademlia, instance2 *Kademlia, t *testing.T) {
	key := NewRandomID()
	value := []byte("hello world!")
	instance1.DoStore(&instance2.SelfContact, key, value)
	// Found value with given key
	assertStringEqual(
		"OK: Found value: "+string(value),
		instance2.LocalFindValue(key),
		"Instance 2 store wrong value!",
		t)
	assertStringEqual(
		"OK: Found value: "+string(value),
		instance1.DoFindValue(&instance2.SelfContact, key),
		"DoFindValue from Instance 1 retrieve wrong value!",
		t)
	// Not found value and got nodes instead
	assertContains(
		instance1.DoFindValue(&instance2.SelfContact, NewRandomID()),
		"OK: Found nodes:",
		"DoFindValue from Instance 1 retrieve wrong value!",
		t)
}

// =============================== Iterative ==============================
func SetUpNetwork() ([]*Kademlia, []uint16) {
	ports := make([]uint16, 0, 50)
	count := 5
	for i := 7894; i < 7894+count; i++ {
		ports = append(ports, uint16(i))
	}
	var instance []*Kademlia
	for _, p := range ports {
		instance = append(instance, NewKademlia("localhost:"+strconv.Itoa(int(p))))
	}
	fmt.Printf("Testing with %d nodes:\n", len(ports))
	fmt.Printf("Node 0, port %d: " + instance[0].NodeID.AsString() + "\n", ports[0])
	for i := 1; i < len(ports); i++ {
		instance[i].DoPing(net.IPv4(127, 0, 0, 1), ports[i-1])
		fmt.Printf("Node %d, port %d: "+instance[i].NodeID.AsString()+"\n", i, ports[i])
	}
	fmt.Println("Done creating network")
	return instance, ports
}

/*
 * Try find existing node
 * - for each node, try to find other node id and should get it in return list.
 */
func test_DoIterativeFindNodeSucc(instance []*Kademlia, drop map[int]bool, t *testing.T) {
	N := len(instance)
	for i := 0; i < N; i++ {
		if drop[i] {
			continue;
		}
		fmt.Println("Testing: " + instance[i].NodeID.AsString())
		for j := 0; j < N; j++ {
			if j == i || drop[j] {
				continue
			}
			list := instance[i].DoIterativeFindNode(instance[j].NodeID)
			assertContains(
				list,
				instance[j].NodeID.AsString(),
				fmt.Sprintf("Cannot find node %d from node %d", j, i),
				t)
		}
	}
}

/*
 * Try find not existing node and test number of return contacts
 */
func test_DoIterativeFindNodeFail(instance []*Kademlia, t *testing.T) {
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
 * - call LocalFindValue on each other nodes and should get it
 * - then call iterativeFindNode on this node and should find it
 */
func test_DoIterativeStoreFindSucc(instance []*Kademlia, drop map[int]bool, t *testing.T) {
	N := len(instance)
	for i := 0; i < N; i++ {
		if drop[i] {
			continue;
		}

		key := NewRandomID()
		value := []byte(key.AsString())
		fmt.Println("Testing: " + instance[i].NodeID.AsString())
		instance[i].DoIterativeStore(key, value)

		if N <= k {
			for j := 0; j < N; j++ {
				if j == i || drop[j] {
					continue
				}
				res := instance[j].LocalFindValue(key)
				assertStringEqual(
					"OK: Found value: "+string(value),
					res,
					fmt.Sprintf("Cannot find value in %d stored by %d", j, i),
					t)
			}
			result := instance[i].DoIterativeFindValue(key)
			assertContains(
				result,
				string(value),
				fmt.Sprintf("Cannot iterative find value from node %d", i),
				t)
		}
	}
}

/*
 * Try iterative find not existing key and test number of return contacts
 */
func test_DoIterativeStoreFindFail(instance []*Kademlia, t *testing.T) {
	notexistid := NewRandomID()
	result := instance[0].DoIterativeFindValue(notexistid)
	assertNotContains(
		result,
		notexistid.AsString(),
		"Key-Value not found but return number of contacts not match",
		t)
}

func TestMain(t *testing.T) {
	instance, ports := SetUpNetwork()

	testPing(instance[0], instance[1], t)
	test_StoreFind(instance[0], instance[1], t)
	
	drop := make(map[int]bool)
	for i := 0; i < len(ports); i++ {
		drop[i] = false
	}

	// All nodes are active
	test_DoIterativeFindNodeSucc(instance, drop, t)
	test_DoIterativeFindNodeFail(instance, t)
	test_DoIterativeStoreFindSucc(instance, drop, t)
	test_DoIterativeStoreFindFail(instance, t)

	// Kill some node using IP tables
	dropping := []int {0}

	for _, d := range dropping {
		drop[d] = true
		fmt.Println(fmt.Sprintf("Stopping node %d at port %d", d, ports[d]))
		cmd := []string{	"iptables",
							"-I", "INPUT", "-p", "tcp", 
							"--dport", fmt.Sprintf("%d", ports[d]), "-j", "DROP"}
		if err := exec.Command("sudo", cmd...).Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	test_DoIterativeFindNodeSucc(instance, drop, t)
	test_DoIterativeFindNodeFail(instance, t)
	test_DoIterativeStoreFindSucc(instance, drop, t)

	fmt.Println("Reset ip tables")
	if err := exec.Command("sudo", "iptables", "-F").Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}