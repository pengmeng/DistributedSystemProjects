package kademlia

import (
	"container/list"
)

type KBuckets struct {
	SelfContact Contact
	SelfId      ID
	Lists       [b]*list.List
	updateCh    chan *Contact
	findCh      chan ID
	resCh       chan []Contact
}

// =============== Public API ========================
func (kb *KBuckets) Update(c Contact) {
	kb.updateCh <- &c
}

func (kb KBuckets) Find(nodeId ID) []Contact {
	kb.findCh <- nodeId
	result := <-kb.resCh
	return result
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
	kbuckets.findCh = make(chan ID)
	kbuckets.resCh = make(chan []Contact)
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

func (kb *KBuckets) handleContact() {
	for {
		select {
		case con := <-kb.updateCh:
			index := con.NodeID.Xor(kb.SelfId).PrefixLen()
			if index == 160 {
				break
			}
			if ele, err := kb.find_element(con.NodeID); err != nil {
				kb.add(index, con)
			} else {
				kb.update(index, ele)
			}
		case nodeId := <-kb.findCh:
			index := nodeId.Xor(kb.SelfId).PrefixLen()
			if result, err := kb.find_contact(nodeId); err != nil {
				l := make([]Contact, 0, k)
				copy2array(&l, kb.Lists[index])
				kb.resCh <- l
			} else {
				l := []Contact{*result}
				kb.resCh <- l
			}
		}
	}
}

func copy2array(s *[]Contact, l *list.List) {
	for each := l.Front(); each != nil; each = each.Next() {
		*s = append(*s, *each.Value.(*Contact))
	}
}
