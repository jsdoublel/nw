package app

import (
	"fmt"
	"log"
	"net/url"
)

// create application on first startup, scrapping all user information
//
// TODO: multithread
func createApp(username string) (*Application, error) {
	log.Print("scrape user lists")
	headers, err := ScapeUserLists(username)
	if err != nil {
		return nil, err
	}
	log.Print("scrape user watchlist")
	watchlist, err := CreateWatchlist(username)
	if err != nil {
		return nil, err
	}
	log.Print("scrape user watched films")
	films, err := CreateWatchedFilms(username)
	if err != nil {
		return nil, err
	}
	fs := FilmStore{Films: make(map[int]*FilmRecord)}
	fs.RegisterSet(watchlist)
	fs.RegisterSet(films)
	return &Application{
		Username:     username,
		ListHeaders:  headers,
		Watchlist:    watchlist,
		WatchedFilms: films,
		TrackedLists: make(map[string]*FilmList),
		FilmStore:    fs,
	}, nil
}

func CreateWatchlist(username string) (map[int]*Film, error) {
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
	return filmListToMap(&watchlist), nil
}

func CreateWatchedFilms(username string) (map[int]*Film, error) {
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
	return filmListToMap(&films), nil
}

func filmListToMap(filmList *FilmList) map[int]*Film {
	filmSet := make(map[int]*Film)
	for _, f := range filmList.Films {
		filmSet[f.LBxdID] = f
	}
	return filmSet
}
