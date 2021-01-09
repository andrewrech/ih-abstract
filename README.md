[![GoDoc](https://godoc.org/github.com/andrewrech/ih-abstract?status.svg)](https://godoc.org/github.com/andrewrech/ih-abstract) [![](https://goreportcard.com/badge/github.com/andrewrech/ih-abstract)](https://goreportcard.com/report/github.com/andrewrech/ih-abstract) ![](https://img.shields.io/badge/docker-andrewrech/ih-abstract:0.0.4-blue?style=plastic&logo=docker)

# ih-abstract

## Description

`ih-abstract` streams input raw pathology results to the immune.health.report R package for report generation and quality assurance. The input is .csv data or direct streaming from a Microsoft SQL driver-compatible database. The output is filtered .csv/.txt files for incremental new report generation and quality assurance.

Optionally, Immune Health-specific filtering can be turned off to use ih-abstract as a general method to retrieve arbitrary or incremental pathology results.

## Installation

See [Releases](https://github.com/andrewrech/ih-abstract/releases).

```zsh
go get -u -v github.com/andrewrech/ih-abstract
```

Cross compiling for Alpine requires disabling CGO:

```zsh
 echo env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ih-abstract
```

## Usage

See `ih-abstract -h` or [documentation](https://github.com/andrewrech/ih-abstract/blob/main/docs.md).

## Authors

- [Andrew J. Rech](mailto:rech@rech.io)

## License

GNU Lesser General Public License v3.0
