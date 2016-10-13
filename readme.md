#Dead Simple LRU Cache

An embeddable Key/Value in memory store for golang.

####Main Characteristics:

  - Size Limits to limit the ammount of memory usage.
  - Allows for heavy concurrent access by a huge number of Goroutines.
  - Expiration for items.
  - Strongly tested.

####Motivation

A high number of services (micro or plain old SOA) just take data from a datastore (or various), transform it to json or xml and then send it through the network. It is common practice to use a key/value in memory store (like memcached or redis) to cache the results. Using this you can avoid the network roundtrip to the kv store, which might be suitable in some cases.

####What's with the name?

The project started as an almost textbook implementation of an LRU Cache for strings. It has grown to be a little more complex since then, but the name stayed, cause it's still pretty simple.

##Usage

###Create Cache

```go
ds = dscache.New(Maxsize unit64)
```

  Where Maxsize is the Size of the cache in bytes.

####Examples
```go
// Initialize a dscache with a size of 4 GB and default options

ds = dscache.New(4 * dscache.GB)

// Initialize a dscache with a size of 200 MB and default options

ds = dscache.New(200 * dscache.MB)
```


###Set Item

```go
ds.Set(key string, value string, expire time.Duration)
```

####Examples
```go
// Expiration in one day

ds.Set("item:17897", "Json string...", 24 * time.Hour)

// Expiration in 30 minutes

ds.Set("item:17897", "Json string...", 30 * time.Minute)
```

###Get Item

```go
item, ok := ds.Get(key string)
```

####Example
```go
// Get an item from cache and verify it exists on cache

item, ok := ds.Get("item:17897")

if !ok {

  // item was not found on cache

  // fetch item from external db

}
```

###Purge Item

```go
ds.Purge(key string)
```

####Example

```go
ds.Purge("item:17897")
```

##Advanced (Custom) configuration

```go
ds := dscache.Custom(maxsize uint64, numberOfLists int, gcWorkerSleep time.Duration, workerSleep time.Duration, getListNumber func(string) int)
```

- maxsize

    Size of the cache in bytes.
- numberOfLists

    Number of internal buckets used by Dscache. To prevent serialization with concurrent accesses, Dscache splits it's keys into buckets that can be accessed independently, so that when one routine is accessing one bucket, another one could access another bucket simultaneously. The recommended number of buckets should be 4 or 8 * number of cores in your CPU. the default number is 32.
- gcWorkerSleep  

    Since Golang is garbage collected it is not possible to control the exact amount of memory that a program is using. Also, when Golang allocates some system memory it does not immediately release it to the system when not in use.

    Consider the following scenario, Dscache is being used heavily and it has filled its capacity, new sets come in and Dscache frees up its space to save the new objects. Till the moment that the Golang runtime runs a GC event, the runtime will be storing both the items on Dscache and the dereferenced old objects. This can create a situation where a lot of memory is allocated that is not actually used by Dscache.

    To prevent this, Dscache runs a worker goroutine that calls runtime.GC() and forces a Garbage collection event every _gcWorkerSleep_. This is done to minimize the ammount of uncollected garbage and the memory allocations that they imply.

    The default value is 1 Second. But it can be changed to fit your needs.

    If you don't wish to use the dscache garbage collector worker, set it to 0 and this behavior will not run. This is recommended if you are forcing a GC event in other parts of your program.
- workerSleep

  Dscache runs a worker for every bucket that iterates through the elements starting from the last used and frees them if they have expired. This runs every _workerSleep_ . To prevent this from happening (say you have very long expire times) use 0.
- getListNumber

  You can create a custom function to decide which bucket to send your items to. This will be dependent of the type of keys you are using and the number of buckets. Set to nil to use default.

####Examples
```go
// Custom dscache

ds = dscache.New(8 * dscache.GB, 128, time.Second, time.Second, nil)

// Custom dscache with special function for numerical keys. ie: "item:187896"

var splitBy100 = func (key string) {

  return int(key[len(key)-1]-48)+ ((key[len(key)-2]-48)*10)) % 100

}

ds = dscache.New(2 * dscache.GB, 100, time.Second, time.Second, splitBy100)
```


##A Note on Memory Usage

__Warning__: appart from the size of Dscache, you must also consider the amount of memory used by your program, dscache goroutines and unused garbage. Don't set Dscache to use all of your system memory. It is suggested that when you set Dscache size, that you consider at least 30% to 40% more memory for all of this. (If you have 10GB free to use by Dscache, set maxsize to 6GB).

Please look at [https://github.com/emluque/dscache/tree/master/simulation] to see actual results of variations on this.
