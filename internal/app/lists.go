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

// Saves list to be tracked
func (app *Application) AddList(filmList *FilmList) error {
	if filmList.Films == nil {
		if err := app.AddListFromUrl(filmList.Url); err != nil {
			return err
		}
		return nil
	}
	app.FilmStore.RegisterList(filmList)
	app.TrackedLists[filmList.Url] = filmList
	return nil
}

func (app *Application) RemoveList(filmList *FilmList) error {
	fl, ok := app.TrackedLists[filmList.Url]
	if !ok {
		return ErrListNotTracked
	}
	delete(app.TrackedLists, fl.Url)
	app.FilmStore.DeregisterList(fl)
	return nil
}

func (app *Application) IsListTracked(url string) bool {
	_, ok := app.TrackedLists[url]
	return ok
}

func (app *Application) AddListFromUrl(url string) error {
	if _, ok := app.TrackedLists[url]; ok {
		return ErrDuplicateList
	}
	if !strings.Contains(url, "/list/") {
		return fmt.Errorf("%w, not a regular letterboxd list", ErrInvalidUrl)
	}
	list, err := ScrapeFilmList(url)
	if err != nil { // TODO: make goroutine
		return fmt.Errorf("could not add list %s, %w", list.Name, err)
	}
	app.FilmStore.RegisterList(&list)
	app.TrackedLists[url] = &list
	return nil
}

func (app *Application) ToggleOrderedList(filmList FilmList) {
	if fl, ok := app.TrackedLists[filmList.Url]; ok {
		fl.ToggleOrdered()
		return
	}
	panic(fmt.Sprintf("Tried to toggle ordered on untracked film list %s", filmList.Name))
}

// Get next film to watch from list
func (app *Application) NextWatchFromList(filmList FilmList) (Film, error) {
	fl, ok := app.TrackedLists[filmList.Url]
	if !ok {
		return Film{}, ErrListNotTracked
	}
	if len(fl.Films) == 0 {
		return Film{}, ErrListEmpty
	}
	if fl.NextFilm != nil && !app.Watched(fl.NextFilm) {
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
		if _, ok := app.WatchedFilms[f.LBxdID]; !ok {
			fl.NextFilm = f
			return *f, nil
		}
	}
	fl.NextFilm = nil
	return Film{}, fmt.Errorf("%w, no unwatched films in %s", ErrNoValidFilm, fl.Name)
}
