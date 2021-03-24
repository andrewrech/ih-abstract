package main

import (
	"log"
	"os"
	"strings"
)

// ih-abstract streams input raw pathology results to the immune.health.report R package for report generation and quality assurance. The input is .csv data or direct streaming from a Microsoft SQL driver-compatible database. The output is filtered .csv files for incremental new report generation and quality assurance.
// Optionally, Immune Health filtering can be turned off to use ih-abstract as a general method to retrieve arbitrary or incremental pathology results.
func main() {
	usage()

	log.Println("starting")

	f := flagParse()

	if *f.example {
		printConf()
		os.Exit(0)
	}

	mainInner(f, os.Stdin)
}

// mainInner facilitates testing by allowing parameters to be passed to the main program code path.
func mainInner(f flags, in *os.File) {
	// parallel process completion signals
	parallelProcesses := 9
	doneSignals := make([]chan struct{}, parallelProcesses)

	// result communication channels
	allResults := make(map[string](chan []string))      // unfiltered
	filteredResults := make(map[string](chan []string)) // from filtering
	msiResults := make(map[string](chan []string))      // msi strings
	pdl1Results := make(map[string](chan []string))     // pdl1 strings
	var buf int64 = 2e7
	diffResults := make(chan []string, buf) // results to diff

	// read raw input data
	r := read(f, in)
	doneSignals[0] = r.done

	// if no filter
	// write all results and diff results
	if *f.noFilter {
		doneSignals[1] = make(chan struct{})
		allResults["results"] = make(chan []string, buf)
		allResults["results"], diffResults, doneSignals[1] = splitCh(r.out)
	}

	// if immune health filter
	// write filtered results, diff results, immune health results
	if !*f.noFilter {
		filteredResults, doneSignals[2] = filterResults(r.out, r.header)
		// exclude intermediate channels
		// containing 'diff' in channel name
		// these are sent to DiffUnq below, not written out
		for i, c := range filteredResults {
			if strings.Contains(i, "diff") {
				continue
			}
			allResults[i] = c
		}

		diffResults = filteredResults["diff"]

		// determine unique strings
		pdl1Results, doneSignals[3] = DiffUnq(filteredResults["pdl1-to-diff"], "pdl1")
		msiResults, doneSignals[4] = DiffUnq(filteredResults["msi-to-diff"], "msi")

		// write unique strings
		doneSignals[5] = Write([]string{"unique-result"}, pdl1Results)
		doneSignals[6] = Write([]string{"unique-result"}, msiResults)
	}

	// diff
	if *f.old != "" {
		allResults["results-increment"], doneSignals[7] = Diff(f.old, diffResults, r.header)
	}

	doneSignals[8] = Write(r.header, allResults)

	// wait for all parallel processes to finish
	for _, signal := range doneSignals {
		if signal != nil {
			<-signal
		}
	}
}
