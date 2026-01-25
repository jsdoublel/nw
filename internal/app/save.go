package app

import (
	"bytes"
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

	lastUserFile = "lastusername.txt"
	saveExt      = ".json"
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

// Retrieve username if it has not been set using a variety of means. askUser
// is a function that can be used to ask the user in some way to enter their
// username if all else fails. GetUser also saves and loads most recently used
// username. Returns an error if no username can ultimately be retrieved.
func GetUser(username *string, askUser func() string) error {
	if *username == "" {
		*username = Config.Username
	}
	lastUsernameFile := filepath.Join(NWDataPath, lastUserFile)
	if *username == "" {
		if content, err := os.ReadFile(lastUsernameFile); err == nil {
			*username = string(bytes.TrimSpace(content))
		}
	}
	if *username == "" && askUser != nil {
		*username = askUser()
	}
	if *username != "" {
		if err := os.WriteFile(lastUsernameFile, []byte(*username), 0o0666); err != nil {
			log.Printf("error storing last username, %s", err)
		}
	} else {
		return errors.New("no username provided")
	}
	return nil
}

// Get save path name from username
func savePath(username string) string {
	return filepath.Join(NWDataPath, username+saveExt)
}

// Post JSON unmarshal setup
func (app *Application) rehydrate() {
	if app.ApiKey == "" {
		app.ApiKey = getAPIKey()
	}
	for _, list := range app.TrackedLists {
		list.watched = app.WatchedFilms
	}
	app.NWQueue.makeLastUpdate()
	app.NWQueue.watchedFilms = app.WatchedFilms
	app.NWQueue.watchlist = app.Watchlist
	app.NWQueue.store = &app.FilmStore
}

func getAPIKey() string {
	if Config.ApiKey != "" {
		return Config.ApiKey
	}
	return os.Getenv("TMDB_API_KEY")
}

// ----- Update user data etc.

// Create application on first startup, scrapping all user information.
func CreateApp(username string) (*Application, error) {
	app := &Application{
		Username:     username,
		TrackedLists: make(map[string]*FilmList),
		FilmStore:    FilmStore{Films: make(map[int]*FilmRecord)},
		ApiKey:       getAPIKey(),
	}
	return app, nil
}

// Updates all of the user's watchlist, watched films, and lists
//
// Argument "check," when true, checks whether previous data has expired---if
// it has not, nothing is done.
func (app *Application) UpdateUserData(check bool) error {
	if check && time.Since(app.UserDataChecked) < userDataExpireTime {
		return nil
	}
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
		app.NWQueue.watchlist = app.Watchlist
		app.NWQueue.watchedFilms = app.WatchedFilms
		if err := app.NWQueue.UpdateWatched(); err != nil {
			log.Print(err)
		}
	} else {
		var err error
		if app.NWQueue, err = app.MakeNextWatch(); err != nil {
			return err
		}
	}
	if err := app.updateTrackedLists(false); err != nil {
		return err
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
	if app.WatchedFilms != nil {
		app.FilmStore.DeregisterSet(app.WatchedFilms)
	}
	app.FilmStore.RegisterSet(watchedFilms)
	app.WatchedFilms = watchedFilms
	for _, v := range app.TrackedLists {
		v.watched = app.WatchedFilms
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

// Updates the data in tracked film lists. Skips lists where the next up film
// is unwatched (in order avoid excessive overall update times). This behavior
// can be overridden with forceAll.
func (app *Application) updateTrackedLists(forceAll bool) error {
	var lastErr error
	for _, fl := range app.TrackedLists {
		if app.WatchedFilms.InSet(fl.NextFilm) || forceAll {
			if err := app.RefreshList(fl); err != nil {
				lastErr = err
				log.Printf("failed refreshing list %s, %s", fl.Name, err)
			}
		}
	}
	return lastErr
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
