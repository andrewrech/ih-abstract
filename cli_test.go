package main

import "testing"

func TestUsage(t *testing.T) {

	usage()

	_ = flagParse()

	printConf()

}
