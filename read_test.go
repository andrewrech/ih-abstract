package main

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadCSV(t *testing.T) {
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
