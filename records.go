package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"runtime"
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

	signal := make(chan int)

	var counter int64
	stopCounter := make(chan int)
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
			signal <- 1
		}()
	}

	<-r.done

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		<-signal
	}

	stopCounter <- 1

	log.Println("total:", counter, "records")

	return rs
}

// New identifies new Pathology database records based on a record hash.
// For each new record, the corresponding patient identifier to saved to a file.
func New(r *Records, colNames map[string]int, in chan []string) (out chan []string, done chan int) {
	var counter int64

	n := make(map[string]([]string))

	var buf int64 = 2e7

	out = make(chan []string, buf)

	done = make(chan int)

	h := []string{"MRN", "MRNFacility", "MedViewPatientID", "PatientName", "DOB"}

	w := File("new-ids.txt", h)

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

			_, ok := n[l[colNames["MRN"]]] // do not duplicate record

			if ok {
				continue
			}

			k := l[colNames["MRN"]]
			v := []string{
				l[colNames["MRNFacility"]],
				l[colNames["MedViewPatientID"]],
				l[colNames["PatientName"]],
				l[colNames["DOB"]],
			}

			n[k] = v
		}

		for k, v := range n {
			s := append([]string{k}, v...)

			err := w.w.Write(s)
			if err != nil {
				log.Fatalln(err)
			}

			atomic.AddInt64(&counter, 1)
		}

		w.done()

		close(out)
		done <- 1

		log.Println("UIDs with new records:", counter)
	}()

	return out, done
}

// Diff diffs old and new record sets.
func Diff(oldFile *string, in chan []string, header []string) (out chan []string, done chan int) {
	var buf int64 = 2e7

	colNames := headerParse(header)

	out = make(chan []string, buf)

	done = make(chan int)

	go func() {
		if *oldFile == "" {

			log.Println("No existing record set provided; returning all records")

			for l := range in {
				out <- l
			}

			close(out)
			done <- 1
		}

		if *oldFile != "" {
			var records Records

			r := &records

			r.Store = make(Store)

			r = Existing(oldFile)

			out, done = New(r, colNames, in)
		}
	}()

	return out, done
}
