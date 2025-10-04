package filmdata

import (
	"fmt"
	"os"

	"github.com/cyruzin/golang-tmdb"
)

var TMDBClient *tmdb.Client = nil

func tmdbClient() (*tmdb.Client, error) {
	if TMDBClient != nil {
		return TMDBClient, nil
	}
	var err error
	TMDBClient, err = tmdb.Init(os.Getenv("TMDB_API"))
	if err != nil {
		return nil, fmt.Errorf("could not connect to API, %s", err)
	}
	TMDBClient.SetClientAutoRetry()
	return TMDBClient, nil
}

func TMDBFilm(id int) (*tmdb.MovieDetails, error) {
	client, err := tmdbClient()
	if err != nil {
		return nil, err
	}
	film, err := client.GetMovieDetails(id, nil)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve film, %s", err)
	}
	return film, nil
}
