package app

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

var (
	ErrDuplicateList  = errors.New("duplicate list")
	ErrListEmpty      = errors.New("list is empty")
	ErrNoValidFilm    = errors.New("no valid film")
	ErrListNotTracked = errors.New("list not tracked")
)

type WatchChecker interface {
	Watched(*Film) bool
}

// Film list that user might track
type FilmList struct {
	Name     string       // name of list on letterboxd
	Desc     string       // description of list
	Url      string       // letterboxd list url
	NumFilms int          // number of films in list
	Ordered  bool         // is the list ordered
	NextFilm *Film        // the next film to be suggested
	Films    []*Film      // films in list (can be nil)
	store    WatchChecker // for checking whether film is watched
}

// Changed Ordered status; clears NextFilm
func (fl *FilmList) ToggleOrdered() {
	fl.Ordered = !fl.Ordered
	fl.NextFilm = nil // next film needs to be recalculated
}

// Suggest next unwatched film to watch from list.
//
// Returns NextFilm if it has not been watched and it has been set. Otherwise
// recalculate the next film to watch. If ordered, it is simply the first
// unwatched film; otherwise, list is shuffled and first unwatched film is
// selected.
//
// If the list is empty, the function returns ErrListEmpty. If all the films
// are watched, then ErrNoValidFilm is returned.
func (fl *FilmList) NextWatch() (Film, error) {
	if len(fl.Films) == 0 {
		return Film{}, ErrListEmpty
	}
	if fl.NextFilm != nil && !fl.store.Watched(fl.NextFilm) {
		return *fl.NextFilm, nil
	}
	var tmpList []*Film
	if !fl.Ordered {
		tmpList = make([]*Film, len(fl.Films))
		copy(tmpList, fl.Films)
		rand.Shuffle(len(fl.Films), func(i, j int) {
			tmpList[i], tmpList[j] = tmpList[j], tmpList[i]
		})
	} else {
		tmpList = fl.Films
	}
	for _, f := range tmpList {
		if !fl.store.Watched(f) {
			fl.NextFilm = f
			return *f, nil
		}
	}
	fl.NextFilm = nil
	return Film{}, fmt.Errorf("%w, no unwatched films in %s", ErrNoValidFilm, fl.Name)
}

// ----- Global list tracking

// Saves list in map of list tracked by the user.
func (app *Application) AddList(filmList *FilmList) error {
	if filmList.Films == nil && filmList.NumFilms > 0 {
		if err := app.AddListFromUrl(filmList.Url); err != nil {
			return err
		}
		return nil
	}
	filmList.store = app.WatchedFilms
	app.FilmStore.RegisterList(filmList)
	app.TrackedLists[filmList.Url] = filmList
	return nil
}

// Remove list from the map of list traced by the user.
func (app *Application) RemoveList(filmList *FilmList) error {
	fl, ok := app.TrackedLists[filmList.Url]
	if !ok {
		return ErrListNotTracked
	}
	delete(app.TrackedLists, fl.Url)
	app.FilmStore.DeregisterList(fl)
	return nil
}

// Checks if list is traced by user.
func (app *Application) IsListTracked(url string) bool {
	_, ok := app.TrackedLists[url]
	return ok
}

// Starts tracking the list corresponding to the given url.
func (app *Application) AddListFromUrl(url string) error {
	if _, ok := app.TrackedLists[url]; ok {
		return ErrDuplicateList
	}
	if !strings.Contains(url, "/list/") {
		return fmt.Errorf("%w, not a regular letterboxd list", ErrInvalidUrl)
	}
	list, err := ScrapeFilmList(url) // TODO: make goroutine
	if err != nil {
		return fmt.Errorf("could not add list %s, %w", list.Name, err)
	}
	return app.AddList(&list)
}
