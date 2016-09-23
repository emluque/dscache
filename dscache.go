package dscache

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

type Dscache struct {
	buckets       []*lrucache
	getListNumber func(string) int
	NumGets       uint64
	NumSets       uint64
}

/*

TODO:

- godoc --- Add documentation!
- Write a proper readme.md!

*/

const defaultNumberOfLists = int(32)
const defaultWorkerSleep = time.Second

var defaultGetListNumber = func(key string) int {
	return int(key[len(key)-1]+key[len(key)-2]) % defaultNumberOfLists
}

const (
	B  uint64 = iota
	KB        = 1 << (10 * iota)
	MB
	GB
	TB

//	PB
)

func New(maxsize uint64) *Dscache {

	if maxsize == 0 {
		panic("Building dscache with maxsize of 0.")
	}

	ds := new(Dscache)
	ds.buckets = make([]*lrucache, defaultNumberOfLists, defaultNumberOfLists)
	for i := 0; i < defaultNumberOfLists; i++ {
		ds.buckets[i] = newLRUCache(maxsize/uint64(defaultNumberOfLists), defaultWorkerSleep)
	}
	ds.getListNumber = defaultGetListNumber
	return ds
}

func Custom(maxsize uint64, numberOfLists int, gcWorkerSleep time.Duration, workerSleep time.Duration, getListNumber func(string) int) *Dscache {

	if maxsize == 0 {
		panic("Building dscache with maxsize of 0.")
	}

	if gcWorkerSleep > 0 && gcWorkerSleep < time.Second/5 {
		panic("Building dscache with gcWorkerSleep < 1/5 of a Second.")
	}

	if numberOfLists == 0 {
		numberOfLists = defaultNumberOfLists
	}

	if getListNumber == nil {
		getListNumber = defaultGetListNumber
	}

	if workerSleep == 0 {
		workerSleep = defaultWorkerSleep
	}

	ds := new(Dscache)
	ds.buckets = make([]*lrucache, numberOfLists, numberOfLists)
	for i := 0; i < numberOfLists; i++ {
		ds.buckets[i] = newLRUCache(maxsize/uint64(numberOfLists), workerSleep)
	}
	ds.getListNumber = getListNumber

	if gcWorkerSleep > 0 {
		go gcWorker(gcWorkerSleep)
	}
	return ds

}

func (ds *Dscache) Set(key, payload string, expires time.Duration) error {
	list := ds.getListNumber(key)
	atomic.AddUint64(&ds.NumSets, 1)
	return ds.buckets[list].set(key, payload, expires)
}

func (ds *Dscache) Get(key string) (string, bool) {
	list := ds.getListNumber(key)
	payload, ok := ds.buckets[list].get(key)
	if ok {
		atomic.AddUint64(&ds.NumGets, 1)
	}
	return payload, ok
}

func (ds *Dscache) Purge(key string) bool {
	list := ds.getListNumber(key)
	return ds.buckets[list].purge(key)
}

func gcWorker(gcSleepTime time.Duration) {
	for {
		time.Sleep(gcSleepTime)
		runtime.GC()
	}
}

func (ds *Dscache) Inspect() {
	for i := 0; i < len(ds.buckets); i++ {
		ds.buckets[i].mu.Lock()
		fmt.Println("Bucket: ", i, " -- Maxize: ", ds.buckets[i].maxsize, " -- Size: ", ds.buckets[i].size)
		ds.buckets[i].mu.Unlock()
	}
}

func (ds *Dscache) Verify() {
	for i := 0; i < len(ds.buckets); i++ {
		err := ds.buckets[i].verifyEndAndStart()
		if err != nil {
			fmt.Println(err)
		}
		err = ds.buckets[i].verifySize()
		if err != nil {
			fmt.Println(err)
		}
		err = ds.buckets[i].verifyUniqueKeys()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (ds *Dscache) NumObjects() uint32 {
	numObjects := uint32(0)
	for i := 0; i < len(ds.buckets); i++ {
		//		numObjects += ds.buckets[i].numObjects
		numObjects += uint32(len(ds.buckets[i].keys))
	}
	return numObjects
}

func (ds *Dscache) FailureRate() float64 {
	g := ds.NumGets
	s := ds.NumSets
	t := g + s
	return float64(s) / float64(t)
}
