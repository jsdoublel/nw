package app

import (
	"errors"
	"testing"
)

func TestGetRSSItems(t *testing.T) {
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
