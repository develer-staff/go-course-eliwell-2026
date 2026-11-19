// Slide "Standard Library Highlights", Go 1.23 to 1.27. Needs Go 1.27.
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

func main() {
	// 1.23: iterator helpers.
	m := map[string]int{"press": 1013, "temp": 21, "hum": 48}
	keys := slices.Sorted(maps.Keys(m))
	fmt.Println("1.23 sorted keys:      ", keys)

	// 1.24: os.Root, file access that cannot escape a directory.
	dir, _ := os.MkdirTemp("", "root")
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"setpoint":4}`), 0o644)
	root, err := os.OpenRoot(dir)
	if err != nil {
		panic(err)
	}
	defer root.Close()

	if f, err := root.Open("config.json"); err == nil {
		fmt.Println("1.24 root.Open ok:     ", f.Name())
		f.Close()
	}
	_, err = root.Open("../../etc/passwd")
	fmt.Println("1.24 root.Open escape: ", err)

	// 1.25: WaitGroup.Go does Add(1) and Done() for you.
	var wg sync.WaitGroup
	var mu sync.Mutex
	polled := 0
	for range 5 {
		wg.Go(func() {
			mu.Lock()
			polled++
			mu.Unlock()
		})
	}
	wg.Wait()
	fmt.Println("1.25 wg.Go:            ", polled, "devices polled")

	// 1.26: generic errors.AsType, no target variable.
	_, err = os.Open("missing.cfg")
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		fmt.Println("1.26 errors.AsType:    ", pe.Op, pe.Path)
	}

	// 1.27: encoding/json/v2 is stable.
	b, _ := json.Marshal(m, json.Deterministic(true))
	fmt.Println("1.27 json/v2 Marshal:  ", string(b))

	// v2 matches names case-sensitively; v1 (encoding/json) does not.
	type Reading struct {
		Temp float64 `json:"temp"`
	}
	in := []byte(`{"TEMP": 21.5}`)
	var r1, r2 Reading
	jsonv1.Unmarshal(in, &r1)
	json.Unmarshal(in, &r2)
	fmt.Printf(`1.27 "TEMP" -> temp:    v1=%v v2=%v`, r1.Temp, r2.Temp)
}
