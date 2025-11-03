package app

import "fmt"

// Film list that user might track
type FilmList struct {
	Name     string  // name of list on letterboxd
	Desc     string  // description of list
	Url      string  // letterboxd list url
	NumFilms int     // number of films in list
	Ordered  bool    // is the list ordered
	NextFilm *Film   // the next film to be suggested
	Films    []*Film // films in list (can be nil)
}

// Changed Ordered status
func (fl *FilmList) ToggleOrdered() {
	fl.Ordered = !fl.Ordered
	fl.NextFilm = nil
}

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
