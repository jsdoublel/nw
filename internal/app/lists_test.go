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
			tc.list.store = WatchedFilms{Films: tc.watched}
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
