package app

import (
	"fmt"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"

	m "github.com/jsdoublel/nw/internal/model"
)

const expireTime = 30 * 24 * time.Hour // film records are deleted after 30 days

// Keeps track of all films that are currently in memory so we do not duplicate
// scraping TMDB ids and TMDB api calls.
type FilmStore struct {
	Films map[int]*FilmRecord // Film records index by letterboxd ids
}

type FilmRecord struct {
	m.Film
	TMDBID  int                // tmdb id number
	Details *tmdb.MovieDetails // film details from tmdb
	// These fields are only exported so they can be martialed. Please don't mutate.
	Checked time.Time // last time details were checked
	NRefs   uint      // number of list references
}

// Add film list to be tracked. Films in registered lists will be saved/stored
// in save data as long as they have references.
func (fs *FilmStore) RegisterList(filmList *m.FilmList) {
	for _, film := range filmList.Films {
		fs.register(*film)
	}
}

// Stop tracking list and decrement ref counts as necessary.
func (fs *FilmStore) DeregisterList(filmList *m.FilmList) {
	for _, film := range filmList.Films {
		fs.deregister(*film)
	}
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

// Clear film records that are either not referenced or too old.
func (fs *FilmStore) Clean() {
	for id, fr := range fs.Films {
		if fr.NRefs == 0 || time.Since(fr.Checked) > expireTime {
			delete(fs.Films, id)
		}
	}
}

// register a film to be tracked (or increase ref counter if already registered)
func (fs *FilmStore) register(film m.Film) {
	if fr, ok := fs.Films[film.LBxdID]; ok {
		fr.NRefs++
	} else {
		fs.Films[film.LBxdID] = &FilmRecord{
			Film:    film,
			TMDBID:  0,
			Details: nil,
			Checked: time.Time{}, // zero, we haven't checked
			NRefs:   1,
		}
	}
}

// stop tracking an instance of a film (decrease ref counter)
func (fs *FilmStore) deregister(film m.Film) {
	fr, ok := fs.Films[film.LBxdID]
	if !ok {
		panic(fmt.Sprintf("trying to deregister %s, but it has not been registered", film))
	}
	if fr.NRefs == 0 {
		panic(fmt.Sprintf("cannot decrement number of refs to %s, already 0", film))
	}
	fr.NRefs--
	if fr.NRefs == 0 {
		delete(fs.Films, film.LBxdID)
	}
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
			TMDBID:  tmdbID,
			Details: details,
			Checked: time.Now(),
			NRefs:   0,
		}
	}
	return nil
}
