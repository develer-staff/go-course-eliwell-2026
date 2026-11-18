package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

func record(ctx context.Context, name string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("["+name+"]", "is recording...")
		case <-ctx.Done():
			fmt.Println("["+name+"]", "uploading video...")
			time.Sleep(time.Second)
			fmt.Println("["+name+"]", "uploaded video")
			return nil
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return record(ctx, "device1") })
	g.Go(func() error { return record(ctx, "device2") })

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
