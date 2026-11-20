// Package history stores per-file playback positions so a reopened file can
// resume where it stopped. Pure Go, no GUI dependency.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry is one remembered playback position.
type Entry struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModTime    int64  `json:"mtime"` // unix seconds
	PositionMs int64  `json:"position_ms"`
	DurationMs int64  `json:"duration_ms"`
	UpdatedAt  int64  `json:"updated_at"` // unix seconds
}

// Key identifies a file version. Size and mtime are part of the key so a
// replaced file does not inherit the old position.
func (e Entry) Key() string {
	return fmt.Sprintf("%s|%d|%d", e.Path, e.Size, e.ModTime)
}

// Store is the on-disk history document.
type Store struct {
	Version int              `json:"version"`
	Entries map[string]Entry `json:"entries"`
}

// New returns an empty store.
func New() *Store { return &Store{Version: 1, Entries: map[string]Entry{}} }

// Load reads the store, returning an empty one when the file does not exist.
// A corrupt file is an error: callers should back it up rather than silently
// losing history.
func Load(path string) (*Store, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return New(), nil
	}
	if err != nil {
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("history: parse %s: %w", path, err)
	}
	if s.Entries == nil {
		s.Entries = map[string]Entry{}
	}
	return &s, nil
}

// Save writes the store atomically (tmp file + rename).
func (s *Store) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Get returns the remembered entry for a file version.
func (s *Store) Get(path string, size, mtime int64) (Entry, bool) {
	e, ok := s.Entries[Entry{Path: path, Size: size, ModTime: mtime}.Key()]
	return e, ok
}

// Put records or updates one entry.
func (s *Store) Put(e Entry) { s.Entries[e.Key()] = e }

// Remove drops one file version.
func (s *Store) Remove(path string, size, mtime int64) {
	delete(s.Entries, Entry{Path: path, Size: size, ModTime: mtime}.Key())
}

// Prune keeps only the most recently updated max entries (0 = no-op).
func (s *Store) Prune(max int) {
	if max <= 0 || len(s.Entries) <= max {
		return
	}
	type kv struct {
		k string
		t int64
	}
	all := make([]kv, 0, len(s.Entries))
	for k, e := range s.Entries {
		all = append(all, kv{k, e.UpdatedAt})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].t > all[j].t })
	for _, item := range all[max:] {
		delete(s.Entries, item.k)
	}
}

// Now is the clock used for UpdatedAt; tests replace it.
var Now = func() int64 { return time.Now().Unix() }
