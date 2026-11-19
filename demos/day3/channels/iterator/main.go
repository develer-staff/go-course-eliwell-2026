// Slide "Channel or Iterator?"
package main

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"runtime"
	"strings"
	"time"
)

// Before: a goroutine and a channel per stream.
func linesCh(r io.Reader) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			ch <- sc.Text() // blocks forever if the caller breaks out
		}
	}()
	return ch
}

// Go 1.23+: range over func, no goroutine.
func Lines(r io.Reader) iter.Seq[string] {
	return func(yield func(string) bool) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			if !yield(sc.Text()) {
				return // caller broke out
			}
		}
	}
}

const log = "boot\nconnect\nEND\nshould not be read\n"

func main() {
	base := runtime.NumGoroutine()

	fmt.Println("1) channel version, break at END")
	for line := range linesCh(strings.NewReader(log)) {
		if line == "END" {
			break
		}
		fmt.Println("  ", line)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("   goroutines left behind: %d\n", runtime.NumGoroutine()-base)

	base = runtime.NumGoroutine()
	fmt.Println("2) iterator version, break at END")
	for line := range Lines(strings.NewReader(log)) {
		if line == "END" {
			break // yield returns false, Lines returns
		}
		fmt.Println("  ", line)
	}
	fmt.Printf("   goroutines left behind: %d\n", runtime.NumGoroutine()-base)

	fmt.Println("3) 200,000 lines")
	big := strings.Repeat("sample line\n", 200_000)
	t0 := time.Now()
	n := 0
	for line := range linesCh(strings.NewReader(big)) {
		n += len(line)
	}
	tCh := time.Since(t0)
	t0 = time.Now()
	n = 0
	for line := range Lines(strings.NewReader(big)) {
		n += len(line)
	}
	tIt := time.Since(t0)
	fmt.Printf("   channel %v, iterator %v (%.0fx faster)\n",
		tCh.Round(time.Millisecond), tIt.Round(time.Millisecond), float64(tCh)/float64(tIt))

	fmt.Println("4) standard library iterators (Go 1.24)")
	for l := range strings.Lines("a\nb\n") {
		fmt.Printf("   strings.Lines    %q\n", l)
	}
	for f := range strings.SplitSeq("temp,hum,press", ",") {
		fmt.Printf("   strings.SplitSeq %q\n", f)
	}
}
