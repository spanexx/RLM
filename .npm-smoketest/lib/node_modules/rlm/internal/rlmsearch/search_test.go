package rlmsearch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchDir_FixedString(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("hello world\nsecond line\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := SearchDir(Options{ContextDir: dir, Query: "world", Regex: false, MaxMatches: 10, MaxPerFile: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(res.Matches))
	}
	if res.Matches[0].Line != 1 {
		t.Fatalf("expected line 1, got %d", res.Matches[0].Line)
	}
}

func TestSearchDir_SkipsBinary(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin.dat")
	if err := os.WriteFile(p, []byte{0x00, 0x01, 0x02, 0x03, 'h', 'i'}, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := SearchDir(Options{ContextDir: dir, Query: "hi", Regex: false, MaxMatches: 10, MaxPerFile: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Matches) != 0 {
		t.Fatalf("expected 0 matches (binary skipped), got %d", len(res.Matches))
	}
}
