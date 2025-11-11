package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cyruzin/golang-tmdb"
)

var (
	TMDBClient *tmdb.Client = nil

	ErrNoAPI            = errors.New("could not connect to TMDB api")
	ErrFailedTMDBLookup = errors.New("failed TMDB lookup")
)

func init() {
	var err error
	if TMDBClient, err = tmdb.Init(os.Getenv("TMDB_API")); err != nil {
		log.Printf("%s, %s", ErrNoAPI, err)
	} else {
		TMDBClient.SetClientAutoRetry()
	}
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
		"include_adult": "true", "page": "1",
		"append_to_response": "credits",
	})
	if err != nil {
		return nil, fmt.Errorf("tmdb search failed, %w", err)
	}
	return res.Results, nil
}

func StringFromMovieResult(mr tmdb.MovieResult) (string, error) {
	releaseDate, err := time.Parse("2006-01-02", mr.ReleaseDate)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s (%d)", mr.Title, releaseDate.Year()), nil
}
