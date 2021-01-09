package main

import (
	"log"
	"time"

	_ "github.com/davecgh/go-spew/spew"
)

// count counts processed lines per unit time.
func count(counter *int64, descr string, signal chan int) {
	second := make(chan int)

	counterInterval := 2000000000 // nanoseconds
	t := time.Duration(counterInterval)

	go func() {
		for {
			time.Sleep(t)
			second <- 1
		}
	}()

	go func() {
		for {
			select {
			case <-signal:
				return
			case <-second:
				log.Println(descr, *counter, "records")
			}
		}
	}()
}
