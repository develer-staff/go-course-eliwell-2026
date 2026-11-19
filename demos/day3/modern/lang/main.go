// Slide "Language Changes", Go 1.22 to 1.27. Needs Go 1.27.
package main

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"
)

// 1.24: generic type aliases.
type Set[T comparable] = map[T]struct{}

// 1.26: new(expr) for optional pointer fields.
type Config struct {
	Timeout *time.Duration // nil = default
	Verbose *bool
}

// 1.27: generic methods. Cache is not generic; Get is.
type Cache struct {
	mu sync.Mutex
	m  map[string]any
}

func (c *Cache) Put(k string, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.m == nil {
		c.m = map[string]any{}
	}
	c.m[k] = v
}

func (c *Cache) Get[T any](k string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.m[k].(T)
	return v, ok
}

// 1.23: a custom iterator.
func Countdown(n int) func(yield func(int) bool) {
	return func(yield func(int) bool) {
		for i := n; i > 0; i-- {
			if !yield(i) {
				return
			}
		}
	}
}

func main() {
	// 1.22: range over int, one variable per iteration.
	ch := make(chan int, 10)
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			ch <- i // safe now: no i := i
		})
	}
	wg.Wait()

	close(ch)
	got := slices.Sorted(func(yield func(int) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	})
	fmt.Println("1.22 range over int:   ", got)

	// 1.23: range over functions.
	m := map[string]int{"temp": 21, "hum": 48}
	for k, v := range maps.All(m) {
		fmt.Printf("1.23 maps.All:          %s=%d\n", k, v)
	}
	fmt.Print("1.23 custom iterator:   ")
	for i := range Countdown(3) {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 1.24: generic type alias, interchangeable with the map type.
	seen := Set[string]{}
	seen["plc-01"] = struct{}{}
	var raw map[string]struct{} = seen // same type, no conversion
	fmt.Println("1.24 generic alias:    ", len(raw), "device(s) seen")

	// 1.26: new with an initial value.
	cfg := Config{Timeout: new(30 * time.Second), Verbose: new(true)}
	fmt.Println("1.26 new(expr):        ", *cfg.Timeout, *cfg.Verbose)

	// 1.27: generic methods.
	var c Cache
	c.Put("setpoint", 4.5)
	c.Put("label", "freezer")
	sp, ok := c.Get[float64]("setpoint")
	fmt.Println("1.27 generic method:   ", sp, ok)
	_, ok = c.Get[int]("label")
	fmt.Println("1.27 wrong type:       ", ok)
}
