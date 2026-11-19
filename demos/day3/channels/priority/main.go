// Slide "Priority, Timeouts and Tickers".
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("1) 5 control messages + 1 tick ready at the same time, 1000 runs")
	for _, p := range []bool{false, true} {
		late := 0
		for range 1000 {
			if trial(p) < 5 {
				late++
			}
		}
		fmt.Printf("   priority=%-5v publish ran before all commands: %d/1000\n", p, late)
	}

	fmt.Println("2) the loop on the slide, with a real ticker and a deadline")
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	ctrl := make(chan string, 10)
	go func() {
		for i := range 4 {
			time.Sleep(60 * time.Millisecond)
			ctrl <- fmt.Sprintf("setpoint=%d", 20+i)
		}
	}()
	fmt.Println("   loop ended:", loop(ctx, ctrl))
}

// trial preloads 5 control messages and one tick, then counts how many
// control messages were applied before the first publish.
func trial(priority bool) int {
	ctrl := make(chan string, 5)
	for i := range 5 {
		ctrl <- fmt.Sprint("cmd", i)
	}
	tick := make(chan struct{}, 1)
	tick <- struct{}{}

	applied := 0
	for {
		if priority {
			select {
			case <-ctrl:
				applied++
				continue
			default:
			}
		}
		select {
		case <-ctrl:
			applied++
		case <-tick:
			return applied // first publish
		}
	}
}

func loop(ctx context.Context, ctrl <-chan string) error {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	state := "setpoint=19"
	start := time.Now()
	for {
		// priority: drain control messages first
		select {
		case c := <-ctrl:
			state = c
			fmt.Printf("   %3dms apply   %s\n", time.Since(start).Milliseconds(), c)
			continue
		default:
		}
		select {
		case c := <-ctrl:
			state = c
			fmt.Printf("   %3dms apply   %s\n", time.Since(start).Milliseconds(), c)
		case <-t.C:
			fmt.Printf("   %3dms publish %s\n", time.Since(start).Milliseconds(), state)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
