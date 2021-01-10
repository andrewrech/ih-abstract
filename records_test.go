package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddCheckRecords(t *testing.T) {
	var r Records
	r.Store = make(Store)

	l := []string{
		"10000001",
		"ZZZ, ZZZ",
		"CBC",
		"100",
	}

	t.Run("Add record", func(t *testing.T) {
		err := r.Add(&l)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Check record", func(t *testing.T) {
		exists, err := r.Check(&l)
		if !exists || err != nil {
			t.Fatal("failed to check record")
		}
	})
}

func TestExistingRecords(t *testing.T) {
	f := TestFile

	r := Existing(&f)

	t.Run("existing", func(t *testing.T) {
		diff := cmp.Diff(int(7), len(r.Store))
		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

func BenchmarkExistingRecords(b *testing.B) {
	f := TestFilePhi

	var r *Records

	for i := 0; i < b.N; i++ {
		r = Existing(&f)
	}

	_ = r
}

func TestNewRecords(t *testing.T) {
	f := TestFileOld
	r := Existing(&f)

	in := helperTestReader(TestFile)

	h := helperCorrectHeader()

	colNames := headerParse(h)

	out, done := New(r, colNames, in)

	for l := range out {
		_ = l
	}

	<-done

	t.Run("Detect new data", func(t *testing.T) {
		lines := helperCsvLines("new-ids.txt")

		diff := cmp.Diff(int64(4), lines)
		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

func BenchmarkNewRecords(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := TestFilePhi
		r := Existing(&f)

		in := helperTestReader(TestFile)

		h := helperCorrectHeader()

		colNames := headerParse(h)

		out, done := New(r, colNames, in)

		for l := range out {
			_ = l
		}

		<-done
	}
}
