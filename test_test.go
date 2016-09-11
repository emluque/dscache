package dscache

import (
	"fmt"
	"testing"
	"time"
	"unsafe"
)

/*

type Dscache struct {
	keys        map[string]*node
	listStart   *node
	listEnd     *node
	size        uint64
	maxsize     uint64
	mu          sync.Mutex
	workerSleep time.Duration
}

*/

func TestTest(t *testing.T) {

	ds := New(100000, time.Second)
	ds.Set("aaa", "aaaaa", time.Second)

	n := ds.listStart

	fmt.Println(unsafe.Sizeof(n), unsafe.Alignof(n))
}
