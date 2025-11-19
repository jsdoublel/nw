/*
NW - Next Watch: A TUI utility for selecting films to watch from Letterboxd
(powered by TMDB).

Copyright (C) 2025 James Willson <jsdoublel@gmail.com>

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <https://www.gnu.org/licenses/>.

This software uses TMDB and the TMDB APIs but is not endorsed, certified, or
otherwise approved by TMDB.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jsdoublel/nw/internal/app"
	"github.com/jsdoublel/nw/internal/tui"
)

func parseArgs() string {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "Usage: nw [flags]\n\nFlags:\n") // nolint
		flag.PrintDefaults()
	}
	username := flag.String("u", "", "letterboxd username (overrides config)")
	config := flag.Bool("c", false, "prints expected config path and exits")
	version := flag.Bool("v", false, "prints version and exits")
	help := flag.Bool("h", false, "prints this message and exits")
	flag.Parse()
	if *config {
		fmt.Printf("nw expects config at %s\n", app.ConfigPath())
		os.Exit(0)
	}
	if *version {
		fmt.Printf("nw version %s\n", app.Version)
		os.Exit(0)
	}
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	return *username
}

func main() {
	username := parseArgs()
	if username == "" {
		username = app.Config.Username
	}
	if username == "" {
		fmt.Println("No username provided (try `nw -h` for help)")
		os.Exit(1)
	}
	if err := tui.RunApplicationTUI(username); err != nil {
		fmt.Fprintf(os.Stderr, "nw failed with error, %s\n", err.Error())
		os.Exit(1)
	}
}
