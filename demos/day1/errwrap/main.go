package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
	err := myOpenFile("/path/to/file")
	switch {
	case errors.Is(err, fs.ErrNotExist):
		fmt.Println("file does not exist, creating it")
	case err != nil:
		fmt.Println("other error", err)
	default:
		fmt.Println("success")
	}
}

func myOpenFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		// Using %w wraps the underlying error so errors.Is / errors.As can inspect it.
		// Using %s or %v only formats the text and loses the error chain!
		return fmt.Errorf("in myOpenFile: %s", err)
	}
	defer f.Close()

	return nil
}
