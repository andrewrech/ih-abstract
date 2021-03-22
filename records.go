package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/blake2b"
)

// Store is a blake2b hash map that stores string slices.
type Store map[[blake2b.Size256]byte](struct{})

// Records provides thread safe access to Store.
type Records struct {
	Store
	sync.Mutex
}

// Add adds a record.
func (r *Records) Add(l *[]string) (err error) {
	buf := &bytes.Buffer{}

	var x struct{}

	err = gob.NewEncoder(buf).Encode(l)
	if err != nil {
		return err
	}

	bs := buf.Bytes()

	hash := blake2b.Sum256(bs)

	r.Lock()
	r.Store[hash] = x
	r.Unlock()

	return nil
}

// Check checks that a record exists.
func (r *Records) Check(l *[]string) (exists bool, err error) {
	buf := &bytes.Buffer{}

	err = gob.NewEncoder(buf).Encode(l)
	if err != nil {
		return false, err
	}

	bs := buf.Bytes()

	hash := blake2b.Sum256(bs)

	r.Lock()
	_, ok := r.Store[hash]
	r.Unlock()

	return ok, nil
}

// Existing creates a map of existing records.
func Existing(name *string) (rs *Records) {
	var records Records
	records.Store = make(Store)

	rs = &records

	f, err := os.Open(*name)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("reading file", *name, "to hash map")

	r := readCSV(f)

	signal := make(chan struct{})

	var counter int64
	stopCounter := make(chan struct{})
	count(&counter, "hashed", stopCounter)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for l := range r.out {
				i := l

				err := rs.Add(&i)
				if err != nil {
					log.Fatalln(err)
				}

				atomic.AddInt64(&counter, 1)
			}
			signal <- struct{}{}
		}()
	}

	<-r.done

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		<-signal
	}

	stopCounter <- struct{}{}

	log.Println("total:", counter, "records")

	return rs
}

// New identifies new Pathology database records based on a record hash.
// For each new record, the corresponding patient identifier to saved to a file.
func New(r *Records, header []string, in chan []string, out chan []string, done chan struct{}) {
	var counter int64

	n := make(map[string](struct{}))
	w := File("new-ids.txt", []string{"identifier"})

	id, err := RecordID(header)
	if err != nil {
		log.Fatalln(err)
	}
	colNames := headerParse(header)
	idIdx := colNames[id]

	go func() {
		for l := range in {
			i := l
			out <- i

			exists, err := r.Check(&i)
			if err != nil {
				log.Fatalln(err)
			}

			if exists {
				continue
			}

			_, ok := n[l[idIdx]] // do not duplicate person instance output

			if ok {
				continue
			}

			n[l[idIdx]] = struct{}{}
		}

		for k := range n {
			err := w.w.Write([]string{k})
			if err != nil {
				log.Fatalln(err)
			}

			atomic.AddInt64(&counter, 1)
		}

		w.done()
		close(out)
		close(done)

		log.Println("Person-instances with new records:", counter)
	}()
}

// RecordID gets a single input data column name containing a person-instance identifier.
// The person instance identifier is either an MRN (preferred) or UID.
func RecordID(header []string) (id string, err error) {
	for _, id := range header {
		if strings.Contains(id, "MRN") {
			return id, nil
		}
	}

	for _, id := range header {
		if strings.Contains(id, "UID") {
			return id, nil
		}
	}

	return "", errors.New("cannot identify patient instance column name")
}

// Diff diffs old and new record sets.
func Diff(oldFile *string, in chan []string, header []string) (out chan []string, done chan struct{}) {
	var buf int64 = 2e7
	out = make(chan []string, buf)
	done = make(chan struct{})

	go func() {
		var records Records

		r := &records

		r.Store = make(Store)

		r = Existing(oldFile)

		New(r, header, in, out, done)
	}()

	return out, done
}
