package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFavoritesSeedsDefaultsOnce(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SPORE_CONFIG_DIR", dir)

	favs, err := LoadFavorites()
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultFavorites()
	if len(favs) != len(want) {
		t.Fatalf("first load: got %d favorites, want %d", len(favs), len(want))
	}
	path := filepath.Join(dir, "favorites.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("favorites.json not written: %v", err)
	}

	favs2, err := LoadFavorites()
	if err != nil {
		t.Fatal(err)
	}
	if len(favs2) != len(want) {
		t.Fatalf("second load: got %d, want %d", len(favs2), len(want))
	}

	if err := SaveFavorites(nil); err != nil {
		t.Fatal(err)
	}
	favs3, err := LoadFavorites()
	if err != nil {
		t.Fatal(err)
	}
	if len(favs3) != 0 {
		t.Fatalf("after clear: got %d, want 0 (defaults must not re-seed)", len(favs3))
	}
}
