package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// ErrIllegalDeck is wrapped by every deck-construction violation (FR-SET-001).
var ErrIllegalDeck = errors.New("illegal deck")

// minMainDeck is the smallest legal main deck (FR-SET-001); there is no maximum.
const minMainDeck = 50

// copyLimit maps a rarity to its maximum copies per printed name (FR-SET-001).
var copyLimit = map[string]int{
	"COMMON":   4,
	"UNCOMMON": 3,
	"RARE":     2,
	"MYTHIC":   1,
}

// hqFactions are the factions a Headquarters may belong to (FR-SET-001).
var hqFactions = []string{"CONCORDIA", "SOULLESS", "PATH", "ANIMUS"}

func illegalDeck(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrIllegalDeck, fmt.Sprintf(format, args...))
}

// ValidateDeck returns nil if d is construction-legal under FR-SET-001.
// Otherwise it returns every violation joined (errors.Join), each wrapping
// ErrIllegalDeck and naming the offending card where there is one.
func ValidateDeck(d DeckList) error {
	var errs []error

	if len(d.Main) < minMainDeck {
		errs = append(errs, illegalDeck("main deck has %d cards, minimum is %d", len(d.Main), minMainDeck))
	}

	hqFaction := ""
	switch {
	case d.HQ == nil:
		errs = append(errs, illegalDeck("no Headquarters card"))
	case d.HQ.Type != TypeHeadquarters:
		errs = append(errs, illegalDeck("%s is not a Headquarters card", d.HQ.Name))
	case strings.EqualFold(d.HQ.Faction, "NEUTRAL"):
		errs = append(errs, illegalDeck("Headquarters %s cannot be Neutral", d.HQ.Name))
	case !slices.ContainsFunc(hqFactions, func(f string) bool { return strings.EqualFold(f, d.HQ.Faction) }):
		errs = append(errs, illegalDeck("Headquarters %s has faction %q, must be one of Concordia, Soulless, Path or Animus", d.HQ.Name, d.HQ.Faction))
	default:
		hqFaction = d.HQ.Faction
	}

	counts := map[string]int{}
	limits := map[string]int{}
	var order []string
	for _, c := range d.Main {
		if c == nil {
			errs = append(errs, illegalDeck("main deck contains a nil card"))
			continue
		}
		switch {
		case c.Type == TypeHeadquarters:
			errs = append(errs, illegalDeck("Headquarters %s cannot be in the main deck", c.Name))
		case c.HasSubtype("Scratch Company"):
			errs = append(errs, illegalDeck("Scratch Company %s cannot be in the main deck", c.Name))
		}
		// Without a usable HQ there is no faction to compare against.
		if hqFaction != "" && !strings.EqualFold(c.Faction, hqFaction) && !strings.EqualFold(c.Faction, "NEUTRAL") {
			errs = append(errs, illegalDeck("%s (%s) does not match Headquarters faction %s", c.Name, c.Faction, hqFaction))
		}
		if c.Type == TypeHeadquarters {
			continue // already reported; not a rarity-limited main-deck card
		}
		limit, ok := copyLimit[strings.ToUpper(c.Rarity)]
		if !ok {
			errs = append(errs, illegalDeck("%s has unrecognised rarity %q", c.Name, c.Rarity))
			continue
		}
		if counts[c.Name] == 0 {
			order = append(order, c.Name)
		}
		counts[c.Name]++
		limits[c.Name] = limit
	}

	sort.Strings(order)
	for _, name := range order {
		if counts[name] > limits[name] {
			errs = append(errs, illegalDeck("%s has %d copies, limit is %d", name, counts[name], limits[name]))
		}
	}

	return errors.Join(errs...)
}

// LoadDeck reads a decklist file and resolves its names against cards.
func LoadDeck(path string, cards map[string]*Card) (DeckList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DeckList{}, fmt.Errorf("reading decklist %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	var deck DeckList
	hqSeen := false
	seenCards := map[string]bool{}

	for i, rawLine := range lines {
		lineNum := i + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "HQ:") {
			if hqSeen {
				return DeckList{}, fmt.Errorf("line %d: multiple HQ lines", lineNum)
			}
			cardName := strings.TrimSpace(strings.TrimPrefix(line, "HQ:"))
			if cardName == "" {
				return DeckList{}, fmt.Errorf("line %d: missing HQ card name", lineNum)
			}
			c, ok := cards[cardName]
			if !ok {
				return DeckList{}, fmt.Errorf("line %d: unknown card name %q", lineNum, cardName)
			}
			deck.HQ = c
			hqSeen = true
			continue
		}

		idx := strings.IndexAny(line, " \t")
		if idx == -1 {
			return DeckList{}, fmt.Errorf("line %d: invalid entry format %q (expected <count> <card name>)", lineNum, line)
		}
		countStr := strings.TrimSpace(line[:idx])
		cardName := strings.TrimSpace(line[idx+1:])
		if cardName == "" {
			return DeckList{}, fmt.Errorf("line %d: missing card name", lineNum)
		}

		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			return DeckList{}, fmt.Errorf("line %d: invalid copy count %q (must be a positive integer)", lineNum, countStr)
		}

		if seenCards[cardName] {
			return DeckList{}, fmt.Errorf("line %d: duplicate card %q (lists must be canonical)", lineNum, cardName)
		}
		seenCards[cardName] = true

		c, ok := cards[cardName]
		if !ok {
			return DeckList{}, fmt.Errorf("line %d: unknown card name %q", lineNum, cardName)
		}

		for range count {
			deck.Main = append(deck.Main, c)
		}
	}

	if !hqSeen {
		return DeckList{}, fmt.Errorf("%s: missing HQ line", path)
	}

	return deck, nil
}

// SimDecks loads and validates decklists for the given factions and deck name.
func SimDecks(cards map[string]*Card, decksDir, name string, factions [2]string) ([2]DeckList, error) {
	var decks [2]DeckList
	for i, f := range factions {
		filename := fmt.Sprintf("%s-%s.txt", strings.ToLower(f), name)
		path := filepath.Join(decksDir, filename)
		d, err := LoadDeck(path, cards)
		if err != nil {
			return decks, err
		}
		if err := ValidateDeck(d); err != nil {
			return decks, fmt.Errorf("%s: %w", filename, err)
		}
		decks[i] = d
	}
	return decks, nil
}
