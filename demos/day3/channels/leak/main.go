// Slide "Finding Leaked Goroutines": the production side.
// The goroutineleak profile is on by default in Go 1.27
// (Go 1.26: GOEXPERIMENT=goroutineleakprofile go run ./channels/leak).
//
//	go run ./channels/leak
//	SHOW_LEAK=1 go test ./channels/leak  (see synctest catch the leak)
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// worker waits for one value. Nobody else holds ch, so if the
// caller never sends, this goroutine can never be woken up.
func worker(ch chan int) {
	v := <-ch
	fmt.Println("got", v)
}

// handleRequest forgets to send on the error path.
func handleRequest(fail bool) {
	ch := make(chan int)
	go worker(ch)
	if fail {
		return // leak: ch is unreachable, worker blocked forever
	}
	ch <- 42
}

func main() {
	handleRequest(false)
	for range 3 {
		handleRequest(true)
	}
	fmt.Println("goroutines running:", runtime.NumGoroutine())

	p := pprof.Lookup("goroutineleak")
	if p == nil {
		fmt.Println("goroutineleak profile not available: needs Go 1.27,")
		fmt.Println("or Go 1.26 with GOEXPERIMENT=goroutineleakprofile")
		return
	}
	// In a service: import _ "net/http/pprof", then
	// go tool pprof http://localhost:6060/debug/pprof/goroutineleak
	p.WriteTo(os.Stdout, 1)
}
