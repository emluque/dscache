package dscache

import (
	"testing"
	"time"
)

func TestSize(t *testing.T) {
	var ds = New(100000)
	if ds.size != 0 {
		t.Error("dscache.size not initialized in 0.")
	}
	ds.Set("a", "123", time.Second*10) //4
	if ds.size != (4 * 8) {
		t.Error("dscache.size not adding correctly. Test 1")
	}
	ds.Set("bb", "12345678", time.Second*10) //+10
	if ds.size != (14 * 8) {
		t.Error("dscache.size not adding correctly. Test 2")
	}
	ds.Set("1234567890", "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", time.Second*10) //+100
	if ds.size != (114 * 8) {
		t.Error("dscache.size not adding correctly. Test 3")
	}
	ds.Set("b234567890", "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", time.Second*10) //+1010
	if ds.size != (1124 * 8) {
		t.Error("dscache.size not adding correctly. Test 4")
	}
}

func TestSizeError(t *testing.T) {
	var ds = New(80)
	err := ds.Set("a", "1234567890", time.Second*10) //88
	if err != EMaxsize {
		t.Error("dscache.not returning an error when exceding size. Test 1")
	}
}

func TestBasicGetSet(t *testing.T) {
	var ds = New(100000)
	ds.Set("a", "a", time.Second*10)

	tmp, _ := ds.Get("a")
	if tmp != "a" {
		t.Error("Basic Get and Set Not Working. Test 1.")
	}

	ds.Set("b", "b", time.Second*10)
	ds.Set("c", "c", time.Second*10)

	if tmp, _ = ds.Get("b"); tmp != "b" {
		t.Error("Basic Get and Set Not Working. Test 2.")
	}

	if tmp, _ = ds.Get("c"); tmp != "c" {
		t.Error("Basic Get and Set Not Working. Test 3.")
	}
}

func TestLRUOrderInsert(t *testing.T) {
	var ds = New(100000)
	ds.Set("a", "a", time.Second*10)
	ds.Set("b", "b", time.Second*10)
	ds.Set("c", "c", time.Second*10)

	var start = ds.listStart
	if start.payload != "c" || start.next.payload != "b" || start.next.next.payload != "a" || start.next.next.next != nil {
		t.Error("LRU Order after inserts not correct. Test 1.")
	}

	var end = ds.listEnd
	if end.payload != "a" || end.previous.payload != "b" || end.previous.previous.payload != "c" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts not correct. Test 2.")
	}
}

func TestLRUOrderInsertPlusGet(t *testing.T) {
	var ds = New(100000)
	ds.Set("a", "a", time.Second*10)
	ds.Set("b", "b", time.Second*10)
	ds.Set("c", "c", time.Second*10)
	ds.Get("a")

	var start = ds.listStart
	if start.payload != "a" || start.next.payload != "c" || start.next.next.payload != "b" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus Get not correct. Test 1.")
	}

	var end = ds.listEnd
	if end.payload != "b" || end.previous.payload != "c" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus Get not correct. Test 2.")
	}
}

func TestLRUOrderInsertPlusVariousGets(t *testing.T) {
	var ds = New(100000)
	ds.Set("a", "a", time.Second*10)
	ds.Set("b", "b", time.Second*10)
	ds.Set("c", "c", time.Second*10)
	ds.Get("a")
	ds.Get("b")
	ds.Get("a")

	var start = ds.listStart
	if start.payload != "a" || start.next.payload != "b" || start.next.next.payload != "c" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various Gets not correct. Test 1.")
	}

	var end = ds.listEnd
	if end.payload != "c" || end.previous.payload != "b" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus Various Gets not correct. Test 2.")
	}

	ds.Get("a")
	start = ds.listStart
	if start.payload != "a" || start.next.payload != "b" || start.next.next.payload != "c" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various Gets not correct. Test 3.")
	}

	end = ds.listEnd
	if end.payload != "c" || end.previous.payload != "b" || end.previous.previous.payload != "a" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus Various Gets not correct. Test 4.")
	}

	ds.Get("c")
	start = ds.listStart
	if start.payload != "c" || start.next.payload != "a" || start.next.next.payload != "b" || start.next.next.next != nil {
		t.Error("LRU Order after inserts Plus various Gets not correct. Test 5.")
	}

	end = ds.listEnd
	if end.payload != "b" || end.previous.payload != "a" || end.previous.previous.payload != "c" || end.previous.previous.previous != nil {
		t.Error("LRU Order after inserts Plus Various Gets not correct. Test 2.")
	}
}

func TestMaxsize(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("a", "abc", time.Second*10) //4
	ds.Set("b", "abc", time.Second*10) //4
	ds.Set("c", "abc", time.Second*10) //4
	ds.Set("d", "abc", time.Second*10) //4

	start := ds.listStart
	if start.key != "d" || start.next.key != "c" || start.next.next.key != "b" || start.next.next.next.key != "a" {
		t.Error("Maxsize. Test 1.a")
	}
	end := ds.listEnd
	if end.key != "a" || end.previous.key != "b" || end.previous.previous.key != "c" || end.previous.previous.previous.key != "d" {
		t.Error("Maxsize. Test 1.b.")
	}

	//Now must delete
	ds.Set("e", "abc", time.Second*10)
	start = ds.listStart
	if start.key != "e" || start.next.key != "d" || start.next.next.key != "c" || start.next.next.next.key != "b" || start.next.next.next.next != nil {
		t.Error("Maxsize. Test 2.a")
	}
	end = ds.listEnd
	if end.key != "b" || end.previous.key != "c" || end.previous.previous.key != "d" || end.previous.previous.previous.key != "e" || end.previous.previous.previous.previous != nil {
		t.Error("Maxsize. Test 2.b.")
	}

	ds.Set("f", "abc", time.Second*10)
	start = ds.listStart
	if start.key != "f" || start.next.key != "e" || start.next.next.key != "d" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("Maxsize. Test 2.a")
	}
	end = ds.listEnd
	if end.key != "c" || end.previous.key != "d" || end.previous.previous.key != "e" || end.previous.previous.previous.key != "f" || end.previous.previous.previous.previous != nil {
		t.Error("Maxsize. Test 2.b.")
	}

}

func TestLRUOrderExhaustiveTest0(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 0. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous.key != "a" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 0. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest1(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	tmp, _ := ds.Get("a")
	if tmp != "aaa" {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect Get.")
	}

	//Now it's a->b->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous.key != "a" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 1. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest2(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	tmp, _ := ds.Get("b")
	if tmp != "bbb" {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect Get.")
	}

	//Now it's b->a->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "a" || start.next.next.key != "c" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "b" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest3(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	tmp, _ := ds.Get("c")
	if tmp != "ccc" {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect Get.")
	}

	//Now it's c->a->b->d
	start := ds.listStart
	if start.previous != nil || start.key != "c" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "d" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "c" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestLRUOrderExhaustiveTest4(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	tmp, _ := ds.Get("d")
	if tmp != "ddd" {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect Get.")
	}

	//Now it's d->a->b->c
	start := ds.listStart
	if start.previous != nil || start.key != "d" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "d" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestSetOfExistingElement(t *testing.T) {

	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	ds.Set("d", "new", time.Second*10)

	tmp, _ := ds.Get("d")
	if tmp != "new" {
		t.Error("Set of existing element. Incorrect Set.")
	}

	//Now it's d->a->b->c
	start := ds.listStart
	if start.previous != nil || start.key != "d" || start.next.key != "a" || start.next.next.key != "b" || start.next.next.next.key != "c" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous.key != "d" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}

}

func TestMaxsizeVariousSetsIncludingResets(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("a", "aaa", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("d", "ddd", time.Second*10) //4

	ds.Set("e", "eee", time.Second*10)
	ds.Set("d", "ddd", time.Second*10)
	ds.Set("f", "fff", time.Second*10)
	ds.Set("d", "ddd", time.Second*10)
	ds.Set("g", "ggg", time.Second*10)

	start := ds.listStart
	if start.previous != nil || start.key != "g" || start.next.key != "d" || start.next.next.key != "f" || start.next.next.next.key != "e" || start.next.next.next.next != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "e" || end.previous.key != "f" || end.previous.previous.key != "d" || end.previous.previous.previous.key != "g" || end.previous.previous.previous.previous != nil {
		t.Error("LRU Order Exhaustive Test. Test 3. Incorrect ListEnd.")
	}

}

func TestLPurgeExhaustiveTest1(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	ds.Purge("a")

	//Now it's b->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Purge Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous != nil {
		t.Error("Purge Exhaustive Test. Test 1. Incorrect ListEnd.")
	}
}

func TestLPurgeExhaustiveTest2(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	ds.Purge("b")

	//Now it's a->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Purge Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Purge Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestLPurgeExhaustiveTest3(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	ds.Purge("c")

	//Now it's a->b->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Purge Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Purge Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestLPurgeExhaustiveTest4(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c
	ds.Purge("d")

	//Now it's a->b->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next != nil {
		t.Error("Purge Exhaustive Test. Test 4. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Purge Exhaustive Test. Test 4. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest1(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second/5)  //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := ds.Get("d")
	if ok {
		t.Error("Expire Exhaustive Test. Test 1. Did not expire.")
	}

	//Should be a->b->c
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "c" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 1. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "c" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 1. Incorrect ListEnd.")
	}

}

func TestExpireExhaustiveTest2(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second/5)  //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := ds.Get("c")
	if ok {
		t.Error("Expire Exhaustive Test. Test 2. Did not expire.")
	}

	//Should be a->b->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "b" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 2. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "b" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 2. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest3(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second/5)  //4
	ds.Set("a", "aaa", time.Second*10) //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := ds.Get("b")
	if ok {
		t.Error("Expire Exhaustive Test. Test 3. Did not expire.")
	}

	//Should be a->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "a" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 3. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "a" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 3. Incorrect ListEnd.")
	}
}

func TestExpireExhaustiveTest4(t *testing.T) {
	var ds = New(128)                  //16*8
	ds.Set("d", "ddd", time.Second*10) //4
	ds.Set("c", "ccc", time.Second*10) //4
	ds.Set("b", "bbb", time.Second*10) //4
	ds.Set("a", "aaa", time.Second/5)  //4

	//Currently it's a->b->c->d
	time.Sleep(time.Second / 2)

	_, ok := ds.Get("a")
	if ok {
		t.Error("Expire Exhaustive Test. Test 4. Did not expire.")
	}

	//Should be b->c->d
	start := ds.listStart
	if start.previous != nil || start.key != "b" || start.next.key != "c" || start.next.next.key != "d" || start.next.next.next != nil {
		t.Error("Expire Exhaustive Test. Test 4. Incorrect ListStart.")
	}
	end := ds.listEnd
	if end.next != nil || end.key != "d" || end.previous.key != "c" || end.previous.previous.key != "b" || end.previous.previous.previous != nil {
		t.Error("Expire Exhaustive Test. Test 4. Incorrect ListEnd.")
	}
}
