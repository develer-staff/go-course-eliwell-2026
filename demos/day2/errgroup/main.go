package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	// errgroup.WithContext derives a context that is automatically cancelled
	// as soon as ANY goroutine launched with g.Go returns a non-nil error.
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(4) // Concurrency limit (optional)

	devices := []struct {
		id    string
		delay time.Duration
		fail  bool
	}{
		{"device-1", 20 * time.Millisecond, false},
		{"device-2", 40 * time.Millisecond, true},   // Fails at 40ms -> cancels ctx!
		{"device-3", 200 * time.Millisecond, false}, // Aborts immediately at 40ms via ctx.Done()
	}

	results := make([]string, len(devices))

	for i, d := range devices {
		g.Go(func() error {
			status, err := fetchDeviceStatus(ctx, d.id, d.delay, d.fail)
			if err != nil {
				fmt.Printf("[%s] stopped with error: %v\n", d.id, err)
				return err
			}
			results[i] = status
			fmt.Printf("[%s] completed: %s\n", d.id, status)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("\ngroup failed fast with first error:", err)
		if errors.Is(err, context.Canceled) {
			fmt.Println("(cancelled)")
		}
		return
	}

	fmt.Println("all devices ok:", results)
}

func fetchDeviceStatus(ctx context.Context, id string, delay time.Duration, fail bool) (string, error) {
	select {
	case <-time.After(delay):
		if fail {
			return "", fmt.Errorf("device %s unreachable", id)
		}
		return fmt.Sprintf("%s: OK", id), nil
	case <-ctx.Done():
		return "", fmt.Errorf("device %s cancelled: %w", id, ctx.Err())
	}
}
