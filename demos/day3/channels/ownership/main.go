// Slide "Ownership Through Channel Direction".
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"runtime"
	"time"
)

func main() {
	base := runtime.NumGoroutine()

	fmt.Println("1) finite sensor: range ends when the owner closes")
	for s := range readings(context.Background(), &Sensor{max: 3}) {
		fmt.Printf("   #%d %.1f°C\n", s.Seq, s.Celsius)
	}

	fmt.Println("2) endless sensor, consumer stops after 3 and cancels")
	ctx, cancel := context.WithCancel(context.Background())
	ch := readings(ctx, &Sensor{max: -1})
	for range 3 {
		<-ch
	}
	cancel()
	time.Sleep(20 * time.Millisecond)
	fmt.Printf("   goroutines above baseline: %d\n", runtime.NumGoroutine()-base)

	fmt.Println("3) same, without ctx: the producer leaks")
	ch2 := readingsNoCtx(&Sensor{max: -1})
	for range 3 {
		<-ch2
	}
	time.Sleep(20 * time.Millisecond)
	fmt.Printf("   goroutines above baseline: %d (leaked, blocked on send)\n", runtime.NumGoroutine()-base)

	// close(ch) // compile error: invalid operation: cannot close receive-only channel
}

type Sample struct {
	Seq     int
	Celsius float64
}

// Sensor returns max samples, then io.EOF (max < 0: never ends).
type Sensor struct{ n, max int }

func (s *Sensor) Read() (Sample, error) {
	if s.max >= 0 && s.n >= s.max {
		return Sample{}, io.EOF
	}
	s.n++
	time.Sleep(2 * time.Millisecond)
	return Sample{s.n, 20 + rand.Float64()*5}, nil
}

// readings is the code on the slide: the owner creates, sends and closes.
// The caller gets a receive-only channel: close(ch) would not compile.
func readings(ctx context.Context, s *Sensor) <-chan Sample {
	out := make(chan Sample)
	go func() {
		defer close(out)
		for {
			v, err := s.Read()
			if err != nil {
				return
			}
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// readingsNoCtx forgets the ctx.Done() case: if the consumer stops
// reading, the goroutine blocks on the send forever.
func readingsNoCtx(s *Sensor) <-chan Sample {
	out := make(chan Sample)
	go func() {
		defer close(out)
		for {
			v, err := s.Read()
			if errors.Is(err, io.EOF) {
				return
			}
			out <- v
		}
	}()
	return out
}
