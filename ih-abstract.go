package main

import (
	"log"
	"os"
)

// ih-abstract streams input raw pathology results to the immune.health.report R package for report generation and quality assurance. The input is .csv data or direct streaming from a Microsoft SQL driver-compatible database. The output is filtered .csv/.txt files for incremental new report generation and quality assurance.
// Optionally, Immune Health filtering can be turned off to use ih-abstract as a general method to retrieve arbitrary or incremental pathology results.
func main() {
	usage()

	log.Println("starting")

	flags := flagParse()

	if *flags.example {
		printConf()
		os.Exit(0)
	}

	r := read(flags)

	// no filtering
	// pull data, diff and exit
	if *flags.noFilter {

		WriteRows(r.out, "./results.csv", r.header, r.done)
		<-r.done
	}

	// in contrast
	// filtering specifically for immune then
	// pull data and filter, then diff and exit
	if !*flags.noFilter {

		colNames := headerParse(r.header)

		channels, filterDone := filter(r.out, colNames)

		pdl1Done := DiffUnq(channels["pdl1Ret"], "pdl1")
		msiDone := DiffUnq(channels["msiRet"], "msi")

		var diffDone chan int
		channels["ih"], diffDone = Diff(flags.old, channels["diff"], colNames)

		writeDone := write(r.header, channels)

		// close parallel processes sequentially
		<-r.done
		<-filterDone
		<-pdl1Done
		<-msiDone
		<-diffDone
		<-writeDone
	}

	log.Println("done")
}