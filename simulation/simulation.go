package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"../"
)

const tenChars = "0123456789"
const hundredChars = tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars
const thousandChars = hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars
const tenThousandChars = thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars

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

func runOps(ds *dscache.Dscache, keySize int, keyArr *[7311616]string) {
	for {
		key := keyArr[rand.Intn(keySize)]
		getSet(ds, key)
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

	printConf(*verify, *keySize, *dsMaxSize, *dsLists, *dsGCSleep, *dsWorkerSleep, *numGoRoutines)

	ds := dscache.Custom(uint64(*dsMaxSize*float64(dscache.GB)), *dsLists, time.Duration(float64(time.Second)**dsGCSleep), time.Duration(float64(time.Second)**dsWorkerSleep), nil)

	keyArr := generateKeys()

	for i := 0; i < *numGoRoutines; i++ {
		go runOps(ds, *keySize, &keyArr)
	}

	var i int
	var memStats runtime.MemStats

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		printExit(i, &memStats, ds)
		printConf(*verify, *keySize, *dsMaxSize, *dsLists, *dsGCSleep, *dsWorkerSleep, *numGoRoutines)
		os.Exit(1)
	}()

	for i = 0; i < 10000; i++ {
		if *verify {
			ds.Verify()
		} else {
			printStats(&memStats, ds)
		}
		time.Sleep(time.Second * 1)
	}

}

func printConf(verify bool, keySize int, dsMaxSize float64, dsLists int, dsGCSleep float64, dsWorkerSleep float64, numGoRoutines int) {
	fmt.Println("--------------------------------------------")
	fmt.Println("Verify:\t\t\t\t", verify)
	fmt.Println("-----")
	fmt.Println("keySize:\t\t\t", keySize)
	fmt.Printf("Payload Total:\t\t\t(%dGB, %dGB)\n", keySize*5000/dscache.GB, keySize*10000/dscache.GB)
	fmt.Println("-----")
	fmt.Println("ds.MaxSize:\t\t\t", dsMaxSize, "GB")
	fmt.Println("ds.Lists:\t\t\t", dsLists)
	fmt.Println("ds.GCSleep:\t\t\t", dsGCSleep)
	fmt.Println("ds.Workerleep:\t\t\t", dsWorkerSleep)
	fmt.Println("-----")
	fmt.Println("NumGoRoutines:\t\t\t", numGoRoutines)
	fmt.Println()
}

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
