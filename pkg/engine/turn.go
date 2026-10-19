package engine

import (
	"fmt"
	"math/rand/v2"
)

// DeckList is one player's deck: a Headquarters plus the main deck.
type DeckList struct {
	HQ   *Card
	Main []*Card
}

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
}

// NewGame runs match setup and returns the game at the start of Turn 1's Main Phase.
func NewGame(cfg Config) (*Game, error) {
	g := &Game{rng: rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9E3779B97F4A7C15))}
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
	for _, p := range g.Players {
		if cfg.Mulligan == nil {
			continue
		}
		var aside []*Instance
		for _, id := range cfg.Mulligan(p.ID, append([]*Instance(nil), p.Hand...)) {
			if inst := handCard(p, id); inst != nil {
				p.Hand = remove(p.Hand, inst)
				aside = append(aside, inst)
			}
		}
		for range aside {
			g.moveTopToHand(p)
		}
		p.Deck = append(p.Deck, aside...)
		g.shuffle(p)
	}

	g.Active = g.FirstPlayer
	g.Turn = 1
	g.beginTurn()
	return g, nil
}

func (g *Game) shuffle(p *Player) {
	g.rng.Shuffle(len(p.Deck), func(i, j int) { p.Deck[i], p.Deck[j] = p.Deck[j], p.Deck[i] })
}

func (g *Game) moveTopToHand(p *Player) {
	if len(p.Deck) > 0 {
		p.Hand = append(p.Hand, p.Deck[0])
		p.Deck = p.Deck[1:]
	}
}

// beginTurn runs the Ready and Draw phases, then opens the Main Phase (FR-TRN-001).
func (g *Game) beginTurn() {
	g.Phase = PhaseReady
	for _, a := range g.Assets() {
		if a.Controller == g.Active {
			a.Exhausted = false
		}
	}
	// No passive income (FR-RES-001). No M1 card has start-of-turn triggers.

	g.Phase = PhaseDraw
	if !(g.Turn == 1 && g.Active == g.FirstPlayer) { // Turn 1 draw skip (FR-SET-004)
		g.draw(g.Active, 1)
		if g.Outcome.Over {
			return
		}
	}
	g.Phase = PhaseMain
}

// draw draws n cards. Drawing from an empty deck loses the match (FR-WIN-001).
func (g *Game) draw(pid PlayerID, n int) {
	p := g.Players[pid]
	for range n {
		if len(p.Deck) == 0 {
			g.end(Outcome{Over: true, Winner: pid.Opponent(), Reason: "drew from an empty deck"})
			return
		}
		g.moveTopToHand(p)
	}
}

// endTurn runs End Phase sub-phases 4a-4d and starts the next turn (FR-TRN-001).
func (g *Game) endTurn() {
	g.Phase = PhaseEnd
	// 4a/4b: no M1 card has end-of-turn triggers.
	// 4c: pools reset, "until end of turn" modifiers expire.
	for _, p := range g.Players {
		p.Req, p.Tac = 0, 0
	}
	for _, a := range g.Assets() {
		a.TempPower = 0
	}
	// 4d: handover.
	g.Active = g.Active.Opponent()
	g.Turn++
	g.beginTurn()
}
