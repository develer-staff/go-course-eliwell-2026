# Tips

1. `net.Listen`, `Accept`, `go`, `bufio.Scanner`. Count connections in the accept loop.
2. A `client` with an `out chan string`; a `hub` with a `map[*client]bool` and a `select` loop. Only the hub closes `out`.
3. `slices.Sort`, `strings.Join`, `regexp.MustCompile`.
4. `strings.Cut`.
5. `signal.NotifyContext`, `context.AfterFunc` (to close the listener), `sync.WaitGroup.Go`. Every send to the hub must also watch `ctx.Done()`.
6. `select` with `default`.
