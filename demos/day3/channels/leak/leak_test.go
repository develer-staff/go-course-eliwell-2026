package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"testing/synctest"
	"time"
)

// waitValue is the fixed worker: it gives up when ctx expires.
func waitValue(ctx context.Context, ch <-chan int) (int, error) {
	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// A one-minute timeout, tested in microseconds: inside the bubble the
// clock only moves when every goroutine is blocked.
func TestWaitValueTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
		defer cancel()

		errc := make(chan error, 1)
		go func() {
			_, err := waitValue(ctx, make(chan int))
			errc <- err
		}()

		time.Sleep(time.Minute + time.Second) // fake clock: returns at once
		synctest.Wait()                       // until the goroutine is blocked or done

		select {
		case err := <-errc:
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("got %v, want DeadlineExceeded", err)
			}
		default:
			t.Fatal("waitValue still blocked after its timeout")
		}
	})
}

// The leaky worker: synctest fails the test because a goroutine
// is still blocked when the bubble ends.
func TestLeakyWorker(t *testing.T) {
	if os.Getenv("SHOW_LEAK") == "" {
		t.Skip("run with SHOW_LEAK=1 to see synctest report the leak")
	}
	synctest.Test(t, func(t *testing.T) {
		go worker(make(chan int))
	})
}
