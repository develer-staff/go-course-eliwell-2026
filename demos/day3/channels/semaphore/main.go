// Slide "Buffered Channel as a Semaphore".
package main

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

var inFlight, maxSeen atomic.Int32

// poll simulates a slow request to one field device.
func poll(dev string) {
	n := inFlight.Add(1)
	for {
		m := maxSeen.Load()
		if n <= m || maxSeen.CompareAndSwap(m, n) {
			break
		}
	}
	fmt.Println("polling", dev)
	time.Sleep(time.Duration(20+rand.IntN(40)) * time.Millisecond)
	inFlight.Add(-1)
}

func main() {
	var devices []string
	for i := range 20 {
		devices = append(devices, fmt.Sprintf("plc-%02d", i))
	}

	start := time.Now()
	sem := make(chan struct{}, 4) // max 4 in flight

	for _, dev := range devices {
		sem <- struct{}{} // acquire: blocks when 4 are running
		go func() {
			defer func() { <-sem }() // release
			poll(dev)                // dev is per-iteration since Go 1.22
		}()
	}

	// Wait for the last ones: take every slot.
	for range cap(sem) {
		sem <- struct{}{}
	}

	fmt.Printf("polled %d devices in %v\n", len(devices), time.Since(start).Round(time.Millisecond))
	fmt.Printf("max requests in flight: %d (limit %d)\n", maxSeen.Load(), cap(sem))
	// With errors and cancellation: golang.org/x/sync/errgroup and g.SetLimit(4).
}
