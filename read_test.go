package main

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadCSV(t *testing.T) {
	mock := strings.NewReader(`MRN,MRNFacility,MedViewPatientID,PatientName,DOB,Sex,DrawnDate,DiagServiceID,AccessionNumber,HNAMOrderID,OrderTypeLocalID,OrderTypeMnemonic,TestTypeLocalID,TestTypeMnemonic,ResultDate,Value
"1000000001      ","UID",1111111111,"ZZZ, ZZZ",1950006-16 00:00:00.000,M,2020-11-15 05:28:00.000,GL,00000111111111,1111111111,1111111111,CMV,1111111111111,WBC,2014-11-15 05:37:58.000,Test removal on basis of Order
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

func BenchmarkReadCSV(b *testing.B) {
	mock := strings.NewReader(`MRN,MRNFacility,MedViewPatientID,PatientName,DOB,Sex,DrawnDate,DiagServiceID,AccessionNumber,HNAMOrderID,OrderTypeLocalID,OrderTypeMnemonic,TestTypeLocalID,TestTypeMnemonic,ResultDate,Value
"1000000001      ","UID",1111111111,"ZZZ, ZZZ",1950006-16 00:00:00.000,M,2020-11-15 05:28:00.000,GL,00000111111111,1111111111,1111111111,CMV,1111111111111,WBC,2014-11-15 05:37:58.000,Test removal on basis of Order
`)

	r := readCSV(mock)

	<-r.done

	for i := 0; i < b.N; i++ {
		var i int64

		for range r.out {
			i++
		}

	}
}

func TestReadSQLRows(t *testing.T) {
	var config string
	var present bool

	if config, present = os.LookupEnv("IH_ABSTRACT_TEST_CONFIG"); !present {
		t.Skip("IH_ABSTRACT_TEST_CONFIG is unset, skipping real SQL test")
	}

	db, err := connect(config)
	if err != nil {
		log.Fatalln(err)
	}

	vars, err := loadConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	vars.Query = "SELECT TOP (1000) * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2020-01-01'"

	rows, err := db.Query(vars.Query)

	r := readSQLRows(rows)

	var counter int64
	for range r.out {
		counter++
	}

	<-r.done

	log.Println(counter)
}

func TestRead(t *testing.T) {
	var f flags
	sql := false
	f.sql = &sql

	conn, err := os.Open(TestFile)
	if err != nil {
		log.Fatalln(err)
	}

	r := read(f, conn)
	<-r.done

	t.Run("read", func(t *testing.T) {
		for l := range r.out {
			_ = l
		}
	})
}

func BenchmarkRead(b *testing.B) {
	var f flags
	sql := false
	f.sql = &sql

	for i := 0; i < b.N; i++ {

		conn, err := os.Open(TestFile)
		if err != nil {
			log.Fatalln(err)
		}

		r := read(f, conn)
		<-r.done
	}
}
