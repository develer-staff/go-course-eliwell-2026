package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel() // ensure the resources are always cleaned up (internal timer)

	// Fast sensor succeeds within the 100ms timeout.
	if temp, err := pollSensor(ctx, "freezer-01", 30*time.Millisecond); err != nil {
		fmt.Println("freezer-01 error:", err)
	} else {
		fmt.Printf("freezer-01 temperature: %.1f°C\n", temp)
	}

	// Slow sensor exceeds the remaining deadline and exits cleanly via <-ctx.Done().
	if temp, err := pollSensor(ctx, "freezer-02", 200*time.Millisecond); err != nil {
		fmt.Println("freezer-02 error:", err)
	} else {
		fmt.Printf("freezer-02 temperature: %.1f°C\n", temp)
	}
}

// pollSensor simulates polling a remote sensor until a reading arrives or ctx
// is cancelled.
func pollSensor(ctx context.Context, deviceID string, latency time.Duration) (float64, error) {
	select {
	case <-time.After(latency):
		return -18.4, nil
	case <-ctx.Done():
		// Always return ctx.Err() (context.Canceled or context.DeadlineExceeded)
		return 0, fmt.Errorf("poll %s aborted: %w", deviceID, ctx.Err())
	}
}
