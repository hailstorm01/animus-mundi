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

// MatchDecks returns the two seats' decks and their recorded names: with deck,
// the validated decklists data/decks/<faction>-<deck>.txt named by DecklistName;
// without it, each faction's stress deck named by StressName.
func MatchDecks(repoRoot string, cards map[string]*engine.Card, deck string, factions [2]string) ([2]engine.DeckList, [2]string, error) {
	var names [2]string
	if deck != "" {
		decks, err := SimDecks(cards, filepath.Join(repoRoot, "data", "decks"), deck, factions)
		if err != nil {
			return decks, names, err
		}
		for i, f := range factions {
			names[i] = DecklistName(f, deck)
		}
		return decks, names, nil
	}
	var decks [2]engine.DeckList
	for i, f := range factions {
		d, err := FactionDeck(cards, f)
		if err != nil {
			return decks, [2]string{}, err
		}
		decks[i] = d
		names[i] = StressName(f)
	}
	return decks, names, nil
}
