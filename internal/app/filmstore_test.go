package app

import (
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
)

// newEmptyStacks builds an empty NextWatch.Stacks with the same shape
// MakeNextWatch produces (CleanFilmStore indexes it directly by position).
func newEmptyStacks() [][]*Film {
	stacks := make([][]*Film, NumberOfStacks+1)
	stacks[0] = make([]*Film, 1)
	for i := range NumberOfStacks {
		stacks[i+1] = make([]*Film, StackSize)
	}
	return stacks
}

func TestApplicationCleanFilmStore(t *testing.T) {
	testCases := []struct {
		name          string
		stacksSetup   func(stacks [][]*Film)
		trackedLists  map[string]*FilmList
		films         map[int]*FilmRecord
		wantRemaining []int
	}{
		{
			name: "retains film visible in a non-top Stacks slot",
			stacksSetup: func(stacks [][]*Film) {
				stacks[1][0] = &Film{LBxdID: 1}
			},
			films:         map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}}},
			wantRemaining: []int{1},
		},
		{
			name: "retains film visible at top-of-queue slot",
			stacksSetup: func(stacks [][]*Film) {
				stacks[0][0] = &Film{LBxdID: 2}
			},
			films:         map[int]*FilmRecord{2: {Film: Film{LBxdID: 2}}},
			wantRemaining: []int{2},
		},
		{
			name: "retains film visible via tracked list NextFilm",
			trackedLists: map[string]*FilmList{
				"list": {NextFilm: &Film{LBxdID: 3}},
			},
			films:         map[int]*FilmRecord{3: {Film: Film{LBxdID: 3}}},
			wantRemaining: []int{3},
		},
		{
			name:          "removes film not referenced anywhere",
			films:         map[int]*FilmRecord{4: {Film: Film{LBxdID: 4}}},
			wantRemaining: []int{},
		},
		{
			name: "nil Stacks slots do not panic",
			stacksSetup: func(stacks [][]*Film) {
				// Most slots stay nil; only one is populated.
				stacks[3][2] = &Film{LBxdID: 5}
			},
			films: map[int]*FilmRecord{
				5: {Film: Film{LBxdID: 5}},
				6: {Film: Film{LBxdID: 6}},
			},
			wantRemaining: []int{5},
		},
		{
			name: "nil NextFilm in TrackedLists does not panic",
			trackedLists: map[string]*FilmList{
				"empty":  {NextFilm: nil},
				"filled": {NextFilm: &Film{LBxdID: 7}},
			},
			films: map[int]*FilmRecord{
				7: {Film: Film{LBxdID: 7}},
				8: {Film: Film{LBxdID: 8}},
			},
			wantRemaining: []int{7},
		},
		{
			name: "mixed roots, including a film reachable from two roots",
			stacksSetup: func(stacks [][]*Film) {
				stacks[0][0] = &Film{LBxdID: 9}
				stacks[2][1] = &Film{LBxdID: 10}
			},
			trackedLists: map[string]*FilmList{
				"list": {NextFilm: &Film{LBxdID: 9}}, // also visible via Stacks[0][0]
			},
			films: map[int]*FilmRecord{
				9:  {Film: Film{LBxdID: 9}},
				10: {Film: Film{LBxdID: 10}},
				11: {Film: Film{LBxdID: 11}},
			},
			wantRemaining: []int{9, 10},
		},
		{
			name:          "empty film store",
			films:         map[int]*FilmRecord{},
			wantRemaining: []int{},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			stacks := newEmptyStacks()
			if test.stacksSetup != nil {
				test.stacksSetup(stacks)
			}
			films := map[int]*FilmRecord{}
			for id, record := range test.films {
				r := *record
				films[id] = &r
			}
			app := &Application{
				NWQueue:      NextWatch{Stacks: stacks},
				TrackedLists: test.trackedLists,
				FilmStore:    FilmStore{Films: films},
			}
			app.CleanFilmStore()
			if len(app.FilmStore.Films) != len(test.wantRemaining) {
				t.Fatalf("expected %d remaining records, got %d", len(test.wantRemaining), len(app.FilmStore.Films))
			}
			for _, id := range test.wantRemaining {
				if _, ok := app.FilmStore.Films[id]; !ok {
					t.Fatalf("expected record %d to remain", id)
				}
			}
		})
	}
}

func TestFilmStoreLookup(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[int]*FilmRecord
		film     Film
		expected *FilmRecord // test only checks title, and FilmRecord specific fields
		wantErr  bool
	}{
		{
			name: "returns existing record",
			existing: map[int]*FilmRecord{
				1: func() *FilmRecord {
					r := &FilmRecord{Film: Film{LBxdID: 1, Title: "Stored"}}
					r.Details = &tmdb.MovieDetails{ID: 1, Title: "Stored"}
					return r
				}(),
			},
			film:     Film{LBxdID: 1},
			expected: &FilmRecord{Film: Film{LBxdID: 1, Title: "Stored"}},
			wantErr:  false,
		},
		{
			name:     "gets new record",
			existing: map[int]*FilmRecord{},
			film: Film{
				Url:    "https://letterboxd.com/film/dancer-in-the-dark/",
				LBxdID: 2701,
				Title:  "Dancer in the Dark",
				Year:   2000,
			},
			expected: &FilmRecord{
				Film: Film{
					Url:    "https://letterboxd.com/film/dancer-in-the-dark/",
					LBxdID: 2701,
					Title:  "Dancer in the Dark",
					Year:   2000,
				},
			},
		},
		{
			name:     "returns error when retrieval fails",
			existing: map[int]*FilmRecord{},
			film:     Film{LBxdID: 2, Url: "https://example.com/not-letterboxd"},
			expected: nil,
			wantErr:  true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fs := &FilmStore{Films: map[int]*FilmRecord{}}
			for id, record := range test.existing {
				r := *record
				fs.Films[id] = &r
			}
			record, err := fs.Lookup(test.film)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if record != nil {
					t.Fatalf("expected no record but got one")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if record == nil {
				t.Fatalf("expected record but got nil")
			}
			if record.LBxdID != test.film.LBxdID {
				t.Fatalf("expected LBxdID %d, got %d", test.film.LBxdID, record.LBxdID)
			}
			if record.Title != test.expected.Title {
				t.Fatalf("expected Title %s, got %s", test.expected.Title, record.Title)
			}
			if record.Details == nil {
				t.Fatalf("details nil for requested film record, %s", record.Title)
			}
			if record.Details.Title != test.expected.Title {
				t.Fatalf("unexpected title, got %s != want %s", record.Details.Title, test.expected.Title)
			}
		})
	}
}
