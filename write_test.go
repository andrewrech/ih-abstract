package main

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRowWriter(t *testing.T) {
	mock := strings.NewReader(`MRN,MRNFacility,MedViewPatientID,PatientName,DOB,Sex,DrawnDate,DiagServiceID,AccessionNumber,HNAMOrderID,OrderTypeLocalID,OrderTypeMnemonic,TestTypeLocalID,TestTypeMnemonic,ResultDate,Value
"1000000001","UID",1111111111,"ZZZ, ZZZ",1950006-1600:00:00.000,M,2020-11-1505:28:00.000,GL,00000111111111,1111111111,1111111111,CMV,1111111111111,WBC,2014-11-1505:37:58.000,Test removal on basis of Order
`)

	r := readCSV(mock)

	<-r.done

	t.Run("write CSV rows", func(t *testing.T) {
		writeDone := make(chan struct{})

		f := "test-write.csv"

		WriteRows(r.out, f, r.header, true, writeDone)
		defer os.Remove("test-write.csv")

		<-writeDone

		lines := helperCsvLines(f)

		diff := cmp.Diff(int64(2), lines)

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

func TestWrite(t *testing.T) {
	header := []string{"MRN", "MRNFacility", "MedViewPatientID", "PatientName", "DOB", "Sex", "DrawnDate", "DiagServiceID", "AccessionNumber", "HNAMOrderID", "OrderTypeLocalID", "OrderTypeMnemonic", "TestTypeLocalID", "TestTypeMnemonic", "ResultDate", "Value"}
	testStr := []string{"1000000001", "UID", "1111111111", "ZZZ, ZZZ", "1950006-1600:00:00.000", "M", "2020-11-1505:28:00.000", "GL", "00000111111111", "1111111111", "1111111111", "CMV", "1111111111111", "WBC", "2014-11-1505:37:58.000", "Test removal on basis of Order"}

	var buf int64 = 2e7
	in := make(chan []string, buf)

	channels := make(map[string](chan []string))
	channels["test-write"] = in

	in <- testStr
	close(in)

	t.Run("write CSV file from channel", func(t *testing.T) {
		defer os.Remove("test-write.csv")
		done := Write(header, channels)
		<-done

		lines := helperCsvLines("test-write.csv")
		diff := cmp.Diff(int64(2), lines)

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}

func TestWriteIncremental(t *testing.T) {
	header := []string{"MRN", "MRNFacility", "MedViewPatientID", "PatientName", "DOB", "Sex", "DrawnDate", "DiagServiceID", "AccessionNumber", "HNAMOrderID", "OrderTypeLocalID", "OrderTypeMnemonic", "TestTypeLocalID", "TestTypeMnemonic", "ResultDate", "Value"}

	done := WriteIncremental(TestFile, TestFileOld, header)

	<-done

	t.Run("append new results to an existing results file", func(t *testing.T) {
		lines := helperCsvLines("results-all.csv")

		diff := cmp.Diff(int64(21), lines)

		if diff != "" {
			t.Fatalf(diff)
		}
	})
}
