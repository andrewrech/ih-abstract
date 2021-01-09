```
2021/01/09 17:13:26 starting

Select raw data for Immune Health report generation.

USAGE:

  < ih-raw.csv | ih-abstract

DEFAULTS:

  -config string
    	Path to ih-abstract.yml SQL connection configuration file
  -no-filter
    	Save input data to .csv and exit without Immune Health filtering
  -old string
    	Path to existing ih.csv output data from previous run (optional)
  -print-config
    	Print an example configuration file and exit
  -sq
    	Read input from Microsoft SQL database instead of Stdin

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
                          IH_ABSTRACT_TEST_LIVE_CONNECTION; this test is disabled
                          by default.

BENCHMARKING:

  go test -bench=.

```
<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# ih\-abstract

```go
import "github.com/andrewrech/ih-abstract"
```

## Index

- [Constants](<#constants>)
- [func CPD(s string) bool](<#func-cpd>)
- [func Diff(oldFile *string, in chan []string, colNames map[string]int) (out chan []string, done chan int)](<#func-diff>)
- [func DiffUnq(in chan []string, name string) (done chan int)](<#func-diffunq>)
- [func Exclude(s string) bool](<#func-exclude>)
- [func MSI(s string) bool](<#func-msi>)
- [func New(r *Records, colNames map[string]int, in chan []string) (out chan []string, done chan int)](<#func-new>)
- [func PDL1(s string) bool](<#func-pdl1>)
- [func WbcLymph(s string) bool](<#func-wbclymph>)
- [func Whitespace(s []string) []string](<#func-whitespace>)
- [func WriteRows(in chan []string, name string, h []string, done chan int)](<#func-writerows>)
- [func connect(config string) (db *sql.DB, err error)](<#func-connect>)
- [func count(counter *int64, descr string, signal chan int)](<#func-count>)
- [func filter(in chan []string, colNames map[string]int) (channels map[string](chan []string), done chan int)](<#func-filter>)
- [func filterRow(l []string, colNames map[string]int, pat map[string](*regexp.Regexp), channels map[string](chan []string), counter *int64)](<#func-filterrow>)
- [func headerParse(h []string) (colNames map[string]int)](<#func-headerparse>)
- [func locateDefaultConfig() (config string, err error)](<#func-locatedefaultconfig>)
- [func main()](<#func-main>)
- [func printConf()](<#func-printconf>)
- [func usage()](<#func-usage>)
- [func write(h []string, in map[string](chan []string)) (done chan int)](<#func-write>)
- [type Records](<#type-records>)
  - [func Existing(name *string) (rs *Records)](<#func-existing>)
  - [func prevUnq(f string) (r *Records)](<#func-prevunq>)
  - [func (r *Records) Add(l *[]string) (err error)](<#func-records-add>)
  - [func (r *Records) Check(l *[]string) (exists bool, err error)](<#func-records-check>)
- [type Store](<#type-store>)
- [type Writer](<#type-writer>)
  - [func File(name string, h []string) (w Writer)](<#func-file>)
- [type confVars](<#type-confvars>)
  - [func loadConfig(config string) (vars confVars, err error)](<#func-loadconfig>)
- [type flags](<#type-flags>)
  - [func flagParse() (f flags)](<#func-flagparse>)
- [type rawRecords](<#type-rawrecords>)
  - [func DB(config string, db *sql.DB) (r rawRecords)](<#func-db>)
  - [func read(f flags) (r rawRecords)](<#func-read>)
  - [func readCSV(in io.Reader) (r rawRecords)](<#func-readcsv>)
  - [func readSQLRows(rows *sql.Rows) (r rawRecords)](<#func-readsqlrows>)


## Constants

MsiReport is the string form of the regular expression used to match microsatellite instability reports of interest\.

```go
const MsiReport = "[Mm]icrosatellite[ ]+[Ii]nstability"
```

MsiResult is the string form of the regular expression used to extract microsatellite instability results\.

```go
const MsiResult = "[^\\.:]+findings[^\\.]+[Mm]icrosat[^\\.]+."
```

Pdl1Report is the string form of the regular expression used to match PD\-L1 reports of interest\.

```go
const Pdl1Report = "(?i)pd-?l1"
```

Pdl1Result is the string form of the regular expression used to extract PD\-L1 tumor/cancer score results\.

```go
const Pdl1Result = "(?i)(tumor proportion score|combined positive score \\(cps\\)|cps score):? ?[><~]* ?[0-9\\-\\.]+ ?%?"
```

SpacesAndBreaks is the string form of the replace\-all regular expression used to normalize whitespace in pathology report strings of interest\.

```go
const SpacesAndBreaks = `\s+`
```

## func [CPD](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L38>)

```go
func CPD(s string) bool
```

CPD efficiently selects reports that are CPD reports using a lookup table\.

## func [Diff](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L189>)

```go
func Diff(oldFile *string, in chan []string, colNames map[string]int) (out chan []string, done chan int)
```

Diff diffs old and new record sets\.

## func [DiffUnq](<https://github.com/andrewrech/ih-abstract/blob/main/unique.go#L30>)

```go
func DiffUnq(in chan []string, name string) (done chan int)
```

DiffUnq saves unique strings to an output file "\-unq\.txt"\. If the output file already exists\, the file is overwritten and a second output file "\-unq\-new\.txt" is generated\. The second output file contains only new strings not identified previously\.

## func [Exclude](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L82>)

```go
func Exclude(s string) bool
```

Exclude efficiently excludes unwanted report categories using a lookup table\.

## func [MSI](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L66>)

```go
func MSI(s string) bool
```

MSI uses a lookup table to efficiently test if a string should be evaluated via regular expression as a potential PD\-L1 report\.

## func [New](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L119>)

```go
func New(r *Records, colNames map[string]int, in chan []string) (out chan []string, done chan int)
```

New identifies new Pathology database records based on a record hash\. For each new record\, the corresponding patient identifier to saved to a file\.

## func [PDL1](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L48>)

```go
func PDL1(s string) bool
```

PDL1 uses a lookup table to efficiently test if a string should be evaluated via regular expression as a potential PD\-L1 report\.

## func [WbcLymph](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L27>)

```go
func WbcLymph(s string) bool
```

WbcLymph efficiently selects records that are WbcLymph or lymphocyte counts using a lookup table\.

## func [Whitespace](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L115>)

```go
func Whitespace(s []string) []string
```

Whitespace normalizes whitespace in report strings of interest\.

## func [WriteRows](<https://github.com/andrewrech/ih-abstract/blob/main/write.go#L56>)

```go
func WriteRows(in chan []string, name string, h []string, done chan int)
```

WriteRows appends strings to a CSV file using a Writer\.

## func [connect](<https://github.com/andrewrech/ih-abstract/blob/main/connect.go#L11>)

```go
func connect(config string) (db *sql.DB, err error)
```

connect connects to the SQL database\.

## func [count](<https://github.com/andrewrech/ih-abstract/blob/main/utils.go#L11>)

```go
func count(counter *int64, descr string, signal chan int)
```

count counts processed lines per unit time\.

## func [filter](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L166>)

```go
func filter(in chan []string, colNames map[string]int) (channels map[string](chan []string), done chan int)
```

filter filters a raw data input stream row by row\.

## func [filterRow](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L126>)

```go
func filterRow(l []string, colNames map[string]int, pat map[string](*regexp.Regexp), channels map[string](chan []string), counter *int64)
```

filterRow filters a row of input data for matches to patterns of interest\.

## func [headerParse](<https://github.com/andrewrech/ih-abstract/blob/main/read.go#L155>)

```go
func headerParse(h []string) (colNames map[string]int)
```

headerParse parses input data column names\.

## func [locateDefaultConfig](<https://github.com/andrewrech/ih-abstract/blob/main/config.go#L37>)

```go
func locateDefaultConfig() (config string, err error)
```

locateDefaultConfig locates the configuration file in $XDG\_CONFIG\_HOME\, $HOME\, or the current directory\.

## func [main](<https://github.com/andrewrech/ih-abstract/blob/main/ih-abstract.go#L12>)

```go
func main()
```

ih\-abstract streams input raw pathology results to the immune\.health\.report R package for report generation and quality assurance\. The input is \.csv data or direct streaming from a Microsoft SQL driver\-compatible database\. The output is filtered \.csv/\.txt files for incremental new report generation and quality assurance\. Optionally\, Immune Health filtering can be turned off to use ih\-abstract as a general method to retrieve arbitrary or incremental pathology results\.

## func [printConf](<https://github.com/andrewrech/ih-abstract/blob/main/config.go#L14>)

```go
func printConf()
```

printConf prints an example SQL database configuration file

## func [usage](<https://github.com/andrewrech/ih-abstract/blob/main/cli.go#L40>)

```go
func usage()
```

usage prints usage\.

## func [write](<https://github.com/andrewrech/ih-abstract/blob/main/immune-health.go#L241>)

```go
func write(h []string, in map[string](chan []string)) (done chan int)
```

write writes results to output CSV files\.

## type [Records](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L19-L22>)

Records provides thread safe access to Store\.

```go
type Records struct {
    Store
    sync.Mutex
}
```

### func [Existing](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L67>)

```go
func Existing(name *string) (rs *Records)
```

Existing creates a map of existing records\.

### func [prevUnq](<https://github.com/andrewrech/ih-abstract/blob/main/unique.go#L10>)

```go
func prevUnq(f string) (r *Records)
```

prevUnq adds previously identified unique strings from an existing output file to a hash map\.

### func \(\*Records\) [Add](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L25>)

```go
func (r *Records) Add(l *[]string) (err error)
```

Add adds a record\.

### func \(\*Records\) [Check](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L47>)

```go
func (r *Records) Check(l *[]string) (exists bool, err error)
```

Check checks that a record exists\.

## type [Store](<https://github.com/andrewrech/ih-abstract/blob/main/records.go#L16>)

Store is a blake2b hash map that stores string slices\.

```go
type Store map[[blake2b.Size256]byte](struct{})
```

## type [Writer](<https://github.com/andrewrech/ih-abstract/blob/main/write.go#L10-L15>)

Writer contains a file name\, connection\, CSV Writer\, and a 'done' signal to cleanup the connection\.

```go
type Writer struct {
    name string
    conn *os.File
    w    *csv.Writer
    done func()
}
```

### func [File](<https://github.com/andrewrech/ih-abstract/blob/main/write.go#L18>)

```go
func File(name string, h []string) (w Writer)
```

File creates an output CSV write file\.

## type [confVars](<https://github.com/andrewrech/ih-abstract/blob/main/config.go#L27-L34>)

confVars is a struct of configuration variables required for the SQL database connection\.

```go
type confVars struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Host     string `yaml:"host"`
    Port     string `yaml:"port"`
    Database string `yaml:"database"`
    Query    string `yaml:"query"`
}
```

### func [loadConfig](<https://github.com/andrewrech/ih-abstract/blob/main/config.go#L80>)

```go
func loadConfig(config string) (vars confVars, err error)
```

## type [flags](<https://github.com/andrewrech/ih-abstract/blob/main/cli.go#L10-L16>)

flagVars contains variables set by command line flags\.

```go
type flags struct {
    config   *string
    example  *bool
    noFilter *bool
    old      *string
    sql      *bool
}
```

### func [flagParse](<https://github.com/andrewrech/ih-abstract/blob/main/cli.go#L19>)

```go
func flagParse() (f flags)
```

flags parses command line flags\.

## type [rawRecords](<https://github.com/andrewrech/ih-abstract/blob/main/connect.go#L44-L48>)

rawRecords contains a header\, a channel of raw records\, and a channel indicating when raw records have been read\.

```go
type rawRecords struct {
    header []string
    out    chan []string
    done   chan int
}
```

### func [DB](<https://github.com/andrewrech/ih-abstract/blob/main/connect.go#L51>)

```go
func DB(config string, db *sql.DB) (r rawRecords)
```

DB reads records from an Sql database\.

### func [read](<https://github.com/andrewrech/ih-abstract/blob/main/read.go#L15>)

```go
func read(f flags) (r rawRecords)
```

read reads raw input data\.

### func [readCSV](<https://github.com/andrewrech/ih-abstract/blob/main/read.go#L107>)

```go
func readCSV(in io.Reader) (r rawRecords)
```

readCSV reads records from a CSV file\.

### func [readSQLRows](<https://github.com/andrewrech/ih-abstract/blob/main/read.go#L42>)

```go
func readSQLRows(rows *sql.Rows) (r rawRecords)
```

readSQLRows reads rows of strings from an SQL database\.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)