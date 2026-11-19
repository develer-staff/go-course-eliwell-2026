// Slide "nil Channels Disable select Cases".
package main

import (
	"fmt"
	"time"
)

func main() {
	run("flags:", mergeFlags)
	run("nil trick:", merge)

	in, out := make(chan int), make(chan int)
	go relay(in, out)
	go func() {
		for i := range 5 {
			in <- i // never blocks on a slow reader: relay buffers
		}
		close(in)
	}()
	time.Sleep(10 * time.Millisecond) // slow reader
	fmt.Print("relay:     ")
	for v := range out {
		fmt.Print(v, " ")
	}
	fmt.Println()
}

func run(name string, f func(a, b <-chan int, out chan<- int) int) {
	a := producer(3, time.Millisecond)    // finishes fast
	b := producer(20, 5*time.Millisecond) // keeps going for ~100ms
	out := make(chan int)
	res := make(chan int)
	go func() { res <- f(a, b, out) }()
	n := 0
	for range out {
		n++
	}
	fmt.Printf("%-10s received %d values, wake-ups on closed channels: %d\n", name, n, <-res)
}

// producer sends n values, one every delay, then closes the channel.
func producer(n int, delay time.Duration) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := range n {
			time.Sleep(delay)
			ch <- i
		}
	}()
	return ch
}

// mergeFlags tracks closed inputs with booleans. A closed channel is
// always ready, so select keeps waking up on it: a busy loop.
func mergeFlags(a, b <-chan int, out chan<- int) (wasted int) {
	defer close(out)
	aDone, bDone := false, false
	for !aDone || !bDone {
		select {
		case v, ok := <-a:
			if !ok {
				aDone = true
				wasted++
				continue
			}
			out <- v
		case v, ok := <-b:
			if !ok {
				bDone = true
				wasted++
				continue
			}
			out <- v
		}
	}
	return wasted
}

// merge is the code on the slide: a closed input is set to nil,
// and a nil channel is never selected again.
func merge(a, b <-chan int, out chan<- int) (wasted int) {
	defer close(out)
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil // never selected again
				wasted++
				continue
			}
			out <- v
		case v, ok := <-b:
			if !ok {
				b = nil
				wasted++
				continue
			}
			out <- v
		}
	}
	return wasted
}

// relay shows the same trick for back-pressure: the send case is
// enabled only when there is a value to send (send == nil otherwise).
func relay(in <-chan int, out chan<- int) {
	defer close(out)
	var pending []int
	for in != nil || len(pending) > 0 {
		var send chan<- int // nil: send case disabled
		var next int
		if len(pending) > 0 {
			send, next = out, pending[0]
		}
		select {
		case v, ok := <-in:
			if !ok {
				in = nil
				continue
			}
			pending = append(pending, v)
		case send <- next:
			pending = pending[1:]
		}
	}
}
