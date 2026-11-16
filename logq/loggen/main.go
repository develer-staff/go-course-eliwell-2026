// Command loggen writes fake log lines, to try logq on big files.
//
//	go run ./loggen -n 1000000 > big.log
//
// The output depends only on -seed and -n: it is the same on every machine.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand" // not math/rand/v2: this generator's sequence is guaranteed never to change
	"os"
	"time"
)

var (
	levels   = []string{"DEBUG", "INFO", "INFO", "INFO", "WARN", "ERROR"} // INFO is the most common
	services = []string{"auth", "billing", "gateway", "payments", "search"}
	messages = map[string][]string{
		"DEBUG": {"cache miss for key=user:%d", "connection pool size=%d"},
		"INFO":  {"request completed status=200 id=r-%05d", "job %d scheduled", "user %d logged in"},
		"WARN":  {"slow query took %dms", "Timeout approaching for request r-%05d"},
		"ERROR": {"timeout calling bank gateway id=r-%05d", "connection refused to db-%d", "Timeout waiting for upstream id=r-%05d"},
	}
)

func main() {
	n := flag.Int("n", 10000, "number of lines")
	seed := flag.Int64("seed", 1, "random seed")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	t := time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC)
	for i := range *n {
		if i%1000 == 999 {
			fmt.Fprintf(out, "--- logrotate: rotated at line %d ---\n", i+1)
			continue
		}
		t = t.Add(time.Duration(r.Intn(3000)) * time.Millisecond).Truncate(time.Second)
		level := levels[r.Intn(len(levels))]
		msgs := messages[level]
		msg := fmt.Sprintf(msgs[r.Intn(len(msgs))], r.Intn(100000))
		fmt.Fprintf(out, "%s %s %s %s\n", t.Format(time.RFC3339), level, services[r.Intn(len(services))], msg)
	}
}
