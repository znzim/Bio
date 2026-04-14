package views

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCloseFlushesPendingHashes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "views.json")

	store, err := newStore(path, time.Hour)
	if err != nil {
		t.Fatalf("newStore() error = %v", err)
	}

	if _, _, err := store.Add("hash-a"); err != nil {
		t.Fatalf("Add(hash-a) error = %v", err)
	}

	if _, _, err := store.Add("hash-b"); err != nil {
		t.Fatalf("Add(hash-b) error = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	hashes := readHashes(t, path)
	if len(hashes) != 2 {
		t.Fatalf("expected 2 persisted hashes, got %d", len(hashes))
	}
}

func TestStoreFlushLoopPersistsAsynchronously(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "views.json")

	store, err := newStore(path, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("newStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	added, count, err := store.Add("hash-a")
	if err != nil {
		t.Fatalf("Add(hash-a) error = %v", err)
	}

	if !added || count != 1 {
		t.Fatalf("expected first add to increment count, got added=%v count=%d", added, count)
	}

	added, count, err = store.Add("hash-a")
	if err != nil {
		t.Fatalf("Add(duplicate) error = %v", err)
	}

	if added || count != 1 {
		t.Fatalf("expected duplicate add to be ignored, got added=%v count=%d", added, count)
	}

	waitForCount(t, path, 1)
}

func waitForCount(t *testing.T, path string, want int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hashes := readHashes(t, path)
		if len(hashes) == want {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %d persisted hashes in %s", want, path)
}

func readHashes(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}

	var payload persisted
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v", path, err)
	}

	return payload.Hashes
}
