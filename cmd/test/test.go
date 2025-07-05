package main

import (
	"fmt"
	"time"
)

// demo process to test the functions
func main() {
	id := 1
	for {
		fmt.Printf("count %d\n", id)
		id++
		time.Sleep(2 * time.Second)
	}
}
