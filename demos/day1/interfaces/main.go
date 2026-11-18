package main

import (
	"fmt"
	"math"
)

type shape interface {
	perimeter() float64
	area() float64
}

type square struct {
	side float64
}

func (s square) perimeter() float64 { return 4 * s.side }
func (s square) area() float64      { return s.side * s.side }

type circle struct {
	radius float64
}

func (c circle) perimeter() float64 { return math.Pi * c.radius }            // 2r
func (c circle) area() float64      { return math.Pi * c.radius * c.radius } // πr2

func main() {
	shapes := []shape{
		square{side: 4},
		circle{radius: 12},
	}

	for _, s := range shapes {
		fmt.Printf("shape of type %T perimeter: %.2f   area: %.2f\n", s, s.perimeter(), s.area())
	}
}
