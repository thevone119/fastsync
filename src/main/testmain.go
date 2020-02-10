package main

import (
	"fmt"
	"time"
	"utils"
)

func main() {
	c := make(map[string]int)
	for j := 0; j < 1000000; j++ {
		c[fmt.Sprintf("%d", j)] = j
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000000; j++ {
				c[fmt.Sprintf("%d", utils.GetNextUint())]=j
			}
		}()
	}
	for{
		time.Sleep(1*time.Second)
	}

}