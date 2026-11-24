package cardcheck

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/engine"
)

// Every starter deck in data/decks is construction-legal (FR-SET-001), and each
// of the four Headquarters factions has one.
func TestDecks_StarterDecksValidate(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("..", "..", "data", "decks")
	paths, err := filepath.Glob(filepath.Join(dir, "*-starter.txt"))
	if err != nil {
		t.Fatal(err)
	}
	found := make(map[string]bool)
	for _, p := range paths {
		found[filepath.Base(p)] = true
		d, err := engine.LoadDeck(p, all)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
			continue
		}
		if err := engine.ValidateDeck(d); err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
		}
	}
	for _, f := range []string{"animus", "concordia", "path", "soulless"} {
		if !found[f+"-starter.txt"] {
			t.Errorf("data/decks has no %s-starter.txt", f)
		}
	}
}
