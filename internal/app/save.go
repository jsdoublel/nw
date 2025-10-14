package app

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
)

const (
	LatestSaveVersion = 0

	saveDir = "nw"
	saveExt = ".json"
)

type Save struct {
	Application
	Version int // save version, if format changes are made this will be incremented
}

// Save application info to file
func (app *Application) Save() error {
	savePath := savePath(app.User.Name)
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
		return &save.Application, nil
	} else if errors.Is(err, os.ErrNotExist) {
		log.Printf("no save found; creating new user %s", username)
		user, err := makeUser(username)
		if err != nil {
			return nil, err
		}
		return &Application{
			User:      user,
			FilmStore: FilmStore{Films: make(map[int]*FilmRecord)},
		}, nil
	} else {
		return nil, err
	}
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
		if dir, ok := os.LookupEnv("LOCALAPPDATA"); ok { // try a Windows location
			return dir
		}
		panic("HOME is not set")
	}
	dir := filepath.Join(home, "Library", "Application Support") // try macOS location
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return filepath.Join(home, ".local", "share")
}
