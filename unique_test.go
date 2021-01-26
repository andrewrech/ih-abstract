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

	channels, done := DiffUnq(in, "test-diff")

	<-done

	var unq []string

	for l := range channels["test-diff-unique-strings"] {
		unq = append(unq, l...)
	}

	var unqNew []string
	for l := range channels["test-diff-unique-strings-new"] {
		unqNew = append(unqNew, l...)
	}

	t.Run("Detect filtering of unique strings", func(t *testing.T) {
		diff := cmp.Diff(unq, []string{"A", "B", "C"})

		if diff != "" {
			t.Fatalf(diff)
		}
	})

	t.Run("Detect filtering of new strings compared to an existing output file", func(t *testing.T) {
		diff := cmp.Diff(unqNew, []string{"C"})

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}
