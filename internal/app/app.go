package app

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

var (
	NWDataPath string

	ErrDuplicateList  error = errors.New("duplicate list")
	ErrListNotTracked       = errors.New("list is not tracked")
)

func init() {
	NWDataPath = filepath.Join(getDirBase(), "nw")
}

type Application struct {

	// ----- stuff from letterboxd

	Username     string      // username on letterboxd
	ListHeaders  []*FilmList // lists that belong to user on letterboxd (without scrapped films)
	Watchlist    *FilmList   // users letterboxd watchlist
	WatchedFilms *FilmList   // users list of watched films on letterboxd

	// ----- tracked by app

	TrackedLists map[string]*FilmList // lists tracked in this program; urls are keys
	FilmStore    FilmStore            // central structure that stores local film information
}

func (app *Application) Shutdown() {
	app.FilmStore.Clean()
	if err := app.Save(); err != nil {
		log.Printf("application save had error %s", err)
	}
}

// Look for data directory location. First check custom NW_DATA_HOME variable,
// then XDG location, then tries a Windows and macOS location. Finally, if all
// of those fails it returns the default XDG location (i.e., ~/.local/share).
//
// Will panic if HOME is not set and it cannot find LOCALAPPDATA.
func getDirBase() string {
	if dir, ok := os.LookupEnv("NW_DATA_HOME"); ok {
		return dir
	}
	if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return dir
	}
	home, ok := os.LookupEnv("HOME")
	if !ok {
		if dir, ok := os.LookupEnv("LOCALAPPDATA"); ok { // try a Windows location
			return dir
		}
		panic("HOME is not set")
	}
	dir := filepath.Join(home, "Library", "Application Support") // try macOS location
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return filepath.Join(home, ".local", "share")
}
