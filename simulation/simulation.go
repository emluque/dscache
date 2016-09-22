package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"

	"../"
)

/*

Simulate actual usage to test memory consumption.

*/

var tenChars = "0123456789"
var hundredChars = tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars
var thousandChars = hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars
var tenThousandChars = thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars

/*
  key number: 7311616
*/
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

func getSet(ds *dscache.Dscache, key string, failures *uint64) {
	_, ok := ds.Get(key)
	if !ok {
		rand.Seed(time.Now().UnixNano())
		randomLength := rand.Intn(5000) + 4999
		str := tenThousandChars[0:randomLength] + "  "
		ds.Set(key, str, time.Second*60)
		atomic.AddUint64(failures, 1)
	}
}

func main() {

	ds := dscache.Custom(4*dscache.GB, 32, time.Second+time.Second/2, time.Second/2, nil)

	keyArr := generateKeys()

	var failures uint64
	numberOfOps := 100000000
	numberOfRoutines := 128
	var runOps = func(ds *dscache.Dscache, keyArr *[7311616]string, failures *uint64) {
		for i := 0; i < numberOfOps; i++ {
			key := keyArr[rand.Intn(7311616)]
			getSet(ds, key, failures)
			/*			if i%100 == 0 {
							time.Sleep(time.Second / 5)
						}
			*/
		}
	}

	for i := 0; i < numberOfRoutines; i++ {
		go runOps(ds, &keyArr, &failures)
	}

	var memStats runtime.MemStats
	for i := 0; i < 10000; i++ {
		//		ds.Verify()
		runtime.ReadMemStats(&memStats)

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
		fmt.Println("-----")
		fmt.Println("NextGC:\t\t", memStats.NextGC)
		fmt.Println("LastGC:\t\t", memStats.LastGC)
		fmt.Println("NumGC:\t\t", memStats.NumGC)

		time.Sleep(time.Second * 1)
	}

}
