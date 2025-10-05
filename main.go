package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jsdoublel/lbxdr/filmdata"
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
	lists, err := filmdata.ScapeUserLists(username)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	fmt.Println(lists)
	// urlList, err := filmdata.ScrapeList(url)
	// if err != nil {
	// 	panic(err)
	// }
	// // fmt.Println(urlList)
	// for _, url := range urlList {
	// 	id, err := filmdata.ScrapeFilmID(url)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	film, err := filmdata.TMDBFilm(id)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("Title: %s, Year:  %s\n", film.Title, film.ReleaseDate)
	// }
}
