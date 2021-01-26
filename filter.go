package main

import (
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
)

// Pdl1Report is the string form of the regular expression used to match PD-L1 reports of interest.
const Pdl1Report = "(?i)pd-?l1"

// MsiReport is the string form of the regular expression used to match microsatellite instability reports of interest.
const MsiReport = "[Mm]icrosatellite[ ]+[Ii]nstability"

// Pdl1Result is the string form of the regular expression used to extract PD-L1 tumor/cancer score results.
const Pdl1Result = "(?i)(tumor proportion score|combined positive score \\(cps\\)|cps score):? ?[><~]* ?[0-9\\-\\.]+ ?%?"

// MsiResult is the string form of the regular expression used to extract microsatellite instability results.
const MsiResult = "[^\\.:]+findings[^\\.]+[Mm]icrosat[^\\.]+."

// SpacesAndBreaks is the string form of the replace-all regular expression used to normalize whitespace in pathology report strings of interest.
const SpacesAndBreaks = `\s+`

// WbcLymph efficiently selects records that are WbcLymph or lymphocyte counts using a lookup table.
func WbcLymph(s string) bool {
	wbc := map[string]bool{
		"WBC":       true,
		"WBC Corr":  true,
		"Lymph Man": true,
	}

	return wbc[s]
}

// CPD efficiently selects reports that are CPD reports using a lookup table.
func CPD(s string) bool {
	cpd := map[string]bool{
		"Solid Tumor NGS Report":   true,
		"Fusion Transcript Report": true,
	}

	return cpd[s]
}

// PDL1 uses a lookup table to efficiently test if a string should be evaluated via regular expression as a potential PD-L1 report.
func PDL1(s string) bool {
	pats := []string{
		"PD",
		"Pd",
		"pD",
		"pd",
	}

	for _, i := range pats {
		if strings.Contains(s, i) {
			return true
		}
	}

	return false
}

// MSI uses a lookup table to efficiently test if a string should be evaluated via regular expression as a potential PD-L1 report.
func MSI(s string) bool {
	pats := []string{
		"Microsatellite",
		"microsatellite",
	}

	for _, i := range pats {
		if strings.Contains(s, i) {
			return true
		}
	}

	return false
}

// Exclude efficiently excludes unwanted report categories using a lookup table.
func Exclude(s string) bool {
	excl := map[string]bool{
		"CMV":                                    true,
		"RVP":                                    true,
		"HIVQNT":                                 true,
		"HCVQNT":                                 true,
		"SDIFF":                                  true,
		"CFPLUS Report":                          true,
		"Case - HIV Quantitation":                true,
		"Case - Respiratory Virus Panel":         true,
		"Case - Epstein-Barr Virus Quantitation": true,
		"HBV DNA":                                true,
		"Case - Cytomegalovirus Quantitation":    true,
		"HCVGENO":                                true,
		"BME Post Report":                        true,
		"Case - HCV Quantitation":                true,
		"BCR Quant Report":                       true,
		"HyperCoag Report":                       true,
		"CML Report":                             true,
		"TCRPCR Report":                          true,
		"FLT3 Report":                            true,
		"Case - Cystic Fibrosis":                 true,
		"BRAF Report":                            true,
		"BRCA1/BRCA2/ESR1 Report":                true,
		"Heme NGS Report":                        true,
		"SPAD":                                   true,
		"Immunophen Report":                      true,
	}

	return excl[s]
}

// Whitespace normalizes whitespace in report strings of interest.
func Whitespace(s []string) []string {
	r := regexp.MustCompile(SpacesAndBreaks)

	for i := range s {
		s[i] = strings.Trim(r.ReplaceAllString(s[i], " "), " \r\n")
	}

	return s
}

// filterRow filters a row of input data for matches to patterns of interest.
func filterRow(l []string, colNames map[string]int, pat map[string](*regexp.Regexp), channels map[string](chan []string), counter *int64) {
	switch {

	case Exclude(l[colNames["OrderTypeMnemonic"]]):

	// WBC are sent directly to output
	// WBC are not counted as 'new data'
	case WbcLymph(l[colNames["TestTypeMnemonic"]]):
		channels["wbc"] <- l
		channels["results"] <- l

	case CPD(l[colNames["OrderTypeMnemonic"]]):
		channels["cpd"] <- l
		// CPD reports, PD-L1 reports, and MSI reports count
		// as "new" data and trigger a new report
		channels["results"] <- l
		channels["diff"] <- l

	case PDL1(l[colNames["Value"]]):
		if pat["pdl1Report"].MatchString(l[colNames["Value"]]) {
			channels["results"] <- l
			channels["pdl1"] <- l
			channels["diff"] <- l

			pdl1Result := pat["pdl1Result"].FindAllString(l[colNames["Value"]], 10)
			channels["pdl1-to-diff"] <- Whitespace(pdl1Result)
		}

	case MSI(l[colNames["Value"]]):
		if pat["msiReport"].MatchString(l[colNames["Value"]]) {
			channels["results"] <- l
			channels["diff"] <- l
			channels["msi"] <- l

			msiResult := pat["msiResult"].FindAllString(l[colNames["Value"]], 10)
			channels["msi-to-diff"] <- Whitespace(msiResult)
		}
	}

	atomic.AddInt64(counter, 1)
}

// filterResults filters a raw data input stream row by row.
func filterResults(in chan []string, header []string) (results map[string](chan []string), done chan struct{}) {
	done = make(chan struct{})

	var buf int64 = 1e7

	// channels contains communication of rows
	// between goroutines processing data
	results = make(map[string](chan []string))

	// other channels for filtering data are closed in this function
	resultTypes := []string{
		"results",
		"diff",
		"wbc",
		"cpd",
		"pdl1",
		"msi",
		"pdl1-to-diff",
		"msi-to-diff",
	}

	for _, name := range resultTypes {
		results[name] = make(chan []string, buf)
	}

	ioCores := 2 // save cores for I/O

	nProc := runtime.GOMAXPROCS(0) - ioCores

	// run at least two filtering processes
	if nProc < 2 {
		nProc = 2
	}

	signal := make(chan struct{}, nProc)

	// create patterns to use for filtering
	pat := make(map[string](*regexp.Regexp))

	pat["pdl1Report"] = regexp.MustCompile(Pdl1Report)
	pat["msiReport"] = regexp.MustCompile(MsiReport)
	pat["pdl1Result"] = regexp.MustCompile(Pdl1Result)
	pat["msiResult"] = regexp.MustCompile(MsiResult)

	colNames := headerParse(header)

	var counter int64

	// filter records on each core
	for i := 0; i < nProc; i++ {
		go func() {
			for l := range in {
				filterRow(l, colNames, pat, results, &counter)
			}
			signal <- struct{}{}
		}()
	}

	stopCounter := make(chan struct{})
	count(&counter, "filtered", stopCounter)

	// wait and close
	go func() {
		for i := 0; i < nProc; i++ {
			<-signal
		}

		stopCounter <- struct{}{}

		log.Println("total filtered:", counter, "records")

		for _, name := range resultTypes {
			close(results[name])
		}

		done <- struct{}{}
	}()

	return results, done
}
