package dscache

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestDscacheBasicGetSet(t *testing.T) {

	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316368, 32, 0, 0, getListNumber)

	ds.Set("a", "a", time.Second*10)

	tmp, _ := ds.Get("a")
	if tmp != "a" {
		t.Error("DSCache Basic Get and Set Not Working. Test 1.")
	}

	ds.Set("b", "b", time.Second*10)
	ds.Set("c", "c", time.Second*10)

	if tmp, _ = ds.Get("b"); tmp != "b" {
		t.Error("DSCache Basic Get and Set Not Working. Test 2.")
	}

	if tmp, _ = ds.Get("c"); tmp != "c" {
		t.Error("DSCache Basic Get and Set Not Working. Test 3.")
	}
}

func TestDscacheSetOfExistingElement(t *testing.T) {
	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316368, 32, 0, 0, getListNumber)

	ds.Set("d", "ddd", time.Second*10) //4 + 8
	ds.Set("c", "ccc", time.Second*10) //4 + 8
	ds.Set("b", "bbb", time.Second*10) //4 + 8
	ds.Set("a", "aaa", time.Second*10) //4 + 8

	ds.Set("d", "new", time.Second*10)

	tmp, _ := ds.Get("d")
	if tmp != "new" {
		t.Error("DSCache Set of existing element. Incorrect Set.")
	}
}

func TestDscachePurge(t *testing.T) {
	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316368, 32, 0, 0, getListNumber)

	ds.Set("d", "ddd", time.Second*10) //4 + 8
	ds.Set("c", "ccc", time.Second*10) //4 + 8
	ds.Set("b", "bbb", time.Second*10) //4 + 8
	ds.Set("a", "aaa", time.Second*10) //4 + 8

	ds.Purge("a")

	_, ok := ds.Get("a")
	if ok {
		t.Error("DSCache Purge. Not Purged.")
	}

}

func TestDscacheExpire(t *testing.T) {
	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316368, 32, 0, 0, getListNumber)

	ds.Set("d", "ddd", time.Second/5)  //12
	ds.Set("c", "ccc", time.Second*10) //12
	ds.Set("b", "bbb", time.Second*10) //12
	ds.Set("a", "aaa", time.Second*10) //12

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := ds.Get("d")
	if ok {
		t.Error("Dscache Expire. Did not expire.")
	}
}

/*
	BENCHMARKS
*/

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes 316k

	It should fit into processor cache.

*/

func Benchmark_Get_1(b *testing.B) {

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

	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316*KB, 32, 0, time.Second/2, getListNumber)

	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		ds.Set(key, val, time.Second*10)
		keyArr[c] = key
		c++
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(17576)]
		go ds.Get(key)
	}
}

/*
	#Keys = 140608
	Payload Size = [65*3 + 8, 122*3 + 8] = [203, 374]
	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size to Fit everything in memory ~ 52 Mb
*/

func Benchmark_Get_2(b *testing.B) {

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

	ds := Custom(100*MB, 32, 0, time.Second/2, nil)

	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		ds.Set(key, val, time.Second*10)
		keyArr[c] = key
		c++
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go ds.Get(key)
	}
}

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes 316k

	It should fit into processor cache.

*/

func Benchmark_Set_1(b *testing.B) {

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
	ds := Custom(316*KB, 32, 0, time.Second/2, nil)
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

func Benchmark_Set_2(b *testing.B) {

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
	ds := Custom(100*MB, 32, 0, time.Second/2, nil)
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

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes ~316k

	cache size ~210k

*/

func Benchmark_GetSet_1(b *testing.B) {

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
			ds.Set(key, testMap[key], time.Second*10)
			atomic.AddUint64(failures, 1)
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(210*KB, 32, 0, time.Second/2, nil)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		ds.Set(key, val, time.Second*10)
		keyArr[c] = key
		c++
	}

	failures := uint64(0)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go benchGetSet1(ds, key, testMap, &failures)
	}

	//	failureRate := float64(failures) / float64(b.N)
	//	fmt.Printf("Failure Rate: %.4f -- ds.size: %d \n", failureRate, ds.size)

}

/*
	#Keys = 140608
	Payload Size = 10000 + 8
	Total size of all keys ~ 1.4G

	Ascci 'A' = 65
	Ascci 'z' = 122

	Cache size: ~100 Meg

*/

func Benchmark_GetSet_2(b *testing.B) {

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
					count++
				}
			}
		}
		return keyArr
	}

	var getSet = func(ds *Dscache, key string, failures *uint64) {
		_, ok := ds.Get(key)
		if !ok {
			ds.Set(key, tenThousandChars, time.Second*10)
			*failures++
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(100*MB, 32, 0, time.Second/2, nil)
	keyArr := generateKeys()
	for i := range keyArr {
		ds.Set(keyArr[i], tenThousandChars, time.Second*10)
	}

	b.StartTimer()

	failures := uint64(0)

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go getSet(ds, key, &failures)
	}

	//	failureRate := float64(failures) / float64(b.N)
	//	fmt.Printf("Failure Rate: %.4f -- ds.size: %d \n", failureRate, ds.size)

}

/*
	#Keys = 140608
	Payload Size = 10000
	Total size of everything ~ 1.4G

	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: ~500 Meg

*/

func Benchmark_GetSet_3(b *testing.B) {

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
					count++
				}
			}
		}
		return keyArr
	}

	var getSet = func(ds *Dscache, key string) {
		_, ok := ds.Get(key)
		if !ok {
			ds.Set(key, tenThousandChars, time.Second*10)
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(500*MB, 32, 0, time.Second/2, nil)
	keyArr := generateKeys()
	for i := range keyArr {
		ds.Set(keyArr[i], tenThousandChars, time.Second*10)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go getSet(ds, key)
	}

}

/*
	#Keys = 140608
	Payload Size = 10000
	Total size of all keys ~ 1.4G

	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: 2G

*/

func Benchmark_GetSet_4(b *testing.B) {

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
					count++
				}
			}
		}
		return keyArr
	}

	var getSet = func(ds *Dscache, key string) {
		_, ok := ds.Get(key)
		if !ok {
			ds.Set(key, tenThousandChars, time.Second*10)
		}
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(2*GB, 32, 0, time.Second/2, nil)
	keyArr := generateKeys()
	for i := range keyArr {
		ds.Set(keyArr[i], tenThousandChars, time.Second*10)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go getSet(ds, key)
	}

}

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes ~316k

	cache size ~210k

*/

func Benchmark_GetSetPurge_1(b *testing.B) {

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

	count := 0
	var benchGetSet1 = func(ds *Dscache, key string, testMap map[string]string, failures *uint64) {
		if count%100 == 0 {
			ds.Purge(key)
		} else {
			_, ok := ds.Get(key)
			if !ok {
				ds.Set(key, testMap[key], time.Second*10)
				atomic.AddUint64(failures, 1)
			}
		}
		count++
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(210*KB, 32, 0, time.Second/2, nil)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		ds.Set(key, val, time.Second*10)
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
	Total size of all keys ~ 1.4G

	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: ~500 Meg

*/

func Benchmark_GetSetPurge_3(b *testing.B) {

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
					count++
				}
			}
		}
		return keyArr
	}

	count := 0
	var getSet = func(ds *Dscache, key string) {
		if count%50 == 0 {
			ds.Purge(key)
		} else {
			_, ok := ds.Get(key)
			if !ok {
				ds.Set(key, tenThousandChars, time.Second*10)
			}
		}
		count++
	}

	b.StopTimer()
	rand.Seed(time.Now().UnixNano())
	ds := Custom(500*MB, 32, 0, time.Second/2, nil)
	keyArr := generateKeys()
	for i := range keyArr {
		ds.Set(keyArr[i], tenThousandChars, time.Second*10)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		key := keyArr[rand.Intn(140608)]
		go getSet(ds, key)
	}

}

/*
	#Keys = 17576
	Payload Size = 10 + 8
	Cache size to Fit everything in memory ~ 316368 bytes ~316k

	cache size ~210k

*/

func Benchmark_GetSetExpire_1(b *testing.B) {

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
	ds := Custom(210*KB, 32, 0, time.Second/2, nil)
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
	Total size of all keys ~ 1.4G

	Ascci 'A' = 65
	Ascci 'z' = 122
	Cache size: ~500 Meg

	Expire: 0.3
*/

func Benchmark_GetSetExpire_3(b *testing.B) {

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
					count++
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
	ds := Custom(500*MB, 32, 0, time.Second/2, nil)
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
