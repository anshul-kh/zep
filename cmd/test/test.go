package main

import (
	"fmt"
	"time"
)

func main() {
	id := 1
	for {
		fmt.Printf("count %d\n", id)
		id++
		time.Sleep(2 * time.Second)
	}
}
