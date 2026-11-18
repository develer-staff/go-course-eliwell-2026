package main

import (
	"fmt"
	"maps"
)

func main() {
	// basic operations
	m := make(map[string]int)

	m["k1"] = 7
	m["k2"] = 13

	fmt.Println("len:", len(m))

	delete(m, "k2")
	fmt.Println("map:", m)

	clear(m)

	_, exists := m["k2"]
	fmt.Println(`m["k2"] exists`, exists)

	// iteration
	m2 := map[string]int{"foo": 1, "bar": 2}
	for k, v := range m2 {
		fmt.Println(k, "=>", v)
	}

	// "maps" package
	fmt.Println(maps.Values(m))
}
