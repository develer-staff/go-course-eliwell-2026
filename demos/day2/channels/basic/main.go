package main

import "fmt"

func main() {
	fmt.Println("--- 1. Buffered Channel & Range ---")
	ch := make(chan int, 3)

	// Producer sends 3 values and closes the channel.
	// Golden Rule: ONLY the sender closes a channel to signal "no more values".
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i * 10
		}
		close(ch)
	}()

	// for-range drains the channel until it is closed and empty.
	for v := range ch {
		fmt.Println("received via range:", v)
	}

	/* ----------------------------------------------- */
	fmt.Println("\n--- 2. Receiving from a Closed Channel ---")
	// Receiving from a closed channel NEVER blocks: it immediately returns
	// the zero value of T and ok == false.
	val, ok := <-ch
	fmt.Printf("after close: val=%d, ok=%v\n", val, ok)

	/* ----------------------------------------------- */
	fmt.Println("\n--- 3. Nil Channel Behavior ---")
	var nilCh chan int // zero value of a channel is nil
	select {
	case <-nilCh:
		fmt.Println("never happens (receive on nil channel blocks forever)")
	default:
		fmt.Println("nil channel is not ready (blocks forever without default)")
	}
}
