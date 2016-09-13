package dscache

import (
	"errors"
	"sync"
	"time"
)

type node struct {
	key            string
	payload        string
	previous, next *node
	size           uint64
	validTill      time.Time
}

type lrucache struct {
	keys        map[string]*node
	listStart   *node
	listEnd     *node
	size        uint64
	maxsize     uint64
	workerSleep time.Duration
	mu          sync.Mutex
}

var ErrMaxsize = errors.New("Value is Bigger than Allowed Maxsize")

func newLRUCache(maxsize uint64, workerSleep time.Duration) *lrucache {
	lru := new(lrucache)
	lru.keys = make(map[string]*node)
	lru.size = 0
	lru.maxsize = maxsize
	lru.workerSleep = workerSleep
	go lru.worker()
	return lru
}

func (lru *lrucache) set(key, payload string, expires time.Duration) error {

	//Verify Size
	nodeSize := (uint64(len(key)) + uint64(len(payload))) + 8 //Size of node structure is 8
	if nodeSize > lru.maxsize {
		//Node Exceeds Maxsize
		return ErrMaxsize
	}

	lru.mu.Lock()

	//Check to see if it was already set
	old, ok := lru.keys[key]
	if ok {
		//Key exists
		oldSize := old.size
		old.payload = payload
		old.size = nodeSize
		old.validTill = time.Now().Add(expires)
		lru.size = lru.size - oldSize + old.size
		lru.sendToTop(old)
	} else {
		//create and add Node
		n := new(node)
		n.key = key
		n.payload = payload
		n.size = nodeSize
		n.validTill = time.Now().Add(expires)
		lru.keys[key] = n
		lru.size = lru.size + nodeSize
		lru.sendToTop(n)
	}

	if lru.size > lru.maxsize {
		lru.resize()
	}
	lru.mu.Unlock()
	return nil
}

func (lru *lrucache) get(key string) (string, bool) {
	lru.mu.Lock()
	n, ok := lru.keys[key]
	if !ok {
		//It doesn't exist
		lru.mu.Unlock()
		return "", false
	}
	if n.validTill.Before(time.Now()) {
		//It has expired
		lru.delete(n)
		lru.mu.Unlock()
		return "", false
	}
	lru.sendToTop(n)
	lru.mu.Unlock()
	return n.payload, true
}

func (lru *lrucache) purge(key string) bool {
	lru.mu.Lock()
	n, ok := lru.keys[key]
	if !ok {
		lru.mu.Unlock()
		return false
	}
	lru.delete(n)
	lru.mu.Unlock()
	return true
}

func (lru *lrucache) worker() {
	for {
		end := lru.listEnd
		for end != nil {
			if end.validTill.Before(time.Now()) {
				lru.mu.Lock()
				if end != nil {
					nend := end.previous
					lru.delete(end)
					end = nend
				}
				lru.mu.Unlock()
			} else {
				end = end.previous
			}
		}
		time.Sleep(lru.workerSleep)
	}
}

func (lru *lrucache) sendToTop(n *node) {
	var listStart = lru.listStart
	if listStart == nil {
		lru.listStart = n
		lru.listEnd = n
		return
	}
	if listStart == n {
		return
	}
	if n.previous != nil {
		n.previous.next = n.next
		if n.next != nil {
			n.next.previous = n.previous
		}
	}
	if lru.listEnd == n {
		lru.listEnd = n.previous
	}
	n.next = listStart
	n.previous = nil
	listStart.previous = n
	lru.listStart = n
}

func (lru *lrucache) resize() {
	if lru.size > lru.maxsize {
		//Shrink lisk
		for lru.size > lru.maxsize {
			end := lru.listEnd
			lru.delete(end)
		}
	}
}

func (lru *lrucache) delete(n *node) {
	if n.next != nil {
		n.next.previous = n.previous
	}
	if n.previous != nil {
		n.previous.next = n.next
	}
	if n == lru.listStart {
		lru.listStart = n.next
	}
	if n == lru.listEnd {
		lru.listEnd = n.previous
	}
	delete(lru.keys, n.key)
	n.previous = nil
	n.next = nil
	lru.size -= n.size
}
