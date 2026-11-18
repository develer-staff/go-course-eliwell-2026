package main

import "strconv"

func main() {}

func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
