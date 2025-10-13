package app

import (
	"fmt"
	"log"

	m "github.com/jsdoublel/nw/internal/model"
)

type Application struct {
	User      User
	FilmStore FilmStore
}

func (app *Application) Shutdown() {
	app.FilmStore.Clean()
	if err := app.Save(); err != nil {
		log.Printf("application save had error %s", err)
	}
}

// Saves list to be tracked
func (app *Application) AddList(filmList *m.FilmList) error {
	if filmList.Films == nil {
		if list, err := m.ScrapeFilmList(filmList.Url); err != nil { // TODO: make goroutine
			return fmt.Errorf("could not add list %s", list.Name)
		}
	}
	app.FilmStore.RegisterList(filmList)
	app.User.Lists = append(app.User.Lists, filmList)
	return nil
}
