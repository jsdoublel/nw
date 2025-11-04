package app

import "fmt"

// A set of films that can be queried (such as a watchlist or watched films)
type FilmSet struct {
	Url   string        // letterboxd list url
	Films map[int]*Film // letterboxd id -> film struct
}

// Struct storing data for film,
type Film struct {
	LBxdID int    // letterboxd film id (used as unique identifier here)
	Url    string // letterboxd url
	Title  string // film title
	Year   uint   // release year
}

func (f Film) String() string {
	return fmt.Sprintf("%s (%d)", f.Title, f.Year)
}
