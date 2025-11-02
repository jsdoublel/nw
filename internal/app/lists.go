package app

import (
	"fmt"
	"strings"
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

func (app *Application) IsListTracked(filmList *FilmList) bool {
	_, ok := app.TrackedLists[filmList.Url]
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
		fl.Ordered = !fl.Ordered
		return
	}
	panic(fmt.Sprintf("Tried to toggle ordered on untracked film list %s", filmList.Name))
}
