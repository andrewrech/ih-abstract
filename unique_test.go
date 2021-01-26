package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiffUnq(t *testing.T) {
	l := []string{"A", "B", "C", "A"}

	in := make(chan []string)

	go func() {
		in <- l
		close(in)
	}()

	done := DiffUnq(in, "test-diff")

	<-done

	conn, err := os.Open(TestFileDiffUnq)
	if err != nil {
		log.Fatalln(err)
	}

	r := csv.NewReader(conn)
	r.LazyQuotes = true

	_, err = r.Read() // discard header
	if err != nil {
		log.Fatalln(err)
	}

	for x, n := range [][]string{{"A"}, {"B"}, {"C"}} {
		x := x
		n := n

		t.Run(fmt.Sprintln("unique string", x), func(t *testing.T) {
			i, err := r.Read()
			if err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(n, i)

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}

	t.Run("Detect filtering of old unique strings", func(t *testing.T) {
		lines := helperCsvLines("test-diff-unq-new.txt")

		diff := cmp.Diff(int64(4), lines)
		if diff != "" {
			t.Fatalf(diff)
		}
	})
}
