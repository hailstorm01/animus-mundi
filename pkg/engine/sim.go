package engine

import (
	"fmt"
	"sort"
	"strings"
)

// FactionDeck builds a 50-card simulation deck from the faction's executable
// cards (plus executable Neutral cards), cycling them to reach 50. The result
// ignores rarity copy limits, so it is a stress-test deck, not a legal one.
// The faction's first Headquarters is used with its printed abilities ignored.
func FactionDeck(cards map[string]*Card, faction string) (DeckList, error) {
	var names []string
	for name := range cards {
		names = append(names, name)
	}
	sort.Strings(names)

	var deck DeckList
	var pool []*Card
	for _, name := range names {
		c := cards[name]
		switch {
		case c.Type == TypeHeadquarters && strings.EqualFold(c.Faction, faction) && deck.HQ == nil:
			deck.HQ = c
		case c.Supported() && !c.HasSubtype("Scratch Company") &&
			(strings.EqualFold(c.Faction, faction) || strings.EqualFold(c.Faction, "NEUTRAL")):
			pool = append(pool, c)
		}
	}
	if deck.HQ == nil {
		return deck, fmt.Errorf("no Headquarters for faction %s", faction)
	}
	if len(pool) == 0 {
		return deck, fmt.Errorf("faction %s has no cards the kernel can execute yet", faction)
	}
	for len(deck.Main) < 50 {
		deck.Main = append(deck.Main, pool[len(deck.Main)%len(pool)])
	}
	return deck, nil
}

// SimStats aggregates many playouts.
type SimStats struct {
	Games          int
	Wins           [2]int
	Draws          int
	FirstPlayerWin int
	TotalTurns     int
	Reasons        map[string]int
}

// Simulate plays games random-policy matches between two decks.
func Simulate(decks [2]DeckList, games int, seed uint64, maxTurns int) (SimStats, error) {
	st := SimStats{Reasons: map[string]int{}}
	for i := range uint64(games) {
		s := seed + i
		g, err := NewGame(Config{Seed: s, Decks: decks,
			Choosers: [2]Chooser{RandomChooser(s*2 + 1), RandomChooser(s*2 + 2)}})
		if err != nil {
			return st, err
		}
		res, err := Playout(g, [2]Policy{RandomPolicy(s*2+1, 0.1), RandomPolicy(s*2+2, 0.1)}, maxTurns)
		if err != nil {
			return st, fmt.Errorf("seed %d: %w", s, err)
		}
		st.Games++
		st.TotalTurns += res.Turns
		st.Reasons[res.Outcome.Reason]++
		if res.Outcome.Draw {
			st.Draws++
			continue
		}
		st.Wins[res.Outcome.Winner]++
		if res.Outcome.Winner == res.FirstPlayer {
			st.FirstPlayerWin++
		}
	}
	return st, nil
}
