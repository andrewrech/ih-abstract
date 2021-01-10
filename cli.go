package main

import (
	"flag"
	"fmt"
	"os"
)

// flagVars contains variables set by command line flags.
type flags struct {
	config   *string
	example  *bool
	noFilter *bool
	old      *string
	sql      *bool
}

// flags parses command line flags.
func flagParse() (f flags) {

	config := flag.String("config", "", "Path to ih-abstract.yml SQL connection configuration file")
	example := flag.Bool("print-config", false, "Print an example configuration file and exit")
	noFilter := flag.Bool("no-filter", false, "Save input data to .csv and exit without Immune Health filtering")
	old := flag.String("old", "", "Path to existing ih.csv output data from previous run (optional)")
	sql := flag.Bool("sql", false, "Read input from Microsoft SQL database instead of Stdin")

	flag.Parse()

	f.config = config
	f.example = example
	f.noFilter = noFilter
	f.old = old
	f.sql = sql

	return

}

// usage prints usage.
func usage() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nSelect raw data for Immune Health report generation.\n")
		fmt.Fprintf(os.Stderr, "\nUSAGE:\n\n")
		fmt.Fprintf(os.Stderr, "  < ih-raw.csv | ih-abstract\n")
		fmt.Fprintf(os.Stderr, "\nDEFAULTS:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
DETAILS:

  ih-abstract streams input raw pathology results to the immune.health.report R package
  for report generation and quality assurance. The input is .csv data or direct streaming
  from a Microsoft SQL driver-compatible database. The output is filtered .csv/.txt files
  for incremental new report generation and quality assurance.

  Optionally, Immune Health filtering can be turned off to use ih-abstract as a general
  method to retrieve arbitrary or incremental pathology results.

  Quality assurance output consists of files containing
    1) unique, and
    2) new (never-before-seen)
  Immune Health report results for manual review.

  Dependencies are vendored and consist of the Go standard library and
  Microsoft SQL driver.

OUTPUT:

  Output for report generation:

    ih.csv:               input for the Immune Health R package that contains
                          all raw data required to generate reports
    new-ids.txt:           list of patient identifiers for which new raw data
                          exists vs. previous run

 Output for quality assurance:

    pdl1.csv:          potential PD-L1 reports
    msi.csv:           potential MSI reports
    cpd.csv:           potential CPD reports
    wbc.csv:           white blood cell counts

    pdl1-unq.txt:      unique PD-L1 strings
    pdl1-unq-new.txt:  unique PD-L1 strings, new vs. previous run
    msi-unq.txt:       unique MSI strings
    msi-unq-new.txt:   unique MSI strings, new vs. previous run

CONFIGURATION FILE:

  See 'ih-abstract --print-config'. Paths searched by default:

      $XDG_CONFIG_HOME/ih-abstract/ih-abstract.yml
      $HOME/.ih-abstract.yml
      ./ih-abstract.yml


TESTING:

 go test
                          Note: some integration tests require restricted
                          PHI-containing data. Data is available within our
                          organization upon request. To test the live server
                          connection, set environment variable
                          IH_ABSTRACT_TEST_CONFIG to a test configuration
                          file path. These tests is disabled by default.

BENCHMARKING:

  go test -bench=.

`)
	}
}
