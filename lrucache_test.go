package dscache

import (
	"testing"
	"time"
)

func TestSize(t *testing.T) {
	var lru = newLRUCache(100000, 0)
	if lru.size != 0 {
		t.Error("lrucache.size not initialized in 0.")
	}

	var expectedSize uint64

	lru.set("a", "123", time.Second*10) //4
	expectedSize = (4 + 8)

	if lru.size != expectedSize {
		t.Error("lrucache.size not adding correctly. Test 1")
	}

	lru.set("bb", "12345678", time.Second*10) //+10
	expectedSize += (10 + 8)

	if lru.size != expectedSize {
		t.Error("lrucache.size not adding correctly. Test 2")
	}

	lru.set("1234567890", "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", time.Second*10) //+100
	expectedSize += (100 + 8)

	if lru.size != expectedSize {
		t.Error("lrucache.size not adding correctly. Test 3")
	}

	lru.set("b234567890", "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", time.Second*10) //+1010
	expectedSize += (1010 + 8)

	if lru.size != expectedSize {
		t.Error("lrucache.size not adding correctly. Test 4")
	}
}

func TestSizeError(t *testing.T) {
	var lru = newLRUCache(16, 0)
	err := lru.set("a", "1234567890", time.Second*10) //10 + 8
	if err != ErrMaxsize {
		t.Error("lrucache not returning an error when exceding size. Test 1")
	}
}

func TestBasicGetSet(t *testing.T) {
	var lru = newLRUCache(100000, 0)
	lru.set("a", "a", time.Second*10)

	tmp, _ := lru.get("a")
	if tmp != "a" {
		t.Error("lrucache basic get and set Not Working. Test 1.")
	}

	lru.set("b", "b", time.Second*10)
	lru.set("c", "c", time.Second*10)

	if tmp, _ = lru.get("b"); tmp != "b" {
		t.Error("lrucache basic get and set Not Working. Test 2.")
	}

	if tmp, _ = lru.get("c"); tmp != "c" {
		t.Error("lrucache basic get and set Not Working. Test 3.")
	}
}

func TestLRUOrderInsert(t *testing.T) {
	var lru = newLRUCache(100000, 0)
	lru.set("a", "a", time.Second*10)
	lru.set("b", "b", time.Second*10)
	lru.set("c", "c", time.Second*10)

	var start = lru.listStart
	if start.payload != "c" || start.next.payload != "b" || start.next.next.payload != "a" || start.next.next.next != nil {
		t.Error("LRU Order after inserts not correct. Test 1.")
	}

	var end = lru.listEnd
	if end.payload != "a" || end.previous.payload != "b" || end.previous.previous.payload != "c" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts not correct. Test 2.")
	}
}

func TestLRUOrderInsertPluSGet(t *testing.T) {
	var lru = newLRUCache(100000, 0)
	lru.set("a", "a", time.Second*10)
	lru.set("b", "b", time.Second*10)
	lru.set("c", "c", time.Second*10)
	lru.get("a")

	var start = lru.listStart
	if start.payload != "a" || start.next.payload != "c" || start.next.next.payload != "b" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus get not correct. Test 1.")
	}

	var end = lru.listEnd
	if end.payload != "b" || end.previous.payload != "c" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus get not correct. Test 2.")
	}
}

func TestLRUOrderInsertPlusVariousGets(t *testing.T) {
	var lru = newLRUCache(100000, 0)
	lru.set("a", "a", time.Second*10)
	lru.set("b", "b", time.Second*10)
	lru.set("c", "c", time.Second*10)
	lru.get("a")
	lru.get("b")
	lru.get("a")

	var start = lru.listStart
	if start.payload != "a" || start.next.payload != "b" || start.next.next.payload != "c" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various gets not correct. Test 1.")
	}

	var end = lru.listEnd
	if end.payload != "c" || end.previous.payload != "b" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus various gets not correct. Test 2.")
	}

	lru.get("a")
	start = lru.listStart
	if start.payload != "a" || start.next.payload != "b" || start.next.next.payload != "c" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various gets not correct. Test 3.")
	}

	end = lru.listEnd
	if end.payload != "c" || end.previous.payload != "b" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus various gets not correct. Test 4.")
	}

	lru.get("c")
	start = lru.listStart
	if start.payload != "c" || start.next.payload != "a" || start.next.next.payload != "b" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various gets not correct. Test 5.")
	}

	end = lru.listEnd
	if end.payload != "b" || end.previous.payload != "a" || end.previous.previous.payload != "c" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus Various gets not correct. Test 2.")
	}
}

func TestMaxsize(t *testing.T) {
	var lru = newLRUCache(48, 0)                //
	lru.set("a", "abc", time.Second*10) //4 + 8
	lru.set("b", "abc", time.Second*10) //4 + 8
	lru.set("c", "abc", time.Second*10) //4 + 8
	lru.set("d", "abc", time.Second*10) //4 + 8

	start := lru.listStart
	if start.key != "d" || start.next.key != "c" || start.next.next.key != "b" || start.next.next.next.key != "a" {
		t.Error("Maxsize. Test 1.a")
	}
	end := lru.listEnd
	if end.key != "a" || end.previous.key != "b" || end.previous.previous.key != "c" || end.previous.previous.previous.key != "d" {
		t.Error("Maxsize. Test 1.b.")
	}

	//Now must delete
	lru.set("e", "abc", time.Second*10)
	start = lru.listStart
	if start.key != "e" || start.next.key != "d" || start.next.next.key != "c" || start.next.next.next.key != "b" || start.next.next.next.next != nil {
		t.Error("Maxsize. Test 2.a")
	}
	end = lru.listEnd
	if end.key != "b" || end.previous.key != "c" || end.previous.previous.key != "d" || end.previous.previous.previous.key != "e" || end.previous.previous.previous.previous != nil {
		t.Error("Maxsize. Test 2.b.")
	}

	lru.set("f", "abc", time.Second*10)
	start = lru.listStart
	if start.key != "f" || start.next.key != "e" || start.next.next.key != "d" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("Maxsize. Test 2.a")
	}
	end = lru.listEnd
	if end.key != "c" || end.previous.key != "d" || end.previous.previous.key != "e" || end.previous.previous.previous.key != "f" || end.previous.previous.previous.previous != nil {
		t.Error("Maxsize. Test 2.b.")
	}

}

func TestLRUOrderExhaustiveTest0(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 0. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous.key != "a" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 0. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest1(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	tmp, _ := lru.get("a")
	if tmp != "aaa" {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect get.")
	}

	//Now it's a->b->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous.key != "a" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest2(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	tmp, _ := lru.get("b")
	if tmp != "bbb" {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect get.")
	}

	//Now it's b->a->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "a" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "b" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest3(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	tmp, _ := lru.get("c")
	if tmp != "ccc" {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect get.")
	}

	//Now it's c->a->b->d
	start := lru.listStart
	if start.previous != nil || start.key != "c" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "c" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest4(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	tmp, _ := lru.get("d")
	if tmp != "ddd" {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect get.")
	}

	//Now it's d->a->b->c
	start := lru.listStart
	if start.previous != nil || start.key != "d" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "d" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestSetOfExistingElement(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	lru.set("d", "new", time.Second*10)

	tmp, _ := lru.get("d")
	if tmp != "new" {
		t.Error("Set of existing element. Incorrect set.")
	}

	//Now it's d->a->b->c
	start := lru.listStart
	if start.previous != nil || start.key != "d" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("Set of existing element. Test 3. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "d" || end.previous.previous.previous.previous != nil {
		t.Error("Set of existing element. Test 3. Incorrect ListEnd.")
	}

}

func TestMaxsizeVariousSetsIncludingResets(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12*4
	lru.set("a", "aaa", time.Second*10) //12
	lru.set("b", "bbb", time.Second*10) //12
	lru.set("c", "ccc", time.Second*10) //12
	lru.set("d", "ddd", time.Second*10) //12

	lru.set("e", "eee", time.Second*10)
	lru.set("d", "ddd", time.Second*10)
	lru.set("f", "fff", time.Second*10)
	lru.set("d", "ddd", time.Second*10)
	lru.set("g", "ggg", time.Second*10)

	start := lru.listStart
	if start.previous != nil || start.key != "g" || start.next.key != "d" || start.next.next.key != "f" || start.next.next.next.key != "e" || start.next.next.next.next != nil {
		t.Error("Maxsize Various Sets Including Resets. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "e" || end.previous.key != "f" || end.previous.previous.key != "d" || end.previous.previous.previous.key != "g" || end.previous.previous.previous.previous != nil {
		t.Error("Maxsize Various Sets Including Resets. Incorrect ListEnd.")
	}

}

func TestPurgeExhaustiveTest1(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	lru.purge("a")

	//Now it's b->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("purge Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous != nil {
		t.Error("purge Exhaustive Test. Test 1. Incorrect ListEnd.")
	}
}

func TestPurgeExhaustiveTest2(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	lru.purge("b")

	//Now it's a->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("purge Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("purge Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestPurgeExhaustiveTest3(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c->d
	lru.purge("c")

	//Now it's a->b->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("purge Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("purge Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestPurgeExhaustiveTest4(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12 * 4
	lru.set("d", "ddd", time.Second*10) //4 + 8
	lru.set("c", "ccc", time.Second*10) //4 + 8
	lru.set("b", "bbb", time.Second*10) //4 + 8
	lru.set("a", "aaa", time.Second*10) //4 + 8

	//Currently it's a->b->c
	lru.purge("d")

	//Now it's a->b->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next != nil {
		t.Error("purge Exhaustive Test. Test 4. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("purge Exhaustive Test. Test 4. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest1(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12*4
	lru.set("d", "ddd", time.Second/5)  //12
	lru.set("c", "ccc", time.Second*10) //12
	lru.set("b", "bbb", time.Second*10) //12
	lru.set("a", "aaa", time.Second*10) //12

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := lru.get("d")
	if ok {
		t.Error("Expire Exhaustive Test. Test 1. Did not expire.")
	}

	//Should be a->b->c
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 1. Incorrect ListEnd.")
	}

}

func TestExpireExhaustiveTest2(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12*4
	lru.set("d", "ddd", time.Second*10) //4
	lru.set("c", "ccc", time.Second/5)  //4
	lru.set("b", "bbb", time.Second*10) //4
	lru.set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := lru.get("c")
	if ok {
		t.Error("Expire Exhaustive Test. Test 2. Did not expire.")
	}

	//Should be a->b->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest3(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12*4
	lru.set("d", "ddd", time.Second*10) //4
	lru.set("c", "ccc", time.Second*10) //4
	lru.set("b", "bbb", time.Second/5)  //4
	lru.set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := lru.get("b")
	if ok {
		t.Error("Expire Exhaustive Test. Test 3. Did not expire.")
	}

	//Should be a->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest4(t *testing.T) {
	var lru = newLRUCache(48, 0)                //12*4
	lru.set("d", "ddd", time.Second*10) //4
	lru.set("c", "ccc", time.Second*10) //4
	lru.set("b", "bbb", time.Second*10) //4
	lru.set("a", "aaa", time.Second/5)  //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := lru.get("a")
	if ok {
		t.Error("Expire Exhaustive Test. Test 4. Did not expire.")
	}

	//Should be b->c->d
	start := lru.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 4. Incorrect ListStart.")
	}
	end := lru.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 4. Incorrect ListEnd.")
	}
}
