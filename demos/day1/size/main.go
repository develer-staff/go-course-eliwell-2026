package main

import (
	"fmt"
	"unsafe"
)

type Suboptimal struct {
	a bool
	b int64
	c bool
}

type Optimal struct {
	b int64
	a bool
	c bool
}

func main() {
	var opt Optimal
	var subopt Suboptimal

	fmt.Printf("%T is %d bytes\n", subopt, unsafe.Sizeof(subopt))
	fmt.Printf("%T is %d bytes\n", opt, unsafe.Sizeof(opt))
}
