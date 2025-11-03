package app

import (
	"log"
	"time"
)

const userDataExpireTime = time.Hour * 24

func (app *Application) Watched(film *Film) bool {
	_, ok := app.WatchedFilms[film.LBxdID]
	return ok
}

// Updates all of the user's watchlist, watched films, and lists
func (app *Application) CheckUserData() error {
	log.Print("updating user data...")
	if err := app.updateWatchlist(); err != nil {
		return err
	}
	if err := app.updateWatchedFilms(); err != nil {
		return err
	}
	if err := app.updateListHeaders(); err != nil {
		return err
	}
	app.UserDataChecked = time.Now()
	return nil
}

func (app *Application) updateWatchlist() error {
	log.Print("updating watchlist")
	watchlist, err := CreateWatchlist(app.Username)
	if err != nil {
		return err
	}
	app.FilmStore.DeregisterSet(app.Watchlist)
	app.FilmStore.RegisterSet(watchlist)
	app.Watchlist = watchlist
	return nil
}

func (app *Application) updateWatchedFilms() error {
	log.Print("updating watched films")
	watchedFilms, err := CreateWatchedFilms(app.Username)
	if err != nil {
		return err
	}
	app.FilmStore.DeregisterSet(app.WatchedFilms)
	app.FilmStore.RegisterSet(watchedFilms)
	app.WatchedFilms = watchedFilms
	return nil
}

func (app *Application) updateListHeaders() error {
	log.Print("updating user's lists")
	headers, err := ScapeUserLists(app.Username)
	if err != nil {
		return err
	}
	app.ListHeaders = headers
	return nil
}
