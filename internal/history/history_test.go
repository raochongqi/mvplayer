package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "history.json")

	s := New()
	s.Put(Entry{Path: "/v/a.mkv", Size: 10, ModTime: 100, PositionMs: 5000, DurationMs: 9000, UpdatedAt: 1})
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	e, ok := got.Get("/v/a.mkv", 10, 100)
	if !ok || e.PositionMs != 5000 || e.DurationMs != 9000 {
		t.Fatalf("roundtrip mismatch: %+v ok=%v", e, ok)
	}
}

func TestLoadMissingIsEmpty(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Entries) != 0 {
		t.Fatalf("want empty, got %d", len(s.Entries))
	}
}

func TestKeyIncludesSizeAndMtime(t *testing.T) {
	s := New()
	s.Put(Entry{Path: "/v/a.mkv", Size: 10, ModTime: 100, PositionMs: 1})
	if _, ok := s.Get("/v/a.mkv", 11, 100); ok {
		t.Fatal("size change must miss")
	}
	if _, ok := s.Get("/v/a.mkv", 10, 101); ok {
		t.Fatal("mtime change must miss")
	}
	if _, ok := s.Get("/v/a.mkv", 10, 100); !ok {
		t.Fatal("same version must hit")
	}
}

func TestPruneKeepsNewest(t *testing.T) {
	s := New()
	for i := 0; i < 5; i++ {
		s.Put(Entry{Path: "/v/" + string(rune('a'+i)) + ".mkv", Size: 1, ModTime: 1, UpdatedAt: int64(i)})
	}
	s.Prune(2)
	if len(s.Entries) != 2 {
		t.Fatalf("want 2, got %d", len(s.Entries))
	}
	if _, ok := s.Get("/v/e.mkv", 1, 1); !ok {
		t.Fatal("newest entry pruned")
	}
	if _, ok := s.Get("/v/a.mkv", 1, 1); ok {
		t.Fatal("oldest entry kept")
	}
}

func TestCorruptFileIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("want error for corrupt file")
	}
}

func TestRemove(t *testing.T) {
	s := New()
	s.Put(Entry{Path: "/v/a.mkv", Size: 1, ModTime: 1})
	s.Remove("/v/a.mkv", 1, 1)
	if len(s.Entries) != 0 {
		t.Fatalf("remove failed: %d", len(s.Entries))
	}
}
