package sim

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0033 Acceptance Case 1: one deck-naming encode and decode.

func deckNames(d engine.DeckList) (hq string, main []string) {
	if d.HQ != nil {
		hq = d.HQ.Name
	}
	for _, c := range d.Main {
		main = append(main, c.Name)
	}
	return hq, main
}

func TestDeckNames_Case1_RoundTrip(t *testing.T) {
	root := filepath.Join("..", "..")
	cards, err := engine.LoadCards(filepath.Join(root, "data", "cards"))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	if got := DecklistName("ANIMUS", "starter"); got != "data/decks/animus-starter.txt" {
		t.Errorf("DecklistName = %q, want data/decks/animus-starter.txt", got)
	}
	if got := StressName("SOULLESS"); got != "stress:SOULLESS" {
		t.Errorf("StressName = %q", got)
	}
	factions := [2]string{"ANIMUS", "PATH"}
	want, err := SimDecks(cards, filepath.Join(root, "data", "decks"), "starter", factions)
	if err != nil {
		t.Fatalf("SimDecks: %v", err)
	}
	for i, f := range factions {
		got, err := ResolveDeck(root, cards, DecklistName(f, "starter"))
		if err != nil {
			t.Fatalf("ResolveDeck %s: %v", f, err)
		}
		gh, gm := deckNames(got)
		wh, wm := deckNames(want[i])
		if gh != wh || !reflect.DeepEqual(gm, wm) {
			t.Errorf("%s: ResolveDeck differs from SimDecks", f)
		}
	}
	for _, f := range []string{"SOULLESS", "CONCORDIA"} {
		got, err := ResolveDeck(root, cards, StressName(f))
		if err != nil {
			t.Fatalf("ResolveDeck stress %s: %v", f, err)
		}
		exp, err := FactionDeck(cards, f)
		if err != nil {
			t.Fatalf("FactionDeck %s: %v", f, err)
		}
		gh, gm := deckNames(got)
		eh, em := deckNames(exp)
		if gh != eh || !reflect.DeepEqual(gm, em) {
			t.Errorf("%s: stress ResolveDeck differs from FactionDeck", f)
		}
	}
	bad := "data/decks/nobody-nothing.txt"
	if _, err := ResolveDeck(root, cards, bad); err == nil || !strings.Contains(err.Error(), "nobody-nothing.txt") {
		t.Errorf("unknown decklist: want error naming the path, got %v", err)
	}
}
