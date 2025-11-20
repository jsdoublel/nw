package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cyruzin/golang-tmdb"
)

var (
	TMDBClient *tmdb.Client = nil

	ErrNoAPI            = errors.New("could not connect to TMDB api")
	ErrFailedTMDBLookup = errors.New("failed TMDB lookup")
)

func (app *Application) ApiInit() { // prefers config key if valid
	var err error
	if TMDBClient, err = tmdb.Init(app.ApiKey); err != nil {
		return
	}
	TMDBClient.SetClientAutoRetry()
}

func TMDBFilm(id int) (*tmdb.MovieDetails, error) {
	if TMDBClient == nil {
		return nil, ErrNoAPI
	}
	film, err := TMDBClient.GetMovieDetails(id, map[string]string{
		"append_to_response": "credits",
	})
	if err != nil {
		return nil, fmt.Errorf("%w, with id %d, %w", ErrFailedTMDBLookup, id, err)
	}
	return film, nil
}

// Queries TMDB for movies matching the given search string.
func SearchFilms(query string) ([]tmdb.MovieResult, error) {
	if TMDBClient == nil {
		return nil, ErrNoAPI
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	res, err := TMDBClient.GetSearchMovies(q, map[string]string{
		"include_adult": "false", "page": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("tmdb search failed, %w", err)
	}
	return res.Results, nil
}

func ReleaseYear(mr tmdb.MovieResult) (int, error) {
	releaseDate, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return 0, fmt.Errorf("error parsing date for film results %s, %s", mr.Title, mr.ReleaseDate)
	}
	return releaseDate.Year(), nil
}
