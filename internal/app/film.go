package app

import "fmt"

// Struct storing data for film,
type Film struct {
	LBxdID int    // letterboxd film id (used as unique identifier here)
	Url    string // letterboxd url
	Title  string // film title
	Year   uint   // release year
}

func (f Film) String() string {
	if f.Year == 0 {
		return f.Title
	}
	return fmt.Sprintf("%s (%d)", f.Title, f.Year)
}

type FilmsSet map[int]*Film

func (fs FilmsSet) InSet(film *Film) bool {
	_, ok := fs[film.LBxdID]
	return ok
}

func (fs FilmsSet) IsZero() bool {
	return fs == nil
}
