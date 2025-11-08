/*
NW - Next Watch: A TUI utility for selecting films to watch from letterboxd.

Copyright (C) 2025 James Willson <jamessw156@gmail.com>

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <https://www.gnu.org/licenses/>.
*/
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
		fmt.Fprintf(os.Stderr, "nw failed with error, %s\n", err.Error())
	}
}
