package app

import (
	"errors"
	"fmt"
	"log"
	"os"

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
