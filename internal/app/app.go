package app

import (
	"errors"
	"fmt"
	"log"
)

var ErrDuplicateList error = errors.New("duplicate list")

type Application struct {

	// ----- stuff from letterboxd

	Username     string      // username on letterboxd
	ListHeaders  []*FilmList // lists that belong to user on letterboxd (without scrapped films)
	Watchlist    *FilmList   // users letterboxd watchlist
	WatchedFilms *FilmList   // users list of watched films on letterboxd

	// ----- tracked by app

	TrackedLists map[string]*FilmList // lists tracked in this program; urls are keys
	FilmStore    FilmStore            // centeral structure that stores local film information
}

func (app *Application) Shutdown() {
	app.FilmStore.Clean()
	if err := app.Save(); err != nil {
		log.Printf("application save had error %s", err)
	}
}

// Saves list to be tracked
func (app *Application) AddList(filmList *FilmList) error {
	if _, ok := app.TrackedLists[filmList.Url]; ok {
		return ErrDuplicateList
	}
	var list FilmList
	if filmList.Films == nil {
		var err error
		if list, err = ScrapeFilmList(filmList.Url); err != nil { // TODO: make goroutine
			return fmt.Errorf("could not add list %s, %w", list.Name, err)
		}
	}
	app.FilmStore.RegisterList(&list)
	app.TrackedLists[filmList.Url] = &list
	return nil
}

func (app *Application) IsListTracked(filmList *FilmList) bool {
	_, ok := app.TrackedLists[filmList.Url]
	return ok
}
