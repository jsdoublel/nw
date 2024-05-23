package filmdata

import (
	"fmt"
	"os"

	"github.com/cyruzin/golang-tmdb"
)

var TMDB_CLIENT *tmdb.Client = nil

func tmdbClient() (*tmdb.Client, error) {
	if TMDB_CLIENT != nil  {
		return TMDB_CLIENT, nil
	}
	TMDB_CLIENT, err := tmdb.Init(os.Getenv("TMDB_API"))
	if err != nil {
		return nil, fmt.Errorf("Error: Could not connect to API. %s\n", err)
	}
	TMDB_CLIENT.SetClientAutoRetry()
	return TMDB_CLIENT, nil
}

func TMDBFilm(id int) (*tmdb.MovieDetails, error) {
	client, err := tmdbClient()
	if err != nil {
		return nil, err
	}
	film, err := client.GetMovieDetails(id, nil)
	if err != nil {
		return nil, fmt.Errorf("Error: Could not retrieve film. %s\n", err)
	}
	return film, nil
}

