Dead Simple LRU Cache

Intention

An embeddable Key Value in memory Store that can be used to replace memcached when working on microservices. A high number of microservices just take data from a datastore (or various), transform it to json or xml and then send it through the network. It is common practice to use a key/value in memory store (like memcached or redis) to cache the results. Using this you can avoid the network roundtrip to the kv store, which might be suitable in some cases.

Now Seriously, what is it?

A Dead Simple, almost textbook implementation of an LRU Cache for strings.

It's so simple! Why are you publishing this?

Regardless of it's simplicity, based on some benchmarking I found out that this performs better than the alternatives I saw out there.

Usage

Create Cache

dscache = dscache.New(Maxsize)

where Maxsize is the number of bytes in

Set Item

dscache.Set(key, value, expire)

Get Item

item, ok := dscache.Get(key)
if !ok {
  //item was not found on cache
  //fetch item from external db

}

Purge Item

dscache.Purge(key)

NO MAN TENES QUE AGREGARLE UN EXPIRE Y UN PURGE A ESTO
