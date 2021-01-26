package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
)

// connect connects to an SQL database.
func connect(config string) (db *sql.DB, err error) {

	if config == "" {
		config, err = locateDefaultConfig()

		if err != nil {
			log.Fatalln(err)
		}
	}

	vars, err := loadConfig(config)
	if err != nil {
		return db, err
	}

	c := url.Values{}
	c.Add("Database", vars.Database)
	c.Add("Trusted_Connection", "Yes")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(vars.Username, vars.Password),
		Host:     fmt.Sprintf("%s:%s", vars.Host, vars.Port),
		RawQuery: c.Encode(),
	}

	db, err = sql.Open("sqlserver", u.String())

	return db, err

}

// rawRecords contains a header, a channel of raw records, and a channel indicating when raw records have been read.
type rawRecords struct {
	header []string
	out    chan []string
	done   chan struct{}
}

// DB reads records from an Sql database.
func DB(config string, db *sql.DB) (r rawRecords) {
	defer db.Close()

	var err error

	if config == "" {
		config, err = locateDefaultConfig()

		if err != nil {
			log.Fatalln(err)
		}
	}

	vars, err := loadConfig(config)

	// sql query
	rows, err := db.Query(vars.Query)

	if err != nil {
		log.Fatalln("failed to run query", err)
	}

	r = readSQLRows(rows)

	return r
}
