package sim

import (
	"fmt"
	"path/filepath"
	"strings"

	"animus-mundi/pkg/engine"
)

// DecklistName is the recorded name of a decklist: its path from the repo
// root, slash-separated.
func DecklistName(faction, deck string) string {
	return fmt.Sprintf("data/decks/%s-%s.txt", strings.ToLower(faction), deck)
}

// StressName is the recorded name of a faction's stress deck.
func StressName(faction string) string { return "stress:" + faction }

// ResolveDeck turns a recorded deck name back into a deck: a stress name
// builds FactionDeck, any other name is a decklist path under repoRoot,
// loaded with LoadDeck (not validated, as replay does today).
func ResolveDeck(repoRoot string, cards map[string]*engine.Card, name string) (engine.DeckList, error) {
	if faction, ok := strings.CutPrefix(name, "stress:"); ok {
		return FactionDeck(cards, faction)
	}
	return engine.LoadDeck(filepath.Join(repoRoot, filepath.FromSlash(name)), cards)
}

// SimDecks loads and validates decklists for the given factions and deck name.
func SimDecks(cards map[string]*engine.Card, decksDir, name string, factions [2]string) ([2]engine.DeckList, error) {
	var decks [2]engine.DeckList
	for i, f := range factions {
		filename := strings.TrimPrefix(DecklistName(f, name), "data/decks/")
		file := filepath.Join(decksDir, filename)
		d, err := engine.LoadDeck(file, cards)
		if err != nil {
			return decks, err
		}
		if err := engine.ValidateDeck(d); err != nil {
			return decks, fmt.Errorf("%s: %w", filename, err)
		}
		decks[i] = d
	}
	return decks, nil
}
