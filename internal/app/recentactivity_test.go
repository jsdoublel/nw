package app

import (
	"errors"
	"testing"
	"time"
)

func TestGetRSSItems(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tests := []struct {
		name     string
		username string
		wantLen  int
		wantGuid string
		wantLink string
		wantErr  error
	}{
		{
			name:     "elihayes",
			username: "elihayes",
			wantLen:  100,
			wantGuid: "letterboxd-watch-106293524",
			wantLink: "https://letterboxd.com/elihayes/film/parental-leave/",
		},
		{
			name:     "non-existent user",
			username: "thisusershouldnotexist_1234567890",
			wantErr:  ErrRetreivingRSS,
		},
	}

	for _, tt := range tests {
		time.Sleep(2 * time.Second)
		t.Run(tt.name, func(t *testing.T) {
			items, err := getRSSItems(tt.username)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("getRSSItems() expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("getRSSItems() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("getRSSItems() unexpected error: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("getRSSItems() len = %v, want %v", len(items), tt.wantLen)
			}
			if len(items) > 0 {
				if items[0].Guid != tt.wantGuid {
					t.Errorf("items[0].Guid = %v, want %v", items[0].Guid, tt.wantGuid)
				}
				if items[0].Link != tt.wantLink {
					t.Errorf("items[0].Link = %v, want %v", items[0].Link, tt.wantLink)
				}
			}
		})
	}
}

func TestUpdateRecentActivity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	stalker := Film{LBxdID: 51062, Title: "Stalker", Url: "https://letterboxd.com/film/stalker/"}
	solaris := Film{LBxdID: 51528, Title: "Solaris", Url: "https://letterboxd.com/film/solaris/"}

	tests := []struct {
		name             string
		initialActivity  RecentActivity
		updatedActivity  []RSSItem
		initialWatchlist FilmsSet
		initialWatched   FilmsSet
		wantWatchlistLen int
		wantWatchedLen   int
		wantInWatched    []int
	}{
		{
			name:            "new watch activity moves film from watchlist to watched",
			initialActivity: RecentActivity{},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-watch-stalker-123", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{51062: &stalker},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 0,
			wantWatchedLen:   1,
			wantInWatched:    []int{51062},
		},
		{
			name:            "multiple updates processed",
			initialActivity: RecentActivity{},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-watch-solaris-1", Link: "https://letterboxd.com/user/film/solaris/"},
				{Guid: "letterboxd-review-stalker-1", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{51062: &stalker, 51528: &solaris},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 0,
			wantWatchedLen:   2,
			wantInWatched:    []int{51062, 51528},
		},
		{
			name:            "breaks loop on non-watch/review item",
			initialActivity: RecentActivity{},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-watch-solaris-1", Link: "https://letterboxd.com/user/film/solaris/"},
				{Guid: "letterboxd-list-some-list", Link: "https://letterboxd.com/user/list/some-list/"},
				{Guid: "letterboxd-watch-stalker-1", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{51062: &stalker, 51528: &solaris},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 1,
			wantWatchedLen:   1,
			wantInWatched:    []int{51528},
		},
		{
			name:            "film not in watchlist still added to watched",
			initialActivity: RecentActivity{},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-watch-stalker-123", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 0,
			wantWatchedLen:   1,
			wantInWatched:    []int{51062},
		},
		{
			name: "duplicate activity does nothing",
			initialActivity: RecentActivity{
				Activity: []RSSItem{
					{Guid: "letterboxd-watch-stalker-123", Link: "https://letterboxd.com/user/film/stalker/"},
				},
			},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-watch-stalker-123", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{51062: &stalker},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 1,
			wantWatchedLen:   0,
			wantInWatched:    []int{},
		},
		{
			name:            "review activity also counts",
			initialActivity: RecentActivity{},
			updatedActivity: []RSSItem{
				{Guid: "letterboxd-review-stalker-456", Link: "https://letterboxd.com/user/film/stalker/"},
			},
			initialWatchlist: FilmsSet{51062: &stalker},
			initialWatched:   FilmsSet{},
			wantWatchlistLen: 0,
			wantWatchedLen:   1,
			wantInWatched:    []int{51062},
		},
	}

	for _, tt := range tests {
		time.Sleep(2 * time.Second)
		t.Run(tt.name, func(t *testing.T) {
			if tt.initialWatchlist == nil {
				tt.initialWatchlist = make(FilmsSet)
			}
			placeholdersCount := 30
			for i := 1000; i < 1000+placeholdersCount; i++ {
				tt.initialWatchlist[i] = &Film{LBxdID: i, Title: "Placeholder"}
			}

			app := &Application{
				Watchlist:      tt.initialWatchlist,
				WatchedFilms:   tt.initialWatched,
				RecentActivity: tt.initialActivity,
				TrackedLists:   make(map[string]*FilmList),
				FilmStore:      FilmStore{Films: make(map[int]*FilmRecord)},
			}

			app.FilmStore.RegisterSet(tt.initialWatchlist)
			for _, fr := range app.FilmStore.Films {
				fr.Checked = time.Now()
				fr.ReleaseDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			var err error
			app.NWQueue, err = app.MakeNextWatch()
			if err != nil {
				t.Fatalf("MakeNextWatch() failed: %v", err)
			}

			err = app.updateRecentActivity(tt.updatedActivity)
			if err != nil {
				t.Fatalf("updateRecentActivity() unexpected error: %v", err)
			}

			wantWlLen := tt.wantWatchlistLen + placeholdersCount
			if len(app.Watchlist) != wantWlLen {
				t.Errorf("Watchlist len = %v, want %v", len(app.Watchlist), wantWlLen)
			}
			if len(app.WatchedFilms) != tt.wantWatchedLen {
				t.Errorf("WatchedFilms len = %v, want %v", len(app.WatchedFilms), tt.wantWatchedLen)
			}
			for _, id := range tt.wantInWatched {
				if _, ok := app.WatchedFilms[id]; !ok {
					t.Errorf("Film ID %v not found in WatchedFilms", id)
				}
			}
		})
	}
}
