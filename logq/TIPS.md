# Tips

1. `os.Args`, `os.Open`, `bufio.Scanner`, `strings.Contains`, `defer`.
2. Both `os.Stdin` and `*os.File` are an `io.Reader`.
3. `strings.SplitN`, `time.Parse(time.RFC3339, …)`, `iota`, `fmt.Stringer`, `flag`.
4. `encoding/json`, struct tags (`omitempty`, `omitzero`, `-`), `encoding.TextMarshaler`, `bufio.Writer`. `os.Exit` skips deferred calls.
5. `slices.Sorted(maps.Keys(m))`, `slices.SortStableFunc`, `cmp.Compare`.
6. `gzip.NewReader` wraps an `io.Reader`.
