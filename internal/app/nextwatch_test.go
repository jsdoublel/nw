package app

import (
	"errors"
	"testing"
)

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
				WatchedFilms: WatchedFilms{Films: watched},
			}
			for i := 0; i < tc.films; i++ {
				id := i + 1
				app.Watchlist[id] = &Film{LBxdID: id}
			}
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
			nw := NextWatch{
				Stacks:       make([][]*Film, NumberOfStacks+1),
				watchedFilms: WatchedFilms{Films: map[int]*Film{}},
				watchlist:    make(map[int]*Film),
			}
			nextID := 1
			nw.Stacks[0] = []*Film{{LBxdID: nextID}}
			nw.watchlist[nextID] = nw.Stacks[0][0]
			for i := 1; i <= NumberOfStacks; i++ {
				nw.Stacks[i] = make([]*Film, StackSize)
				for j := range StackSize {
					nextID++
					f := &Film{LBxdID: nextID}
					nw.Stacks[i][j] = f
					nw.watchlist[nextID] = f
				}
			}
			nextID++
			extra := &Film{LBxdID: nextID}
			nw.watchlist[nextID] = extra
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
			target := &Film{LBxdID: 99}
			watched := make(map[int]*Film)
			if tc.watchedTarget {
				watched[target.LBxdID] = target
			}
			nw := NextWatch{
				Stacks:       make([][]*Film, NumberOfStacks+1),
				watchedFilms: WatchedFilms{Films: watched},
				watchlist:    make(map[int]*Film),
			}
			nw.Stacks[0] = []*Film{{LBxdID: 1}}
			nw.watchlist[1] = nw.Stacks[0][0]
			for i := 1; i <= NumberOfStacks; i++ {
				nw.Stacks[i] = make([]*Film, StackSize)
				for j := range StackSize {
					var f *Film
					if i == NumberOfStacks && j == StackSize-1 {
						f = target
					} else {
						id := (i*StackSize + j + 1) * 10
						f = &Film{LBxdID: id}
					}
					nw.Stacks[i][j] = f
					nw.watchlist[f.LBxdID] = f
				}
			}
			extra := &Film{LBxdID: 1000}
			nw.watchlist[1000] = extra
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
