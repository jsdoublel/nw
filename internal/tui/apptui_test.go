package tui

import (
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

// TestFlockSingleInstanceLock pins the exact flock contention semantics
// RunApplicationTUI's single-instance lock depends on: TryLock returns
// (false, nil) on contention, not an error, and the lock frees up once
// Unlock is called. Deliberately does not exercise RunApplicationTUI itself
// -- that would require mocking a real tea.Program and the
// AskQuestion/GetUser/Load chain for a package with no existing test
// infrastructure, for logic that's otherwise thin plumbing around a
// well-established library.
func TestFlockSingleInstanceLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "nw.lock")

	lock1 := flock.New(lockPath)
	locked1, err := lock1.TryLock()
	if err != nil {
		t.Fatalf("first TryLock returned error: %v", err)
	}
	if !locked1 {
		t.Fatalf("expected first TryLock to succeed")
	}

	lock2 := flock.New(lockPath)
	locked2, err := lock2.TryLock()
	if err != nil {
		t.Fatalf("second TryLock returned error: %v", err)
	}
	if locked2 {
		t.Fatalf("expected second TryLock to fail while first lock is held")
	}

	if err := lock1.Unlock(); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}

	locked3, err := lock2.TryLock()
	if err != nil {
		t.Fatalf("third TryLock returned error: %v", err)
	}
	if !locked3 {
		t.Fatalf("expected TryLock to succeed after first lock released")
	}
	defer func() { _ = lock2.Unlock() }()
}
