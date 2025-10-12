package model

import (
	"fmt"
)

// User film list
type FilmList struct {
	Name  string  // name of list on letterboxd
	Url   string  // letterboxd list url
	Films []*Film // films in list (can be nil)
}

func (fl FilmList) String() string {
	return fmt.Sprintf("{%s: %s}", fl.Name, fl.Url)
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
