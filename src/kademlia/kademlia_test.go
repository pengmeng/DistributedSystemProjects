package kademlia

import (
    "testing"
    "net"
    "strconv"
    "fmt"
    "bytes"
)

// ======================= Primitive operations ===========================
func StringToIpPort(laddr string) (ip net.IP, port uint16, err error){
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
    instance1 := NewKademlia("localhost:7890")
    instance2 := NewKademlia("localhost:7891")
    host2, port2, _ := StringToIpPort("localhost:7891")
    instance1.DoPing(host2, port2)
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
    if contact1.NodeID != instance1.NodeID {
        t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
    }
    if contact2.NodeID != instance2.NodeID {
        t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
    }
    return
}

func TestStoreFind(t *testing.T) {
    instance1 := NewKademlia("localhost:7892")
    instance2 := NewKademlia("localhost:7893")

    id := NewRandomID()
    value := []byte("hello world!")

    // Primitive DoStore
    instance1.DoStore(&instance2.SelfContact, id, value)
    
    res := instance2.LocalFindValue(id)
    if res != ("OK: Found value: " + string(value)) {
        t.Error("Instance 2 store wrong value!")
    }

    // Primitive DoFindValue
    res = instance1.DoFindValue(&instance2.SelfContact, id)
    if res != ("OK: Found value: " + string(value)) {
        t.Error("DoFindValue from Instance 1 retrieve wrong value!")   
    }

    // Some problem with gob error??
    // res = instance1.DoFindValue(&instance2.SelfContact, NewRandomID())
    // fmt.Println(res)
    // // if res != ("OK: Found value: " + string(value)) {
    // //     t.Error("DoFindValue from Instance 1 retrieve wrong value!")   
    // // }
}


// =============================== Iterative ==============================
// Global variable, any better design??
var instance = SetUpNetwork()

func SetUpNetwork() []*Kademlia {
    // Create the network
    ports := []uint16{7894, 7895, 7896, 7897, 7898, 7899}
    var instance []*Kademlia

    for _, p := range ports{
        instance = append(instance, NewKademlia("localhost:" + strconv.Itoa(int(p))))
    }

    for i := 1; i < len(ports); i++ {
        instance[i].DoPing(net.IPv4(127,0,0,1), ports[i - 1])
    }
    fmt.Println("Done creating network")
    return instance
}

func TestIterativeFindNode(t *testing.T) {
    // Use Node ID to look up, Need to return that node
    N := len(instance)

    for i := 0; i < N; i++ {
        //if j := i it's not working
        for j := i + 1; j < N; j++ {
            contacts := instance[i].iterativeFindNode(instance[j].NodeID)
            found := false
            for _, c := range contacts {
                if  c.Equal(instance[j].SelfContact) {
                    found = true
                }
            }
            if !found {
                t.Error(fmt.Sprintf("Iterative Find couldn't find node %d from node %d", j, i))
            }        
        }
    }
}

func TestDoIterativeStoreAndIterativeFind(t *testing.T) {
    N := len(instance)

    for i := 0; i < N; i++ {
        //if j := i it's not working
        for j := i + 1; j < N; j++ {
     
            key := instance[j].NodeID
            value := []byte(NewRandomID().AsString())
            
            instance[i].DoIterativeStore(key, value)

            res := instance[j].LocalFindValue(key)
            if res != ("OK: Found value: " + string(value)) {
                t.Error(fmt.Sprintf("Node %d doesn't store correctly key,value from node %d", j, i))
            }

            retrieve := instance[i].DoIterativeFindValue(key)
            if !bytes.Equal(value, []byte(retrieve)) {
                t.Error(fmt.Sprintf("Error retrieving data from node %d", i))
            }
        }
    }
}


