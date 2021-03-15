package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
)

// printConf prints an example SQL database configuration file
func printConf() {
	config := `---
username: username
password: password
host: host
port: 443
database: database
query: SELECT TOP 100 FROM table
...`
	fmt.Println(config)
}

// confVars is a struct of configuration variables required for the SQL database connection.
type confVars struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Query    string `yaml:"query"`
}

// locateDefaultConfig locates the configuration file in $XDG_CONFIG_HOME, $HOME, or the current directory.
func locateDefaultConfig() (config string, err error) {
	defaultConfigDir, defaultConfigDirSet := os.LookupEnv("XDG_CONFIG_HOME")

	// first check $XDG_CONFIG_HOME
	if defaultConfigDirSet {
		var XDGDirConfig strings.Builder
		XDGDirConfig.WriteString(defaultConfigDir)
		XDGDirConfig.WriteString("/")
		XDGDirConfig.WriteString("ih-abstract")
		XDGDirConfig.WriteString("/")
		XDGDirConfig.WriteString("ih-abstract.yml")
		config = XDGDirConfig.String()

		if _, err := os.Stat(config); !os.IsNotExist(err) {
			return config, nil
		}
	}

	// then check $HOME
	homeDir, homeDirSet := os.LookupEnv("HOME")
	if homeDirSet {
		var homeDirConfig strings.Builder
		homeDirConfig.WriteString(homeDir)
		homeDirConfig.WriteString("/")
		homeDirConfig.WriteString(".ih-abstract.yml")
		config = homeDirConfig.String()

		if _, err := os.Stat(config); !os.IsNotExist(err) {
			return config, nil
		}

	}

	// then check the current working directory
	if _, err := os.Stat(".ih-abstract.yml"); !os.IsNotExist(err) {
		return ".ih-abstract.yml", nil
	}

	return "", errors.New("Cannot locate SQL database configuration file at default locations $XDG_CONFIG_HOME/ih-abstract/ih-abstract.yml, $HOME/.ih-abstract.yml, and ./ih-abstract.yml")
}

func loadConfig(config string) (vars confVars, err error) {
	y, err := ioutil.ReadFile(config)
	if err != nil {
		return vars, err
	}

	err = yaml.Unmarshal(y, &vars)
	if err != nil {
		return vars, err
	}

	return vars, nil
}
