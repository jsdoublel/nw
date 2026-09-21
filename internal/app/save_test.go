package app

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ----- Test helpers

// setTestDataDir points NWDataPath at a fresh temp directory for the
// duration of the test. t.Setenv("NW_DATA_HOME", ...) does not work for
// this: NWDataPath is set exactly once in app.go's init(), which always
// runs before any test, so nothing re-reads the env var afterward.
func setTestDataDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := NWDataPath
	NWDataPath = dir
	t.Cleanup(func() { NWDataPath = old })
	return dir
}

// withTestConfig sets Config.Username/Config.ApiKey for the duration of the
// test and restores the previous Config afterward. Needed because Config,
// like NWDataPath, is a package-level var populated once at init().
func withTestConfig(t *testing.T, username, apiKey string) {
	t.Helper()
	old := Config
	Config = config{Username: username, ApiKey: apiKey}
	t.Cleanup(func() { Config = old })
}

// writeLegacySaveFixture writes s as plain, uncompressed JSON -- the
// pre-gzip save format -- for migration tests.
func writeLegacySaveFixture(t *testing.T, dir, filename string, s Save) {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), b, 0o644); err != nil {
		t.Fatalf("write legacy fixture failed: %v", err)
	}
}

// writeCanonicalSaveFixture writes s as a gzip-compressed save file,
// matching what Save() actually produces.
func writeCanonicalSaveFixture(t *testing.T, dir, filename string, s Save) {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	out, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	gw := gzip.NewWriter(out)
	if _, err := gw.Write(b); err != nil {
		t.Fatalf("gzip write failed: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close failed: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("file close failed: %v", err)
	}
}

// readGzipSave reads and decodes a gzip-compressed save file at path.
func readGzipSave(t *testing.T, path string) Save {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open save: %v", err)
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func() { _ = gz.Close() }()
	b, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("failed to read save: %v", err)
	}
	var saved Save
	if err := json.Unmarshal(b, &saved); err != nil {
		t.Fatalf("failed to unmarshal save: %v", err)
	}
	return saved
}

// writeLastUsernameFile seeds lastusername.txt with the given raw content.
func writeLastUsernameFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, lastUserFile), []byte(content), 0o644); err != nil {
		t.Fatalf("write lastusername.txt failed: %v", err)
	}
}

// ----- Save / Load

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
					1: {Film: Film{LBxdID: 1, Title: "Stored", Url: "https://example.com/film"}},
				}},
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			setTestDataDir(t)
			if err := test.app.Save(); err != nil {
				t.Fatalf("save returned error: %v", err)
			}
			path := savePath(test.app.Username)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("save file missing: %v", err)
			}
			saved := readGzipSave(t, path)
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
						7: {Film: Film{LBxdID: 7, Title: "Loaded"}},
					}},
				},
			},
			wantFilm: "Loaded",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			dir := setTestDataDir(t)
			writeCanonicalSaveFixture(t, dir, filepath.Base(savePath(test.user)), test.content)
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

// TestLoadMigratesLegacySave covers migrateLegacySave and Load's use of it,
// including a regression case for a bug caught during review: a migration
// failure must propagate as an error, not be silently treated the same as
// "no legacy save found" (which would create a fresh empty account and
// orphan the user's real data).
func TestLoadMigratesLegacySave(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		setup    func(t *testing.T, dir string)
		check    func(t *testing.T, dir string, app *Application, loadErr error)
	}{
		{
			name:     "exact case legacy file migrated",
			username: "dave",
			setup: func(t *testing.T, dir string) {
				writeLegacySaveFixture(t, dir, "dave.json", Save{Application: Application{
					Username:  "dave",
					FilmStore: FilmStore{Films: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1, Title: "Migrated"}}}},
				}})
			},
			check: func(t *testing.T, dir string, app *Application, loadErr error) {
				if loadErr != nil {
					t.Fatalf("unexpected error: %v", loadErr)
				}
				if got := app.FilmStore.Films[1].Title; got != "Migrated" {
					t.Fatalf("expected migrated content, got %s", got)
				}
				saved := readGzipSave(t, filepath.Join(dir, "dave.json.gz"))
				if got := saved.FilmStore.Films[1].Title; got != "Migrated" {
					t.Fatalf("canonical file content mismatch: got %s", got)
				}
				if _, err := os.Stat(filepath.Join(dir, "dave.json")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("expected legacy file to be removed, err=%v", err)
				}
			},
		},
		{
			name:     "different case legacy file migrated",
			username: "eve",
			setup: func(t *testing.T, dir string) {
				writeLegacySaveFixture(t, dir, "Eve.json", Save{Application: Application{
					Username:  "eve",
					FilmStore: FilmStore{Films: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1, Title: "CaseMigrated"}}}},
				}})
			},
			check: func(t *testing.T, dir string, app *Application, loadErr error) {
				if loadErr != nil {
					t.Fatalf("unexpected error: %v", loadErr)
				}
				if got := app.FilmStore.Films[1].Title; got != "CaseMigrated" {
					t.Fatalf("expected migrated content, got %s", got)
				}
				if _, err := os.Stat(filepath.Join(dir, "eve.json.gz")); err != nil {
					t.Fatalf("expected canonical file to exist after migration: %v", err)
				}
				if _, err := os.Stat(filepath.Join(dir, "Eve.json")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("expected legacy file to be removed, err=%v", err)
				}
			},
		},
		{
			name:     "no legacy or canonical file creates new empty app",
			username: "frank",
			setup:    func(t *testing.T, dir string) {},
			check: func(t *testing.T, dir string, app *Application, loadErr error) {
				if loadErr != nil {
					t.Fatalf("unexpected error: %v", loadErr)
				}
				if app.Username != "frank" {
					t.Fatalf("expected username frank, got %s", app.Username)
				}
				if _, err := os.Stat(filepath.Join(dir, "frank.json.gz")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("expected no save file to be created by Load alone, err=%v", err)
				}
			},
		},
		{
			name:     "canonical file present skips migration",
			username: "gina",
			setup: func(t *testing.T, dir string) {
				writeCanonicalSaveFixture(t, dir, "gina.json.gz", Save{Application: Application{
					Username:  "gina",
					FilmStore: FilmStore{Films: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1, Title: "Canonical"}}}},
				}})
				writeLegacySaveFixture(t, dir, "gina.json", Save{Application: Application{
					Username:  "gina",
					FilmStore: FilmStore{Films: map[int]*FilmRecord{1: {Film: Film{LBxdID: 1, Title: "Legacy"}}}},
				}})
			},
			check: func(t *testing.T, dir string, app *Application, loadErr error) {
				if loadErr != nil {
					t.Fatalf("unexpected error: %v", loadErr)
				}
				if got := app.FilmStore.Films[1].Title; got != "Canonical" {
					t.Fatalf("expected canonical content to win, got %s", got)
				}
				if _, err := os.Stat(filepath.Join(dir, "gina.json")); err != nil {
					t.Fatalf("expected legacy file to remain untouched: %v", err)
				}
			},
		},
		{
			name:     "migration failure returns error, does not silently create new account",
			username: "hank",
			setup: func(t *testing.T, dir string) {
				if runtime.GOOS == "windows" {
					t.Skip("permission bits behave differently on windows")
				}
				if os.Geteuid() == 0 {
					t.Skip("running as root ignores directory permissions")
				}
				writeLegacySaveFixture(t, dir, "hank.json", Save{Application: Application{Username: "hank"}})
				if err := os.Chmod(dir, 0o555); err != nil {
					t.Fatalf("chmod failed: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
			},
			check: func(t *testing.T, dir string, app *Application, loadErr error) {
				if loadErr == nil {
					t.Fatalf("expected error from Load when migration fails")
				}
				if !errors.Is(loadErr, ErrFailedMigration) {
					t.Fatalf("expected ErrFailedMigration, got %v", loadErr)
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := setTestDataDir(t)
			tc.setup(t, dir)
			app, err := Load(tc.username)
			tc.check(t, dir, app, err)
		})
	}
}

// TestApplicationSaveBackupRotation covers Save()'s atomic-write and
// backup-rotation behavior. Each case uses its own username so rows stay
// independent despite sharing the table-driven shape.
func TestApplicationSaveBackupRotation(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, dir string) string // returns the canonical save path
		check func(t *testing.T, dir, path string)
	}{
		{
			name: "first save creates no backup",
			setup: func(t *testing.T, dir string) string {
				app := &Application{Username: "iris1", FilmStore: FilmStore{Films: map[int]*FilmRecord{
					1: {Film: Film{LBxdID: 1, Title: "First"}},
				}}}
				if err := app.Save(); err != nil {
					t.Fatalf("save failed: %v", err)
				}
				return savePath("iris1")
			},
			check: func(t *testing.T, dir, path string) {
				if _, err := os.Stat(path + ".bak"); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("did not expect a .bak after the first save, err=%v", err)
				}
			},
		},
		{
			name: "second save rotates first save's content into .bak",
			setup: func(t *testing.T, dir string) string {
				username := "iris2"
				first := &Application{Username: username, FilmStore: FilmStore{Films: map[int]*FilmRecord{
					1: {Film: Film{LBxdID: 1, Title: "First"}},
				}}}
				if err := first.Save(); err != nil {
					t.Fatalf("first save failed: %v", err)
				}
				second := &Application{Username: username, FilmStore: FilmStore{Films: map[int]*FilmRecord{
					1: {Film: Film{LBxdID: 1, Title: "Second"}},
				}}}
				if err := second.Save(); err != nil {
					t.Fatalf("second save failed: %v", err)
				}
				return savePath(username)
			},
			check: func(t *testing.T, dir, path string) {
				bak := readGzipSave(t, path+".bak")
				if got := bak.FilmStore.Films[1].Title; got != "First" {
					t.Fatalf("expected .bak to hold first save's content, got %s", got)
				}
				cur := readGzipSave(t, path)
				if got := cur.FilmStore.Films[1].Title; got != "Second" {
					t.Fatalf("expected canonical path to hold second save's content, got %s", got)
				}
			},
		},
		{
			name: "no leftover temp files after saving",
			setup: func(t *testing.T, dir string) string {
				username := "iris3"
				app := &Application{Username: username, FilmStore: FilmStore{Films: map[int]*FilmRecord{
					1: {Film: Film{LBxdID: 1, Title: "Only"}},
				}}}
				if err := app.Save(); err != nil {
					t.Fatalf("save failed: %v", err)
				}
				return savePath(username)
			},
			check: func(t *testing.T, dir, path string) {
				matches, err := filepath.Glob(filepath.Join(dir, "*.tmp-*"))
				if err != nil {
					t.Fatalf("glob failed: %v", err)
				}
				if len(matches) != 0 {
					t.Fatalf("expected no leftover temp files, found %v", matches)
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := setTestDataDir(t)
			path := tc.setup(t, dir)
			tc.check(t, dir, path)
		})
	}
}

// ----- GetUser / validUser

func TestValidUser(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		want     bool
	}{
		{name: "plain alnum", username: "alice123", want: true},
		{name: "contains forward slash", username: "al/ice", want: false},
		{name: "contains backslash", username: `al\ice`, want: false},
		{name: "empty string", username: "", want: true},
		{name: "contains both separators", username: `a/b\c`, want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validUser(tc.username); got != tc.want {
				t.Fatalf("validUser(%q) = %v, want %v", tc.username, got, tc.want)
			}
		})
	}
}

// TestGetUser covers source precedence (explicit param -> Config.Username ->
// lastusername.txt -> askUser), that the returned username is always
// lowercased regardless of source, and two current-behavior quirks worth
// pinning down rather than assuming away: the write-back to
// lastusername.txt happens with the raw, unlowered value, and an invalid
// username still gets persisted there even though GetUser ultimately
// errors for that call.
func TestGetUser(t *testing.T) {
	mustNotAskUser := func() string { panic("askUser should not be called") }

	testCases := []struct {
		name             string
		explicitUsername string
		configUsername   string
		lastUsernameFile string // "" means don't seed the file
		askUser          func() string
		wantErr          bool
		wantUsername     string
		wantFileContent  string // "" means expect no file to exist
	}{
		{
			name:             "explicit param wins over everything, gets lowercased",
			explicitUsername: "Alice",
			configUsername:   "configuser",
			lastUsernameFile: "fileuser",
			askUser:          mustNotAskUser,
			wantUsername:     "alice",
			wantFileContent:  "Alice", // write-back happens before lowering
		},
		{
			name:             "Config.Username used when explicit empty",
			explicitUsername: "",
			configUsername:   "Bob",
			askUser:          mustNotAskUser,
			wantUsername:     "bob",
			wantFileContent:  "Bob",
		},
		{
			name:             "lastusername.txt used when explicit and Config empty",
			explicitUsername: "",
			lastUsernameFile: "  CAROL  \n",
			askUser:          mustNotAskUser,
			wantUsername:     "carol",
			wantFileContent:  "carol", // this source is lowered/trimmed at read time
		},
		{
			name:             "askUser used as last resort",
			explicitUsername: "",
			askUser:          func() string { return "Dave" },
			wantUsername:     "dave",
			wantFileContent:  "Dave",
		},
		{
			name:             "nothing available and no askUser returns error, writes no file",
			explicitUsername: "",
			askUser:          nil,
			wantErr:          true,
			wantFileContent:  "",
		},
		{
			name:             "invalid username still written to lastusername.txt (documents current quirk)",
			explicitUsername: "Bad/Name",
			askUser:          mustNotAskUser,
			wantErr:          true,
			wantFileContent:  "Bad/Name", // raw, unlowered value persisted despite the error
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := setTestDataDir(t)
			withTestConfig(t, tc.configUsername, "")
			if tc.lastUsernameFile != "" {
				writeLastUsernameFile(t, dir, tc.lastUsernameFile)
			}
			username := tc.explicitUsername
			err := GetUser(&username, tc.askUser)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if username != tc.wantUsername {
				t.Fatalf("expected username %q, got %q", tc.wantUsername, username)
			}
			content, readErr := os.ReadFile(filepath.Join(dir, lastUserFile))
			if tc.wantFileContent == "" {
				if !errors.Is(readErr, os.ErrNotExist) {
					t.Fatalf("expected no lastusername.txt, err=%v", readErr)
				}
			} else {
				if readErr != nil {
					t.Fatalf("expected lastusername.txt to exist: %v", readErr)
				}
				if string(content) != tc.wantFileContent {
					t.Fatalf("expected lastusername.txt content %q, got %q", tc.wantFileContent, content)
				}
			}
		})
	}
}

// ----- Tracked lists

func TestUpdateTrackedLists(t *testing.T) {
	watchedFilm := &Film{LBxdID: 1, Title: "Seen"}
	unwatchedFilm := &Film{LBxdID: 2, Title: "Next"}
	testCases := []struct {
		name                   string
		forceAll               bool
		wantUnwatchedRefreshed bool
	}{
		{
			name:                   "refreshes lists with watched next film",
			forceAll:               false,
			wantUnwatchedRefreshed: false,
		},
		{
			name:                   "force refresh refreshes all lists",
			forceAll:               true,
			wantUnwatchedRefreshed: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &Application{
				TrackedLists: make(map[string]*FilmList),
				FilmStore:    FilmStore{Films: map[int]*FilmRecord{}},
				WatchedFilms: FilmsSet{
					watchedFilm.LBxdID: watchedFilm,
				},
			}
			watchedList := &FilmList{
				Url: "https://letterboxd.com/bad-watched",
				Films: []*Film{
					{LBxdID: 11, Title: "Tracked"},
				},
				NextFilm: watchedFilm,
			}
			unwatchedList := &FilmList{
				Url: "https://letterboxd.com/bad-unwatched",
				Films: []*Film{
					{LBxdID: 22, Title: "Other"},
				},
				NextFilm: unwatchedFilm,
			}
			if err := app.AddList(watchedList); err != nil {
				t.Fatalf("add watched list: %v", err)
			}
			if err := app.AddList(unwatchedList); err != nil {
				t.Fatalf("add unwatched list: %v", err)
			}
			watchedList.watched = nil
			unwatchedList.watched = nil
			err := app.updateTrackedLists(tc.forceAll)
			if err == nil || !errors.Is(err, ErrInvalidUrl) {
				t.Fatalf("expected ErrInvalidUrl, got %v", err)
			}
			if watchedList.watched == nil || !watchedList.watched.InSet(watchedFilm) {
				t.Fatalf("watched list should have refreshed")
			}
			if tc.wantUnwatchedRefreshed {
				if unwatchedList.watched == nil || !unwatchedList.watched.InSet(watchedFilm) {
					t.Fatalf("unwatched list was not refreshed when forced")
				}
			} else if unwatchedList.watched != nil {
				t.Fatalf("unwatched list should not refresh")
			}
			if !app.IsListTracked(watchedList.Url) || !app.IsListTracked(unwatchedList.Url) {
				t.Fatalf("lists should remain tracked")
			}
		})
	}
}

func TestUpdateTrackedListsNilNextFilm(t *testing.T) {
	testCases := []struct {
		name     string
		forceAll bool
	}{
		{
			name:     "handles nil NextFilm without panic",
			forceAll: false,
		},
		{
			name:     "handles nil NextFilm with forceAll without panic",
			forceAll: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := &Application{
				TrackedLists: make(map[string]*FilmList),
				WatchedFilms: FilmsSet{},
			}
			listWithNilNext := &FilmList{
				Name:     "Empty or Completed List",
				Url:      "https://letterboxd.com/user/list/empty/",
				NextFilm: nil,
			}
			app.TrackedLists[listWithNilNext.Url] = listWithNilNext
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("updateTrackedLists panicked with nil NextFilm: %v", r)
				}
			}()
			_ = app.updateTrackedLists(tc.forceAll)
		})
	}
}
