package kademlia

import (
	"container/list"
	"sort"
)

type KBuckets struct {
	SelfContact Contact
	SelfId      ID
	Lists       [b]*list.List
	updateCh    chan *Contact
	removeCh    chan ID
	//channels for find a single contact
	findCh chan ID
	resCh  chan *Contact
	//channels for find k closest contacts with an ID
	closestCh    chan ID
	closestResCh chan []Contact
}

// =============== Public API ========================
func (kb *KBuckets) Update(c Contact) {
	kb.updateCh <- &c
}

func (kb *KBuckets) Remove(nodeId ID) {
	kb.removeCh <- nodeId
}

func (kb KBuckets) Find(nodeId ID) []Contact {
	kb.closestCh <- nodeId
	result := <-kb.closestResCh
	return result
}

func (kb KBuckets) FindThree(nodeId ID) []Contact {
	result := kb.Find(nodeId)
	length := 3
	if len(result) < 3 {
		length = len(result)
	}
	return result[:length]
}

func (kb *KBuckets) FindOne(nodeId ID) (*Contact, error) {
	kb.findCh <- nodeId
	if result := <-kb.resCh; result != nil {
		return result, nil
	} else {
		return nil, &NotFoundError{nodeId, "Not Found"}
	}
}

// =======================================================

func BuildKBuckets(self Contact) *KBuckets {
	kbuckets := new(KBuckets)
	kbuckets.SelfContact = self
	kbuckets.SelfId = self.NodeID
	for i := 0; i < b; i++ {
		kbuckets.Lists[i] = list.New()
	}
	kbuckets.updateCh = make(chan *Contact)
	kbuckets.removeCh = make(chan ID)
	kbuckets.findCh = make(chan ID)
	kbuckets.resCh = make(chan *Contact)
	kbuckets.closestCh = make(chan ID)
	kbuckets.closestResCh = make(chan []Contact)
	go kbuckets.handleContact()
	return kbuckets
}

func (kb KBuckets) find_contact(nodeId ID) (*Contact, error) {
	ele, err := kb.find_element(nodeId)
	if ele != nil {
		return ele.Value.(*Contact), err
	} else {
		return nil, err
	}
}

func (kb KBuckets) find_element(nodeId ID) (*list.Element, error) {
	dis := nodeId.Xor(kb.SelfId)
	index := dis.PrefixLen()
	for each := kb.Lists[index].Front(); each != nil; each = each.Next() {
		if nodeId == each.Value.(*Contact).NodeID {
			return each, nil
		}
	}
	return nil, &NotFoundError{nodeId, "Not Found"}
}

func (kb *KBuckets) add(index int, con *Contact) {
	l := kb.Lists[index]
	if l.Len() == k {
		node := l.Front().Value.(*Contact)
		if _, err := PingHelper(kb.SelfContact, node.Host, node.Port); err != nil {
			l.Remove(l.Front())
			l.PushBack(con)
		} else {
			l.MoveToBack(l.Front())
		}
	} else {
		l.PushBack(con)
	}
}

func (kb *KBuckets) update(index int, node *list.Element) {
	l := kb.Lists[index]
	l.MoveToBack(node)
}

func (kb *KBuckets) remove(index int, node *list.Element) {
	l := kb.Lists[index]
	l.Remove(node)
}

func (kb *KBuckets) handleContact() {
	for {
		select {
		case con := <-kb.updateCh:
			index := con.NodeID.Xor(kb.SelfId).PrefixLen()
			if index == 160 {
				continue
			}
			if ele, err := kb.find_element(con.NodeID); err != nil {
				kb.add(index, con)
			} else {
				kb.update(index, ele)
			}
		case nodeId := <-kb.removeCh:
			index := nodeId.Xor(kb.SelfId).PrefixLen()
			if ele, err := kb.find_element(nodeId); err == nil {
				kb.remove(index, ele)
			}
		case nodeId := <-kb.findCh:
			if result, err := kb.find_contact(nodeId); err != nil {
				kb.resCh <- nil
			} else {
				kb.resCh <- result
			}
		case nodeId := <-kb.closestCh:
			index := nodeId.Xor(kb.SelfId).PrefixLen()
			l := make([]Contact, 0, k)
			kb.feedWithCLosest(&l, index, nodeId)
			sort.Stable(ContactArray{l, nodeId})
			kb.closestResCh <- l
		}
	}
}

func (kb *KBuckets) feedWithCLosest(s *[]Contact, index int, nodeId ID) {
	for i := index; i < b; i++ {
		copy2array(s, kb.Lists[i], nodeId)
		if len(*s) == k {
			return
		}
	}
	for i := index - 1; i >= 0; i-- {
		copy2array(s, kb.Lists[i], nodeId)
		if len(*s) == k {
			return
		}
	}
}

func copy2array(s *[]Contact, l *list.List, nodeId ID) {
	for each := l.Front(); each != nil; each = each.Next() {
		if len(*s) == k {
			break
		}
		if !nodeId.Equals(each.Value.(*Contact).NodeID) {
			*s = append(*s, *each.Value.(*Contact))
		}
	}
}

type ContactArray struct {
	Array []Contact
	Id    ID
}

func (c ContactArray) Len() int {
	return len(c.Array)
}

func (c ContactArray) Swap(i, j int) {
	c.Array[i], c.Array[j] = c.Array[j], c.Array[i]
}

func (c ContactArray) Less(i, j int) bool {
	dis_i := c.Array[i].NodeID.Xor(c.Id).PrefixLen()
	dis_j := c.Array[j].NodeID.Xor(c.Id).PrefixLen()
	return dis_i > dis_j
}
