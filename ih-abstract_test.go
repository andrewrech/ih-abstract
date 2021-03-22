package main

import (
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestFile is a test CSV file containing simulated / de-identified input data.
const TestFile = "test.csv"

// TestFileOld is a test CSV file containing simulated / de-identified outdated input data.
// Some records are removed to simulate old data.
const TestFileOld = "test_old.csv"

// TestFilePhi is a test CSV file containing real input data.
// This file is available within our organization upon request.
const TestFilePhi = "test_phi.csv"

// TestFilePhiOld is a test CSV file containing real outdated input data.
// This file is available within our organization upon request.
const TestFilePhiOld = "test_phi_old.csv"

// TestFilePhiGeneric is a test CSV file containing real input data.
// The file contains generic pathology data.
// This file is available within our organization upon request.
const TestFilePhiGeneric = "test_phi_generic.csv"

// TestFilePhiGenericOld is a test CSV file containing real outdated input data.
// The file contains generic pathology data.
// This file is available within our organization upon request.
const TestFilePhiGenericOld = "test_phi_generic_old.csv"

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

	// change to previous directory
	err = os.Chdir(previousDir)
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitVal)
}

func cleanupTestFull() {
	testFiles := []string{
		"cpd.csv",
		"msi-unique-strings.csv",
		"msi-unique-strings-new.csv",
		"msi.csv",
		"new-ids.tst",
		"pdl1-unique-strings.csv",
		"pdl1-unique-strings-new.csv",
		"pdl1.csv",
		"results-increment.csv",
		"results.csv",
		"wbc.csv",
	}

	for _, f := range testFiles {
		if _, err := os.Stat(f); err == nil {
			err := os.Remove(f)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func innerTest(f flags, newFile string) {
	conn, err := os.Open(newFile)
	if err != nil {
		log.Fatalln(err)
	}

	mainInner(f, conn)
}

func TestFullFilter(t *testing.T) {
	cleanupTestFull()

	config := ""
	example := false
	noFilter := false
	old := TestFileOld
	sql := false

	var f flags
	f.config = &config
	f.example = &example
	f.noFilter = &noFilter
	f.old = &old
	f.sql = &sql

	defer cleanupTestFull()
	innerTest(f, TestFile)

	tests := map[string]struct {
		input string
		want  int64
	}{
		"integration: cpd.csv":                     {input: "cpd.csv", want: int64(2)},
		"integration: msi-unique-strings-new.csv":  {input: "msi-unique-strings-new.csv", want: int64(3)},
		"integration: msi-unique-strings.csv":      {input: "msi-unique-strings.csv", want: int64(3)},
		"integration: pdl1-unique-strings-new.csv": {input: "pdl1-unique-strings-new.csv", want: int64(2)},
		"integration: pdl1-unique-strings.csv":     {input: "pdl1-unique-strings.csv", want: int64(2)},
		"integration: pdl1.csv":                    {input: "pdl1.csv", want: int64(5)},
		"integration: results-increment.csv":       {input: "results-increment.csv", want: int64(8)},
		"integration: results.csv":                 {input: "results.csv", want: int64(12)},
		"integration: wbc.csv":                     {input: "wbc.csv", want: int64(5)},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {
			got := helperCsvLines(tc.input)

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestFullNoFilter(t *testing.T) {
	cleanupTestFull()

	config := ""
	example := false
	noFilter := true
	old := TestFileOld
	sql := false

	var f flags
	f.config = &config
	f.example = &example
	f.noFilter = &noFilter
	f.old = &old
	f.sql = &sql

	defer cleanupTestFull()
	innerTest(f, TestFile)

	tests := map[string]struct {
		input string
		want  int64
	}{
		"integration: results-increment.csv": {input: "results-increment.csv", want: int64(13)},
		"integration: results.csv":           {input: "results.csv", want: int64(13)},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {
			got := helperCsvLines(tc.input)

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestPHIFilter(t *testing.T) {
	err := os.Chdir("./phi")
	if err != nil {
		log.Fatalln(err)
	}

	cleanupTestFull()

	err = os.Link("pdl1-unique-strings.csv-test", "pdl1-unique-strings.csv")
	if err != nil {
		log.Fatalln(err)
	}

	err = os.Link("msi-unique-strings.csv-test", "msi-unique-strings.csv")
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		err := os.Chdir("../")
		if err != nil {
			log.Fatalln(err)
		}
	}()

	config := ""
	example := false
	noFilter := false
	old := TestFilePhiOld
	sql := false

	var f flags
	f.config = &config
	f.example = &example
	f.noFilter = &noFilter
	f.old = &old
	f.sql = &sql

	defer cleanupTestFull()
	innerTest(f, TestFilePhi)

	tests := map[string]struct {
		input string
		want  int64
	}{
		"integration: cpd.csv":                     {input: "cpd.csv", want: int64(559)},
		"integration: msi-unique-strings-new.csv":  {input: "msi-unique-strings-new.csv", want: int64(1)},
		"integration: msi-unique-strings.csv":      {input: "msi-unique-strings.csv", want: int64(3)},
		"integration: pdl1-unique-strings-new.csv": {input: "pdl1-unique-strings-new.csv", want: int64(1)},
		"integration: pdl1-unique-strings.csv":     {input: "pdl1-unique-strings.csv", want: int64(16)},
		"integration: pdl1.csv":                    {input: "pdl1.csv", want: int64(118)},
		"integration: results-increment.csv":       {input: "results-increment.csv", want: int64(725)},
		"integration: results.csv":                 {input: "results.csv", want: int64(334924)},
		"integration: wbc.csv":                     {input: "wbc.csv", want: int64(334200)},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {
			got := helperCsvLines(tc.input)

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

func TestPHINoFilter(t *testing.T) {
	err := os.Chdir("./phi")
	if err != nil {
		log.Fatalln(err)
	}

	cleanupTestFull()

	defer func() {
		err := os.Chdir("../")
		if err != nil {
			log.Fatalln(err)
		}
	}()

	config := ""
	example := false
	noFilter := true
	old := TestFilePhiGenericOld
	sql := false

	var f flags
	f.config = &config
	f.example = &example
	f.noFilter = &noFilter
	f.old = &old
	f.sql = &sql

	// defer cleanupTestFull()
	innerTest(f, TestFilePhiGeneric)

	tests := map[string]struct {
		input string
		want  int64
	}{
		"integration: results-increment.csv": {input: "results-increment.csv", want: int64(50001)},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {
			got := helperCsvLines(tc.input)

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
