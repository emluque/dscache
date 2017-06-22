// Copyright 2016 Emiliano Martínez Luque. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dscache

import (
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// Dscache Base Structure
type Dscache struct {
	buckets         []*lrucache
	getBucketNumber func(string) int
	numGets         uint64
	numRequests     uint64
	numSets         uint64
}

// Default Number of Buckets in Dscache
const defaultNumberOfBuckets = int(32)

// Default Duration of sleep for expiring items workers
const defaultWorkerSleep = time.Second

// ErrCreateMaxsizeOfZero Returned when attemping to creat DSCache with a maxsize of 0
var ErrCreateMaxsizeOfZero = errors.New("Building dscache with maxsize of 0")

// ErrCreateWorkerSleep Returned when attemping to creat DSCache with a gcWorkerSleep lower than 1/5
var ErrCreateGCWorkerSleep = errors.New("Building dscache with gcWorkerSleep < 1/5 of a Second")

// Function that creates the Default Get Bucket Number Function
//
// The default getBucketNumber function
// Taken from http://www.partow.net/programming/hashfunctions/index.html#BKDRHashFunction
var defaultGetBucketNumber = func(numBuckets int) func(string) int {
	return func(key string) int {
		seed := uint64(131)
		hash := uint64(0)
		for _, r := range key {
			hash = hash*seed + uint64(r)
		}

		return int(hash) % numBuckets
	}
}

// Byte Sizes Constants
//
// Accesible through dscache.GB, dscache.MB, etc.
const (
	B  uint64 = iota
	KB        = 1 << (10 * iota)
	MB
	GB
	TB

//	PB
)

// New DSCache with Default values
//
// @param 	maxsize		Maxsize of cache in Bytes
func New(maxsize uint64) (*Dscache, error) {

	if maxsize == 0 {
		return nil, ErrCreateMaxsizeOfZero
	}

	ds := new(Dscache)
	ds.buckets = make([]*lrucache, defaultNumberOfBuckets, defaultNumberOfBuckets)
	for i := 0; i < defaultNumberOfBuckets; i++ {
		ds.buckets[i] = newLRUCache(maxsize/uint64(defaultNumberOfBuckets), defaultWorkerSleep)
	}
	ds.getBucketNumber = defaultGetBucketNumber(defaultNumberOfBuckets)
	return ds, nil
}

// Custom Constructor
//
// @param	maxsize	Maxsize of cache in Bytes
// @param	numberOfBuckets	Number of Bucktets in Dscache
//		Suggested Use number of CPU Cores * 8
//		default: 32
// @param	gcWorkerSleep	Time to sleep bettween calls to GC
//		0 to disable GC Worker
//		default: 1 Second
// @param	workerSleep	Time to sleep for expiration workers
//		0 to disable Expiration Worker
//		default: 1 Second
// @param	getBucketNumber	function to calculate the bucket number from a key
func Custom(maxsize uint64, numberOfBuckets int, gcWorkerSleep time.Duration, workerSleep time.Duration, getBucketNumber func(string) int) (*Dscache, error) {

	if maxsize == 0 {
		return nil, ErrCreateMaxsizeOfZero
	}

	if gcWorkerSleep > 0 && gcWorkerSleep < time.Second/5 {
		return nil, ErrCreateGCWorkerSleep
	}

	if numberOfBuckets == 0 {
		numberOfBuckets = defaultNumberOfBuckets
	}

	if getBucketNumber == nil {
		getBucketNumber = defaultGetBucketNumber(numberOfBuckets)
	}

	if workerSleep == 0 {
		workerSleep = defaultWorkerSleep
	}

	ds := new(Dscache)
	ds.buckets = make([]*lrucache, numberOfBuckets, numberOfBuckets)
	for i := 0; i < numberOfBuckets; i++ {
		ds.buckets[i] = newLRUCache(maxsize/uint64(numberOfBuckets), workerSleep)
	}
	ds.getBucketNumber = getBucketNumber

	if gcWorkerSleep > 0 {
		go gcWorker(gcWorkerSleep)
	}
	return ds, nil

}

// Set element
//
// @param key element key
//
// @param payload element payload
//
// @param expires Time.Duration ie: For how much time should it be valid
func (ds *Dscache) Set(key, payload string, expires time.Duration) error {
	Bucket := ds.getBucketNumber(key)
	atomic.AddUint64(&ds.numSets, 1)
	return ds.buckets[Bucket].set(key, payload, expires)
}

// Get element
//
// @param key element key
func (ds *Dscache) Get(key string) (string, bool) {
	Bucket := ds.getBucketNumber(key)
	payload, ok := ds.buckets[Bucket].get(key)
	if ok {
		atomic.AddUint64(&ds.numGets, 1)
	}
	atomic.AddUint64(&ds.numRequests, 1)
	return payload, ok
}

// Purge (delete) element
//
// @param key element key
func (ds *Dscache) Purge(key string) bool {
	Bucket := ds.getBucketNumber(key)
	return ds.buckets[Bucket].purge(key)
}

// Garbage Collection Worker
func gcWorker(gcSleepTime time.Duration) {
	for {
		time.Sleep(gcSleepTime)
		runtime.GC()
	}
}

/*
func (ds *Dscache) Inspect() {
	for i := 0; i < len(ds.buckets); i++ {
		ds.buckets[i].mu.Lock()
		fmt.Println("Bucket: ", i, " -- Maxize: ", ds.buckets[i].maxsize, " -- Size: ", ds.buckets[i].size)
		ds.buckets[i].mu.Unlock()
	}
}
*/

// Verify all of the lits on buckets for inconsistencies.
//
// Used for testing
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

// NumGets Number of Gets the cache has had
func (ds *Dscache) NumGets() uint64 {
	numGets := atomic.LoadUint64(&ds.numGets)
	return numGets
}

// NumRequests Number of Requests the cache has had
func (ds *Dscache) NumRequests() uint64 {
	numRequests := atomic.LoadUint64(&ds.numRequests)
	return numRequests
}

// NumSets Number of Sets the cache has had
func (ds *Dscache) NumSets() uint64 {
	numSets := atomic.LoadUint64(&ds.numSets)
	return numSets
}

// NumObjects Number of Objects in Cache
func (ds *Dscache) NumObjects() uint32 {
	numObjects := uint32(0)
	for i := 0; i < len(ds.buckets); i++ {
		ds.buckets[i].mu.Lock()
		numObjects += uint32(len(ds.buckets[i].keys))
		ds.buckets[i].mu.Unlock()
	}
	return numObjects
}

// NumEvictions Number of Evictions from Cache
func (ds *Dscache) NumEvictions() uint64 {
	numEvictions := uint64(0)
	for i := 0; i < len(ds.buckets); i++ {
		ne := atomic.LoadUint64(&ds.buckets[i].NumEvictions)
		numEvictions += ne
	}
	return numEvictions
}

// HitRate Gets/Tries
func (ds *Dscache) HitRate() float64 {
	return float64(ds.NumGets()) / float64(ds.NumRequests())
}
