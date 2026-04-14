package views

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultFlushInterval = 2 * time.Second

type Store struct {
	mu               sync.RWMutex
	path             string
	hashes           map[string]struct{}
	version          uint64
	persistedVersion uint64
	flushInterval    time.Duration
	stopCh           chan struct{}
	doneCh           chan struct{}
	closeOnce        sync.Once
}

type persisted struct {
	Hashes []string `json:"hashes"`
}

func NewStore(path string) (*Store, error) {
	return newStore(path, defaultFlushInterval)
}

func newStore(path string, flushInterval time.Duration) (*Store, error) {
	if flushInterval <= 0 {
		flushInterval = defaultFlushInterval
	}

	store := &Store{
		path:          path,
		hashes:        make(map[string]struct{}),
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			go store.flushLoop()
			return store, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		go store.flushLoop()
		return store, nil
	}

	var payload persisted
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	for _, hash := range payload.Hashes {
		trimmed := strings.TrimSpace(hash)
		if trimmed == "" {
			continue
		}
		store.hashes[trimmed] = struct{}{}
	}

	store.version = uint64(len(store.hashes))
	store.persistedVersion = store.version
	go store.flushLoop()

	return store, nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.hashes)
}

func (s *Store) Add(hash string) (bool, int, error) {
	trimmed := strings.TrimSpace(hash)
	if trimmed == "" {
		return false, s.Count(), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hashes[trimmed]; exists {
		return false, len(s.hashes), nil
	}

	s.hashes[trimmed] = struct{}{}
	s.version++

	return true, len(s.hashes), nil
}

func HashIP(ip, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Store) Close() error {
	var err error

	s.closeOnce.Do(func() {
		close(s.stopCh)
		<-s.doneCh
		err = s.flush()
	})

	return err
}

func (s *Store) flushLoop() {
	ticker := time.NewTicker(s.flushInterval)
	defer func() {
		ticker.Stop()
		close(s.doneCh)
	}()

	for {
		select {
		case <-ticker.C:
			if err := s.flush(); err != nil {
				log.Printf("view store flush error: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) flush() error {
	hashes, version := s.snapshot()
	if hashes == nil {
		return nil
	}

	if err := s.persistSnapshot(hashes); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.persistedVersion < version {
		s.persistedVersion = version
	}

	return nil
}

func (s *Store) snapshot() ([]string, uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.version == s.persistedVersion {
		return nil, s.persistedVersion
	}

	hashes := make([]string, 0, len(s.hashes))
	for hash := range s.hashes {
		hashes = append(hashes, hash)
	}

	sort.Strings(hashes)

	return hashes, s.version
}

func (s *Store) persistSnapshot(hashes []string) error {
	payload := persisted{Hashes: hashes}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}
