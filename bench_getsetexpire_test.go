package dscache

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

//import "fmt"

/*
	#Keys = 140608
	Payload Size = [65*3, 122*3] = [195, 366]
	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: ~1 Meg

	Expires: 0.3 Second

*/

func BenchmarkGetSetExpire(b *testing.B) {

	var generateValue = func(strLen int) string {
		rand.Seed(time.Now().UnixNano())
		const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
		result := make([]byte, strLen)
		for i := 0; i < strLen; i++ {
			result[i] = chars[rand.Intn(len(chars))]
		}
		return string(result)
	}

	var generateKeysPlusValues = func() map[string]string {
		var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		testMap := make(map[string]string)
		for i := 0; i < len(letters); i++ {
			for j := 0; j < len(letters); j++ {
				for k := 0; k < len(letters); k++ {
					var tmpKey = letters[i:i+1] + letters[j:j+1] + letters[k:k+1]
					tmpVal := generateValue(i + j + k)
					testMap[tmpKey] = tmpVal
				}
			}
		}
		return testMap
	}

	var benchGetSet1 = func(ds *Dscache, key string, testMap map[string]string, failures *uint64) {
		_, ok := ds.Get(key)
		if !ok {
			ds.Set(key, testMap[key], time.Second*3/10)
			atomic.AddUint64(failures, 1)
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := New(1000000, time.Second/2)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		ds.Set(key, val, time.Second*3/10)
		keyArr[c] = key
		c++
	}

	failures := uint64(0)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go benchGetSet1(ds, key, testMap, &failures)
	}

	/*	failureRate := float64(failures) / float64(b.N)

		fmt.Printf("Failure Rate: %.4f -- ds.size: %d \n", failureRate, ds.size)
	*/
}

/*
	#Keys = 140608
	Payload Size = 10000
	Total size of all keys ~ 1G

	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: ~500 Meg

	Expire: 0.3
*/

func BenchmarkGetSetExpire3(b *testing.B) {

	tenChars := "012345678"
	hundredChars := tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars + tenChars
	thousandChars := hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars + hundredChars
	tenThousandChars := thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars + thousandChars

	var generateKeys = func() [140608]string {
		var letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var keyArr [140608]string
		count := 0
		for i := 0; i < len(letters); i++ {
			for j := 0; j < len(letters); j++ {
				for k := 0; k < len(letters); k++ {
					var tmpKey = letters[i:i+1] + letters[j:j+1] + letters[k:k+1]
					keyArr[count] = tmpKey
				}
			}
		}
		return keyArr
	}

	var getSet = func(ds *Dscache, key string) {
		_, ok := ds.Get(key)
		if !ok {
			ds.Set(key, tenThousandChars, time.Second*3/10)
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := New(500000000, time.Second/2)
	keyArr := generateKeys()
	for i := range keyArr {
		ds.Set(keyArr[i], tenThousandChars, time.Second*3/10)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go getSet(ds, key)
	}

}
