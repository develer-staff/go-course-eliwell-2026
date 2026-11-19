// Slide "close() as a Broadcast".
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("1) close(quit) wakes up every worker")
	s := newServer()
	jobs := make(chan Job)
	for id := range 4 {
		s.wg.Go(func() {
			s.worker(id, jobs)
		})
	}
	for j := range 6 {
		jobs <- Job(j)
	}
	s.Stop()
	s.Stop() // no "close of closed channel" panic thanks to sync.Once
	s.wg.Wait()

	fmt.Println("2) context.WithCancel: the same broadcast, built in")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for id := range 3 {
		wg.Go(func() {
			<-ctx.Done() // Done() returns a channel that is closed on cancel

			fmt.Printf("  listener %d: %v\n", id, ctx.Err())
		})
	}
	time.Sleep(10 * time.Millisecond)
	cancel()
	cancel() // also safe to call twice
	wg.Wait()
}

type Job int

func (j Job) Run(worker int) {
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("  worker %d ran job %d\n", worker, j)
}

type Server struct {
	quit chan struct{}
	once sync.Once // or: stop := sync.OnceFunc(func() { close(quit) })
	wg   sync.WaitGroup
}

func newServer() *Server { return &Server{quit: make(chan struct{})} }

// Stop can be called any number of times: close runs once.
func (s *Server) Stop() {
	s.once.Do(func() { close(s.quit) })
}

func (s *Server) worker(id int, jobs <-chan Job) {
	for {
		select {
		case <-s.quit: // every worker sees the close at once
			fmt.Printf("  worker %d: quit\n", id)
			return
		case j := <-jobs:
			j.Run(id)
		}
	}
}
