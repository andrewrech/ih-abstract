package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb" // sql driver
)

// read reads raw input data.
func read(f flags, in *os.File) (r rawRecords) {
	log.Println("initializing records map")

	if *f.sql {
		log.Println("reading Sql")

		db, err := connect(*f.config)
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		r = DB(*f.config, db)
	}

	if !(*f.sql) {
		log.Println("reading Stdin")

		r = readCSV(in)
	}

	return r
}

// readSQLRows reads rows of strings from an SQL database.
func readSQLRows(rows *sql.Rows) (r rawRecords) {
	var buf int64 = 2e7

	// initialize channels
	r.out = make(chan []string, buf)
	r.done = make(chan struct{})

	var err error
	r.header, err = rows.Columns()
	if err != nil {
		log.Fatalln(err)
	}

	// slice of byte slices of correct length
	rawResult := make([]sql.NullString, len(r.header))
	// string slice of correct length
	result := make([]string, len(r.header))
	// destination interface
	dest := make([]interface{}, len(r.header))

	// add pointers to destination
	for i := range result {
		dest[i] = &rawResult[i]
	}

	var counter int64
	stopCounter := make(chan struct{})
	count(&counter, "read (sql)", stopCounter)

	go func() {
		for rows.Next() {
			// fill destination
			err = rows.Scan(dest...)
			if err != nil {
				log.Fatalln(err)
			}

			counter++

			for i, raw := range rawResult {
				// handle nil type with conversion to ""
				if raw.Valid {
					result[i] = raw.String
				}
			}

			r.out <- result
		}

		err = rows.Err()
		if err != nil {
			log.Fatalln("error encountered during iteration:", rows.Err())
		}

		close(r.out)
		stopCounter <- struct{}{}
		r.done <- struct{}{}
	}()

	return r
}

// readCSV reads records from a CSV file.
func readCSV(in io.Reader) (r rawRecords) {
	var buf int64 = 2e7

	// initialize channels
	r.out = make(chan []string, buf)
	r.done = make(chan struct{})

	reader := csv.NewReader(in)
	reader.LazyQuotes = true

	var err error
	r.header, err = reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	var counter int64
	stopCounter := make(chan struct{})
	count(&counter, "read (csv)", stopCounter)

	// process records
	go func() {
		for {
			l, err := reader.Read()

			counter++

			switch {
			case errors.Is(err, io.EOF):
				r.done <- struct{}{}

				close(r.out)

			case err != nil:
				log.Fatal(err)

			default:
				r.out <- l
			}
		}
	}()

	stopCounter <- struct{}{}

	return r
}

// headerParse parses input data column names.
func headerParse(h []string) (colNames map[string]int) {
	colNames = make(map[string]int)

	for i, s := range h {
		colNames[s] = i
	}

	return
}
