package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestFile is a test CSV file containing simulated / de-identified input data.
const TestFile = "test.csv"

// TestFileOld is a test CSV file containing simulated / de-identified input data.
// Some records are removed to simulate old data.
const TestFileOld = "test_old.csv"

// TestFileDiffUnq contains simulated "old" unique strings that should not be present in new unique string .csv output.
const TestFileDiffUnq = "test-diff-unq.csv"

// TestFilePhi is a test CSV file containing real input data.
// This file is available within our organization upon request.
const TestFilePhi = "phi/test_phi.csv"

// TestFileOldPhi is a test CSV file containing real outdated input data.
// This file is available within our organization upon request.
const TestFileOldPhi = "phi/test_old_phi.csv"

func TestMain(m *testing.M) {
	previousDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	err = os.Chdir("./testdata")
	if err != nil {
		log.Println(err)
	}

	exitVal := m.Run()

	// delete test output if it exists
	testOutput := []string{
		"cpd.csv",
		"msi-unq-new.txt",
		"msi-unq.txt",
		"msi.csv",
		"new-ids.txt",
		"pdl1-unq-new.txt",
		"pdl1-unq.txt",
		"pdl1.csv",
		"wbc.csv",
		"ih.csv",
		"test-diff-unq.txt",
		"test-diff-unq-new.txt",
	}

	for _, f := range testOutput {
		_, err := os.Stat(f)

		if err == nil {
			err := os.Remove(f)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	// change to previous directory
	err = os.Chdir(previousDir)
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitVal)
}

func helperTestReader(f string) (out chan []string) {
	conn, err := os.Open(f)
	if err != nil {
		log.Fatalln(err)
	}

	r := csv.NewReader(conn)
	r.LazyQuotes = true

	in := make(chan []string, 1000000)

	// discard header for testing purposes
	_, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {
			l, err := r.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalln(err)
			}
			in <- l
		}
		close(in)

		conn.Close()
	}()

	return in
}

func helperCorrectHeader() (h []string) {
	h = []string{
		"MRN",
		"MRNFacility",
		"MedViewPatientID",
		"PatientName",
		"DOB",
		"Sex",
		"DrawnDate",
		"DiagServiceID",
		"AccessionNumber",
		"HNAMOrderID",
		"OrderTypeLocalID",
		"OrderTypeMnemonic",
		"TestTypeLocalID",
		"TestTypeMnemonic",
		"ResultDate",
		"Value",
		"Status",
	}

	return
}

func helperCsvLines(f string) int64 {
	var counter int64

	conn, err := os.Open(f)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		conn.Close()
	}()

	r := csv.NewReader(conn)
	r.LazyQuotes = true

	for {
		_, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		atomic.AddInt64(&counter, 1)
	}

	return counter
}

func BenchmarkFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		in := helperTestReader(TestFilePhi)

		colNames := make(map[string]int)

		colNames["OrderTypeMnemonic"] = 11
		colNames["TestTypeMnemonic"] = 13
		colNames["Value"] = 15

		out, filterDone := filter(in, colNames)

		<-filterDone

		_ = out
	}
}

func TestFilter(t *testing.T) {
	colNames := make(map[string]int)

	colNames["OrderTypeMnemonic"] = 11
	colNames["TestTypeMnemonic"] = 13
	colNames["Value"] = 15
	colNames["Status"] = 16

	in := helperTestReader(TestFile)

	out, filterDone := filter(in, colNames)

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

	colNames := headerParse(r.header)

	channels, filterDone := filter(r.out, colNames)

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
