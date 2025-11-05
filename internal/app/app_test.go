package app

import (
	"os"
	"path/filepath"
	"testing"
)

func setEnv(t *testing.T, key, value string, present bool) {
	prev, existed := os.LookupEnv(key)
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, prev); err != nil {
				t.Fatalf("Setenv failed, %s", err)
			}
		} else {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("Unsetenv failed, %s", err)
			}
		}
	})
	if present {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set %s: %v", key, err)
		}
	} else {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset %s: %v", key, err)
		}
	}
}

func TestGetDirBase(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func(*testing.T) (string, bool)
	}{
		{
			name: "prefers NW_DATA_HOME",
			prepare: func(t *testing.T) (string, bool) {
				tmp := t.TempDir()
				setEnv(t, "NW_DATA_HOME", tmp, true)
				setEnv(t, "XDG_DATA_HOME", "", false)
				setEnv(t, "HOME", "", false)
				setEnv(t, "LOCALAPPDATA", "", false)
				return tmp, false
			},
		},
		{
			name: "falls back to XDG_DATA_HOME",
			prepare: func(t *testing.T) (string, bool) {
				tmp := t.TempDir()
				setEnv(t, "NW_DATA_HOME", "", false)
				setEnv(t, "XDG_DATA_HOME", tmp, true)
				setEnv(t, "HOME", "", false)
				setEnv(t, "LOCALAPPDATA", "", false)
				return tmp, false
			},
		},
		{
			name: "uses mac application support when present",
			prepare: func(t *testing.T) (string, bool) {
				home := t.TempDir()
				target := filepath.Join(home, "Library", "Application Support")
				if err := os.MkdirAll(target, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				setEnv(t, "NW_DATA_HOME", "", false)
				setEnv(t, "XDG_DATA_HOME", "", false)
				setEnv(t, "HOME", home, true)
				setEnv(t, "LOCALAPPDATA", "", false)
				return target, false
			},
		},
		{
			name: "falls back to XDG default when mac path missing",
			prepare: func(t *testing.T) (string, bool) {
				home := t.TempDir()
				setEnv(t, "NW_DATA_HOME", "", false)
				setEnv(t, "XDG_DATA_HOME", "", false)
				setEnv(t, "HOME", home, true)
				setEnv(t, "LOCALAPPDATA", "", false)
				return filepath.Join(home, ".local", "share"), false
			},
		},
		{
			name: "uses LOCALAPPDATA when home missing",
			prepare: func(t *testing.T) (string, bool) {
				local := filepath.Join(t.TempDir(), "AppData", "Local")
				setEnv(t, "NW_DATA_HOME", "", false)
				setEnv(t, "XDG_DATA_HOME", "", false)
				setEnv(t, "HOME", "", false)
				setEnv(t, "LOCALAPPDATA", local, true)
				return local, false
			},
		},
		{
			name: "panics when no home or localappdata",
			prepare: func(t *testing.T) (string, bool) {
				setEnv(t, "NW_DATA_HOME", "", false)
				setEnv(t, "XDG_DATA_HOME", "", false)
				setEnv(t, "HOME", "", false)
				setEnv(t, "LOCALAPPDATA", "", false)
				return "", true
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want, wantPanic := tc.prepare(t)
			defer func() {
				if r := recover(); r != nil {
					if !wantPanic {
						t.Fatalf("unexpected panic: %v", r)
					}
				} else if wantPanic {
					t.Fatal("expected panic but did not get one")
				}
			}()
			got := getDirBase()
			if wantPanic {
				return
			}
			if got != want {
				t.Fatalf("getDirBase() = %q, want %q", got, want)
			}
		})
	}
}
