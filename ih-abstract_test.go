package main

import (
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestUsage(t *testing.T) {

	usage()

	_ = flagParse()

	printConf()

}

func TestLocateConfig(t *testing.T) {

	os.Setenv("XDG_CONFIG_HOME", "ih-abstractTestdirectory")
	os.Setenv("HOME", "ih-abstractTestdirectory")

	f, _ := locateDefaultConfig()

	if f != "" {
		t.Error("", f)
	}

}

func TestLoadConfig(t *testing.T) {

	vars, err := loadConfig("ih-abstract.yml")
	if err != nil {
		log.Fatalln(err)
	}

	tests := map[string]struct {
		got  string
		want string
	}{
		"Username": {got: vars.Username, want: "username"},
		"Password": {got: vars.Password, want: "password"},
		"Host":     {got: vars.Host, want: "host"},
		"Port":     {got: vars.Port, want: "443"},
		"Database": {got: vars.Database, want: "database"},
		"Query":    {got: vars.Query, want: "SELECT TOP 100 FROM table"},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {

			got := tc.got

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}

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

// helperCsvLines counts lines of a CSV file.
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

// TestReadLiveRecord tests that a record can be read from the live PHI-containing SQL databae.
func TestReadLiveRecord(t *testing.T) {

	if _, present := os.LookupEnv("IH_ABSTRACT_TEST_LIVE_CONNECTION"); !present {
		t.Skip("IH_ABSTRACT_TEST_LIVE_CONNECTION is unset, skipping connection test")
	}

	db, err := connect("../secrets/ih-abstract.yml")
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	r := DB("ih-abstract.yml", db)

	for l := range r.out {
		_ = l
	}

	<-r.done

}

func TestSql(t *testing.T) {
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
		diff := cmp.Diff(int64(7), counter)
		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

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

func TestRead(t *testing.T) {
	mock := strings.NewReader(`MRN,MRNFacility,MedViewPatientID,PatientName,DOB,Sex,DrawnDate,DiagServiceID,AccessionNumber,HNAMOrderID,OrderTypeLocalID,OrderTypeMnemonic,TestTypeLocalID,TestTypeMnemonic,ResultDate,Value,Status
"1000000001      ","UID       ",1111111111,"ZZZ, ZZZ",1950006-16 00:00:00.000,M,2020-11-15 05:28:00.000,GL,00000111111111,1111111111,1111111111,CMV,1111111111111,WBC,2014-11-15 05:37:58.000,Test removal on basis of Order,A
`)

	r := readCSV(mock)

	<-r.done

	t.Run("read CSV rows", func(t *testing.T) {
		var i int64

		for range r.out {
			i++
		}

		diff := cmp.Diff(int64(1), i)

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

func TestRowWriter(t *testing.T) {
	mock := strings.NewReader(`MRN,MRNFacility,MedViewPatientID,PatientName,DOB,Sex,DrawnDate,DiagServiceID,AccessionNumber,HNAMOrderID,OrderTypeLocalID,OrderTypeMnemonic,TestTypeLocalID,TestTypeMnemonic,ResultDate,Value,Status
"1000000001      ","UID       ",1111111111,"ZZZ, ZZZ",1950006-16 00:00:00.000,M,2020-11-15 05:28:00.000,GL,00000111111111,1111111111,1111111111,CMV,1111111111111,WBC,2014-11-15 05:37:58.000,Test removal on basis of Order,A
`)

	r := readCSV(mock)

	<-r.done

	t.Run("write CSV rows", func(t *testing.T) {
		writeDone := make(chan int)

		f := "test-write.csv"

		WriteRows(r.out, f, r.header, writeDone)
		defer os.Remove("test-write.csv")

		<-writeDone

		lines := helperCsvLines(f)

		diff := cmp.Diff(int64(2), lines)

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

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
