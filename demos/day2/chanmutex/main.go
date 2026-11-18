package main

import (
	"fmt"
	"sync"
)

// 1. Use sync.Mutex when multiple goroutines need to read/update shared in-memory state.
type DeviceMetrics struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewDeviceMetrics() *DeviceMetrics {
	return &DeviceMetrics{counts: make(map[string]int)}
}

func (m *DeviceMetrics) Inc(deviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[deviceID]++
}

func (m *DeviceMetrics) Snapshot() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]int, len(m.counts))
	for k, v := range m.counts {
		out[k] = v
	}
	return out
}

func main() {
	metrics := NewDeviceMetrics()

	// 2. Use a Channel to distribute work / transfer data ownership across pipeline stages.
	events := make(chan string, 10)

	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			for devID := range events {
				metrics.Inc(devID)
			}
		})
	}

	for _, id := range []string{"dev-1", "dev-2", "dev-1", "dev-3", "dev-1", "dev-2"} {
		events <- id
	}
	close(events)

	wg.Wait()
	fmt.Println("aggregated metrics:", metrics.Snapshot())
}
