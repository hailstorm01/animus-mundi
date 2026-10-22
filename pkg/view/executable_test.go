package view_test

import (
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

// Acceptance Case 8: Executable flag.
// A new test builds a game from the starter decks with a fixed seed.
// Every CardView in the seat's hand has Executable == c.Supported() for that card.
// Seed 1 is chosen: seat 0's initial hand contains Despoiler Larva (supported)
// as well as Shifting Impossibility, Veil Render, Rapturous Sermon, Sinuous Growth,
// and Animoceps Spores (unsupported).
func TestView_Case8_ExecutableFlag(t *testing.T) {
	decks := starterDecks(t)
	g, err := engine.NewGame(engine.Config{
		Seed:  1,
		Decks: decks,
	})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	v := view.PlayerView(g, 0)
	if len(v.You.Hand) == 0 {
		t.Fatalf("seat 0 has empty hand")
	}

	hasSupported := false
	hasUnsupported := false

	// Compare each CardView in seat 0's hand against the actual game card instance.
	for i, cv := range v.You.Hand {
		inst := g.Players[0].Hand[i]
		if cv.Card != inst.Card.Name {
			t.Fatalf("card %d: view name %q != instance card %q", i, cv.Card, inst.Card.Name)
		}
		expected := inst.Card.Supported()
		if cv.Executable != expected {
			t.Errorf("card %d (%s): cv.Executable = %v, want %v", i, cv.Card, cv.Executable, expected)
		}
		if expected {
			hasSupported = true
		} else {
			hasUnsupported = true
		}
	}

	if !hasSupported {
		t.Errorf("seed 1: expected at least one supported card in seat 0 hand")
	}
	if !hasUnsupported {
		t.Errorf("seed 1: expected at least one unsupported card in seat 0 hand")
	}
}
