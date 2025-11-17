package app

import (
	"errors"
	"testing"
)

func TestFilmListNextWatch(t *testing.T) {
	testCases := []struct {
		name    string
		list    FilmList
		watched map[int]*Film
		want    Film
		wantErr error
	}{
		{
			name:    "returns error when list empty",
			list:    FilmList{},
			watched: map[int]*Film{},
			want:    Film{},
			wantErr: ErrListEmpty,
		},
		{
			name: "returns cached next film when unwatched",
			list: func() FilmList {
				f := &Film{LBxdID: 1, Title: "Cached", Year: 2000}
				return FilmList{Films: []*Film{f}, NextFilm: f}
			}(),
			watched: map[int]*Film{},
			want:    Film{LBxdID: 1, Title: "Cached", Year: 2000},
			wantErr: nil,
		},
		{
			name: "returns first unwatched in ordered list",
			list: FilmList{
				Ordered: true,
				Films: []*Film{
					{LBxdID: 1, Title: "Seen"},
					{LBxdID: 2, Title: "Next"},
					{LBxdID: 3, Title: "Later"},
				},
			},
			watched: map[int]*Film{
				1: {LBxdID: 1, Title: "Seen"},
			},
			want:    Film{LBxdID: 2, Title: "Next"},
			wantErr: nil,
		},
		{
			name: "returns error when all watched",
			list: FilmList{
				Ordered: true,
				Films: []*Film{
					{LBxdID: 1, Title: "Seen"},
				},
			},
			watched: map[int]*Film{
				1: {LBxdID: 1, Title: "Seen"},
			},
			want:    Film{},
			wantErr: ErrNoValidFilm,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.list.watched = tc.watched
			got, err := tc.list.NextWatch()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.LBxdID != tc.want.LBxdID {
				t.Fatalf("got %d want %d", got.LBxdID, tc.want.LBxdID)
			}
			if tc.list.NextFilm == nil {
				t.Fatal("expected NextFilm to be set")
			}
			if tc.list.NextFilm.LBxdID != got.LBxdID {
				t.Fatalf("NextFilm id %d does not match result %d", tc.list.NextFilm.LBxdID, got.LBxdID)
			}
		})
	}
}

func TestFilmListToggleOrdered(t *testing.T) {
	testCases := []struct {
		name        string
		list        FilmList
		wantOrdered bool
	}{
		{
			name:        "toggles from unordered to ordered",
			list:        FilmList{NextFilm: &Film{LBxdID: 1}},
			wantOrdered: true,
		},
		{
			name: "toggles from ordered to unordered",
			list: FilmList{
				Ordered: true,
			},
			wantOrdered: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.list.ToggleOrdered()
			if tc.list.Ordered != tc.wantOrdered {
				t.Fatalf("Ordered flag = %v want %v", tc.list.Ordered, tc.wantOrdered)
			}
			if tc.list.NextFilm != nil {
				t.Fatal("expected NextFilm to be nil after toggle")
			}
		})
	}
}

func TestApplicationRefreshList(t *testing.T) {
	watched := &Film{LBxdID: 5, Title: "Seen"}
	testCases := []struct {
		name        string
		list        *FilmList
		setup       func(app *Application, list *FilmList) error
		wantErr     error
		wantTracked bool
		wantWatched bool
	}{
		{
			name: "returns error when list not tracked",
			list: &FilmList{
				Url: "https://letterboxd.com/list/missing/",
			},
			setup:       func(app *Application, list *FilmList) error { return nil },
			wantErr:     ErrListNotTracked,
			wantTracked: false,
			wantWatched: false,
		},
		{
			name: "restores list when refresh fails",
			list: &FilmList{
				Url: "https://letterboxd.com/badrefresh",
				Films: []*Film{
					{LBxdID: 77, Title: "Tracked"},
				},
			},
			setup: func(app *Application, list *FilmList) error {
				if err := app.AddList(list); err != nil {
					return err
				}
				list.watched = nil
				return nil
			},
			wantErr:     ErrInvalidUrl,
			wantTracked: true,
			wantWatched: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &Application{
				TrackedLists: make(map[string]*FilmList),
				FilmStore:    FilmStore{Films: map[int]*FilmRecord{}},
				WatchedFilms: FilmsSet{
					watched.LBxdID: watched,
				},
			}
			if err := tc.setup(app, tc.list); err != nil {
				t.Fatalf("setup: %v", err)
			}
			err := app.RefreshList(tc.list)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
			} else {
				if err == nil || !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			}
			if tracked := app.IsListTracked(tc.list.Url); tracked != tc.wantTracked {
				t.Fatalf("IsListTracked=%v want %v", tracked, tc.wantTracked)
			}
			if tc.wantWatched {
				if tc.list.watched == nil || !tc.list.watched.InSet(watched) {
					t.Fatalf("expected watched pointer restored")
				}
			} else if tc.list.watched != nil {
				if tc.list.watched.InSet(watched) {
					t.Fatalf("did not expect watched pointer to be set")
				}
			}
		})
	}
}
