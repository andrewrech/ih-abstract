package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"sync/atomic"
	"testing"
)

// TestFile is a test CSV file containing simulated / de-identified input data.
const TestFile = "test.csv"

// TestFileOld is a test CSV file containing simulated / de-identified input data.
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
