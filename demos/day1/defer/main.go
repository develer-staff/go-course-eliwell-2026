package main

import (
	"fmt"
	"os"
)

func main() {
	lifoOrder()

	argEvaluationVsClosure()

	processFilesBad([]string{
		"defer/main.go",
		"escape/main.go",
		"size/main.go",
	})

	processFilesGood([]string{
		"defer/main.go",
		"escape/main.go",
		"size/main.go",
	})
}

func lifoOrder() {
	fmt.Println("open file A")
	defer fmt.Println("closing A")

	fmt.Println("open file B")
	defer fmt.Println("closing B")

	fmt.Println("inside function body")
}

func argEvaluationVsClosure() {
	status := "initializing"

	// Nuance 1: Arguments to a deferred function call are evaluated IMMEDIATELY.
	defer fmt.Println("direct arg evaluation:", status)

	// Nuance 2: A closure captures the variable by reference, seeing its final value.
	defer func() {
		fmt.Println("closure capture:      ", status)
	}()

	status = "done"
}

// Pitfall: calling defer f.Close() directly inside a for-loop keeps all files
// open until the entire function exits.
func processFilesBad(paths []string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		defer f.Close() // Runs as soon as inspectFile returns!

		info, err := f.Stat()
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		fmt.Printf("inspected %s (%d bytes)\n", p, info.Size())
	}
}

// Solution: extract a small helper function so defer runs at the end of each
// iteration.
func processFilesGood(paths []string) {
	for _, p := range paths {
		if err := inspectFile(p); err != nil {
			fmt.Println("error:", err)
			return
		}
	}
}

func inspectFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close() // Runs as soon as inspectFile returns!

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("inspected %s (%d bytes)\n", path, info.Size())
	return nil
}
