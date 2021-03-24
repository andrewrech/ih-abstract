package main

import (
	"io"
	"log"
	"os"
	"time"

	_ "github.com/davecgh/go-spew/spew"
)

// count counts processed lines per unit time.
func count(counter *int64, descr string, signal chan struct{}) {
	second := make(chan struct{})

	counterInterval := 2000000000 // nanoseconds
	t := time.Duration(counterInterval)

	go func() {
		for {
			time.Sleep(t)
			second <- struct{}{}
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

// splitCh splits a []string channel into two channels, sending results from the input channel onto both output channels
func splitCh(in chan []string) (out1 chan []string, out2 chan []string, done chan struct{}) {
	var buf int64 = 2e7

	out1 = make(chan []string, buf)
	out2 = make(chan []string, buf)
	done = make(chan struct{})

	go func() {
		for l := range in {
			out1 <- l
			out2 <- l
		}

		close(out1)
		close(out2)
		done <- struct{}{}
	}()

	return
}

// copyFileContents copies the contents of the file named src to the file named by dst.
func copyFileContents(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		log.Fatalln(err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		err := out.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	_, err = io.Copy(out, in)

	if err != nil {
		log.Fatalln(err)
	}

	err = out.Sync()

	if err != nil {
		log.Fatalln(err)
	}
}
