package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	m "github.com/jsdoublel/nw/internal/model"
	w "github.com/jsdoublel/nw/internal/web"
)

const (
	LatestSaveVersion = 0

	saveDir = "nw"
	saveExt = ".json"
)

type User struct {
	Name        string
	ListHeaders []*m.FilmList
	Watchlist   *m.FilmList
	Films       *m.FilmList
	// nwQueue   NextWatch
}

type Save struct {
	User
	Version int // save version, if format changes are made this will be incremented
}

// Save user info to file
func (u *User) Save() error {
	savePath := savePath(u.Name)
	if _, err := os.Stat(filepath.Dir(savePath)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
			return err
		}
	}
	bytes, err := json.Marshal(Save{User: *u, Version: LatestSaveVersion})
	if err != nil {
		return err
	}
	if err = os.WriteFile(savePath, bytes, 0o644); err != nil {
		return err
	}
	return nil
}

// Creates user struct. First tries to load user from save file; otherwise, it
// creates new user.
func LoadUser(username string) (*User, error) {
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
		return &save.User, nil
	} else if os.IsNotExist(err) {
		log.Printf("no save found; creating new user %s", username)
		return makeUser(username)
	} else {
		return nil, err
	}
}

// creates user, scrapping all user information
//
// TODO: multithread
func makeUser(username string) (*User, error) {
	headers, err := w.ScapeUserLists(username)
	if err != nil {
		return nil, err
	}
	wlUrl, err := url.JoinPath(w.LetterboxdUrl, username, "watchlist")
	if err != nil {
		return nil, err
	}
	watchlist, err := w.ScrapeFilmList(wlUrl)
	if err != nil {
		return nil, err
	} else if watchlist.Name != "" {
		return nil, fmt.Errorf("watchlist had unexpected name %s", watchlist.Name)
	}
	watchlist.Name = "Watchlist"
	fUrl, err := url.JoinPath(w.LetterboxdUrl, username, "films")
	if err != nil {
		return nil, err
	}
	films, err := w.ScrapeFilmList(fUrl)
	if err != nil {
		return nil, err
	} else if films.Name != "" {
		return nil, fmt.Errorf("films list had unexpected name %s", watchlist.Name)
	}
	films.Name = "Watched"
	return &User{
		Name:        username,
		ListHeaders: headers,
		Watchlist:   &watchlist,
		Films:       &films,
	}, nil
}

// Get save path name from username
func savePath(username string) string {
	return filepath.Join(getSaveDirBase(), saveDir, username+saveExt)
}

// Look for save data directory location. First check custom NW_DATA_HOME
// variable, then XDG location, then tries a Windows and macOS location.
// Finally, if all of those fails it returns the default XDG location (i.e.,
// ~/.local/share).
//
// Will panic if HOME is not set and it cannot find LOCALAPPDATA.
func getSaveDirBase() string {
	if dir, ok := os.LookupEnv("NW_DATA_HOME"); ok {
		return dir
	}
	if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return dir
	}
	home, ok := os.LookupEnv("HOME")
	if !ok {
		if dir, ok := os.LookupEnv("APPDATA"); ok { // try a Windows location
			return dir
		}
		panic("HOME is not set")
	}
	dir := filepath.Join(home, "Library", "Application Support") // try macOS location
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return filepath.Join(os.Getenv("HOME"), ".local", "share")
}
