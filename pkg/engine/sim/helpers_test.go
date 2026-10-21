package sim

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"animus-mundi/pkg/engine"
)

// Copies of the engine test helpers (scenario_test.go, deck_test.go and
// deck_load_test.go) that the moved tests use. Those are unexported test code
// in package engine, so another package cannot import them.

func hq(power, health int) *engine.Card {
	return &engine.Card{Name: fmt.Sprintf("Test HQ %d/%d", power, health), Type: engine.TypeHeadquarters,
		Power: power, Health: health, ReqYield: 1, TacYield: 2, DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
}

func unit(name string, power, health int, kw engine.Keywords) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeUnit, Power: power, Health: health, ReqYield: 1, TacYield: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater}, Keywords: kw}
}

func structure(name string, power, health int, kw engine.Keywords) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeStructure, Power: power, Health: health, ReqYield: 2, TacYield: 2,
		DeployZones: []engine.Zone{engine.ZoneMainTheater}, Keywords: kw}
}

func deckHQ(faction string) *engine.Card {
	return &engine.Card{Name: "Deck HQ " + faction, Type: engine.TypeHeadquarters, Faction: faction, Rarity: "HQ",
		Power: 3, Health: 20, DeployZones: []engine.Zone{engine.ZoneHeadquarters}}
}

func deckCard(name, faction, rarity string) *engine.Card {
	return &engine.Card{Name: name, Type: engine.TypeUnit, Faction: faction, Rarity: rarity, Power: 1, Health: 1,
		DeployZones: []engine.Zone{engine.ZoneReserves, engine.ZoneMainTheater}}
}

func writeDeckFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	p := filepath.Join(dir, filename)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write deck file %s: %v", p, err)
	}
	return p
}

// scene is the small part of the engine's scenario scene that
// TestProgress_Case2_Preferences needs, built from NewGame and exported state.
// Instance and asset IDs start above the range NewGame uses.
type scene struct {
	g      *engine.Game
	nextID int
}

// newScene starts a match where player 0 takes Turn 1, both HQs are 3/20 and
// both decks hold 20 vanilla 1/1 fillers.
func newScene(t *testing.T) *scene {
	t.Helper()
	filler := make([]*engine.Card, 20)
	for i := range filler {
		filler[i] = unit("Filler", 1, 1, engine.Keywords{})
	}
	g, err := engine.NewGame(engine.Config{
		Seed:      1,
		Decks:     [2]engine.DeckList{{HQ: hq(3, 20), Main: filler}, {HQ: hq(3, 20), Main: filler}},
		PlayFirst: func(decider engine.PlayerID) bool { return decider == 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	return &scene{g: g, nextID: 100000}
}

func (s *scene) id() int {
	s.nextID++
	return s.nextID
}

// put places a Readied asset directly onto the battlefield.
func (s *scene) put(p engine.PlayerID, zone engine.Zone, c *engine.Card) *engine.Asset {
	a := &engine.Asset{ID: s.id(), Inst: &engine.Instance{ID: s.id(), Card: c, Owner: p}, Controller: p,
		Zone: zone, Armor: c.Keywords.Armor}
	switch zone {
	case engine.ZoneMainTheater:
		s.g.Theater = append(s.g.Theater, a)
	case engine.ZoneReserves:
		s.g.Players[p].Reserves = append(s.g.Players[p].Reserves, a)
	}
	return a
}

// give puts a card into p's hand.
func (s *scene) give(p engine.PlayerID, c *engine.Card) *engine.Instance {
	inst := &engine.Instance{ID: s.id(), Card: c, Owner: p}
	s.g.Players[p].Hand = append(s.g.Players[p].Hand, inst)
	return inst
}
