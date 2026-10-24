// Match setup: Config and NewGame (FR-SET-001 to FR-SET-004).

package engine

import (
	"fmt"
	"math/rand/v2"
)

// Config describes a match. Nil callbacks take the FR default for a timeout.
type Config struct {
	Seed  uint64
	Decks [2]DeckList

	// PlayFirst is asked of the randomly chosen decider (FR-SET-003).
	// Nil means Play First, the timeout default.
	PlayFirst func(decider PlayerID) bool

	// Mulligan returns the IDs of hand cards to set aside (FR-SET-002).
	// Nil means keep all 6, the timeout default.
	Mulligan func(p PlayerID, hand []*Instance) []int

	// Choosers answer each seat's Decisions, indexed by PlayerID. A nil
	// entry means that seat always times out (FR-MOD-001 §6).
	Choosers [2]Chooser

	Observer  Observer  // optional; nil means only Game.Log
	DeckNames [2]string // optional; recorded in MatchStart
}

// NewGame runs match setup and returns the game at the start of Turn 1's Main Phase.
func NewGame(cfg Config) (*Game, error) {
	g := &Game{rng: rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9E3779B97F4A7C15))}
	g.choosers = cfg.Choosers
	g.obs = cfg.Observer
	for i := range g.Players {
		pid := PlayerID(i)
		deck := cfg.Decks[i]
		if deck.HQ == nil || deck.HQ.Type != TypeHeadquarters {
			return nil, fmt.Errorf("player %d: deck needs a Headquarters card", i)
		}
		p := &Player{ID: pid}
		g.Players[i] = p
		hq := &Asset{ID: g.newID(), Inst: &Instance{ID: g.newID(), Card: deck.HQ, Owner: pid}, Controller: pid}
		g.placeOnBattlefield(hq, ZoneHeadquarters)
		for _, c := range deck.Main {
			p.Deck = append(p.Deck, &Instance{ID: g.newID(), Card: c, Owner: pid})
		}
	}

	// First-player determination happens before hands are drawn (FR-SET-003).
	decider := PlayerID(g.rng.IntN(2))
	playFirst := true
	if cfg.PlayFirst != nil {
		playFirst = cfg.PlayFirst(decider)
	}
	g.FirstPlayer = decider
	if !playFirst {
		g.FirstPlayer = decider.Opponent()
	}

	// Shuffle, draw 6, single partial mulligan (FR-SET-002).
	for _, p := range g.Players {
		g.shuffle(p)
		for range 6 {
			g.moveTopToHand(p)
		}
	}
	var setAside [2][]string
	var setAsideInsts [2][]int
	for _, p := range g.Players {
		if cfg.Mulligan == nil {
			continue
		}
		var aside []*Instance
		for _, id := range cfg.Mulligan(p.ID, append([]*Instance(nil), p.Hand...)) {
			if inst := handCard(p, id); inst != nil {
				g.takeCard(inst)
				aside = append(aside, inst)
				setAside[p.ID] = append(setAside[p.ID], inst.Card.Name)
				setAsideInsts[p.ID] = append(setAsideInsts[p.ID], inst.ID)
			}
		}
		for range aside {
			g.moveTopToHand(p)
		}
		for _, inst := range aside {
			g.putCard(inst, ZoneDeck, true)
		}
		g.shuffle(p)
	}

	g.emitMatchStart(cfg, decider, playFirst, setAside, setAsideInsts)

	g.Active = g.FirstPlayer
	g.Turn = 1
	g.beginTurn()
	g.settled()
	return g, nil
}
