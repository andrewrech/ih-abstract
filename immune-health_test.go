package main

import (
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		in := helperTestReader(TestFilePhi)

		header := helperCorrectHeader()
		out, filterDone := filter(in, header)

		<-filterDone

		_ = out
	}
}

func TestFilter(t *testing.T) {

	in := helperTestReader(TestFile)

	header := helperCorrectHeader()
	out, filterDone := filter(in, header)

	<-filterDone

	ts := map[string]struct {
		input string
		want  int64
	}{
		"diff":           {input: string("diff"), want: int64(5)},
		"filter (CPD)":   {input: string("cpd"), want: int64(1)},
		"filter (WBC)":   {input: string("wbc"), want: int64(1)},
		"filter (PD-L1)": {input: string("pdl1"), want: int64(3)},
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

func csvOutTestHelper(oldFile, newFile string) {
	existing := Existing(&oldFile)

	f, err := os.Open(newFile)
	if err != nil {
		log.Fatalln(err)
	}

	r := readCSV(f)

	channels, filterDone := filter(r.out, r.header)

	colNames := headerParse(r.header)

	var diffDone chan int
	channels["ih"], diffDone = New(existing, colNames, channels["diff"])

	pdl1Done := DiffUnq(channels["pdl1Ret"], "pdl1")

	msiDone := DiffUnq(channels["msiRet"], "msi")

	writeDone := write(r.header, channels)

	<-r.done
	<-filterDone
	<-pdl1Done
	<-msiDone
	<-diffDone
	<-writeDone
}

func TestCsvOut(t *testing.T) {
	csvOutTestHelper(TestFileOld, TestFile)

	ts := map[string]struct {
		input string
		want  int64
	}{
		"ih.csv":           {input: string("ih.csv"), want: int64(6)},
		"pdl1.csv":         {input: string("pdl1.csv"), want: int64(4)},
		"wbc.csv":          {input: string("wbc.csv"), want: int64(2)},
		"msi.csv":          {input: string("msi.csv"), want: int64(2)},
		"cpd.csv":          {input: string("cpd.csv"), want: int64(2)},
		"msi-unq.txt":      {input: string("msi-unq.txt"), want: int64(2)},
		"msi-unq-new.txt":  {input: string("msi-unq-new.txt"), want: int64(2)},
		"pdl1-unq.txt":     {input: string("pdl1-unq.txt"), want: int64(2)},
		"pdl1-unq-new.txt": {input: string("pdl1-unq-new.txt"), want: int64(2)},
		"new-ids.txt":      {input: string("new-ids.txt"), want: int64(4)},
	}

	for x, n := range ts {
		x := x
		n := n

		t.Run(x, func(t *testing.T) {
			lines := helperCsvLines(n.input)

			diff := cmp.Diff(n.want, lines)

			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func BenchmarkCsvOut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		csvOutTestHelper(TestFileOldPhi, TestFilePhi)
	}
}
