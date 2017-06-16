# DSCache

An embeddable Key/Value in memory store for golang.

#### Main Characteristics:

  - Size Limits to limit the ammount of memory usage.
  - Allows for concurrent access by different Goroutines.
  - LRU Eviction and Expiration for items.
  - Strongly tested.

#### Motivation

A high number of services (micro or plain old SOA) just take data from a datastore (or various), transform it to json or xml and then send it through the network. It is common practice to use a key/value in memory store (like memcached or redis) to cache the results. Using this you can avoid the network roundtrip to the kv store, which might be suitable in some cases.

#### What's the Use Case for this?

In general the use of a cache is recommended when values are expensive to compute or to retrive, thus making it better to store them for later use. So it makes sense when you are willing to spend some memory to improve speed if you are expecting keys will get queried more than once.

The Use of DSCache is particularly suited when:
  . You can fit all of your items in memory and you need time based expiration for the items.
  . When you have a small number of keys which are queried disproportionally more than the rest of the set and that you can fit into your box's memory. (Ie: 10% of keys get 50% of requests so it makes sense to cache them locally like you would with Memcached or Redis).
  . As part of a Hierarchy of Caches, again if your request distribution is very uneven, you could have the more common keys on DSCache, and the rest on an external store like Memcached). The reason for this is that a retrieval from DSCache is on th order of the nanoseconds where even on a local version of Memcached you will be considering access on the order or Milliseconds because of the time spent context switching by the OS and the cost of networking. 


## A Note on Memory Usage

__Warning__: appart from the size of Dscache, you must also consider the amount of memory used by your program, dscache goroutines and unused garbage. Don't set Dscache to use all of your system memory. It is suggested that when you set Dscache size, that you consider at least 30% to 40% more memory for all of this. (If you have 10GB free to use by Dscache, set maxsize to 6GB).

Please look at [https://github.com/emluque/dscache/tree/master/simulation] to see actual results of variations on this.

## Usage

### Create Cache

```go
ds = dscache.New(Maxsize unit64)
```

  Where Maxsize is the Size of the cache in bytes.

#### Examples
```go
// Initialize a dscache with a size of 4 GB and default options

ds = dscache.New(4 * dscache.GB)

// Initialize a dscache with a size of 200 MB and default options

ds = dscache.New(200 * dscache.MB)
```


### Set Item

```go
ds.Set(key string, value string, expire time.Duration)
```

#### Examples
```go
// Expiration in one day

ds.Set("item:17897", "Json string...", 24 * time.Hour)

// Expiration in 30 minutes

ds.Set("item:17897", "Json string...", 30 * time.Minute)
```

### Get Item

```go
item, ok := ds.Get(key string)
```

#### Example
```go
// Get an item from cache and verify it exists on cache

item, ok := ds.Get("item:17897")

if !ok {

  // item was not found on cache

  // fetch item from external db

}
```

### Purge Item

```go
ds.Purge(key string)
```

#### Example

```go
ds.Purge("item:17897")
```

## Advanced (Custom) configuration

```go
ds := dscache.Custom(maxsize uint64, numberOfLists int, gcWorkerSleep time.Duration, workerSleep time.Duration, getListNumber func(string) int)
```

- maxsize

    Size of the cache in bytes.
- numberOfBuckets

    Number of internal buckets used by Dscache. To prevent serialization with concurrent accesses, Dscache splits it's keys into buckets that can be accessed independently, so that when one routine is accessing one bucket, another one could access another bucket simultaneously. The recommended number of buckets should be 4 or 8 * number of cores in your CPU. the default number is 32.
- gcWorkerSleep  

    Since Golang is garbage collected it is not possible to control the exact amount of memory that a program is using. Also, when Golang allocates some system memory it does not immediately release it to the system when not in use.

    Consider the following scenario, Dscache is being used heavily and it has filled its capacity, new sets come in and Dscache frees up its space to save the new objects. Till the moment that the Golang runtime runs a GC event, the runtime will be storing both the items on Dscache and the dereferenced old objects. This can create a situation where a lot of memory is allocated that is not actually used by Dscache.

    To prevent this, Dscache runs a worker goroutine that calls runtime.GC() and forces a Garbage collection event every _gcWorkerSleep_. This is done to minimize the ammount of uncollected garbage and the memory allocations that they imply.

    The default value is 1 Second. But it can be changed to fit your needs.

    If you don't wish to use the dscache garbage collector worker, set it to 0 and this behavior will not run. This is recommended if you are forcing a GC event in other parts of your program.
- workerSleep

  Dscache runs a worker for every bucket that iterates through the elements starting from the last used and frees them if they have expired. This runs every _workerSleep_ . To prevent this from happening (say you have very long expire times) use 0.
- getBucketNumber

  You can create a custom function to decide which bucket to send your items to. This will be dependent of the type of keys you are using and the number of buckets. Set to nil to use default. Default hashing function is: http://www.partow.net/programming/hashfunctions/index.html#BKDRHashFunction
 
#### Examples
```go
// Custom dscache

ds = dscache.New(8 * dscache.GB, 128, time.Second, time.Second, nil)

// Custom dscache with special function for numerical keys. ie: "item:187896"
var numericFormat = func (key string) {
  index := strings.LastIndex(key, ":")
  numericString := key[index, len(key)]

  num, _ := strconv.Atoi(numericString)
  return num % 256

}

ds = dscache.New(2 * dscache.GB, 256, time.Second, time.Second, numericFormat)
```

## Statistics
```go
// Number of Objects currently stored on the Cache
numObjects := ds.NumObjects()

// Current Hit Rate (Number of succesful gets / total number of requests)
hitRate := ds.HitRate()

//Number of Total Evictions
numEvictions := ds.NumEvictions()


```

