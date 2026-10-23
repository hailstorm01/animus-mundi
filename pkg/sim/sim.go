// Package sim drives matches on the rules kernel: the seat policies
// (RandomPolicy, ProgressPolicy), Playout for one game, Simulate and
// SimulateWith for many, the simulation decks (FactionDeck, SimDecks,
// ResolveDeck) and the rule monitors (Monitors, Violation, Coverage) that
// check named FRD 01 rules in every simulated game.
//
// It is a client of the engine's exported API: it imports engine only from
// this module, and engine never imports it. cardtown lint fences checks its
// imports. CheckInvariants and RandomChooser stay in engine.
package sim

import (
	"fmt"
	"sort"
	"strings"

	"animus-mundi/pkg/engine"
)

// FactionDeck builds a 50-card simulation deck from the faction's executable
// cards (plus executable Neutral cards), cycling them to reach 50. The result
// ignores rarity copy limits, so it is a stress-test deck, not a legal one.
// The faction's first Headquarters is used with its printed abilities ignored.
func FactionDeck(cards map[string]*engine.Card, faction string) (engine.DeckList, error) {
	var names []string
	for name := range cards {
		names = append(names, name)
	}
	sort.Strings(names)

	var deck engine.DeckList
	var pool []*engine.Card
	for _, name := range names {
		c := cards[name]
		switch {
		case c.Type == engine.TypeHeadquarters && strings.EqualFold(c.Faction, faction) && deck.HQ == nil:
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
	Coverage       Coverage // rule monitor checks summed over games; nil when NoMonitors
}

// SimOptions are the optional extras of SimulateWith.
type SimOptions struct {
	DeckNames  [2]string                                   // recorded in each match-start
	Observer   func(game int, seed uint64) engine.Observer // optional; one observer per game, game counts from 1
	Policy     func(seed uint64) Policy                    // nil means RandomPolicy(seed, DefaultEndBias), today's behavior
	NoMonitors bool                                        // true runs without rule monitors; nothing else changes
}

// SeatSeed is the seed of seat's chooser and policy in the game with seed:
// seed*2+1 for seat 0 and seed*2+2 for seat 1.
func SeatSeed(seed uint64, seat engine.PlayerID) uint64 { return seed*2 + 1 + uint64(seat) }

// DefaultEndBias is RandomPolicy's chance of ending the turn when other actions remain.
const DefaultEndBias = 0.1

// fanout passes every call to its observers in order.
type fanout []engine.Observer

func (f fanout) Event(e engine.Event) {
	for _, o := range f {
		o.Event(e)
	}
}

func (f fanout) Settled(s engine.Snapshot) {
	for _, o := range f {
		o.Settled(s)
	}
}

// Simulate plays games random-policy matches between two decks.
func Simulate(decks [2]engine.DeckList, games int, seed uint64, maxTurns int) (SimStats, error) {
	return SimulateWith(decks, games, seed, maxTurns, SimOptions{})
}

// SimulateWith is Simulate with the options: deck names for the match-start
// record and one observer per game. Observing does not change the result.
func SimulateWith(decks [2]engine.DeckList, games int, seed uint64, maxTurns int, opt SimOptions) (SimStats, error) {
	st := SimStats{Reasons: map[string]int{}}
	if !opt.NoMonitors {
		st.Coverage = Coverage{}
	}
	for i := range uint64(games) {
		s := seed + i
		cfg := engine.Config{Seed: s, Decks: decks, DeckNames: opt.DeckNames,
			Choosers: [2]engine.Chooser{engine.RandomChooser(SeatSeed(s, 0)), engine.RandomChooser(SeatSeed(s, 1))}}
		var obs fanout
		if opt.Observer != nil {
			obs = append(obs, opt.Observer(int(i)+1, s))
		}
		var mon *Monitors
		if !opt.NoMonitors {
			mon = NewMonitors(decks, int(i)+1)
			obs = append(obs, mon)
		}
		switch len(obs) {
		case 0:
		case 1:
			cfg.Observer = obs[0]
		default:
			cfg.Observer = obs
		}
		g, err := engine.NewGame(cfg)
		if err != nil {
			return st, err
		}
		pol := opt.Policy
		if pol == nil {
			pol = func(seed uint64) Policy { return RandomPolicy(seed, DefaultEndBias) }
		}
		res, err := Playout(g, [2]Policy{pol(SeatSeed(s, 0)), pol(SeatSeed(s, 1))}, maxTurns)
		if mon != nil {
			if v := mon.Violation(); v != nil {
				return st, v
			}
			st.Coverage.add(mon.Coverage())
		}
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
