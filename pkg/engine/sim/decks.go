package sim

import (
	"fmt"
	"path/filepath"
	"strings"

	"animus-mundi/pkg/engine"
)

// SimDecks loads and validates decklists for the given factions and deck name.
func SimDecks(cards map[string]*engine.Card, decksDir, name string, factions [2]string) ([2]engine.DeckList, error) {
	var decks [2]engine.DeckList
	for i, f := range factions {
		filename := fmt.Sprintf("%s-%s.txt", strings.ToLower(f), name)
		path := filepath.Join(decksDir, filename)
		d, err := engine.LoadDeck(path, cards)
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
