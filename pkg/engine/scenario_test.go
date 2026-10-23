package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

// Scenario helpers. Rules tests are named TestFR_<DOM>_<NNN>_<Behavior> so
// `cardtown lint rules` can trace them back to docs/frd/01_Game_Rules_FRD.md.

func hq(power, health int) *Card {
	return &Card{Name: fmt.Sprintf("Test HQ %d/%d", power, health), Type: TypeHeadquarters,
		Power: power, Health: health, ReqYield: 1, TacYield: 2, DeployZones: []Zone{ZoneHeadquarters}}
}

func unit(name string, power, health int, kw Keywords) *Card {
	return &Card{Name: name, Type: TypeUnit, Power: power, Health: health, ReqYield: 1, TacYield: 1,
		DeployZones: []Zone{ZoneReserves, ZoneMainTheater}, Keywords: kw}
}

func structure(name string, power, health int, kw Keywords) *Card {
	return &Card{Name: name, Type: TypeStructure, Power: power, Health: health, ReqYield: 2, TacYield: 2,
		DeployZones: []Zone{ZoneMainTheater}, Keywords: kw}
}

func filler(n int) []*Card {
	out := make([]*Card, n)
	for i := range out {
		out[i] = unit("Filler", 1, 1, Keywords{})
	}
	return out
}

type scene struct {
	t *testing.T
	g *Game

	asks map[PlayerID][]Decision // decisions put to each seat's scripted chooser
}

// newScene starts a match where player 0 takes Turn 1, both HQs are 3/20 and
// both decks hold 20 vanilla 1/1 fillers.
func newScene(t *testing.T) *scene {
	t.Helper()
	g, err := NewGame(Config{
		Seed:      1,
		Decks:     [2]DeckList{{HQ: hq(3, 20), Main: filler(20)}, {HQ: hq(3, 20), Main: filler(20)}},
		PlayFirst: func(decider PlayerID) bool { return decider == 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	return &scene{t: t, g: g}
}

// put places a Readied asset directly onto the battlefield, bypassing the stack.
func (s *scene) put(p PlayerID, zone Zone, c *Card) *Asset {
	a := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: c, Owner: p}, Controller: p, Armor: c.Keywords.Armor}
	s.g.placeOnBattlefield(a, zone)
	return a
}

// putScratch places a face-down, Exhausted Scratch Company in p's Reserves,
// built as the kernel's Deploy Scratch Company does, and returns it.
func (s *scene) putScratch(p PlayerID) *Asset {
	a := &Asset{ID: s.g.newID(), Inst: &Instance{ID: s.g.newID(), Card: unit("Filler", 1, 1, Keywords{}), Owner: p},
		Controller: p, FaceDown: true, Exhausted: true}
	s.g.placeOnBattlefield(a, ZoneReserves)
	return a
}

// onTop puts new instances of cards on top of p's deck, cards[0] on top.
func (s *scene) onTop(p PlayerID, cards ...*Card) []*Instance {
	insts := make([]*Instance, len(cards))
	for i, c := range cards {
		insts[i] = &Instance{ID: s.g.newID(), Card: c, Owner: p}
	}
	pl := s.g.Players[p]
	pl.Deck = append(append([]*Instance{}, insts...), pl.Deck...)
	return insts
}

// give puts a card into p's hand.
func (s *scene) give(p PlayerID, c *Card) *Instance {
	inst := &Instance{ID: s.g.newID(), Card: c, Owner: p}
	s.g.Players[p].Hand = append(s.g.Players[p].Hand, inst)
	return inst
}

func (s *scene) do(a Action) {
	s.t.Helper()
	if err := s.g.Apply(a); err != nil {
		s.t.Fatalf("%s rejected: %v", a.Kind, err)
	}
}

func (s *scene) reject(a Action) {
	s.t.Helper()
	if err := s.g.Apply(a); !errors.Is(err, ErrIllegal) {
		s.t.Fatalf("%s should be illegal, got err=%v", a.Kind, err)
	}
}

func (s *scene) attack(att, def *Asset) Action {
	return Action{Kind: ActAttack, Player: att.Controller, Asset: att.ID, Target: def.ID}
}

func (s *scene) cleave(att, def, sec *Asset) Action {
	a := s.attack(att, def)
	a.Secondary = sec.ID
	return a
}

func (s *scene) endTurn() { s.do(Action{Kind: ActEndTurn, Player: s.g.Active}) }

// moveTo moves battlefield asset a to zone z (Main Theater or Reserves),
// keeping the zone lists consistent. It sets state and asserts nothing.
func (s *scene) moveTo(a *Asset, z Zone) {
	s.g.removeFromBattlefield(a)
	s.g.placeOnBattlefield(a, z)
}

func (s *scene) alive(a *Asset) bool { return s.g.Asset(a.ID) != nil }

func (s *scene) wantDamage(a *Asset, want int) {
	s.t.Helper()
	if a.Damage != want {
		s.t.Errorf("%s damage = %d, want %d", a, a.Damage, want)
	}
}

var (
	realCardsOnce sync.Once
	realCards     map[string]*Card
	realCardsErr  error
)

// realCard returns a shallow copy of the named card from data/cards, loaded
// once per test binary.
func realCard(name string) (*Card, error) {
	realCardsOnce.Do(func() {
		realCards, realCardsErr = LoadCards(filepath.Join("..", "..", "data", "cards"))
	})
	if realCardsErr != nil {
		return nil, realCardsErr
	}
	c, ok := realCards[name]
	if !ok {
		return nil, fmt.Errorf("realCard: no card named %q in data/cards", name)
	}
	cp := *c
	return &cp, nil
}

// card returns the real repository card, or fails the test.
func (s *scene) card(name string) *Card {
	s.t.Helper()
	c, err := realCard(name)
	if err != nil {
		s.t.Fatalf("%v", err)
	}
	return c
}

// script installs a chooser for seat p that returns answers in order and
// records every Decision it is asked. Asked past the end of the script, it
// reports an error naming the prompt and returns a timeout.
func (s *scene) script(p PlayerID, answers ...Answer) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	next := 0
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		if next >= len(answers) {
			s.t.Errorf("player %d chooser asked past the end of its script: %q", p, d.Prompt)
			return Answer{Timeout: true}
		}
		next++
		return answers[next-1]
	}
}

// asked returns the decisions seat p's scripted chooser was asked, in order.
func (s *scene) asked(p PlayerID) []Decision { return s.asks[p] }
