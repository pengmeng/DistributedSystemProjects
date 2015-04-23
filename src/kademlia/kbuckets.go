package kademlia

import (
	"container/list"
)

type KBuckets struct {
	SelfId ID
	Lists  [b]*list.List
}

func BuildKBuckets(selfId ID) *KBuckets {
	kbuckets := new(KBuckets)
	kbuckets.SelfId = selfId
	for i := 0; i < b; i++ {
		kbuckets.Lists[i] = list.New()
	}
	return kbuckets
}

func (kb KBuckets) Find(nodeId ID) (*Contact, error) {
	ele, err := kb.find(nodeId)
	if ele != nil {
		return ele.Value.(*Contact), err
	} else {
		return nil, err
	}
}

func (kb KBuckets) find(nodeId ID) (*list.Element, error) {
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
		// need to ping front node
		l.Remove(l.Front())
		l.PushBack(con)
	} else {
		l.PushBack(con)
	}
}

func (kb *KBuckets) update(index int, node *list.Element) {
	l := kb.Lists[index]
	l.MoveToBack(node)
}

func (kb *KBuckets) HandleContact(conCh chan *Contact) {
	for {
		select {
		case con := <-conCh:
			index := con.NodeID.Xor(kb.SelfId).PrefixLen()
			if index == 160 {
				break
			}
			if ele, err := kb.find(con.NodeID); err != nil {
				kb.add(index, con)
			} else {
				kb.update(index, ele)
			}
		}
	}
}
