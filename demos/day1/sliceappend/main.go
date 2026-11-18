package main

import (
	"fmt"
	"slices"
)

func main() {
	// 1. Sub-slicing shares the same underlying array.
	base := make([]int, 3, 4) // len=3, cap=4: [0, 0, 0, _]
	base[0], base[1], base[2] = 10, 20, 30

	fmt.Printf("initial base: %v (len=%d, cap=%d)\n", base, len(base), cap(base))

	// Appending within capacity mutates the shared backing array!
	shared := append(base, 99)
	shared[0] = 999
	fmt.Printf("after append within cap:\n")
	fmt.Printf("  base:   %v (len=%d, cap=%d) <- base[0] was mutated!\n", base, len(base), cap(base))
	fmt.Printf("  shared: %v (len=%d, cap=%d)\n", shared, len(shared), cap(shared))

	// 2. Appending beyond capacity allocates a brand new backing array.
	grown := append(shared, 555)
	grown[0] = -1
	fmt.Printf("\nafter append beyond cap (reallocation):\n")
	fmt.Printf("  shared: %v (len=%d, cap=%d) <- untouched\n", shared, len(shared), cap(shared))
	fmt.Printf("  grown:  %v (len=%d, cap=%d) <- new backing array\n", grown, len(grown), cap(grown))

	// 3. Safe copy with slices.Clone when isolation is needed.
	isolated := slices.Clone(base)
	isolated[0] = 42
	fmt.Printf("\nslices.Clone isolation: base[0]=%d, isolated[0]=%d\n", base[0], isolated[0])
}
