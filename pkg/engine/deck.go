package engine

import (
	"errors"
	"fmt"
	"sort"
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
