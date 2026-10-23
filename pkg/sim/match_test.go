package sim

import (
	"path/filepath"
	"reflect"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0047 Acceptance Case 1: one match-setup scheme.

func matchRoot(t *testing.T) (string, map[string]*engine.Card) {
	t.Helper()
	root := filepath.Join("..", "..")
	cards, err := engine.LoadCards(filepath.Join(root, "data", "cards"))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	return root, cards
}

func sameDecks(t *testing.T, got, want [2]engine.DeckList) {
	t.Helper()
	for i := range got {
		gh, gm := deckNames(got[i])
		wh, wm := deckNames(want[i])
		if gh != wh || !reflect.DeepEqual(gm, wm) {
			t.Errorf("seat %d deck = %s %v, want %s %v", i, gh, gm, wh, wm)
		}
	}
}

func TestMatchSetup_Case1a_Decklists(t *testing.T) {
	root, cards := matchRoot(t)
	factions := [2]string{"ANIMUS", "PATH"}
	want, err := SimDecks(cards, filepath.Join(root, "data", "decks"), "starter", factions)
	if err != nil {
		t.Fatalf("SimDecks: %v", err)
	}
	got, names, err := MatchDecks(root, cards, "starter", factions)
	if err != nil {
		t.Fatalf("MatchDecks: %v", err)
	}
	sameDecks(t, got, want)
	if wantNames := [2]string{"data/decks/animus-starter.txt", "data/decks/path-starter.txt"}; names != wantNames {
		t.Errorf("names = %v, want %v", names, wantNames)
	}
}

func TestMatchSetup_Case1b_StressDecks(t *testing.T) {
	root, cards := matchRoot(t)
	factions := [2]string{"ANIMUS", "PATH"}
	var want [2]engine.DeckList
	for i, f := range factions {
		d, err := FactionDeck(cards, f)
		if err != nil {
			t.Fatalf("FactionDeck(%s): %v", f, err)
		}
		want[i] = d
	}
	got, names, err := MatchDecks(root, cards, "", factions)
	if err != nil {
		t.Fatalf("MatchDecks: %v", err)
	}
	sameDecks(t, got, want)
	if wantNames := [2]string{"stress:ANIMUS", "stress:PATH"}; names != wantNames {
		t.Errorf("names = %v, want %v", names, wantNames)
	}
}

func TestMatchSetup_Case1c_MissingDeck(t *testing.T) {
	root, cards := matchRoot(t)
	_, names, err := MatchDecks(root, cards, "nonexistent", [2]string{"ANIMUS", "PATH"})
	if err == nil {
		t.Fatal("MatchDecks(nonexistent) returned no error")
	}
	if names != [2]string{} {
		t.Errorf("names on error = %v, want the zero value", names)
	}
}

func TestMatchSetup_Case1d_SeedAndBias(t *testing.T) {
	if got := SeatSeed(5, 0); got != 11 {
		t.Errorf("SeatSeed(5, 0) = %d, want 11", got)
	}
	if got := SeatSeed(5, 1); got != 12 {
		t.Errorf("SeatSeed(5, 1) = %d, want 12", got)
	}
	if DefaultEndBias != 0.1 {
		t.Errorf("DefaultEndBias = %v, want 0.1", DefaultEndBias)
	}
}
