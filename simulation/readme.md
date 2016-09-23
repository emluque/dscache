###dscache/simulation

Script to run simulations of dscache usage to test how it uses memory

###usage

go run simulation.go

###Flags:

  -verify boolean
    true 		verify all buckets of dscache every Second
    false 	print memory stats every second

  -keySize int
    Number of keys to be used.
      Considering each key may take a paylod from 5000 to 10000 chars,
      the number of possible keys deterimines the total size of all cacheable
      elements. Which combined with dsMaxSize (the size of the cache) will deterimine
      get failure rate.
      Maximum is 7311616

  -dsMaxSize float64
    Maximum size in GB of the cache.

  -dsLists	int
    Number of buckets in dscache.

  -dsGCSleep float64
    Seconds to wait before running GC worker in dscache.

  -dsWorkerSleep float64
  Seconds to wait before running expiration cleanup worker in each bucket.

  -numGoRoutines int
    Number of goroutines to be running get/set operations.

###results

####Test with different Cache Sizes and Default GC Time
| dsMaxSize | keySize| Payload Est. | NumObjects | GC Sleep | Sys Alloc |
| 1GB | 2m | 13GB | 141k | 1 sec | 1.94GB |
| 2GB | 2m | 13GB | 283k | 1 sec | 3.1GB |
| 4GB | 2m | 13GB | 566k | 1 sec | 5.4GB |
| 6GB | 2m | 13GB | 849k | 1 sec | 7.6GB |
| 8GB | 2m | 13GB | 1130k | 1 sec | 9.96GB |
| 10GB | 2m | 13GB | 1413k | 1 sec | 12.25GB |


####Test with 4GB Cache and different GC times
| dsMaxSize | keySize| Payload Est. | NumObjects | GC Sleep | Sys Alloc |
| 4GB | 800k | 5GB | 566k | 1 sec (default) | 5.35GB |
| 4GB | 800k | 5GB | 566k | 0.5 sec | 5.31GB |
| 4GB | 800k | 5GB | 566k | 1.5 sec | 5.5GB |
| 4GB | 800k | 5GB | 566k | 2 sec | 5.7GB |
| 4GB | 800k | 5GB | 566k | 3 sec | 6.0GB |
| 4GB | 800k | 5GB | 566k | No Forced GC | 9.44GB |

####Test with 1GB Cache and different GC times
| dsMaxSize | keySize| Payload Est. | NumObjects | GC Sleep | Sys Alloc |
| 1GB | 800k | 5GB | 141k | 1 sec (default) | 1.9GB |
| 1GB | 800k | 5GB | 141k | 0.5 sec | 1.76GB |
| 1GB | 800k | 5GB | 141k | 1.5 sec | 2.15GB |
| 1GB | 800k | 5GB | 141k | 2 sec | 2.3GB |
| 1GB | 800k | 5GB | 141k | 3 sec | 2.71GB |
| 1GB | 800k | 5GB | 141k | No Forced GC | 2.76GB |



__Note__: All Results after 1 minute of running.
