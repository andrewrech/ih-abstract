package main

import (
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
)

func TestDBLive(t *testing.T) {
	var config string
	var present bool

	if config, present = os.LookupEnv("IH_ABSTRACT_TEST_CONFIG"); !present {
		t.Skip("IH_ABSTRACT_TEST_CONFIG is unset, skipping connection test")
	}

	db, err := connect(config)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	r := DB(config, db)

	var counter int64
	for range r.out {
		counter++
	}

	t.Run("Read 5 records from live database", func(t *testing.T) {
		if counter != 5 {
			t.Fatalf("failed to read 5 records from live Immune Health SQL database")
		}
	})

	<-r.done
}

func TestDBMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	in := helperTestReader(TestFile)

	h := helperCorrectHeader()

	rows := sqlmock.NewRows(h)

	entry := make([]driver.Value, len(h))

	for l := range in {
		for i := range l {
			entry[i] = driver.Value(l[i])
		}

		rows.AddRow(entry...)
	}

	// run SQL test
	query := "SELECT"
	mock.ExpectQuery(query).WillReturnRows(rows)

	r := DB("ih-abstract.yml", db)

	var counter int64

	for range r.out {
		atomic.AddInt64(&counter, 1)
	}

	<-r.done

	t.Run("Read from mock SQL", func(t *testing.T) {
		diff := cmp.Diff(int64(13), counter)
		if diff != "" {
			t.Fatalf(diff)
		}
	})
}
