package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jsdoublel/nw/internal/app"
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
	application, err := app.Load(username)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	fmt.Println(application.User.ListHeaders)
	fmt.Println(application.User.Watchlist.Films)
	if err = application.Save(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
