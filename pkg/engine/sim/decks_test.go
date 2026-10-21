package sim

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

func TestSimDecks_Case13_BuildsValidDecks(t *testing.T) {
	cards, err := engine.LoadCards(filepath.Join("..", "..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	decksDir := filepath.Join("..", "..", "..", "data", "decks")
	decks, err := SimDecks(cards, decksDir, "starter", [2]string{"ANIMUS", "PATH"})
	if err != nil {
		t.Fatalf("SimDecks failed: %v", err)
	}
	if decks[0].HQ == nil || decks[0].HQ.Name != "Animoceps Rager" {
		t.Errorf("expected P0 HQ Animoceps Rager, got %v", decks[0].HQ)
	}
	if len(decks[0].Main) != 50 {
		t.Errorf("expected P0 50 main cards, got %d", len(decks[0].Main))
	}
	if decks[1].HQ == nil || decks[1].HQ.Name != "Eternal Watchman" {
		t.Errorf("expected P1 HQ Eternal Watchman, got %v", decks[1].HQ)
	}
	if len(decks[1].Main) != 50 {
		t.Errorf("expected P1 50 main cards, got %d", len(decks[1].Main))
	}
}

func TestSimDecks_Case14_MissingFileNamesPath(t *testing.T) {
	cards, err := engine.LoadCards(filepath.Join("..", "..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	decksDir := filepath.Join("..", "..", "..", "data", "decks")
	_, err = SimDecks(cards, decksDir, "starter", [2]string{"SOULLESS", "PATH"})
	if err == nil {
		t.Fatal("expected error for missing soulless-starter.txt, got nil")
	}
	if !strings.Contains(err.Error(), "soulless-starter.txt") {
		t.Errorf("expected error to name missing file 'soulless-starter.txt', got: %v", err)
	}
}

func TestSimDecks_Case15_InvalidDeckFails(t *testing.T) {
	cards := map[string]*engine.Card{
		"Deck HQ ANIMUS": deckHQ("ANIMUS"),
		"Trooper":        deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	dir := t.TempDir()
	writeDeckFile(t, dir, "animus-invalid.txt", "HQ: Deck HQ ANIMUS\n10 Trooper\n")
	_, err := SimDecks(cards, dir, "invalid", [2]string{"ANIMUS", "ANIMUS"})
	if err == nil {
		t.Fatal("expected error when ValidateDeck fails on loaded deck, got nil")
	}
	if !errors.Is(err, engine.ErrIllegalDeck) {
		t.Errorf("expected error to wrap engine.ErrIllegalDeck, got: %v", err)
	}
}

func TestSimDecks_Case16_FactionDeckStressDecksUnchanged(t *testing.T) {
	cards, err := engine.LoadCards(filepath.Join("..", "..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	for _, f := range []string{"SOULLESS", "CONCORDIA"} {
		d, err := FactionDeck(cards, f)
		if err != nil {
			t.Fatalf("FactionDeck(%s) failed: %v", f, err)
		}
		if d.HQ == nil {
			t.Errorf("expected HQ for %s", f)
		}
		if len(d.Main) != 50 {
			t.Errorf("expected 50 cards for %s stress deck, got %d", f, len(d.Main))
		}
	}
}
