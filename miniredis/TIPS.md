# Tips

1. `net.Listen`, `Accept`, `go`, `strings.Fields`, `sync.Mutex`, `strconv.Quote`.
2. `strconv.Atoi`, `path.Match`, `slices.Sort`.
3. `map[string]any`, type assertions, `func get[T any](…)`.
4. `slices.Sorted(maps.Keys(h))`.
5. A second map `map[string]time.Time`, `time.Ticker`.
6. `signal.NotifyContext`, `context.AfterFunc` (to close the listener and the connections), `sync.WaitGroup.Go`.
7. Closing a channel wakes every receiver. `context.WithTimeout`. Don't wait with the mutex locked.
