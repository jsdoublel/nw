package app

import (
	"fmt"
	"log"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

const filmExpireTime = 30 * 24 * time.Hour // film records are deleted after 30 days

// Keeps track of all films that are currently in memory so we do not duplicate
// scraping TMDB ids and TMDB api calls.
type FilmStore struct {
	Films map[int]*FilmRecord // Film records index by letterboxd ids
}

type FilmRecord struct {
	Film
	TMDBID      int                // tmdb id number
	Details     *tmdb.MovieDetails // film details from tmdb
	ReleaseDate time.Time          // release date (according to tmdb)
	Watched     bool               // film is recorded as watched

	// These fields are only exported so they can be martialed. Please don't mutate.

	Checked time.Time // last time details were checked
	NRefs   uint      // number of list references
}

// Add film list to be tracked. Films in registered lists will be saved/stored
// in save data as long as they have references.
func (fs *FilmStore) RegisterList(filmList *FilmList) {
	for _, film := range filmList.Films {
		fs.register(*film)
	}
}

// Stop tracking list and decrement ref counts as necessary.
func (fs *FilmStore) DeregisterList(filmList *FilmList) {
	for _, film := range filmList.Films {
		fs.deregister(*film)
	}
}

// Add film set to be tracked (such as watchlist or watched films). Films in
// registered set will be saved/stored in save data as long as they have
// references.
func (fs *FilmStore) RegisterSet(filmSet map[int]*Film) {
	for _, film := range filmSet {
		fs.register(*film)
	}
}

// Stop tracking set and decrement ref counts as necessary.
func (fs *FilmStore) DeregisterSet(filmSet map[int]*Film) {
	for _, film := range filmSet {
		fs.deregister(*film)
	}
}

// Get cached film record, retrieve if necessary
//
// Returns error if it needs to retrieve details and fails.
func (fs *FilmStore) Lookup(film Film) (*FilmRecord, error) {
	if f, ok := fs.Films[film.LBxdID]; ok && time.Since(f.Checked) < filmExpireTime {
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
		if fr.NRefs == 0 {
			delete(fs.Films, id)
		}
	}
}

// register a film to be tracked (or increase ref counter if already registered)
func (fs *FilmStore) register(film Film) {
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
func (fs *FilmStore) deregister(film Film) {
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
func (fs *FilmStore) retrieve(film Film) error {
	fr, ok := fs.Films[film.LBxdID]
	if !ok { // new tmp record if one does not exist
		fr = &FilmRecord{
			Film:  film,
			NRefs: 0,
		}
		fs.Films[film.LBxdID] = fr
	}
	if fr.Details != nil && time.Since(fr.Checked) < filmExpireTime {
		return nil
	}
	if fr.TMDBID == 0 {
		var err error
		fr.TMDBID, err = ScrapeFilmID(film.Url)
		if err != nil {
			return fmt.Errorf("couldn't get TMDB id, %w", err)
		}
	}
	var err error
	fr.Details, err = TMDBFilm(fr.TMDBID)
	if err != nil {
		return err
	}
	fr.ReleaseDate, err = time.Parse("2006-01-02", fr.Details.ReleaseDate)
	if err != nil {
		log.Printf("failed to parse release date %s as time", fr.Details.ReleaseDate)
	}
	fr.Checked = time.Now()
	return nil
}
