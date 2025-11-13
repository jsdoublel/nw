package app

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

var NWDataPath string

func init() {
	NWDataPath = filepath.Join(getDataDirBase(), "nw")
	if _, err := os.Stat(NWDataPath); os.IsNotExist(err) {
		if err := os.MkdirAll(NWDataPath, 0o755); err != nil {
			log.Printf("could not create directory at %s", NWDataPath)
		}
	}
}

type Application struct {

	// ----- stuff from letterboxd

	Username     string      // username on letterboxd
	ListHeaders  []*FilmList // lists that belong to user on letterboxd (without scrapped films)
	Watchlist    FilmsSet    // users letterboxd watchlist
	WatchedFilms FilmsSet    // users list of watched films on letterboxd

	// ----- tracked by app

	NWQueue         NextWatch
	TrackedLists    map[string]*FilmList // lists tracked in this program; urls are keys
	FilmStore       FilmStore            // central structure that stores local film information
	UserDataChecked time.Time            // last time watchlist, watched films, etc. were checked

	// ----- tracked processes
	DiscordRPC DiscordRPC
}

// Run application shutdown tasks (e.g., write save).
func (app *Application) Shutdown() {
	app.StopDiscordRPC()
	app.FilmStore.Clean()
	if err := app.Save(); err != nil {
		log.Printf("application save had error %s", err)
	}
}

// Gets location for nw data folder (used for save data and logging)
func getDataDirBase() string {
	if dir, ok := os.LookupEnv("NW_DATA_HOME"); ok {
		return dir
	}
	return xdg.DataHome
}
