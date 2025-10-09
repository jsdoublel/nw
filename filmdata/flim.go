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
