package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApplicationSave(t *testing.T) {
	testCases := []struct {
		name string
		app  Application
	}{
		{
			name: "writes save file",
			app: Application{
				Username: "alice",
				FilmStore: FilmStore{Films: map[int]*FilmRecord{
					1: {Film: Film{LBxdID: 1, Title: "Stored", Url: "https://example.com/film"}, NRefs: 1, Checked: time.Now()},
				}},
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("NW_DATA_HOME", tempDir)
			if err := test.app.Save(); err != nil {
				t.Fatalf("save returned error: %v", err)
			}
			path := savePath(test.app.Username)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("save file missing: %v", err)
			}
			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read save: %v", err)
			}
			var saved Save
			if err := json.Unmarshal(bytes, &saved); err != nil {
				t.Fatalf("failed to unmarshal save: %v", err)
			}
			if saved.Version != LatestSaveVersion {
				t.Fatalf("expected version %d, got %d", LatestSaveVersion, saved.Version)
			}
			stored, ok := saved.FilmStore.Films[1]
			if !ok {
				t.Fatalf("expected film record to be stored")
			}
			if stored.Title != "Stored" {
				t.Fatalf("expected stored title, got %s", stored.Title)
			}
		})
	}
}

func TestLoadReturnsSavedData(t *testing.T) {
	testCases := []struct {
		name     string
		user     string
		content  Save
		wantFilm string
	}{
		{
			name: "loads existing save",
			user: "bob",
			content: Save{
				Version: LatestSaveVersion,
				Application: Application{
					Username: "bob",
					FilmStore: FilmStore{Films: map[int]*FilmRecord{
						7: {Film: Film{LBxdID: 7, Title: "Loaded"}, NRefs: 2},
					}},
				},
			},
			wantFilm: "Loaded",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tempDir := t.TempDir()
			t.Setenv("NW_DATA_HOME", tempDir)
			path := savePath(test.user)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir failed: %v", err)
			}
			bytes, err := json.Marshal(test.content)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if err := os.WriteFile(path, bytes, 0o644); err != nil {
				t.Fatalf("write failed: %v", err)
			}
			app, err := Load(test.user)
			if err != nil {
				t.Fatalf("load returned error: %v", err)
			}
			record, ok := app.FilmStore.Films[7]
			if !ok {
				t.Fatalf("expected film record with id 7")
			}
			if record.Title != test.wantFilm {
				t.Fatalf("expected film title %s, got %s", test.wantFilm, record.Title)
			}
		})
	}
}
