package model

import "fmt"

type FilmList struct {
	Name    string
	ListUrl string
	Films   []*Film
}

func (fl FilmList) String() string {
	return fmt.Sprintf("{%s: %s}", fl.Name, fl.ListUrl)
}

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
