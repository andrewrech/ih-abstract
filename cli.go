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
	old := flag.String("old", "", "Path to existing results.csv output data from last run (optional)")
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
		fmt.Fprintf(os.Stderr, "  < results-raw.csv | ih-abstract\n")
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

    results.csv:                     all results
    results-increment.csv:           new results since last run
    new-ids.txt:                     patient identifiers with new results since last run

 Output for quality assurance:

    pdl1.csv:                        potential PD-L1 reports
    msi.csv:                         potential MSI reports
    cpd.csv:                         potential CPD reports
    wbc.csv:                         white blood cell counts

    pdl1-unique-strings.txt:         unique PD-L1 strings
    pdl1-unique-strings-new.txt:     unique PD-L1 strings, new vs. last run
    msi-unique-strings.txt:          unique MSI strings
    msi-unique-strings-new.txt:      unique MSI strings, new vs. last run

CONFIGURATION FILE:

  See 'ih-abstract --print-config'. Paths searched by default:

      $XDG_CONFIG_HOME/ih-abstract/ih-abstract.yml
      $HOME/.ih-abstract.yml
      ./ih-abstract.yml


TESTING:

 go test
                          NOTE: some integration tests require restricted
                          PHI-containing data. Data is available within our
                          organization upon request.

													NOTE: To test the live server connection, set
                          environment variable IH_ABSTRACT_TEST_CONFIG to
                          a test configuration file path.
                          These tests is disabled by default.

BENCHMARKING:

  go test -bench=.
                          NOTE: some benchmarks require restricted
                          organization VPN access. Access is available within our
                          organization upon request. To test the live server
                          connection, set environment variable
                          IH_ABSTRACT_TEST_CONFIG to a test configuration
                          file path. These tests is disabled by default.

`)
	}
}
