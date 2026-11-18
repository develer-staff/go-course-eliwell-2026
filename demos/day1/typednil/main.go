package main

import "fmt"

type CustomErr struct{}

func (e *CustomErr) Error() string { return "boom" }

func getError() error {
	var p *CustomErr = nil // typed nil pointer
	return p               // Returns error interface with itab != nil!
}

func main() {
	fmt.Println(getError() == nil) // false
}
