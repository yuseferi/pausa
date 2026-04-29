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
		return nil, fmt.Errorf("read config: %w", err)
	}
	var loaded Config
	if jerr := json.Unmarshal(data, &loaded); jerr != nil {
		// Corrupted file - keep defaults but report.
		return s, fmt.Errorf("parse config (using defaults): %w", jerr)
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
	s.cfg = c
	s.exists = true
	subs := append([]chan Config(nil), s.subscribers...)
	s.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- c:
		default: // drop if subscriber is slow; latest update will follow
		}
	}
	return c, nil
}

// Subscribe returns a channel that receives the latest config every time it
// changes. The buffer is small (1); use a separate goroutine to drain.
// The channel is closed when ctx-equivalent cleanup happens via Close.
func (s *Store) Subscribe() <-chan Config {
	ch := make(chan Config, 1)
	s.mu.Lock()
	s.subscribers = append(s.subscribers, ch)
	s.mu.Unlock()
	return ch
}

// Close releases subscriber channels.
func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.subscribers = nil
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
