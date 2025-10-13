package app

import (
	"fmt"
	"net/url"

	m "github.com/jsdoublel/nw/internal/model"
)

type User struct {
	Name        string
	ListHeaders []*m.FilmList
	Lists       []*m.FilmList
	Watchlist   *m.FilmList
	Films       *m.FilmList
	// nwQueue   NextWatch
}

// creates user, scrapping all user information
//
// TODO: multithread
func makeUser(username string) (User, error) {
	headers, err := m.ScapeUserLists(username)
	if err != nil {
		return User{}, err
	}
	wlUrl, err := url.JoinPath(m.LetterboxdUrl, username, "watchlist")
	if err != nil {
		return User{}, err
	}
	watchlist, err := m.ScrapeFilmList(wlUrl)
	if err != nil {
		return User{}, err
	} else if watchlist.Name != "" {
		return User{}, fmt.Errorf("watchlist had unexpected name %s", watchlist.Name)
	}
	watchlist.Name = "Watchlist"
	fUrl, err := url.JoinPath(m.LetterboxdUrl, username, "films")
	if err != nil {
		return User{}, err
	}
	films, err := m.ScrapeFilmList(fUrl)
	if err != nil {
		return User{}, err
	} else if films.Name != "" {
		return User{}, fmt.Errorf("films list had unexpected name %s", watchlist.Name)
	}
	films.Name = "Watched"
	return User{
		Name:        username,
		ListHeaders: headers,
		Watchlist:   &watchlist,
		Films:       &films,
	}, nil
}
