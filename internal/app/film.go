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
	return fmt.Sprintf("%s (%d)", f.Title, f.Year)
}
