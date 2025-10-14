package app

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cyruzin/golang-tmdb"
)

var TMDBClient *tmdb.Client = nil

var ErrNoAPI error = errors.New("could not connect to TMDB API")
var ErrFailedTMDBLookup error = errors.New("failed TMDB lookup")

func init() {
	var err error
	if TMDBClient, err = tmdb.Init(os.Getenv("TMDB_API")); err != nil {
		log.Printf("could not connect to TMDB API, %s", err)
	}
	TMDBClient.SetClientAutoRetry()
}

func TMDBFilm(id int) (*tmdb.MovieDetails, error) {
	if TMDBClient == nil {
		return nil, ErrNoAPI
	}
	film, err := TMDBClient.GetMovieDetails(id, nil)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrFailedTMDBLookup, err)
	}
	return film, nil
}
