package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- string) {
	for j := range jobs {
		time.Sleep(20 * time.Millisecond) // simulate work
		results <- fmt.Sprintf("worker %d processed job %d", id, j)
	}
}

func main() {
	const (
		numWorkers = 3
		numJobs    = 7
	)

	jobs := make(chan int, numJobs)
	results := make(chan string, numJobs)

	var wg sync.WaitGroup

	// Spawn a fixed number of workers (bounded concurrency).
	for w := 1; w <= numWorkers; w++ {
		workerID := w
		// Go 1.25+: wg.Go automatically calls wg.Add(1) before spawning
		// and defer wg.Done() when the goroutine finishes.
		wg.Go(func() {
			worker(workerID, jobs, results)
		})
	}

	// Enqueue all jobs and close the jobs channel so workers exit their range loop.
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Wait for all workers to finish, then close results.
	wg.Wait()
	close(results)

	for res := range results {
		fmt.Println(res)
	}
}
