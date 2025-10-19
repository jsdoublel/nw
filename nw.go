package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

func runNW() int {
	username := parseArgs()
	f, err := tea.LogToFile(filepath.Join(app.NWDataPath, "nw.log"), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error %s\n", err)
		}
	}()
	log.Print("nw starting...")
	application, err := app.Load(username)
	if err != nil {
		log.Printf("error: %v", err)
		return 1
	}
	addList := tui.MakeAddListPane(application)
	p := tea.NewProgram(addList, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("error: %v", err)
	}
	if err = application.Save(); err != nil {
		log.Printf("error: %v", err)
		return 1
	}
	return 0
}

func main() { os.Exit(runNW()) }
