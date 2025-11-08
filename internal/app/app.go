package app

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

var NWDataPath string

func init() {
	NWDataPath = filepath.Join(getDirBase(), "nw")
	if _, err := os.Stat(NWDataPath); os.IsNotExist(err) {
		if err := os.MkdirAll(NWDataPath, 0o755); err != nil {
			log.Printf("could not create directory at %s", NWDataPath)
		}
	}
}

type Application struct {

	// ----- stuff from letterboxd

	Username     string        // username on letterboxd
	ListHeaders  []*FilmList   // lists that belong to user on letterboxd (without scrapped films)
	Watchlist    map[int]*Film // users letterboxd watchlist
	WatchedFilms WatchedFilms  // users list of watched films on letterboxd

	// ----- tracked by app

	NWQueue         NextWatch
	TrackedLists    map[string]*FilmList // lists tracked in this program; urls are keys
	FilmStore       FilmStore            // central structure that stores local film information
	UserDataChecked time.Time            // last time watchlist, watched films, etc. were checked
}

// Tasks to run on application startup goes here (e.g., checking letterboxd for updated data).
func (app *Application) Init() error {
	if time.Since(app.UserDataChecked) > userDataExpireTime {
		if err := app.UpdateUserData(); err != nil {
			return err
		}
		app.UserDataChecked = time.Now()
	}
	return nil
}

// Run application shutdown tasks (e.g., write save).
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
