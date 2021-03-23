package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		in := helperTestReader(TestFilePhi)

		header := helperCorrectHeader()
		out, filterDone := filterResults(in, header)

		<-filterDone

		_ = out
	}
}

func TestFilter(t *testing.T) {
	in := helperTestReader(TestFile)

	header := helperCorrectHeader()

	out, filterDone := filterResults(in, header)

	<-filterDone

	ts := map[string]struct {
		input string
		want  int64
	}{
		"diff":           {input: string("diff"), want: int64(8)},
		"filter (CPD)":   {input: string("cpd"), want: int64(1)},
		"filter (WBC)":   {input: string("wbc"), want: int64(4)},
		"filter (PD-L1)": {input: string("pdl1"), want: int64(6)},
		"filter (MSI)":   {input: string("msi"), want: int64(1)},
	}

	for x, n := range ts {
		x := x
		n := n

		t.Run(x, func(t *testing.T) {
			var i int64

			for range out[n.input] {
				i++
			}

			diff := cmp.Diff(n.want, i)

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
