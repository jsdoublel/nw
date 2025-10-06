package userdata

import (
	"encoding/json"
	"log"
	"os"
	"path"

	fd "github.com/jsdoublel/nw/filmdata"
)

const (
	LatestSaveVersion = 0

	saveDir = "nw"
	saveExt = ".json"
)

type User struct {
	Name        string
	ListHeaders []fd.FilmListHeader
	// watchlist FilmList
	// films     FilmList
	// nwQueue   NextWatch
}

type Save struct {
	User
	Version int // save version, if format changes are made this will be incremented
}

// Save user info to file
func (u *User) Save() error {
	savePath := savePath(u.Name)
	if _, err := os.Stat(path.Dir(savePath)); os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(savePath), 0o755); err != nil {
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
		log.Println("no save found; creating new user")
		return makeUser(username)
	} else {
		return nil, err
	}
}

func makeUser(username string) (*User, error) {
	headers, err := fd.ScapeUserLists(username)
	if err != nil {
		return nil, err
	}
	return &User{Name: username, ListHeaders: headers}, nil
}

// Get save path name from username
func savePath(username string) string {
	return path.Join(getSaveDirBase(), saveDir, username+saveExt)
}

// Look for save data directory location. First check XDG location, then try
// default XDG location (i.e., ~/.local/share); otherwise, use HOME.
//
// Currently assumes HOME is set (as per POSIX).
func getSaveDirBase() string {
	if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return dir
	}
	home := os.Getenv("HOME")
	if home == "" {
		panic("HOME is not set")
	}
	return path.Join(os.Getenv("HOME"), ".local", "share")
}
