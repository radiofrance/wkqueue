package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/inhies/go-bytesize"
)

type BigObject [1000 * 1000]byte

func main() {
	ch := make(chan interface{}, 2000)
	// producer
	go func() {
		for {
			ch <- BigObject{}
		}
	}()

	// consumer
	var baseMem runtime.MemStats
	var deltaMax float64
	runtime.ReadMemStats(&baseMem)
	go func() {
		var currMem runtime.MemStats
		for {
			runtime.ReadMemStats(&currMem)

			inChan := int64(len(ch) * len(BigObject{}))
			inMemory := int64(currMem.HeapInuse - baseMem.HeapInuse)
			delta := float64(inMemory) / float64(inChan)

			if len(ch) == cap(ch) && delta > deltaMax {
				deltaMax = delta
			}

			fmt.Printf(
				"In Chan: %s | In Memory: %s | Delta max: %0.3f%%\n",
				bytesize.ByteSize(inChan).String(),
				bytesize.ByteSize(inMemory).String(),
				deltaMax*100,
			)
		}
	}()

	panic(http.ListenAndServe(":6060", nil))
}
