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
		diff := cmp.Diff(int(12), len(r.Store))
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

	header := helperCorrectHeader()

	var buf int64 = 2e7
	out := make(chan []string, buf)
	done := make(chan struct{})

	New(r, header, in, out, done)

	for l := range out {
		_ = l
	}

	<-done

	t.Run("Detect new data", func(t *testing.T) {
		lines := helperCsvLines("new-ids.txt")

		diff := cmp.Diff(int64(11), lines)
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

		header := helperCorrectHeader()

		var buf int64 = 2e7
		out := make(chan []string, buf)
		done := make(chan struct{})

		New(r, header, in, out, done)

		for l := range out {
			_ = l
		}

		<-done
	}
}
