package filmdata

import "fmt"

type Film struct {
	Url     string
	Name    string
	Year    uint
	TMDBID  uint
	Details *FilmDetails
}

type FilmDetails struct {
}

func (f Film) String() string {
	return fmt.Sprintf("%s (%d)", f.Name, f.Year)
}

// func MakeFilm(filmUrl string) (*Film, error) {
// 	film := Film{Url: filmUrl}
// 	qd, err := FilmQuickDetails(filmUrl)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get film details, %w", err)
// 	}
// 	film.Name = qd.Name
// 	film.Year = uint(qd.ReleaseYear)
// 	return &film, nil
// }
