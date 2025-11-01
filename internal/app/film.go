package app

import "fmt"

// User film list
type FilmList struct {
	Name     string  // name of list on letterboxd
	Desc     string  // description of list
	Url      string  // letterboxd list url
	NumFilms int     // number of films in list
	Ordered  bool    // is the list ordered
	Films    []*Film // films in list (can be nil)
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
