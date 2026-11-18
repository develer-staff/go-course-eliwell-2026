// Slide "select keyword": random choice among ready cases, and default
// for non-blocking try-send, try-receive and polling.
package main

import (
	"fmt"
	"time"
)

func main() {
	// 1. When several cases are ready, select picks one uniformly at random.
	a, b := make(chan int, 1000), make(chan int, 1000)
	for i := range 1000 {
		a <- i
		b <- i
	}
	counts := map[string]int{}
	for range 1000 {
		select {
		case <-a:
			counts["a"]++
		case <-b:
			counts["b"]++
		}
	}
	fmt.Println("1) both always ready, 1000 selects:", counts)

	// 2. Try-send: drop the event instead of blocking when the buffer is full.
	events := make(chan string, 2)
	for _, e := range []string{"e1", "e2", "e3", "e4"} {
		select {
		case events <- e:
			fmt.Println("2) queued ", e)
		default:
			fmt.Println("2) dropped", e, "(buffer full)")
		}
	}

	// 3. Try-receive: take a value if there is one, otherwise move on.
	for range 3 {
		select {
		case e := <-events:
			fmt.Println("3) got", e)
		default:
			fmt.Println("3) nothing ready, moving on")
		}
	}

	// 4. Polling: do other work while waiting for a signal.
	// (Real code should block on select instead of sleeping in a loop.)
	done := make(chan struct{})
	go func() {
		time.Sleep(35 * time.Millisecond)
		close(done)
	}()
	polls := 0
	for {
		select {
		case <-done:
			fmt.Printf("4) done signalled after %d polls\n", polls)
			return
		default:
			polls++
			time.Sleep(10 * time.Millisecond) // "other work"
		}
	}
}
