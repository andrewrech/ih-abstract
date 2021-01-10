package main

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
