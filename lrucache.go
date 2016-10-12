package dscache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Node Structure for Doubly Linked List
type node struct {
	key            string
	payload        string
	previous, next *node
	size           uint64
	validTill      time.Time
}

// LRUCache structure
type lrucache struct {
	mu        sync.Mutex
	keys      map[string]*node
	listStart *node
	listEnd   *node
	size      uint64

	maxsize      uint64
	workerSleep  time.Duration
	nodeBaseSize uint64
}

// ErrMaxsize Used when a key + payload is bigger than allowed LRU Cache size
var ErrMaxsize = errors.New("Value is Bigger than Allowed Maxsize")

// newLRUCache Constructor
func newLRUCache(maxsize uint64, workerSleep time.Duration) *lrucache {
	lru := new(lrucache)
	lru.keys = make(map[string]*node)
	lru.size = 0
	lru.maxsize = maxsize
	lru.workerSleep = workerSleep
	lru.nodeBaseSize = lru.calculateBaseNodeSize()
	go lru.worker()
	return lru
}

// set an element
func (lru *lrucache) set(key, payload string, expires time.Duration) error {

	// Verify Size
	nodeSize := (uint64(len(key)) + uint64(len(payload))) + lru.nodeBaseSize // Size of node structure is 8
	if nodeSize > lru.maxsize {
		// Node Exceeds Maxsize
		return ErrMaxsize
	}

	lru.mu.Lock()
	defer lru.mu.Unlock()

	// Check to see if it was already set
	old, ok := lru.keys[key]
	if ok {
		// Key exists
		oldSize := old.size
		old.payload = payload
		old.size = nodeSize
		old.validTill = time.Now().Add(expires)
		diff := int64(nodeSize) - int64(oldSize)
		if diff > 0 {
			atomic.AddUint64(&lru.size, uint64(diff))
		} else {
			atomic.AddUint64(&lru.size, ^uint64(diff-1))
		}
		lru.sendToTop(old)
	} else {
		// create and add Node
		n := new(node)
		n.key = key
		n.payload = payload
		n.size = nodeSize
		n.validTill = time.Now().Add(expires)
		lru.keys[key] = n
		atomic.AddUint64(&lru.size, nodeSize)
		lru.sendToTop(n)
	}

	if lru.size > lru.maxsize {
		lru.resize()
	}
	return nil
}

// get an element
func (lru *lrucache) get(key string) (string, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	n, ok := lru.keys[key]
	if !ok {
		// It doesn't exist
		return "", false
	}
	if n.validTill.Before(time.Now()) {
		// It has expired
		lru.delete(n)
		return "", false
	}
	lru.sendToTop(n)
	return n.payload, true
}

func (lru *lrucache) purge(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	n, ok := lru.keys[key]
	if !ok {
		return false
	}
	lru.delete(n)
	return true
}

// worker Expiration worker
//
// Expriration Workers go from the bottom of the list to the top
// And delete all elements that have expired.
// Then they wait for the configured time before starting again.
func (lru *lrucache) worker() {
	for {
		lru.mu.Lock()
		end := lru.listEnd
		lru.mu.Unlock()

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
				lru.mu.Lock()
				end = end.previous
				lru.mu.Unlock()
			}
		}

		time.Sleep(lru.workerSleep)
	}
}

// sendToTop promote node to top of list
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

// resize Resise list by size from the bottom
func (lru *lrucache) resize() {
	if lru.size > lru.maxsize {
		// Shrink lisk
		for lru.size > lru.maxsize {
			end := lru.listEnd
			lru.delete(end)
		}
	}
}

// delete Delete node
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
	n.previous = nil
	n.next = nil

	// Test if it's in the keys. (Function might have been called from worker after it has been deleted by another goroutine
	// since worker does not lock the structure all the time. The following situation is pausible: A node is selected by worker
	// in it's iterations, the lock is released, another routine locks and then deletes that node, then worker finds out the node
	// has expired, locks and tries to delete it again, decrementing lru.size 2 times)
	if _, ok := lru.keys[n.key]; ok {
		delete(lru.keys, n.key)
		atomic.AddUint64(&lru.size, ^uint64(n.size-1))
	}
}

// calculateBaseNodeSize Calculate the Byte Size of a single Node
func (lru *lrucache) calculateBaseNodeSize() uint64 {
	n := new(node)
	size := uint64(unsafe.Sizeof(n.key)) + uint64(unsafe.Sizeof(n.payload)) + uint64(unsafe.Sizeof(n.previous)) + uint64(unsafe.Sizeof(n.next)) + uint64(unsafe.Sizeof(n.size)) + uint64(unsafe.Sizeof(n.validTill))
	return size
}

// verifyEndAndStart testing function
//
// For Concurrent tests.
// Verifies that list is the same from listStart to listEnd
func (lru *lrucache) verifyEndAndStart() error {

	lru.mu.Lock()
	defer lru.mu.Unlock()

	start := lru.listStart

	if start != nil {

		// Get to last element of start
		for start.next != nil {
			start = start.next
		}

		end := lru.listEnd

		// Compare them
		for start.previous != nil {
			if end != start {
				return errors.New("listStart does not match order of listEnd")
			}
			end = end.previous
			start = start.previous
		}

	}

	return nil
}

// verifyUniqueKey testing function
//
// For Concurrent tests.
// Verifies that list has all unique keys
func (lru *lrucache) verifyUniqueKeys() error {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	test := make(map[string]bool)
	start := lru.listStart
	for start != nil {
		_, ok := test[start.key]
		if !ok {
			test[start.key] = true
		} else {
			return errors.New("Duplicated Key in listStart")
		}
		start = start.next
	}
	return nil
}

// verifySize testing function
//
// For Concurrent tests.
// Verifies that list size is consistent with actual size
func (lru *lrucache) verifySize() error {

	lru.mu.Lock()
	defer lru.mu.Unlock()

	start := lru.listStart
	realSize := uint64(0)
	sumSize := uint64(0)

	if start != nil {

		// Get to last element of start
		for start.next != nil {
			realSize += uint64(len(start.key)) + uint64(len(start.payload)) + lru.calculateBaseNodeSize()
			sumSize += start.size
			start = start.next
		}

		// Compare them
		if realSize > lru.maxsize {
			err := fmt.Sprintf("realSize: %v  > maxsize: %v --- size: %v --sumSize: %v", realSize, lru.maxsize, lru.size, sumSize)
			return errors.New(err)
		}
		if sumSize > lru.maxsize {
			err := fmt.Sprintf("sumSize: %v  > maxsize: %v --- size: %v ---realSize: %v", sumSize, lru.maxsize, lru.size, realSize)
			return errors.New(err)
		}
	}

	return nil
}
