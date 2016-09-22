package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"../"
)

var tenChars = "0123456789"
var hundredChars = tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars
var thousandChars = hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars
var tenThousandChars = thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars

/*
  Number of Keys: 7311616
	All Payloads Size: [35, 70] GB
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

func getSet(ds *dscache.Dscache, key string) {
	_, ok := ds.Get(key)
	if !ok {
		rand.Seed(time.Now().UnixNano())
		randomLength := rand.Intn(5000) + 4999
		str := tenThousandChars[0:randomLength] + "  "
		ds.Set(key, str, time.Second*60)
	}
}

func main() {

	verify := flag.Bool("verify", false, "Wether to run on Verify or Simulation Mode.")
	keySize := flag.Int("keySize", 7311616, "Number of Keys to use in testing.")
	dsMaxSize := flag.Float64("dsMaxSize", 4.0, "ds Maxsize, in GB, may take floats.")
	dsLists := flag.Int("dsLists", 32, "ds Number Of Lists.")
	dsGCSleep := flag.Float64("dsGCSleep", 1.0, "ds GC Sleep, in Seconds, may take floats.")
	dsWorkerSleep := flag.Float64("dsWorkerSleep", 0.5, "ds Worker Sleep, in Seconds, may take floats.")
	numGoRoutines := flag.Int("numGoRoutines", 64, "Number of Goroutines to be accessing the cache simultaneously.")

	flag.Parse()

	ds := dscache.Custom(uint64(*dsMaxSize*float64(dscache.GB)), *dsLists, time.Duration(float64(time.Second)**dsGCSleep), time.Duration(float64(time.Second)**dsWorkerSleep), nil)

	keyArr := generateKeys()

	var runOps = func(ds *dscache.Dscache, keyArr *[7311616]string) {
		for {
			key := keyArr[rand.Intn(*keySize)]
			getSet(ds, key)
		}
	}

	for i := 0; i < *numGoRoutines; i++ {
		go runOps(ds, &keyArr)
	}

	var memStats runtime.MemStats

	for i := 0; i < 10000; i++ {
		if *verify {
			ds.Verify()
		} else {
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
			fmt.Printf("ds.FailureRate:\t\t%.3f\n", ds.FailureRate())
			fmt.Println("-----")
			fmt.Println("NextGC:\t\t", memStats.NextGC)
			fmt.Println("LastGC:\t\t", memStats.LastGC)
			fmt.Println("NumGC:\t\t", memStats.NumGC)
		}
		time.Sleep(time.Second * 1)
	}

}
