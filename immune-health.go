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
		channels["wbc"] <- l // WBC are
		channels["ih"] <- l

	case CPD(l[colNames["OrderTypeMnemonic"]]):
		channels["cpd"] <- l
		// CPD reports, PD-L1 reports, and MSI reports count
		// as "new" data and trigger a new report
		channels["diff"] <- l

	case PDL1(l[colNames["Value"]]):
		if pat["pdl1Report"].MatchString(l[colNames["Value"]]) {
			channels["pdl1"] <- l
			channels["diff"] <- l

			pdl1Result := pat["pdl1Result"].FindAllString(l[colNames["Value"]], 10)
			channels["pdl1Ret"] <- Whitespace(pdl1Result)
		}

	case MSI(l[colNames["Value"]]):
		if pat["msiReport"].MatchString(l[colNames["Value"]]) {
			channels["msi"] <- l
			channels["diff"] <- l

			msiResult := pat["msiResult"].FindAllString(l[colNames["Value"]], 10)
			channels["msiRet"] <- Whitespace(msiResult)
		}
	}

	atomic.AddInt64(counter, 1)
}

// filter filters a raw data input stream row by row.
func filter(in chan []string, header []string) (channels map[string](chan []string), done chan int) {
	done = make(chan int)

	var buffer int64 = 1e7

	// channels contains communication of rows
	// between goroutines processing data
	channels = make(map[string](chan []string))

	// main output channel to ih.csv
	// is closed by the the csv writer
	// after new records are compared to existing records
	channels["ih"] = make(chan []string, buffer)

	// other channels for filtering data are closed in this function
	names := []string{"diff", "wbc", "cpd", "pdl1", "msi", "pdl1Ret", "msiRet"}

	for _, name := range names {
		channels[name] = make(chan []string, buffer)
	}

	ioCores := 2 // save cores for I/O

	nProc := runtime.GOMAXPROCS(0) - ioCores

	// run at least two filtering processes
	if nProc < 2 {
		nProc = 2
	}

	signal := make(chan int, nProc)

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
			log.Println("starting filtering thread")

			for l := range in {
				filterRow(l, colNames, pat, channels, &counter)
			}
			signal <- 1
		}()
	}

	stopCounter := make(chan int)
	count(&counter, "processed", stopCounter)

	// wait and close
	go func() {
		for i := 0; i < nProc; i++ {
			<-signal

			log.Println(i)
		}

		stopCounter <- 1

		log.Println("total:", counter, "records")

		for _, name := range names {
			close(channels[name])
		}

		done <- 1
	}()

	return channels, done
}

// write writes results to output CSV files.
func write(h []string, in map[string](chan []string)) (done chan int) {
	done = make(chan int)

	nOutputFiles := 5
	signal := make(chan int, nOutputFiles)

	if _, ok := in["ih"]; ok {
		WriteRows(in["ih"], "ih.csv", h, signal)
	}
	if _, ok := in["wbc"]; ok {
		WriteRows(in["wbc"], "wbc.csv", h, signal)
	}
	if _, ok := in["cpd"]; ok {
		WriteRows(in["cpd"], "cpd.csv", h, signal)
	}
	if _, ok := in["pdl1"]; ok {
		WriteRows(in["pdl1"], "pdl1.csv", h, signal)
	}
	if _, ok := in["msi"]; ok {
		WriteRows(in["msi"], "msi.csv", h, signal)
	}

	go func() {
		for i := 0; i < nOutputFiles; i++ {
			<-signal
		}

		done <- 1
	}()

	return done
}
