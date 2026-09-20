package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	LatestSaveVersion  = 0
	userDataExpireTime = time.Hour * 24

	lastUserFile  = "lastusername.txt"
	saveExt       = ".json.gz"
	legacySaveExt = ".json"
)

var (
	ErrLoading         = errors.New("failed to load save")
	ErrSaving          = errors.New("failed to write save")
	ErrFailedMigration = errors.New("failed to migrate legacy save")
)

// Conditions to scrape Letterboxd for user data update
type CheckUpdateCondition int

const (
	UpdateAlwaysCheck = iota
	UpdateExpiredCheck
	UpdateNeverCheck
)

// ----- Save functionality

type Save struct {
	Application
	Version int // save version, if format changes are made this will be incremented
}

// Save application info to file
func (app *Application) Save() error {
	savePath := savePath(app.Username)
	if err := os.Rename(savePath, savePath+".bak"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("failed to backup save, %s", err)
	}
	bytes, err := json.Marshal(Save{Application: *app, Version: LatestSaveVersion})
	if err != nil {
		return fmt.Errorf("failed to marshal save data, %w", err)
	}
	out, err := os.CreateTemp(NWDataPath, filepath.Base(savePath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("%w, could not create file, %w", ErrSaving, err)
	}
	defer func() { _ = out.Close() }()
	gw := gzip.NewWriter(out)
	if _, err := gw.Write(bytes); err != nil {
		return fmt.Errorf("%w, failed to write save file, %w", ErrSaving, err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("%w, could not close gzip writer, %w", ErrSaving, err)
	}
	if err := os.Rename(out.Name(), savePath); err != nil {
		return fmt.Errorf("%w, error overwriting save with tmp, %w", ErrSaving, err)
	}
	log.Printf("application data saved to %s", savePath)
	return nil
}

// Creates application struct. First tries to load user from save file;
// otherwise, it creates new user and filmstore.
func Load(username string) (*Application, error) {
	path := savePath(username)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if migrated, err := migrateLegacySave(username); !migrated { // changes save path if true
			if err != nil {
				return nil, fmt.Errorf("%w, %w", ErrFailedMigration, err)
			}
			log.Printf("no save found; creating new user %s", username)
			app, err := CreateApp(username)
			if err != nil {
				return nil, err
			}
			return app, nil
		}
	} else if err != nil {
		return nil, err
	}
	log.Printf("save found at %s, loading...", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w, could not open save, %w", ErrLoading, err)
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("%w, could not create gzip reader, %w", ErrLoading, err)
	}
	defer func() { _ = gz.Close() }()
	bytes, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("%w, could not read gzip, %w", ErrLoading, err)
	}
	var save Save
	if err := json.Unmarshal(bytes, &save); err != nil {
		return nil, fmt.Errorf("%w, could not unmarshal json, %w", ErrLoading, err)
	}
	app := &save.Application
	app.rehydrate()
	return app, nil
}

// Check data directory for possible save files with legacy names, then moves
// them to the proper name/type.
//
// Legacy files may be in mixed case, and will have the extension .json
func migrateLegacySave(username string) (bool, error) {
	legacyName := fmt.Sprintf("%s%s", username, legacySaveExt) // maybe the wrong case
	canonicalName := fmt.Sprintf("%s%s", username, saveExt)
	entries, err := os.ReadDir(NWDataPath)
	if err != nil {
		return false, fmt.Errorf("failed to read NWDataPath, %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(legacyName, e.Name()) {
			legacySave := filepath.Join(NWDataPath, e.Name())
			bytes, err := os.ReadFile(legacySave)
			if err != nil {
				return false, fmt.Errorf("could not read %s, %w", e.Name(), err)
			}
			out, err := os.Create(filepath.Join(NWDataPath, canonicalName))
			if err != nil {
				return false, fmt.Errorf("could not create %s, %w", canonicalName, err)
			}
			defer func() { _ = out.Close() }()
			gw := gzip.NewWriter(out)
			if _, err := gw.Write(bytes); err != nil {
				return false, fmt.Errorf("could not write %s, %w", canonicalName, err)
			}
			if err := gw.Close(); err != nil {
				return false, fmt.Errorf("could not close %s, %w", canonicalName, err)
			}
			if err := os.Remove(legacySave); err != nil {
				log.Printf("failed to remove legacy save, %s", err.Error())
			}
			return true, nil
		}
	}
	return false, nil
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
			*username = string(bytes.ToLower(bytes.TrimSpace(content)))
		}
	}
	if *username == "" && askUser != nil {
		*username = askUser()
	}
	if *username != "" {
		if err := os.WriteFile(lastUsernameFile, []byte(*username), 0o0644); err != nil {
			log.Printf("error storing last username, %s", err)
		}
	} else {
		return errors.New("no username provided")
	}
	*username = strings.ToLower(*username)
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
func (app *Application) UpdateUserData(check CheckUpdateCondition) (err error) {
	// TODO: Probably best to move this logic to internal/tui
	if check == UpdateNeverCheck || check == UpdateExpiredCheck && (time.Since(app.UserDataChecked) < userDataExpireTime || Config.Features.DisableStartupUpdate) {
		return nil
	}
	log.Printf("updating user data (check=%v)...", check)
	defer func() { log.Printf("updateUserData: finished, err=%v", err) }()
	if err = app.updateListHeaders(); err != nil {
		return err
	}
	if err = app.updateWatchlist(); err != nil {
		return err
	}
	if err = app.updateWatchedFilms(); err != nil {
		return err
	}
	if err = app.updateNextWatchQueue(); err != nil {
		return err
	}
	if err = app.updateTrackedLists(false); err != nil {
		return err
	}
	app.UserDataChecked = time.Now()
	err = app.Save()
	return err
}

// Updates user Next Watch Queue (or creates it if it does not exist)
func (app *Application) updateNextWatchQueue() error {
	if app.NWQueue.Stacks == nil {
		log.Print("creating next watch queue...")
		var err error
		if app.NWQueue, err = app.MakeNextWatch(); err != nil {
			return fmt.Errorf("failed to create next watch queue: %w", err)
		}
		return nil
	}
	log.Print("updating next watch queue...")
	app.NWQueue.watchlist = app.Watchlist
	app.NWQueue.watchedFilms = app.WatchedFilms
	if err := app.NWQueue.UpdateWatched(); err != nil {
		return fmt.Errorf("failed to update next watch queue: %w", err)
	}
	return nil
}

func (app *Application) updateWatchlist() error {
	log.Print("updating watchlist")
	watchlist, err := retrieveWatchlist(app.Username)
	if err != nil {
		return fmt.Errorf("failed to update watchlist: %w", err)
	}
	app.Watchlist = watchlist
	log.Printf("watchlist updated: %d films", len(watchlist))
	return nil
}

func (app *Application) updateWatchedFilms() error {
	log.Print("updating watched films")
	watchedFilms, err := retrieveWatchedFilms(app.Username)
	if err != nil {
		return fmt.Errorf("failed to update watched films: %w", err)
	}
	app.WatchedFilms = watchedFilms
	for _, v := range app.TrackedLists {
		v.watched = app.WatchedFilms
	}
	log.Printf("watched films updated: %d films", len(watchedFilms))
	return nil
}

func (app *Application) updateListHeaders() error {
	log.Print("updating user's lists")
	headers, err := ScrapeUserLists(app.Username)
	if err != nil {
		return fmt.Errorf("failed to update user's lists: %w", err)
	}
	app.ListHeaders = headers
	return nil
}

// Updates the data in tracked film lists. Skips lists where the next up film
// is unwatched (in order avoid excessive overall update times). This behavior
// can be overridden with forceAll.
func (app *Application) updateTrackedLists(forceAll bool) error {
	var lastErr error
	refreshed, skipped, failed := 0, 0, 0
	for _, fl := range app.TrackedLists {
		if fl.NextFilm == nil || app.WatchedFilms.InSet(fl.NextFilm) || forceAll {
			log.Printf("refreshing tracked list %s (nextFilm=%v, forceAll=%v)", fl.Name, fl.NextFilm, forceAll)
			if err := app.RefreshList(fl); err != nil {
				lastErr = err
				failed++
				log.Printf("failed refreshing list %s, %s", fl.Name, err)
			} else {
				refreshed++
			}
		} else {
			skipped++
			log.Printf("skipping tracked list %s refresh (nextFilm=%s not yet watched)", fl.Name, fl.NextFilm)
		}
	}
	log.Printf("updateTrackedLists done: refreshed=%d skipped=%d failed=%d", refreshed, skipped, failed)
	if lastErr != nil {
		return fmt.Errorf("failed to update tracked lists: %w", lastErr)
	}
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
