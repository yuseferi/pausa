package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store loads, persists, and notifies subscribers of configuration changes.
// All operations are safe for concurrent use. Subscribers receive a copy of
// the new config on a buffered channel; slow subscribers may drop updates.
type Store struct {
	path string

	mu          sync.RWMutex
	cfg         Config
	exists      bool
	subscribers []chan Config
	closed      bool
}

// Open loads (or creates) the store at the given path. Missing or invalid
// files fall back to defaults. The returned Store is safe for concurrent
// use.
func Open(path string) (*Store, error) {
	s := &Store{path: path, cfg: Default()}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// First run: leave defaults, mark not-yet-saved.
		return s, nil
	case err != nil:
		// Any other read failure (permissions, I/O error, path is a
		// directory) must still yield a usable defaults store so callers
		// never nil-dereference at startup.
		return s, fmt.Errorf("read config (using defaults): %w", err)
	}
	var loaded Config
	if jerr := json.Unmarshal(data, &loaded); jerr != nil {
		// Corrupted file - keep defaults but report.
		return s, fmt.Errorf("parse config (using defaults): %w", jerr)
	}
	// Migration: MediaCountsAsActivity was added after 1.0.4 with a default
	// of true. Config files written before it existed decode the missing key
	// as false, which would silently keep the old "movies count as idle"
	// behavior. Detect the missing key and opt existing users into the fix;
	// an explicitly saved false is preserved. The migrated value persists
	// on the next Set.
	if !hasJSONKey(data, "idle", "mediaCountsAsActivity") {
		loaded.Idle.MediaCountsAsActivity = true
	}
	loaded.Validate()
	s.cfg = loaded
	s.exists = true
	return s, nil
}

// Get returns a copy of the current configuration.
func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Exists reports whether a config file exists on disk (false on first run).
func (s *Store) Exists() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exists
}

// Set replaces the configuration, validates and atomically persists it,
// and broadcasts the change to all subscribers. Returns the validated
// (possibly clamped) config that was actually saved.
func (s *Store) Set(c Config) (Config, error) {
	c.Validate()

	if err := s.writeAtomic(c); err != nil {
		return c, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = c
	s.exists = true
	if s.closed {
		// Store was closed concurrently; don't touch subscriber channels.
		return c, nil
	}
	// Send while holding the lock. The sends are non-blocking (default
	// drops), so this cannot deadlock, and it guarantees Close cannot
	// close a channel between our snapshot and our send.
	for _, ch := range s.subscribers {
		select {
		case ch <- c:
		default: // drop if subscriber is slow; latest update will follow
		}
	}
	return c, nil
}

// Subscribe returns a channel that receives the latest config every time it
// changes. The buffer is small (1); use a separate goroutine to drain.
// The channel is closed when ctx-equivalent cleanup happens via Close. If
// the store is already closed, the returned channel is closed immediately.
func (s *Store) Subscribe() <-chan Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan Config, 1)
	if s.closed {
		close(ch)
		return ch
	}
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// Close releases subscriber channels. It is idempotent.
func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.subscribers = nil
}

// hasJSONKey reports whether the nested object data[obj][key] is present in
// a raw JSON document. Used for migrating newly added keys: a missing key
// means "written by an older version", as opposed to an explicit zero value.
func hasJSONKey(data []byte, obj, key string) bool {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return false
	}
	raw, ok := top[obj]
	if !ok {
		return false
	}
	var inner map[string]json.RawMessage
	if err := json.Unmarshal(raw, &inner); err != nil {
		return false
	}
	_, ok = inner[key]
	return ok
}

func (s *Store) writeAtomic(c Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// DefaultPath returns the canonical config file location.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "Pausa", "config.json"), nil
}
