package app

import (
	"testing"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

func TestFilmStoreRegisterList(t *testing.T) {
	testCases := []struct {
		name     string
		existing map[int]*FilmRecord
		list     *FilmList
		wantRefs map[int]uint
	}{
		{
			name:     "registers new film",
			existing: map[int]*FilmRecord{},
			list: &FilmList{
				Films: []*Film{{LBxdID: 1, Url: "https://letterboxd.com/film/example", Title: "Example", Year: 2000}},
			},
			wantRefs: map[int]uint{1: 1},
		},
		{
			name: "increments existing reference",
			existing: map[int]*FilmRecord{
				1: {
					Film:    Film{LBxdID: 1, Url: "https://letterboxd.com/film/example", Title: "Example", Year: 2000},
					NRefs:   1,
					Checked: time.Now(),
				},
			},
			list: &FilmList{Films: []*Film{
				{LBxdID: 1, Url: "https://letterboxd.com/film/example", Title: "Example", Year: 2000}},
			},
			wantRefs: map[int]uint{1: 2},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fs := &FilmStore{Films: map[int]*FilmRecord{}}
			for id, record := range test.existing {
				r := *record
				fs.Films[id] = &r
			}
			fs.RegisterList(test.list)
			if len(fs.Films) != len(test.wantRefs) {
				t.Fatalf("expected %d records, got %d", len(test.wantRefs), len(fs.Films))
			}
			for id, refs := range test.wantRefs {
				record, ok := fs.Films[id]
				if !ok {
					t.Fatalf("missing record %d", id)
				}
				if record.NRefs != refs {
					t.Fatalf("expected %d references, got %d", refs, record.NRefs)
				}
				if record.Checked != (time.Time{}) && record.Checked.IsZero() {
					t.Fatalf("checked time should not be zero when set")
				}
			}
		})
	}
}

func TestFilmStoreDeregisterList(t *testing.T) {
	testCases := []struct {
		name       string
		existing   map[int]*FilmRecord
		list       *FilmList
		wantRefs   map[int]uint
		wantPanic  bool
		wantExists map[int]bool
	}{
		{
			name:       "decrements references",
			existing:   map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}, NRefs: 2, Checked: time.Now()}},
			list:       &FilmList{Films: []*Film{{LBxdID: 1}}},
			wantRefs:   map[int]uint{1: 1},
			wantExists: map[int]bool{1: true},
		},
		{
			name:       "removes when count reaches zero",
			existing:   map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}, NRefs: 1, Checked: time.Now()}},
			list:       &FilmList{Films: []*Film{{LBxdID: 1}}},
			wantRefs:   map[int]uint{},
			wantExists: map[int]bool{1: false},
		},
		{
			name:       "panics when film missing",
			existing:   map[int]*FilmRecord{},
			list:       &FilmList{Films: []*Film{{LBxdID: 42}}},
			wantPanic:  true,
			wantRefs:   map[int]uint{},
			wantExists: map[int]bool{},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fs := &FilmStore{Films: map[int]*FilmRecord{}}
			for id, record := range test.existing {
				r := *record
				fs.Films[id] = &r
			}
			defer func() {
				r := recover()
				if test.wantPanic {
					if r == nil {
						t.Fatalf("expected panic but none occurred")
					}
				} else if r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()
			fs.DeregisterList(test.list)
			if test.wantPanic {
				return
			}
			for id, expected := range test.wantExists {
				record, ok := fs.Films[id]
				if expected && !ok {
					t.Fatalf("expected record %d to remain", id)
				}
				if !expected && ok {
					t.Fatalf("expected record %d to be removed", id)
				}
				if expected {
					if record.NRefs != test.wantRefs[id] {
						t.Fatalf("expected %d refs, got %d", test.wantRefs[id], record.NRefs)
					}
				}
			}
		})
	}
}

func TestFilmStoreClean(t *testing.T) {
	testCases := []struct {
		name   string
		record map[int]*FilmRecord
		want   map[int]bool
	}{
		{
			name:   "removes unreferenced film",
			record: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}, Checked: time.Now()}},
			want:   map[int]bool{1: false},
		},
		{
			name:   "retains expired film with refs",
			record: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}, NRefs: 1, Checked: time.Now().Add(-filmExpireTime - time.Second)}},
			want:   map[int]bool{1: true},
		},
		{
			name:   "retains active film",
			record: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1}, NRefs: 1, Checked: time.Now()}},
			want:   map[int]bool{1: true},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fs := &FilmStore{Films: map[int]*FilmRecord{}}
			for id, record := range test.record {
				r := *record
				fs.Films[id] = &r
			}
			fs.Clean()
			for id, expected := range test.want {
				_, ok := fs.Films[id]
				if expected && !ok {
					t.Fatalf("expected record %d", id)
				}
				if !expected && ok {
					t.Fatalf("unexpected record %d", id)
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
					r := &FilmRecord{Film: Film{LBxdID: 1, Title: "Stored"}, NRefs: 1, Checked: time.Now()}
					r.Details = &tmdb.MovieDetails{ID: 1, Title: "Stored"}
					return r
				}(),
			},
			film:     Film{LBxdID: 1},
			expected: &FilmRecord{Film: Film{LBxdID: 1, Title: "Stored"}, NRefs: 1, Checked: time.Now()},
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
				Checked: time.Now(),
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
			if record.NRefs != test.expected.NRefs {
				t.Fatalf("expected NReps %d, got %d", test.expected.NRefs, record.NRefs)
			}
			if record.Checked.IsZero() != test.expected.Checked.IsZero() {
				t.Fatalf("checked time in unexpected state %+v", record.Checked)
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
