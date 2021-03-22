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

	channels[unqRecordsName] = make(chan []string, buf)

	channels[unqRecordsNameNew] = make(chan []string, buf)

	// read previous output
	f := strings.Join([]string{name, "-unique-strings.csv"}, "")
	prevResults := prevUnq(f)

	var records Records
	records.Store = make(Store)
	currentResults := &records

	go func() {
		for l := range in { // for each slice
			for _, s := range l { // each string of slice

				i := []string{s}

				existsCurrent, err := currentResults.Check(&i)
				if err != nil {
					log.Fatalln(err)
				}

				// string does not exist in current records
				if !existsCurrent {
					err = currentResults.Add(&i)
					if err != nil {
						log.Fatalln(err)
					}

					channels[unqRecordsName] <- []string{s}
				}

				// string does not exist in previous records
				existsPrev, err := prevResults.Check(&i)
				if err != nil {
					log.Fatalln(err)
				}

				if !existsPrev {
					err = prevResults.Add(&i)
					if err != nil {
						log.Fatalln(err)
					}

					log.Println("New string:", s)
					channels[unqRecordsNameNew] <- []string{s}
				}
			}
		}

		close(channels[unqRecordsName])
		close(channels[unqRecordsNameNew])
		done <- struct{}{}
	}()

	return channels, done
}
