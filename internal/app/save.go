package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	LatestSaveVersion  = 0
	userDataExpireTime = time.Hour * 24

	saveExt = ".json"
)

// ----- Save functionality

type Save struct {
	Application
	Version int // save version, if format changes are made this will be incremented
}

// Save application info to file
func (app *Application) Save() error {
	savePath := savePath(app.Username)
	if _, err := os.Stat(filepath.Dir(savePath)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
			return err
		}
	}
	bytes, err := json.Marshal(Save{Application: *app, Version: LatestSaveVersion})
	if err != nil {
		return err
	}
	if err = os.WriteFile(savePath, bytes, 0o644); err != nil {
		return err
	}
	return nil
}

// Creates application struct. First tries to load user from save file;
// otherwise, it creates new user and filmstore.
func Load(username string) (*Application, error) {
	save := savePath(username)
	if _, err := os.Stat(save); err == nil {
		log.Printf("save found at %s, loading...", save)
		bytes, err := os.ReadFile(save)
		if err != nil {
			return nil, err
		}
		var save Save
		if err := json.Unmarshal(bytes, &save); err != nil {
			return nil, err
		}
		app := &save.Application
		app.rehydrate()
		return app, nil
	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("no save found; creating new user %s", username)
		app, err := CreateApp(username)
		if err != nil {
			return nil, err
		}
		return app, nil
	} else {
		return nil, err
	}
}

// Get save path name from username
func savePath(username string) string {
	return filepath.Join(NWDataPath, username+saveExt)
}

// Post JSON unmarshal setup
func (app *Application) rehydrate() {
	for _, list := range app.TrackedLists {
		list.store = app.WatchedFilms
	}
	app.NWQueue.watchedFilms = app.WatchedFilms
	app.NWQueue.watchlist = app.Watchlist
}

// ----- Update user data etc.

// Create application on first startup, scrapping all user information.
func CreateApp(username string) (*Application, error) {
	app := &Application{
		Username:     username,
		TrackedLists: make(map[string]*FilmList),
		FilmStore:    FilmStore{Films: make(map[int]*FilmRecord)},
	}
	if err := app.UpdateUserData(); err != nil {
		return nil, err
	}
	var err error
	if app.NWQueue, err = app.MakeNextWatch(); err != nil {
		return nil, err
	}
	return app, nil
}

// Updates all of the user's watchlist, watched films, and lists
func (app *Application) UpdateUserData() error {
	log.Print("updating user data...")
	if err := app.updateListHeaders(); err != nil {
		return err
	}
	if err := app.updateWatchlist(); err != nil {
		return err
	}
	if err := app.updateWatchedFilms(); err != nil {
		return err
	}
	if app.NWQueue.Stacks != nil {
		if err := app.NWQueue.UpdateWatched(); err != nil {
			log.Print(err)
		}
	}
	app.UserDataChecked = time.Now()
	return nil
}

func (app *Application) updateWatchlist() error {
	log.Print("updating watchlist")
	watchlist, err := retrieveWatchlist(app.Username)
	if err != nil {
		return err
	}
	if app.Watchlist != nil {
		app.FilmStore.DeregisterSet(app.Watchlist)
	}
	app.FilmStore.RegisterSet(watchlist)
	app.Watchlist = watchlist
	return nil
}

func (app *Application) updateWatchedFilms() error {
	log.Print("updating watched films")
	watchedFilms, err := retrieveWatchedFilms(app.Username)
	if err != nil {
		return err
	}
	if !app.WatchedFilms.IsZero() {
		app.FilmStore.DeregisterSet(app.WatchedFilms.Films)
	}
	app.FilmStore.RegisterSet(watchedFilms)
	app.WatchedFilms = WatchedFilms{Films: watchedFilms}
	for _, v := range app.TrackedLists {
		v.store = app.WatchedFilms
	}
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

// Retrieve watchlist from letterbxod
func retrieveWatchlist(username string) (map[int]*Film, error) {
	wlUrl, err := url.JoinPath(LetterboxdUrl, username, "watchlist")
	if err != nil {
		return nil, err
	}
	watchlist, err := ScrapeFilmList(wlUrl)
	if err != nil {
		return nil, err
	} else if watchlist.Name != "" {
		return nil, fmt.Errorf("watchlist had unexpected name %s", watchlist.Name)
	}
	watchlist.Name = "Watchlist"
	return filmListToMap(watchlist), nil
}

// Retrieved watched films from letterboxd
func retrieveWatchedFilms(username string) (map[int]*Film, error) {
	fUrl, err := url.JoinPath(LetterboxdUrl, username, "films")
	if err != nil {
		return nil, err
	}
	films, err := ScrapeFilmList(fUrl)
	if err != nil {
		return nil, err
	} else if films.Name != "" {
		return nil, fmt.Errorf("films list had unexpected name %s", films.Name)
	}
	films.Name = "Watched"
	return filmListToMap(films), nil
}

// Convert a FilmList struct to a map from letterboxd ids to films
func filmListToMap(filmList FilmList) map[int]*Film {
	filmSet := make(map[int]*Film)
	for _, f := range filmList.Films {
		filmSet[f.LBxdID] = f
	}
	return filmSet
}
