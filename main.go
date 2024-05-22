package main

import (
	"flag"
	"fmt"
	"os"
)

func parseArgs() string {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage of %s [flags] <positional args>\n",
			os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(),
			"Flags:\n")
		flag.PrintDefaults()
		// List positional args
		fmt.Fprint(flag.CommandLine.Output(),
			"\nPositional:\n\t[url] letterboxd list url\n")
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	return flag.Args()[0]
}

func main() {
	url := parseArgs()
	urlList, err := ScrapeList(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(urlList)
	id, err := ScrapeFilmID(urlList[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
}
