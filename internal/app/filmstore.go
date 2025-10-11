package app

import (
	"fmt"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"

	m "github.com/jsdoublel/nw/internal/model"
)

// Keeps track of all films that are currently in memory so we do not duplicate
// scraping TMDB ids and TMDB api calls.
type FilmStore struct {
	Films map[uint]*FilmRecord // Film records index by letterboxd ids
}

type FilmRecord struct {
	m.Film
	TMDBID  uint               // tmdb id number
	Details *tmdb.MovieDetails // film details from tmdb
	// These fields are only exported so they can be martialed. Please don't mutate.
	Checked time.Time // last time details were checked
	NRefs   uint      // number of list references
}

// Add film list to be tracked. Films in registered lists will be saved/stored
// in save data as long as they have references.
func (fs *FilmStore) AddList(filmList *m.FilmList) error {
	return nil
}

// Stop tracking list and decrement ref counts as necessary.
func (fs *FilmStore) RemoveList(filmList *m.FilmList) error {
	return nil
}

// Get cached film record, retrieve if necessary
//
// Returns error if it needs to retrieve details and fails.
func (fs *FilmStore) Lookup(film m.Film) (*FilmRecord, error) {
	if f, ok := fs.Films[film.LBxdID]; ok {
		return f, nil
	}
	if err := fs.retrieve(film); err != nil {
		return nil, err
	}
	if f, ok := fs.Films[film.LBxdID]; ok {
		return f, nil
	}
	panic("film record not found after adding it even though Add() returned err=nil")
}

// register a film to be tracked
func (fs *FilmStore) register(film *m.Film) error {
	return nil
}

// retrieve film details. This involves scrapping letterboxd for TMDB id and
// then querying TMDB for details.
//
// panics if used when details already existed (use lookup).
func (fs *FilmStore) retrieve(film m.Film) error {
	if f, ok := fs.Films[film.LBxdID]; ok && !f.Checked.IsZero() {
		panic(fmt.Sprintf("tried to retrieve details for %s, but details already stored", film))
	}
	tmdbID, err := m.ScrapeFilmID(film.Url)
	if err != nil {
		return fmt.Errorf("couldn't get TMDB id, %w", err)
	}
	details, err := m.TMDBFilm(tmdbID)
	if err != nil {
		return err
	}
	if fr, ok := fs.Films[film.LBxdID]; ok {
		fr.Details = details
	} else { // temp record, since it has no references it will get cleared
		fs.Films[film.LBxdID] = &FilmRecord{
			Film:    film,
			TMDBID:  uint(tmdbID),
			Details: details,
			Checked: time.Now(),
			NRefs:   0,
		}
	}
	return nil
}

func (fs *FilmStore) Remove(film m.Film) {
}

func (fs *FilmStore) Delete(film m.Film) {
}
