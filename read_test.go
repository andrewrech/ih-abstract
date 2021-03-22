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

func TestReadLiveSQLRows(t *testing.T) {
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

func BenchmarkReadLiveSQLRows(b *testing.B) {
	// Benchmark results:
	// Total records read is a major factor in slowdown
	// Unclear if this is related to connection over VPN or other factors
	// Incrementally reading results would improve performance at least 25x
	// go test -run xxx -bench Live
	// goos: linux
	// goarch: amd64
	// pkg: github.com/andrewrech/ih-abstract
	// cpu: Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz
	// Read_5_since_2020-2             10         145743066 ns/op
	// Read_5_recent_date-2             5         297709119 ns/op
	// Read_5_all_dates-2               9         128749305 ns/op
	// Read_all_from_distant_24_hour_period-2   1 38676233648 ns/op
	// Read_all_from_recent_24_hour_period-2    1 6274454093 ns/op
	// Read_all_from_recent_48_hour_period-2    1 144199276366 ns/op
	// Read_all_from_recent_96_hour_period-2    1 113868611573 ns/op
	var config string
	var present bool

	if config, present = os.LookupEnv("IH_ABSTRACT_TEST_CONFIG"); !present {
		b.Skip("IH_ABSTRACT_TEST_CONFIG is unset, skipping real SQL test")
	}

	db, err := connect(config)
	if err != nil {
		log.Fatalln(err)
	}

	benchmarks := []struct {
		name  string
		input string
	}{
		{
			"Read 5 since 2020",
			"SELECT TOP (5) * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2020-01-01'",
		},
		{
			"Read 5 recent date",
			"SELECT TOP (5) * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2021-03-01'",
		},
		{
			"Read 5 all dates",
			"SELECT TOP (5) * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2010-01-01'",
		},
		{
			"Read all from distant 24 hour period",
			"SELECT * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2016-03-19' AND DrawnDate <= '2016-03-21'",
		},
		{
			"Read all from recent 24 hour period",
			"SELECT * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2016-03-19' AND DrawnDate <= '2016-03-21'",
		},
		{
			"Read all from recent 48 hour period",
			"SELECT * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2021-03-18' AND DrawnDate <= '2021-03-21'",
		},
		{
			"Read all from recent 96 hour period",
			"SELECT * FROM [DMEE_ExtAccess].[immune_health].[LabData] WHERE MRNFacility = 'UID' AND DrawnDate >= '2021-03-16' AND DrawnDate <= '2021-03-21'",
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rows, err := db.Query(bm.input)
				if err != nil {
					log.Fatalln(err)
				}
				r := readSQLRows(rows)

				var counter int64
				for range r.out {
					counter++
				}

				<-r.done

				log.Println(counter)
			}
		})
	}
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
