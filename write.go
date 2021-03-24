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
func File(name string, h []string, trunc bool) (w Writer) {
	var f *os.File
	var err error

	// open file with trunc or append
	if trunc {
		f, err = os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	}
	if !trunc {
		f, err = os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	}
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

// WriteRows appends strings to a CSV file using a Writer.
func WriteRows(in chan []string, name string, h []string, trunc bool, done chan struct{}) {
	w := File(name, h, trunc)

	log.Println("Writing", name)
	go func() {
		for l := range in {
			err := w.w.Write(l)
			if err != nil {
				log.Fatalln(err)
			}

			atomic.AddInt64(&w.counter, 1)
		}

		w.done()
		done <- struct{}{}
		log.Println("Wrote", w.counter, "records to", name)
	}()
}

// Write writes results to output CSV files using a common header.
func Write(h []string, in map[string](chan []string)) (done chan struct{}) {
	done = make(chan struct{})

	nOutputFiles := len(in)
	signal := make(chan struct{}, nOutputFiles)

	for i, c := range in {

		// filename
		var fn strings.Builder
		fn.WriteString(i)
		fn.WriteString(".csv")

		WriteRows(c, fn.String(), h, true, signal)
	}

	go func() {
		for i := 0; i < nOutputFiles; i++ {
			<-signal
		}

		close(done)
	}()

	return done
}

// WriteIncremental appends new unique results to an existing results file.
func WriteIncremental(newFile string, oldFile string, h []string) (done chan struct{}) {
	fn := "results-all.csv"

	copyFileContents(oldFile, fn)

	conn, err := os.Open(newFile)
	if err != nil {
		log.Fatalln(err)
	}

	r := readCSV(conn)

	done = make(chan struct{})
	writeDone := make(chan struct{})

	WriteRows(r.out, fn, h, false, writeDone)

	<-r.done
	<-writeDone

	close(done)
	return done
}
