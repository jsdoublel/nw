package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsdoublel/nw/internal/app"
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
	application, err := app.Load(username)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	addList := tui.MakeAddListPane(application)
	p := tea.NewProgram(addList, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("error: %s", err)
	}
	if err = application.Save(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
