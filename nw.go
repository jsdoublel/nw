package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jsdoublel/nw/internal/tui"
)

func parseArgs() string {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [flags] <positional args>\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprint(flag.CommandLine.Output(), "\nPositional:\n\t[url] letterboxd list url\n")
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	return flag.Args()[0]
}

func main() {
	username := parseArgs()
	if err := tui.RunApplicationTUI(username); err != nil {
		log.Fatalf("nw failed with error, %s", err.Error())
	}
}
