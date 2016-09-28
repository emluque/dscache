// Copyright 2016 Emiliano Martínez Luque. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

/*

	Script to run simulations of dscache usage to test how it uses memory

	Flags:

		-verify boolean
			true 		verify all buckets of dscache every Second
			false 	print memory stats every second

		-keySize int
			Number of keys to be used.
				Considering each key may take a paylod from 5000 to 10000 chars,
				the number of possible keys deterimines the total size of all cacheable
				elements. Which combined with dsMaxSize (the size of the cache) will deterimine
				get failure rate.

		-dsMaxSize float64
			Maximum size in GB of the cache.

		-dsLists	int
			Number of buckets in dscache.

		-dsGCSleep float64
			Seconds to wait before running GC worker in dscache.

		-dsWorkerSleep float64
		Seconds to wait before running expiration cleanup worker in each bucket.

		-numGoRoutines int
			Number of goroutines to be running get/set operations.

		-expires int
			Expire for sets in Seconds. Default 3600 (1 Hour)
*/

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/emluque/dscache"
)

// Create a constant string with 10000 chars
const tenChars = "0123456789"
const hundredChars = tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars
const thousandChars = hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars
const tenThousandChars = thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars

func main() {

	verify := flag.Bool("verify", false, "Wether to run on Verify or Simulation Mode.")
	keySize := flag.Int("keySize", 800000, "Number of Keys to use in testing.")
	dsMaxSize := flag.Float64("dsMaxSize", 4.0, "ds Maxsize, in GB, may take floats.")
	dsLists := flag.Int("dsLists", 32, "ds Number Of Lists.")
	dsGCSleep := flag.Float64("dsGCSleep", 1.0, "ds GC Sleep, in Seconds, may take floats.")
	dsWorkerSleep := flag.Float64("dsWorkerSleep", 0.5, "ds Worker Sleep, in Seconds, may take floats.")
	numGoRoutines := flag.Int("numGoRoutines", 64, "Number of Goroutines to be accessing the cache simultaneously.")
	expires := flag.Int("expires", 3600, "Expire for sets in Seconds.")
	flag.Parse()

	printConf(*verify, *keySize, *dsMaxSize, *dsLists, *dsGCSleep, *dsWorkerSleep, *numGoRoutines, *expires)

	ds := dscache.Custom(uint64(*dsMaxSize*float64(dscache.GB)), *dsLists, time.Duration(float64(time.Second)**dsGCSleep), time.Duration(float64(time.Second)**dsWorkerSleep), nil)

	keyArr := generateKeys()

	// Launch Goroutines that do the actual work.
	for i := 0; i < *numGoRoutines; i++ {
		go runOps(ds, *keySize, &keyArr, time.Duration(*expires)*time.Second)
	}

	var i int
	var memStats runtime.MemStats

	// Register Signal for exiting program. Ctrl C on Linux.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		printExit(i, &memStats, ds)
		printConf(*verify, *keySize, *dsMaxSize, *dsLists, *dsGCSleep, *dsWorkerSleep, *numGoRoutines, *expires)
		os.Exit(1)
	}()

	// Main program, every second either verify the structure or print stats.
	for i = 0; i < 10000; i++ {
		if *verify {
			ds.Verify()
		} else {
			printStats(&memStats, ds)
		}
		time.Sleep(time.Second * 1)
	}

}

// Generate Keys
// Number of Keys: 7311616
//	All Payloads Size: [35, 70] GB
func generateKeys() [7311616]string {
	var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var keyArr [7311616]string
	count := 0
	for i := 0; i < len(letters); i++ {
		for j := 0; j < len(letters); j++ {
			for k := 0; k < len(letters); k++ {
				for l := 0; l < len(letters); l++ {
					var tmpKey = letters[i:i+1] + letters[j:j+1] + letters[k:k+1] + letters[l:l+1]
					keyArr[count] = tmpKey
					count++
				}
			}
		}
	}
	return keyArr
}

// If Key is present get it
// If it's not set it with a string of 5000 to 10001 characters
func getSet(ds *dscache.Dscache, key string, expires time.Duration) {
	_, ok := ds.Get(key)
	if !ok {
		rand.Seed(time.Now().UnixNano())
		randomLength := rand.Intn(5000) + 4999
		str := tenThousandChars[0:randomLength] + "  "
		ds.Set(key, str, expires)
	}
}

// Select a Key randomly from the specied keySize
// Run getSet on it
func runOps(ds *dscache.Dscache, keySize int, keyArr *[7311616]string, expires time.Duration) {
	for {
		key := keyArr[rand.Intn(keySize)]
		getSet(ds, key, expires)
	}
}

// Print configuration
func printConf(verify bool, keySize int, dsMaxSize float64, dsLists int, dsGCSleep float64, dsWorkerSleep float64, numGoRoutines int, expires int) {
	fmt.Println("--------------------------------------------")
	fmt.Println("Verify:\t\t\t\t", verify)
	fmt.Println("-----")
	fmt.Println("keySize:\t\t\t", keySize)
	fmt.Printf("Payload Total:\t\t\t(%dGB, %dGB)\n", keySize*5000/dscache.GB, keySize*10000/dscache.GB)
	fmt.Printf("Payload Est.:\t\t\t%dGB\n", keySize*7500/dscache.GB)
	fmt.Println("-----")
	fmt.Println("ds.MaxSize:\t\t\t", dsMaxSize, "GB")
	fmt.Println("ds.Lists:\t\t\t", dsLists)
	fmt.Println("ds.GCSleep:\t\t\t", dsGCSleep)
	fmt.Println("ds.Workerleep:\t\t\t", dsWorkerSleep)
	fmt.Println("-----")
	fmt.Println("NumGoRoutines:\t\t\t", numGoRoutines)
	fmt.Println("expires:\t\t\t", expires)
	fmt.Println()
}

//Print Stats
func printStats(memStats *runtime.MemStats, ds *dscache.Dscache) {

	runtime.ReadMemStats(memStats)

	fmt.Println("--------------------------------------------")
	fmt.Println("Alloc:\t\t\t", memStats.Alloc)
	fmt.Println("Sys:\t\t\t", memStats.Sys)
	fmt.Println("-----")
	fmt.Println("TotalAlloc:\t\t", memStats.TotalAlloc)
	fmt.Println("-----")
	fmt.Println("HeapAlloc:\t\t", memStats.HeapAlloc)
	fmt.Println("HeapSys:\t\t", memStats.HeapSys)
	fmt.Println("HeapIdle:\t\t", memStats.HeapIdle)
	fmt.Println("HeapInuse:\t\t", memStats.HeapInuse)
	fmt.Println("HeapReleased:\t\t", memStats.HeapReleased)
	fmt.Println("HeapObjects:\t\t", memStats.HeapObjects)
	fmt.Println("-----")
	fmt.Println("ds.NumObjects:\t\t", ds.NumObjects())
	fmt.Println("ds.NumGets:\t\t", ds.NumGets)
	fmt.Println("ds.NumSets:\t\t", ds.NumSets)
	fmt.Printf("ds.FailureRate:\t\t%.3f\n", ds.FailureRate())
	fmt.Println("-----")
	fmt.Println("NextGC:\t\t", memStats.NextGC)
	fmt.Println("LastGC:\t\t", memStats.LastGC)
	fmt.Println("NumGC:\t\t", memStats.NumGC)

}

//Print Exit Message
func printExit(i int, memStats *runtime.MemStats, ds *dscache.Dscache) {

	fmt.Println()
	fmt.Println()
	fmt.Println()
	printStats(memStats, ds)
	fmt.Println("--------------------------------------------")
	fmt.Println()
	fmt.Println("Exiting.")
	fmt.Println("Ran ", i, " times.")
	fmt.Println()

}
