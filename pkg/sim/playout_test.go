package sim

import (
	"path/filepath"
	"sort"
	"testing"

	"animus-mundi/pkg/engine"
)

// keywordDeck exercises every M1 keyword.
func keywordDeck() []*engine.Card {
	pool := []*engine.Card{
		unit("Vanilla", 2, 2, engine.Keywords{}),
		unit("Resister", 1, 3, engine.Keywords{Resistant: 1}),
		unit("Brittle", 3, 2, engine.Keywords{Resistant: -1}),
		unit("Stalwart", 1, 2, engine.Keywords{Stalwart: true}),
		unit("Brawler", 1, 3, engine.Keywords{Aggressive: 2}),
		unit("Ghost", 2, 1, engine.Keywords{Ethereal: true}),
		unit("Warded", 2, 2, engine.Keywords{Ward: true}),
		unit("Sneak", 1, 2, engine.Keywords{Versatile: true}),
		unit("Bodyguard", 0, 4, engine.Keywords{Guardian: true}),
		unit("Plated", 2, 1, engine.Keywords{Armor: 2}),
		unit("Slowpoke", 3, 3, engine.Keywords{DeploysExhausted: true}),
		unit("Berserker", 2, 3, engine.Keywords{Berserk: 2}),
		structure("Bastion", 0, 6, engine.Keywords{Guardian: true}),
	}
	for i, c := range pool {
		c.ReqCost = i % 3
		c.TacCost = i % 2
	}
	var deck []*engine.Card
	for len(deck) < 50 {
		deck = append(deck, pool[len(deck)%len(pool)])
	}
	return deck
}

func TestRandomPlayoutsTerminateAndHoldInvariants(t *testing.T) {
	games := 300
	if testing.Short() {
		games = 30
	}
	for seed := range uint64(games) {
		decks := [2]engine.DeckList{{HQ: hq(3, 20), Main: keywordDeck()}, {HQ: hq(2, 25), Main: keywordDeck()}}
		mon := NewMonitors(decks, int(seed)+1)
		g, err := engine.NewGame(engine.Config{Seed: seed, Decks: decks, Observer: mon})
		if err != nil {
			t.Fatal(err)
		}
		_, err = Playout(g, [2]Policy{RandomPolicy(seed, 0.1), RandomPolicy(seed+1000, 0.1)}, 400)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if v := mon.Violation(); v != nil {
			t.Fatalf("seed %d: %v", seed, v)
		}
	}
}

func TestPlayoutsAreDeterministicPerSeed(t *testing.T) {
	run := func() (engine.Outcome, int) {
		g, _ := engine.NewGame(engine.Config{Seed: 42, Decks: [2]engine.DeckList{{HQ: hq(3, 20), Main: keywordDeck()}, {HQ: hq(3, 20), Main: keywordDeck()}}})
		res, err := Playout(g, [2]Policy{RandomPolicy(1, 0.1), RandomPolicy(2, 0.1)}, 400)
		if err != nil {
			t.Fatal(err)
		}
		return res.Outcome, res.Actions
	}
	o1, n1 := run()
	o2, n2 := run()
	if o1 != o2 || n1 != n2 {
		t.Fatalf("same seed produced different matches: %+v/%d vs %+v/%d", o1, n1, o2, n2)
	}
}

func TestLoadCardsFromRepositoryData(t *testing.T) {
	cards, err := engine.LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) < 150 {
		t.Fatalf("loaded %d cards, expected the full catalog", len(cards))
	}
	var supported []string
	for name, c := range cards {
		if c.Supported() {
			supported = append(supported, name)
		}
	}
	sort.Strings(supported)
	t.Logf("%d of %d cards are executable by the kernel: %v", len(supported), len(cards), supported)
	// KIM-B04 ruling: Tempest Surge prints Berserk 2.
	if c := cards["Tempest Surge"]; c == nil || c.Keywords.Berserk != 2 {
		t.Errorf("Tempest Surge should parse Berserk 2, got %+v", c)
	}
}
