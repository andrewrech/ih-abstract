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

// DiffUnq identifies unique strings from an input stream and compares the unique strings to an existing output file. The function returns 1) unique strings and 2) new strings compared to the existing output file.
func DiffUnq(in chan []string, name string) (channels map[string](chan []string), done chan struct{}) {
	done = make(chan struct{})

	var buf int64 = 1e7

	// channels contains communication of rows
	// between goroutines processing data
	channels = make(map[string](chan []string))

	// add to an existing records map if
	// if CSV output already exists
	unqRecordsName := strings.Join([]string{name, "-unique-strings"}, "")
	unqRecordsNameNew := strings.Join([]string{name, "-unique-strings-new"}, "")

	channels[unqRecordsName] = make(chan []string, buffer)

	channels[unqRecordsNameNew] = make(chan []string, buffer)

	// read previous output
	f := strings.Join([]string{name, "-unique-strings.csv"}, "")
	existing := prevUnq(f)

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
