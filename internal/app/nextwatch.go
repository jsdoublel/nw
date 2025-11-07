package app

import (
	"errors"
	"fmt"
	"iter"
	"math/rand"
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
	watchedFilms WatchedFilms
	watchlist    map[int]*Film
}

// Create NextWatch queue data structure, selecting NumberOfStacks*StackSize+1
// unwatched films from watchlist at random (the plus one is for the next pick
// at the top of the queue).
//
// Returns an error if there is not enough unwatched films in the watchlist.
func (app *Application) MakeNextWatch() (NextWatch, error) {
	pool := make([]*Film, 0, len(app.Watchlist))
	for _, f := range app.Watchlist {
		if _, ok := app.WatchedFilms.Films[f.LBxdID]; !ok {
			pool = append(pool, f)
		}
	}
	if len(pool) < NumberOfStacks*StackSize+1 {
		return NextWatch{}, fmt.Errorf("%w, %d required", ErrNotEnoughFilms, NumberOfStacks*StackSize+1)
	}
	rand.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	stacks := make([][]*Film, NumberOfStacks+1)
	stacks[0] = []*Film{pool[0]}
	for i := range NumberOfStacks {
		stacks[i+1] = make([]*Film, StackSize)
		for j := range StackSize {
			stacks[i+1][j] = pool[StackSize*i+j+1]
		}
	}
	nw := NextWatch{
		Stacks:       stacks,
		watchedFilms: app.WatchedFilms,
		watchlist:    app.Watchlist,
	}
	nw.makeLastUpdate()
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

// Deletes all watched films from queue, replacing them with nil pointers.
func (nw *NextWatch) deleteWatched() {
	for i, j := range nw.Positions() {
		if nw.Stacks[i][j] != nil && nw.watchedFilms.Watched(nw.Stacks[i][j]) {
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
		if !nw.watchedFilms.Watched(f) && !nw.ContainsFilm(*f) {
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
				if poolIdx >= len(pool) {
					return fmt.Errorf("%w, %d required", ErrNotEnoughFilms, NumberOfStacks*StackSize+1)
				}
				nw.Stacks[i][j] = pool[poolIdx]
				nw.lastUpdated[i][j] = true
				poolIdx++
			}
		}
	}
	return nil
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
