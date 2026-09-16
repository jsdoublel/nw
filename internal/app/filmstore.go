package app

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

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
}

// Get cached film record, retrieve if necessary
//
// Returns error if it needs to retrieve details and fails.
func (fs *FilmStore) Lookup(film Film) (*FilmRecord, error) {
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

// Clear film records that are not on the main screen.
func (app *Application) CleanFilmStore() {
	cleaned := 0
	visibleFilms := make([]int, 0, NumberOfStacks*StackSize+1)
	for i, j := range app.NWQueue.Positions() {
		if f := app.NWQueue.Stacks[i][j]; f != nil {
			visibleFilms = append(visibleFilms, f.LBxdID)
		}
	}
	for _, l := range app.TrackedLists {
		if f := l.NextFilm; f != nil {
			visibleFilms = append(visibleFilms, f.LBxdID)
		}
	}
	for id := range app.FilmStore.Films {
		if !slices.Contains(visibleFilms, id) {
			delete(app.FilmStore.Films, id)
			cleaned++
		}
	}
	if cleaned > 0 {
		log.Printf("film store: cleaned %d film record(s), %d remain", cleaned, len(app.FilmStore.Films))
	}
}

// retrieve film details. This involves scrapping letterboxd for TMDB id and
// then querying TMDB for details.
func (fs *FilmStore) retrieve(film Film) error {
	fr := &FilmRecord{Film: film}
	var err error
	if fr.TMDBID, err = ScrapeFilmID(film.Url); err != nil {
		return fmt.Errorf("couldn't get TMDB id, %w", err)
	}
	if fr.Details, err = TMDBFilm(fr.TMDBID); err != nil {
		return err
	}
	if fr.ReleaseDate, err = time.Parse("2006-01-02", fr.Details.ReleaseDate); err != nil {
		log.Printf("failed to parse release date %s as time", fr.Details.ReleaseDate)
	}
	fs.Films[fr.LBxdID] = fr
	return nil
}

func (fd *FilmRecord) DirectorString() string {
	crew := fd.Details.Credits.Crew
	directors := make([]string, 0)
	for _, member := range crew {
		if strings.EqualFold(member.Job, "Director") {
			directors = append(directors, member.Name)
		}
	}
	if len(directors) == 0 {
		return ""
	}
	return fmt.Sprintf("dir. %s", strings.Join(directors, ", "))
}
