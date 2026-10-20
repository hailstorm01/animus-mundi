package engine

import (
	"errors"
	"fmt"
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

func (s *scene) alive(a *Asset) bool { return s.g.Asset(a.ID) != nil }

func (s *scene) wantDamage(a *Asset, want int) {
	s.t.Helper()
	if a.Damage != want {
		s.t.Errorf("%s damage = %d, want %d", a, a.Damage, want)
	}
}

// realCard returns a shallow copy of the named card from data/cards, loaded
// once per test binary.
func realCard(name string) (*Card, error) {
	return nil, errors.New("realCard: not implemented")
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
