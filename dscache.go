package dscache

import "time"

type Dscache struct {
	buckets       []*lrucache
	getListNumber func(string) int
}

const defaultNumberOfLists = int(32)
const defaultWorkerSleep = time.Second

var defaultGetListNumber = func(key string) int {
	return int(key[len(key)-1]+key[len(key)-2]) % defaultNumberOfLists
}

/*

TODO:

- Move benchmarks to only one file - Use Better names - Do pure Set too
- clean up all code
- godoc general

*/

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

func NewConfigured(maxsize uint64, numberOfLists int, workerSleep time.Duration, getListNumber func(string) int) *Dscache {

	if maxsize == 0 {
		panic("Building dscache with maxsize of 0.")
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
	return ds

}

func (ds *Dscache) Set(key, payload string, expires time.Duration) error {
	list := ds.getListNumber(key)
	return ds.buckets[list].set(key, payload, expires)
}

func (ds *Dscache) Get(key string) (string, bool) {
	list := ds.getListNumber(key)
	return ds.buckets[list].get(key)
}

func (ds *Dscache) Purge(key string) bool {
	list := ds.getListNumber(key)
	return ds.buckets[list].purge(key)
}
