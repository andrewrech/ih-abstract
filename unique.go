package main

import (
	"log"
	"os"
	"strings"
)

// prevUnq adds previously identified unique strings from an existing output file to a hash map.
func prevUnq(f string) (r *Records) {
	var records Records
	records.Store = make(Store)

	r = &records

	if _, err := os.Stat(f); err == nil {
		log.Println("reading patterns from existing records file", f)
		r = Existing(&f)
	} else {
		log.Println("existing records file", f, "does not exist, skipping diff")
	}

	return r
}

// DiffUnq saves unique strings to an output file "-unq.txt".
// If the output file already exists, the file is overwritten and a
// second output file "-unq-new.txt" is generated. The second output
// file contains only new strings not identified previously.
func DiffUnq(in chan []string, name string) (done chan int) {
	done = make(chan int)

	// add to an existing records map if
	// if CSV output already exists
	f := strings.Join([]string{name, "-unq.txt"}, "")
	newF := strings.Join([]string{name, "-unq-new.txt"}, "")
	existing := prevUnq(f)

	// create new output files
	newUnique := File(newF, []string{"unique-raw-string"})
	allUnique := File(f, []string{"unique-raw-string-new"})

	n := make(map[string](bool))

	go func() {
		for l := range in {
			for _, i := range l { // each string of slice
				s := i

				// skip if already seen
				if n[s] {
					continue
				}

				n[i] = true // add to map

				err := allUnique.w.Write([]string{s})
				if err != nil {
					log.Fatalln(err)
				}

				// write to file if record does not
				// exist in existing records
				exists, err := existing.Check(&[]string{s})
				if err != nil {
					log.Fatalln(err)
				}

				if !exists {
					log.Println("New string:", s)

					err := newUnique.w.Write([]string{s})
					if err != nil {
						log.Fatalln(err)
					}
				}
			}
		}

		allUnique.done()
		newUnique.done()
		done <- 1
	}()

	return done
}
