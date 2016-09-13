package dscache

import (
	"time"
)

type Dscache struct {
	buckets [NUMBEROFLISTS]*lrucache
}

const NUMBEROFLISTS = 32

/*

TODO:

- Move benchmarks to only one file - Use Better names - Do pure Set too
- clean up all code
- godoc general

*/

func New(maxsize uint64, workerSleep time.Duration) *Dscache {
	ds := new(Dscache)
	for i := 0; i < NUMBEROFLISTS; i++ {
		ds.buckets[i] = newLRUCache(maxsize/NUMBEROFLISTS, workerSleep)
	}
	return ds
}

func (ds *Dscache) Set(key, payload string, expires time.Duration) error {
	list := getListNumberFromKey(key)
	return ds.buckets[list].set(key, payload, expires)
}

func (ds *Dscache) Get(key string) (string, bool) {
	list := getListNumberFromKey(key)
	return ds.buckets[list].get(key)
}

func (ds *Dscache) Purge(key string) bool {
	list := getListNumberFromKey(key)
	return ds.buckets[list].purge(key)
}

func getListNumberFromKey(key string) byte {
	if len(key) > 1 {
		return (key[len(key)-1] + key[len(key)-2]) % NUMBEROFLISTS
	} else {
		return key[len(key)-1] % NUMBEROFLISTS
	}
}
