package main

import (
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLocateConfig(t *testing.T) {

	os.Setenv("XDG_CONFIG_HOME", "ih-abstractTestdirectory")
	os.Setenv("HOME", "ih-abstractTestdirectory")

	f, _ := locateDefaultConfig()

	if f != "" {
		t.Error("", f)
	}

}

func TestLoadConfig(t *testing.T) {

	vars, err := loadConfig("ih-abstract.yml")
	if err != nil {
		log.Fatalln(err)
	}

	tests := map[string]struct {
		got  string
		want string
	}{
		"Username": {got: vars.Username, want: "username"},
		"Password": {got: vars.Password, want: "password"},
		"Host":     {got: vars.Host, want: "host"},
		"Port":     {got: vars.Port, want: "443"},
		"Database": {got: vars.Database, want: "database"},
		"Query":    {got: vars.Query, want: "SELECT TOP 100 FROM table"},
	}

	for name, tc := range tests {
		name := name
		tc := tc

		t.Run(name, func(t *testing.T) {

			got := tc.got

			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}

}
