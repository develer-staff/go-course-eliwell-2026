package main

import "fmt"

func main() {
	header := getHeader()
	fmt.Printf("len %d, cap: %d\n", len(header), cap(header))
}

func getHeader() []byte {
	big := readBigFile()
	return big[:10]
}

func readBigFile() []byte {
	content := make([]byte, 1024*1024*256)
	return content
}
