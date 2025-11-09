package app

import (
	"errors"
	"fmt"
	"iter"
	"log"
	"math/rand"
	"time"
)

const (
	NumberOfStacks = 5
	StackSize      = 5
)

var (
	ErrNotEnoughFilms = errors.New("not enough films in watchlist")
	ErrFilmNotFound   = errors.New("film not found")
)

type NextWatch struct {
	Stacks       [][]*Film
	lastUpdated  [][]bool // position changed in last update
	watchedFilms FilmsSet
	watchlist    FilmsSet
	store        *FilmStore
}

// Create NextWatch queue data structure, selecting NumberOfStacks*StackSize+1
// unwatched films from watchlist at random (the plus one is for the next pick
// at the top of the queue).
//
// Returns an error if there is not enough unwatched films in the watchlist.
func (app *Application) MakeNextWatch() (NextWatch, error) {
	stacks := make([][]*Film, NumberOfStacks+1)
	stacks[0] = make([]*Film, 1)
	for i := range NumberOfStacks {
		stacks[i+1] = make([]*Film, StackSize)
	}
	nw := NextWatch{
		Stacks:       stacks,
		watchedFilms: app.WatchedFilms,
		watchlist:    app.Watchlist,
		store:        &app.FilmStore,
	}
	nw.makeLastUpdate()
	if err := nw.update(); err != nil {
		return NextWatch{}, err
	}
	nw.ClearLastUpdated()
	return nw, nil
}

// Remove stack from Next Watch queue from given stack and stack index.
func (nw *NextWatch) DeleteFilm(film Film) error {
	deleted := false
	for i, j := range nw.Positions() {
		if nw.Stacks[i][j].LBxdID == film.LBxdID {
			nw.Stacks[i][j] = nil
			deleted = true
			break
		}
	}
	if !deleted {
		return fmt.Errorf("%w, %s", ErrFilmNotFound, film.Title)
	}
	return nw.update()
}

// Update Next Watch queue by removing watched films.
func (nw *NextWatch) UpdateWatched() error {
	nw.deleteWatched()
	return nw.update()
}

// Deletes all watched films from queue, replacing them with nil pointers. Also
// removes films no longer in watchlist.
func (nw *NextWatch) deleteWatched() {
	for i, j := range nw.Positions() {
		if nw.Stacks[i][j] != nil &&
			(nw.watchedFilms.InSet(nw.Stacks[i][j]) || !nw.watchlist.InSet(nw.Stacks[i][j])) {
			nw.Stacks[i][j] = nil
		}
	}
}

// Fill empty spots in queue as per random stack logic.
func (nw *NextWatch) update() error {
	if nw.Full() { // do nothing if nw is already full
		return nil
	}
	nw.ClearLastUpdated()
	pool := make([]*Film, 0, len(nw.watchlist))
	for _, f := range nw.watchlist {
		if !nw.watchedFilms.InSet(f) && !nw.ContainsFilm(*f) {
			pool = append(pool, f)
		}
	}
	rand.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	poolIdx := 0
	for !nw.Full() {
		for i, j := range nw.Positions() {
			if nw.Stacks[i][j] == nil && i != NumberOfStacks {
				r := rand.Intn(StackSize)
				nw.Stacks[i][j] = nw.Stacks[i+1][r]
				nw.Stacks[i+1][r] = nil
				nw.lastUpdated[i][j] = true
			} else if nw.Stacks[i][j] == nil {
				for {
					if poolIdx >= len(pool) {
						return fmt.Errorf("%w, %d required", ErrNotEnoughFilms, NumberOfStacks*StackSize+1)
					}
					if nw.filterFilm(*pool[poolIdx]) {
						break
					}
					poolIdx++
				}
				nw.Stacks[i][j] = pool[poolIdx]
				nw.lastUpdated[i][j] = true
				poolIdx++
			}
		}
	}
	return nil
}

// Filter out films we don't want in the Next Watch queue by checking details
// from TMDB.
//
// Filters out TV shows, and unreleased films. Also filters anything that
// cannot be retrieve for any other reason (excluding API errors).
func (nw *NextWatch) filterFilm(film Film) bool {
	f, err := nw.store.Lookup(film)
	if errors.Is(err, ErrFailedTMDBLookup) { // TV shows fail by default
		log.Printf("%s, excluding film %s", err, film)
		return false
	}
	if errors.Is(err, ErrNoAPI) {
		log.Printf("%s, proceeding without checks to add film %s to next watch queue", err, film)
		return true
	}
	if f.ReleaseDate.IsZero() {
		log.Printf("invalid release date for %s, %s, excluding film", film, err)
		return false
	}
	if f.ReleaseDate.After(time.Now()) {
		log.Printf("excluding film %s, it has not been released", film)
		return false
	}
	log.Printf("%s added to next watch queue", film)
	return true
}

// Checks if all stack positions have a film in them
func (nw *NextWatch) Full() bool {
	for i, j := range nw.Positions() {
		if nw.Stacks[i][j] == nil {
			return false
		}
	}
	return true
}

// Checks if Next Watch queue contains a given film (by letterboxd id).
func (nw *NextWatch) ContainsFilm(film Film) bool {
	for i, j := range nw.Positions() {
		if nw.Stacks[i][j] != nil && nw.Stacks[i][j].LBxdID == film.LBxdID {
			return true
		}
	}
	return false
}

func (nw *NextWatch) LastUpdated(i, j int) bool {
	return nw.lastUpdated[i][j]
}

func (nw *NextWatch) ClearLastUpdated() {
	for i, j := range nw.Positions() {
		nw.lastUpdated[i][j] = false
	}
}

func (nw *NextWatch) makeLastUpdate() {
	nw.lastUpdated = make([][]bool, NumberOfStacks+1)
	nw.lastUpdated[0] = []bool{false}
	for i := range NumberOfStacks {
		nw.lastUpdated[i+1] = make([]bool, StackSize)
	}
}

// Iterator over valid i, j pairs in Stacks
func (nw *NextWatch) Positions() iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		if !yield(0, 0) {
			return
		}
		for i := 1; i <= NumberOfStacks; i++ {
			for j := range StackSize {
				if !yield(i, j) {
					return
				}
			}
		}
	}
}
