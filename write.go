package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

// Writer contains a file name, connection, CSV Writer, and a 'done' signal to cleanup the connection.
type Writer struct {
	name    string
	conn    *os.File
	w       *csv.Writer
	counter int64
	done    func()
}

// File creates an output CSV write file.
func File(name string, h []string) (w Writer) {
	f, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalln(err)
	}

	c := csv.NewWriter(f)

	c.UseCRLF = false

	err = c.Write(h)
	if err != nil {
		return Writer{}
	}

	w.name = name
	w.conn = f
	w.w = c

	// done closes the connection
	done := func() {
		c.Flush()

		err := c.Error()
		if err != nil {
			log.Fatalln(err)
		}

		f.Close()
	}

	w.done = done

	return w
}

// Write appends strings to a CSV file using a Writer.
func WriteRows(in chan []string, name string, h []string, done chan int) {
	w := File(name, h)

	go func() {
		for l := range in {
			err := w.w.Write(l)

			if err != nil {
				log.Fatalln(err)
			}

			atomic.AddInt64(&w.counter, 1)
		}

		w.done()
		done <- 1

		log.Println("wrote", w.counter, "records to", name)
	}()
}
