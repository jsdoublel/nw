package app

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func makeTestNextWatch(t *testing.T, totalFilms int, watched map[int]*Film) NextWatch {
	t.Helper()
	if totalFilms < NumberOfStacks*StackSize+1 {
		t.Fatalf("not enough films to build NextWatch: got %d need %d", totalFilms, NumberOfStacks*StackSize+1)
	}
	if watched == nil {
		watched = make(map[int]*Film)
	}
	watchlist := make(map[int]*Film, totalFilms)
	for id := 1; id <= totalFilms; id++ {
		watchlist[id] = &Film{LBxdID: id}
	}
	app := Application{
		Watchlist:    watchlist,
		WatchedFilms: watched,
		FilmStore:    FilmStore{},
	}
	seedFilmStore(t, &app.FilmStore, watchlist)
	nw, err := app.MakeNextWatch()
	if err != nil {
		t.Fatalf("MakeNextWatch returned error: %v", err)
	}
	return nw
}

func seedFilmStore(t *testing.T, store *FilmStore, films map[int]*Film) {
	t.Helper()
	if store.Films == nil {
		store.Films = make(map[int]*FilmRecord, len(films))
	}
	now := time.Now()
	released := now.Add(-24 * time.Hour)
	for _, film := range films {
		store.Films[film.LBxdID] = &FilmRecord{
			Film:        *film,
			Stats:       LBxdStats{Views: 1},
			ReleaseDate: released,
		}
	}
}

func TestApplicationMakeNextWatch(t *testing.T) {
	testCases := []struct {
		name      string
		films     int
		watched   []int
		wantErr   bool
		wantCount int
	}{
		{
			name:      "builds stacks without panic",
			films:     NumberOfStacks*StackSize + 1,
			watched:   nil,
			wantErr:   false,
			wantCount: NumberOfStacks*StackSize + 1,
		},
		{
			name:      "ignores watched films when building queue",
			films:     NumberOfStacks*StackSize + 4,
			watched:   []int{2, 4, 6},
			wantErr:   false,
			wantCount: NumberOfStacks*StackSize + 1,
		},
		{
			name:      "fails when there are not enough unwatched films",
			films:     NumberOfStacks * StackSize,
			watched:   nil,
			wantErr:   true,
			wantCount: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			watched := make(map[int]*Film)
			for _, id := range tc.watched {
				watched[id] = &Film{LBxdID: id}
			}
			app := Application{
				Watchlist:    make(map[int]*Film),
				WatchedFilms: watched,
				FilmStore:    FilmStore{},
			}
			for i := 0; i < tc.films; i++ {
				id := i + 1
				app.Watchlist[id] = &Film{LBxdID: id}
			}
			seedFilmStore(t, &app.FilmStore, app.Watchlist)
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("MakeNextWatch panicked: %v", r)
				}
			}()
			nw, err := app.MakeNextWatch()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error %v", ErrNotEnoughFilms)
				}
				if !errors.Is(err, ErrNotEnoughFilms) {
					t.Fatalf("expected %v got %v", ErrNotEnoughFilms, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			count := 0
			if len(nw.Stacks) != NumberOfStacks+1 {
				t.Fatalf("stack count %d want %d", len(nw.Stacks), NumberOfStacks+1)
			}
			for i := range nw.Stacks {
				if i == 0 && len(nw.Stacks[i]) != 1 {
					t.Fatalf("stack 0 size %d want 1", len(nw.Stacks[i]))
				}
				if i > 0 && len(nw.Stacks[i]) != StackSize {
					t.Fatalf("stack %d size %d want %d", i, len(nw.Stacks[i]), StackSize)
				}
				for j := range nw.Stacks[i] {
					if nw.Stacks[i][j] == nil {
						t.Fatalf("stack %d index %d is nil", i, j)
					}
					for _, watchedID := range tc.watched {
						if nw.Stacks[i][j].LBxdID == watchedID {
							t.Fatalf("watched film %d was enqueued", watchedID)
						}
					}
					count++
				}
			}
			if count != tc.wantCount {
				t.Fatalf("queue contains %d films want %d", count, tc.wantCount)
			}
			if !nw.Full() {
				t.Fatal("expected queue to be full")
			}
		})
	}
}

func TestNextWatchDeleteFilm(t *testing.T) {
	testCases := []struct {
		name       string
		stackNum   int
		stackIndex int
	}{
		{
			name:       "maintains fullness after deleting head",
			stackNum:   0,
			stackIndex: 0,
		},
		{
			name:       "maintains fullness after deleting nested film",
			stackNum:   2,
			stackIndex: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			const totalFilms = NumberOfStacks*StackSize + 2
			nw := makeTestNextWatch(t, totalFilms, nil)
			removed := nw.Stacks[tc.stackNum][tc.stackIndex]
			if err := nw.DeleteFilm(*nw.Stacks[tc.stackNum][tc.stackIndex]); err != nil {
				t.Fatalf("DeleteFilm returned error: %v", err)
			}
			if !nw.Full() {
				t.Fatal("expected queue to be full after delete")
			}
			if nw.Stacks[tc.stackNum][tc.stackIndex] != nil && nw.Stacks[tc.stackNum][tc.stackIndex].LBxdID == removed.LBxdID {
				t.Fatalf("film %d still present at original position", removed.LBxdID)
			}
		})
	}
}

func TestNextWatchUpdateWatched(t *testing.T) {
	testCases := []struct {
		name          string
		watchedTarget bool
	}{
		{
			name:          "removes watched entries and refills",
			watchedTarget: true,
		},
		{
			name:          "leaves queue intact when nothing watched",
			watchedTarget: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			watched := make(map[int]*Film)
			const totalFilms = NumberOfStacks*StackSize + 2
			nw := makeTestNextWatch(t, totalFilms, watched)
			target := nw.Stacks[NumberOfStacks][StackSize-1]
			if target == nil {
				t.Fatal("expected target film to exist in queue")
			}
			if tc.watchedTarget {
				watched[target.LBxdID] = target
			} else {
				delete(watched, target.LBxdID)
			}
			if err := nw.UpdateWatched(); err != nil {
				t.Fatalf("UpdateWatched returned error: %v", err)
			}
			if tc.watchedTarget && nw.ContainsFilm(*target) {
				t.Fatal("watched film remained in queue")
			}
			if !tc.watchedTarget && !nw.ContainsFilm(*target) {
				t.Fatal("unexpected removal of unwatched film")
			}
			if !nw.Full() {
				t.Fatal("expected queue to be full after update")
			}
		})
	}
}

func TestNextWatchUpdateWatchedKeepsOrder(t *testing.T) {
	testCases := []struct {
		name    string
		watched [][2]int // (stack, index) positions to mark watched
	}{
		{
			name:    "watching next pick",
			watched: [][2]int{{0, 0}},
		},
		{
			name:    "watching nested film",
			watched: [][2]int{{2, 1}},
		},
		{
			name:    "watching multiple films in same stack",
			watched: [][2]int{{1, 0}, {1, 3}},
		},
		{
			name:    "watching films across stacks",
			watched: [][2]int{{0, 0}, {3, 2}, {3, 4}, {NumberOfStacks, 0}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for range 20 { // refill is random, so repeat to cover different pulls
				watched := make(map[int]*Film)
				nw := makeTestNextWatch(t, NumberOfStacks*StackSize+10, watched)
				before := make([][]int, len(nw.Stacks))
				for i, j := range nw.Positions() {
					before[i] = append(before[i], nw.Stacks[i][j].LBxdID)
				}
				for _, p := range tc.watched {
					f := nw.Stacks[p[0]][p[1]]
					watched[f.LBxdID] = f
				}
				if err := nw.UpdateWatched(); err != nil {
					t.Fatalf("UpdateWatched returned error: %v", err)
				}
				for i := 1; i <= NumberOfStacks; i++ {
					after := make([]int, 0, StackSize)
					for _, f := range nw.Stacks[i] {
						after = append(after, f.LBxdID)
					}
					retained := slices.DeleteFunc(slices.Clone(before[i]), func(id int) bool { return !slices.Contains(after, id) })
					if !slices.Equal(after[:len(retained)], retained) {
						t.Fatalf("stack %d: retained films %v not kept in order at top, got %v (before %v)", i, retained, after, before[i])
					}
				}
			}
		})
	}
}
