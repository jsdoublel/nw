package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jsdoublel/nw/internal/tui"
)

func parseArgs() string {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage of nw [flags] <positional args>\n")
		fmt.Fprint(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\nPositional:\n\t[username] letterboxd username\n")
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
		fmt.Fprintf(os.Stderr, "nw failed with error, %s", err.Error())
	}
}
