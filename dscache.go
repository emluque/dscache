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

type Dscache struct {
	keys        map[string]*node
	listStart   *node
	listEnd     *node
	size        uint64
	Maxsize     uint64
	mu          sync.Mutex
	workerSleep time.Duration
}

/*

- expire worker (test how that works)
- Better size (size, align) for precise memory usage
- do better (actually more) benchmarks
- clean up all code
- godoc general

*/

var ErrMaxsize = errors.New("Value is Bigger than Allowed Maxsize")

func New(maxsize uint64, workerSleep time.Duration) *Dscache {
	ds := new(Dscache)
	ds.keys = make(map[string]*node)
	ds.Maxsize = maxsize
	ds.size = 0
	ds.workerSleep = workerSleep
	if workerSleep > 0 {
		go ds.worker()
	}
	return ds
}

//TODO: ERROR CHECKING
func (ds *Dscache) Set(key, payload string, expires time.Duration) error {
	//Verify Size
	nodeSize := (uint64(len(key)) + uint64(len(payload))) * 8
	if nodeSize > ds.Maxsize {
		//Node Exceeds Maxsize
		return ErrMaxsize
	}

	ds.mu.Lock()

	//Check to see if it was already set
	old, ok := ds.keys[key]
	if ok {
		//Key exists
		oldSize := old.size
		old.payload = payload
		old.size = nodeSize
		old.validTill = time.Now().Add(expires)
		ds.size = ds.size - oldSize + old.size
		ds.sendToTop(old)
	} else {
		//create and add Node
		n := new(node)
		n.key = key
		n.payload = payload
		n.size = nodeSize
		n.validTill = time.Now().Add(expires)
		ds.keys[key] = n
		ds.size = ds.size + nodeSize
		ds.sendToTop(n)
	}

	if ds.size > ds.Maxsize {
		ds.resize()
	}
	ds.mu.Unlock()
	return nil
}

func (ds *Dscache) Get(key string) (string, bool) {
	ds.mu.Lock()
	n, ok := ds.keys[key]
	if !ok {
		//It doesn't exist
		ds.mu.Unlock()
		return "", false
	}
	if n.validTill.Before(time.Now()) {
		//It has expired
		ds.delete(n)
		ds.mu.Unlock()
		return "", false
	}
	ds.sendToTop(n)
	ds.mu.Unlock()
	return n.payload, true
}

func (ds *Dscache) Purge(key string) {
	ds.mu.Lock()
	n, ok := ds.keys[key]
	if !ok {
		ds.mu.Unlock()
		return
	}
	ds.delete(n)
	ds.mu.Unlock()
	return
}

func (ds *Dscache) worker() {
	for {
		end := ds.listEnd
		for end != nil {
			if end.validTill.Before(time.Now()) {
				ds.mu.Lock()
				if end != nil {
					nend := end.previous
					ds.delete(end)
					end = nend
				}
				ds.mu.Unlock()
			} else {
				end = end.previous
			}
		}
		time.Sleep(ds.workerSleep)
	}
}

func (ds *Dscache) sendToTop(n *node) {
	var listStart = ds.listStart
	if listStart == nil {
		ds.listStart = n
		ds.listEnd = n
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
	if ds.listEnd == n {
		ds.listEnd = n.previous
	}
	n.next = listStart
	n.previous = nil
	listStart.previous = n
	ds.listStart = n
}

func (ds *Dscache) resize() {
	if ds.size > ds.Maxsize {
		//Shrink lisk
		for ds.size > ds.Maxsize {
			end := ds.listEnd
			ds.delete(end)
		}
	}
}

func (ds *Dscache) delete(n *node) {
	if n.next != nil {
		n.next.previous = n.previous
	}
	if n.previous != nil {
		n.previous.next = n.next
	}
	if n == ds.listStart {
		ds.listStart = n.next
	}
	if n == ds.listEnd {
		ds.listEnd = n.previous
	}
	delete(ds.keys, n.key)
	n.previous = nil
	n.next = nil
	ds.size -= n.size
}
