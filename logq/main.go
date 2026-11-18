// Command logq prints the lines of log files that contain a word.
//
//	logq [flags] WORD [FILE...]
package main

import (
	"bufio"
	"cmp"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
)

type options struct {
	word       string
	ignoreCase bool
	minLevel   Level
	service    string
	count      bool
	json       bool
	stats      bool
}

func main() {
	opts, files := parseFlags()

	out := bufio.NewWriter(os.Stdout)
	s := &searcher{
		opts:     opts,
		out:      out,
		json:     json.NewEncoder(out),
		levels:   map[string]int{},
		services: map[string]int{},
	}

	failed := false
	if len(files) == 0 {
		if err := s.search(os.Stdin, "-"); err != nil {
			fmt.Fprintln(os.Stderr, "logq:", err)
			failed = true
		}
	}
	for _, name := range files {
		if err := s.searchFile(name); err != nil {
			fmt.Fprintln(os.Stderr, "logq:", err)
			failed = true
		}
	}
	s.printSummary()

	out.Flush() // not deferred: os.Exit does not run deferred calls
	if failed {
		os.Exit(1)
	}
}

func parseFlags() (options, []string) {
	var opts options
	var level string
	flag.BoolVar(&opts.ignoreCase, "i", false, "ignore case when looking for WORD")
	flag.StringVar(&level, "level", "", "only entries at this `level` or above: debug, info, warn or error")
	flag.StringVar(&opts.service, "service", "", "only entries of this `service`")
	flag.BoolVar(&opts.count, "count", false, "print the number of matching entries instead of the entries")
	flag.BoolVar(&opts.json, "json", false, "print the entries as JSON, one per line")
	flag.BoolVar(&opts.stats, "stats", false, "print the number of matching entries per level and per service")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: logq [flags] WORD [FILE...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	if level != "" {
		l, err := ParseLevel(level)
		if err != nil {
			fmt.Fprintln(os.Stderr, "logq:", err)
			os.Exit(1)
		}
		opts.minLevel = l
	}
	opts.word = flag.Arg(0)
	if opts.ignoreCase {
		opts.word = strings.ToLower(opts.word)
	}
	return opts, flag.Args()[1:]
}

func (o options) match(e Entry) bool {
	text := e.Raw
	if o.ignoreCase {
		text = strings.ToLower(text)
	}
	return strings.Contains(text, o.word) &&
		e.Level >= o.minLevel &&
		(o.service == "" || e.Service == o.service)
}

type searcher struct {
	opts     options
	out      io.Writer
	json     *json.Encoder
	matches  int
	levels   map[string]int
	services map[string]int
}

func (s *searcher) searchFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	var r io.Reader = f
	if strings.HasSuffix(name, ".gz") {
		zr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		defer zr.Close()
		r = zr
	}
	return s.search(r, name)
}

// search reads r line by line. Any io.Reader works: a file, stdin, a gzip
// stream...
func (s *searcher) search(r io.Reader, name string) error {
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		e := parse(sc.Text())
		e.File, e.Line = name, n
		if s.opts.match(e) {
			s.found(e)
		}
	}
	return sc.Err()
}

func (s *searcher) found(e Entry) {
	s.matches++
	s.levels[e.Level.String()]++
	s.services[cmp.Or(e.Service, "-")]++

	switch {
	case s.opts.count, s.opts.stats:
		// printed at the end
	case s.opts.json:
		s.json.Encode(e)
	default:
		fmt.Fprintln(s.out, e.Raw)
	}
}

func (s *searcher) printSummary() {
	switch {
	case s.opts.stats:
		printCounts(s.out, "levels", s.levels)
		printCounts(s.out, "services", s.services)
	case s.opts.count:
		fmt.Fprintln(s.out, s.matches)
	}
}

// printCounts prints the counts from the highest to the lowest, and in
// alphabetical order for equal counts.
func printCounts(w io.Writer, title string, counts map[string]int) {
	fmt.Fprintf(w, "%s:\n", title)
	names := slices.Sorted(maps.Keys(counts))
	slices.SortStableFunc(names, func(a, b string) int { return cmp.Compare(counts[b], counts[a]) })
	for _, name := range names {
		fmt.Fprintf(w, "  %-10s %d\n", name, counts[name])
	}
}
