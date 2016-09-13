package dscache

//import "fmt"
import (
	"errors"
	"math/rand"
	"testing"
	"time"
)

func (lru *lrucache) verifyEndAndStart() error {

	lru.mu.Lock()
	start := lru.listStart

	if start != nil {

		//Get to last element of start
		for start.next != nil {
			start = start.next
		}

		end := lru.listEnd

		//Compare them
		for start.previous != nil {
			if end != start {
				lru.mu.Unlock()
				return errors.New("listStart does not match order of listEnd")
			}
			end = end.previous
			start = start.previous
		}

	}
	lru.mu.Unlock()

	return nil
}

func (lru *lrucache) verifyUniqueKeys() error {
	lru.mu.Lock()
	test := make(map[string]bool)
	start := lru.listStart
	for start != nil {
		_, ok := test[start.key]
		if !ok {
			test[start.key] = true
		} else {
			lru.mu.Unlock()
			return errors.New("Duplicated Key in listStart")
		}
		start = start.next
	}
	lru.mu.Unlock()
	return nil
}

func TestInGoroutines(t *testing.T) {

	var getSet = func(lru *lrucache, key string, val string) {
		_, ok := lru.get(key)
		if !ok {
			lru.set(key, val, time.Second*10)
		}
	}

	var lru = newLRUCache(128, time.Second/2)

	rand.Seed(time.Now().UnixNano())

	var letters = "abcdefghijklmno"
	for i := 0; i < 1000000; i++ {
		key := string(letters[rand.Intn(15)])
		go getSet(lru, key, "abc")
	}

	time.Sleep(5 * time.Second)
	err := lru.verifyEndAndStart()
	if err != nil {
		t.Error(err)
	}
	err = lru.verifyUniqueKeys()
	if err != nil {
		t.Error(err)
	}
}

func TestInGoroutines2(t *testing.T) {

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

	var benchGetSet = func(lru *lrucache, key string, testMap map[string]string) {
		_, ok := lru.get(key)
		if !ok {
			lru.set(key, testMap[key], time.Second*10)
		}
	}

	rand.Seed(time.Now().UnixNano())
	lru := newLRUCache(1280000, time.Second/2)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		lru.set(key, val, time.Second*10)
		keyArr[c] = key
		c++
	}

	for i := 0; i < 1000000; i++ {
		key := keyArr[rand.Intn(140608)]
		go benchGetSet(lru, key, testMap)
	}

	time.Sleep(5 * time.Second)
	err := lru.verifyEndAndStart()
	if err != nil {
		t.Error(err)
	}
	err = lru.verifyUniqueKeys()
	if err != nil {
		t.Error(err)
	}
}

/*
	Same as last but with one purge every 100 operations
*/

func TestInGoroutines3(t *testing.T) {

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
	var benchGetSet = func(lru *lrucache, key string, testMap map[string]string) {
		if count%100 == 0 {
			lru.purge(key)
		} else {
			_, ok := lru.get(key)
			if !ok {
				lru.set(key, testMap[key], time.Second*10)
			}
		}
		count++
	}

	rand.Seed(time.Now().UnixNano())
	lru := newLRUCache(1280000, time.Second/2)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		lru.set(key, val, time.Second*10)
		keyArr[c] = key
		c++
	}

	for i := 0; i < 1000000; i++ {
		key := keyArr[rand.Intn(140608)]
		go benchGetSet(lru, key, testMap)
	}

	time.Sleep(5 * time.Second)
	err := lru.verifyEndAndStart()
	if err != nil {
		t.Error(err)
	}
	err = lru.verifyUniqueKeys()
	if err != nil {
		t.Error(err)
	}
}

/*
	Expire of 1/5 of a second
*/
func TestInGoroutines4(t *testing.T) {

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

	var benchGetSet = func(lru *lrucache, key string, testMap map[string]string) {
		_, ok := lru.get(key)
		if !ok {
			lru.set(key, testMap[key], time.Second/5)
		}
	}

	rand.Seed(time.Now().UnixNano())
	lru := newLRUCache(1280000, time.Second/2)
	testMap := generateKeysPlusValues()
	var keyArr [140608]string
	c := 0
	for key, val := range testMap {
		lru.set(key, val, time.Second/5)
		keyArr[c] = key
		c++
	}

	for i := 0; i < 1000000; i++ {
		key := keyArr[rand.Intn(140608)]
		go benchGetSet(lru, key, testMap)
	}

	time.Sleep(5 * time.Second)
	err := lru.verifyEndAndStart()
	if err != nil {
		t.Error(err)
	}
	err = lru.verifyUniqueKeys()
	if err != nil {
		t.Error(err)
	}
}
