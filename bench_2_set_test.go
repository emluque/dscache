package dscache

import (
	"math/rand"
	"testing"
	"time"
)

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes 316k

	It should fit into processor cache.

*/

func BenchmarkSet1(b *testing.B) {

	var generateKeysPlusValues = func() map[string]string {
		var letters = "abcdefghijklmnopqrstuvwxyz"
		testMap := make(map[string]string)
		for i := 0; i < len(letters); i++ {
			for j := 0; j < len(letters); j++ {
				for k := 0; k < len(letters); k++ {
					var tmpKey = letters[i:i+1] + letters[j:j+1] + letters[k:k+1]
					tmpVal := "1234567890"
					testMap[tmpKey] = tmpVal
				}
			}
		}
		return testMap
	}

	rand.Seed(time.Now().UnixNano())
	ds := Custom(316*KB, 32, time.Second/2, nil)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key := range testMap {
		keyArr[c] = key
		c++
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(17576)]
		go ds.Set(key, testMap[key], time.Second*10)
	}
}

/*
	#Keys = 140608
	Payload Size = [65*3 + 8, 122*3 + 8] = [203, 374]
	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size to Fit everything in memory ~ 52 Mb
*/

func BenchmarkSet2(b *testing.B) {

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

	rand.Seed(time.Now().UnixNano())
	ds := Custom(100*MB, 32, time.Second/2, nil)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key := range testMap {
		keyArr[c] = key
		c++
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		ds.Set(key, testMap[key], time.Second*10)
	}
}
