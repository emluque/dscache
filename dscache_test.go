package dscache

import (
	"testing"
	"time"
)

func TestDscacheBasicGetSet(t *testing.T) {

	var getListNumber = func(key string) int {
		return int(key[len(key)-1]) % 32
	}
	ds := Custom(316368, 32, 0, getListNumber)

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
	ds := Custom(316368, 32, 0, getListNumber)

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
	ds := Custom(316368, 32, 0, getListNumber)

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
	ds := Custom(316368, 32, 0, getListNumber)

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
