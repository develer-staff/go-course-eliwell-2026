// Slide "Request/Response: a Channel of Channels".
// The map is owned by one goroutine: no mutex, and go run -race stays quiet.
//
//	Run with -race
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	gets, incs := make(chan getReq), make(chan incReq)
	go cache(gets, incs)
	ctx := context.Background()

	// 1. 10 goroutines x 100 increments, no lock.
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			for range 100 {
				incs <- incReq{"alarms", 1}
			}
		})
	}
	wg.Wait()

	v, _ := get(ctx, gets, "alarms")
	fmt.Println("1) alarms =", v, "(expected 1000)")

	// 2. A caller that times out does not block the owner.
	tctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	_, err := get(tctx, gets, "slow")
	cancel()
	fmt.Println("2) slow get:", err)

	v, err = get(ctx, gets, "alarms")
	fmt.Println("3) owner still serving: alarms =", v, err)
}

func get(ctx context.Context, gets chan<- getReq, key string) (int, error) {
	reply := make(chan int, 1) // buffer 1: the owner can always deliver
	select {
	case gets <- getReq{key, reply}:
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	select {
	case v := <-reply:
		return v, nil
	case <-ctx.Done():
		return 0, ctx.Err() // caller gives up; the owner is not stuck
	}
}

type getReq struct {
	key   string
	reply chan<- int // the response channel travels inside the request
}

type incReq struct {
	key   string
	delta int
}

// cache owns the map: only this goroutine reads or writes it.
func cache(gets <-chan getReq, incs <-chan incReq) {
	m := map[string]int{}
	for {
		select {
		case r := <-gets:
			if r.key == "slow" {
				time.Sleep(50 * time.Millisecond) // e.g. a device round trip
			}
			r.reply <- m[r.key] // buffered reply: never blocks, even if the caller left
		case r := <-incs:
			m[r.key] += r.delta
		}
	}
}
