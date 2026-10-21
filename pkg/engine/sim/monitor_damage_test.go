package sim

import (
	"path/filepath"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0026: the damage-sequence and targeting monitors, and the event
// facts they read. Each TestDamage_Case<N> is Acceptance Case N.

var (
	damageRules    = []string{"FR-CBT-002", "FR-KWD-002b", "FR-KWD-002c", "FR-KWD-002k", "FR-REP-004"}
	targetingRules = []string{"FR-KWD-002h", "FR-TGT-001", "SVY-R3"}
)

// damageDecks gives both seats cards with the mitigation keywords.
func damageDecks() [2]engine.DeckList {
	pool := []*engine.Card{
		unit("Vanilla", 3, 5, engine.Keywords{}),
		unit("Wall", 2, 9, engine.Keywords{Stalwart: true}),
		unit("Tough", 1, 9, engine.Keywords{Resistant: 1}),
		unit("Plated", 2, 9, engine.Keywords{Armor: 2}),
		unit("Warded", 1, 9, engine.Keywords{Ward: true}),
	}
	return [2]engine.DeckList{{HQ: hq(3, 20), Main: pool}, {HQ: hq(3, 20), Main: pool}}
}

func damageFeed() *feed {
	f := &feed{m: NewMonitors(damageDecks(), 3)}
	f.event(engine.Event{Kind: "match-start", Player: -1, Start: &engine.MatchStart{Seed: feedSeed}})
	return f
}

// powered is a snapshot asset with an effective power and Armor pool.
func powered(id int, card string, owner int, power, armor int) engine.AssetSnap {
	a := snapAsset(id, card, owner, zMT)
	a.Power, a.Armor = power, armor
	return a
}

// dmgEvent is a damage event from asset src to asset tgt.
func dmgEvent(src, tgt, raw int, combat bool, armor, absorbed, amount int) engine.Event {
	return engine.Event{Kind: "damage", Player: -1, Source: src, Target: tgt,
		Raw: raw, Combat: combat, Armor: armor, Absorbed: absorbed, Amount: amount}
}

func combatEvent(att, def int) engine.Event {
	return engine.Event{Kind: "combat", Player: -1, Source: att, Target: def}
}

// damageCase feeds a snapshot, then events, and returns the feed and the Seq
// of the last event sent.
func damageCase(snap engine.Snapshot, evs ...engine.Event) (*feed, int) {
	f := damageFeed()
	f.settle(snap)
	seq := 0
	for _, e := range evs {
		seq = f.event(e)
	}
	return f, seq
}

func TestDamage_Case3_DamageSequence(t *testing.T) {
	// Attacker 10 (power 3) and defender 11 of the given card.
	board := func(card string, armor int) engine.Snapshot {
		return baseSnap(0, 0, powered(10, "Vanilla", 0, 3, 0), powered(11, card, 1, 2, armor))
	}
	cases := []struct {
		name string
		snap engine.Snapshot
		evs  []engine.Event
	}{
		{"Amount ignores Resistant", board("Tough", 0),
			[]engine.Event{dmgEvent(10, 11, 3, false, 0, 0, 3)}},
		{"Stalwart applied to non-combat damage", board("Wall", 0),
			[]engine.Event{dmgEvent(10, 11, 3, false, 0, 0, 1)}},
		{"Stalwart not applied to combat damage", board("Wall", 0),
			[]engine.Event{combatEvent(10, 11), dmgEvent(10, 11, 3, true, 0, 0, 3)}},
		{"Armor absorbs nothing", board("Plated", 2),
			[]engine.Event{dmgEvent(10, 11, 2, false, 2, 0, 2)}},
		{"Armor does not match the snapshot", board("Plated", 2),
			[]engine.Event{dmgEvent(10, 11, 2, false, 1, 1, 1)}},
		{"combat Raw differs from the attacker's power", board("Vanilla", 0),
			[]engine.Event{combatEvent(10, 11), dmgEvent(10, 11, 5, true, 0, 0, 5)}},
		{"third Combat damage after one combat", board("Vanilla", 0),
			[]engine.Event{combatEvent(10, 11), dmgEvent(10, 11, 3, true, 0, 0, 3), dmgEvent(11, 10, 2, true, 0, 0, 2),
				dmgEvent(10, 11, 3, true, 0, 0, 3)}},
	}
	for _, c := range cases {
		f, seq := damageCase(c.snap, c.evs...)
		wantViolation(t, c.name, f.m, "damage-sequence", damageRules, seq)
	}

	// Valid sequences.
	f, _ := damageCase(board("Plated", 2),
		combatEvent(10, 11),
		dmgEvent(11, 10, 2, true, 0, 0, 2),
		dmgEvent(10, 11, 3, true, 2, 2, 1))
	wantClean(t, "combat pair into Armor, reversed order", f.m)
	wantCounts(t, "combat pair", f.m, "damage-sequence", "FR-REP-004", "FR-CBT-002", "FR-KWD-002k")

	f, _ = damageCase(board("Plated", 2),
		dmgEvent(10, 11, 1, false, 2, 1, 0),
		dmgEvent(10, 11, 2, false, 1, 1, 1))
	wantClean(t, "two hits on one Armor pool", f.m)

	f, _ = damageCase(board("Wall", 0), combatEvent(10, 11),
		dmgEvent(10, 11, 3, true, 0, 0, 1), dmgEvent(11, 10, 2, true, 0, 0, 2),
		dmgEvent(10, 11, 3, false, 0, 0, 3))
	wantClean(t, "Stalwart caps only combat", f.m)
	wantCounts(t, "Stalwart", f.m, "damage-sequence", "FR-KWD-002c")

	f, _ = damageCase(board("Tough", 0), dmgEvent(10, 11, 3, false, 0, 0, 2))
	wantClean(t, "Resistant", f.m)
	wantCounts(t, "Resistant", f.m, "damage-sequence", "FR-KWD-002b")
}

func TestDamage_Case3_Targeting(t *testing.T) {
	target := func(player int, refs ...int) engine.Event {
		return engine.Event{Kind: "decision", Player: player,
			Decision: &engine.DecisionRecord{Kind: "Target", Player: player, Refs: refs}}
	}
	board := baseSnap(0, 0, powered(10, "Vanilla", 0, 3, 0), powered(11, "Vanilla", 1, 3, 0), powered(12, "Warded", 1, 1, 0))

	// The first event after the decision breaks the rule; the violation is
	// reported at the decision.
	f, _ := damageCase(board, target(0, 11))
	dseq := f.seq
	f.event(dmgEvent(10, 11, 1, false, 0, 0, 1))
	wantViolation(t, "Target decision followed by damage", f.m, "targeting", targetingRules, dseq)

	f, _ = damageCase(board, target(0, 11))
	dseq = f.seq
	f.event(engine.Event{Kind: "place", Player: 1, Detail: "player 1 places Attack on the stack"})
	wantViolation(t, "place from the other player", f.m, "targeting", targetingRules, dseq)

	f, seq := damageCase(board, target(0, 11, 12))
	wantViolation(t, "opponent's Ward asset offered", f.m, "targeting", targetingRules, seq)

	// Valid.
	f, _ = damageCase(board, target(0, 11), target(0, 11),
		engine.Event{Kind: "place", Player: 0, Detail: "player 0 places ability on the stack"})
	wantClean(t, "two slots then a place", f.m)
	wantCounts(t, "two slots then a place", f.m, "targeting", "FR-TGT-001", "SVY-R3", "FR-KWD-002h")

	// An allied Ward asset is a fine option.
	f, _ = damageCase(board, target(1, 12, 11),
		engine.Event{Kind: "place", Player: 1, Detail: "player 1 places ability on the stack"})
	wantClean(t, "allied Ward asset offered", f.m)

	// The game ends after a Target decision: dropped.
	f, _ = damageCase(board, target(0, 11))
	wantClean(t, "decision at game end", f.m)
}

// Case 2 (sweep): every Deploy, March, Attack and Scratch action is followed,
// past any decisions, by the place event of its item.
func TestDamage_Case2_PlaceFollowsEveryAction(t *testing.T) {
	var cols []*gameCollector
	opt := SimOptions{DeckNames: simDeckNames, Observer: func(game int, s uint64) engine.Observer {
		c := &gameCollector{}
		cols = append(cols, c)
		return c
	}}
	if _, err := SimulateWith(simStarterDecks(t), 20, 1, 400, opt); err != nil {
		t.Fatal(err)
	}
	checked := map[string]int{}
	for g, c := range cols {
		for i, e := range c.events {
			if e.Kind != "action" {
				continue
			}
			kind, _, _ := strings.Cut(e.Detail, " ")
			switch kind {
			case "Deploy", "March", "Attack", "Scratch":
			default:
				continue
			}
			j := i + 1
			for j < len(c.events) && c.events[j].Kind == "decision" {
				j++
			}
			if j >= len(c.events) {
				t.Errorf("game %d: %s action at event %d has no event after it", g+1, kind, e.Seq)
				continue
			}
			p := c.events[j]
			ok := p.Kind == "place" && p.Player == e.Player
			if kind == "Deploy" {
				ok = ok && p.Inst == e.Inst
			} else {
				ok = ok && p.Source == e.Source
			}
			if !ok {
				t.Errorf("game %d: %s action at event %d is followed by %+v, want its place event", g+1, kind, e.Seq, p)
			}
			checked[kind]++
		}
	}
	for _, k := range []string{"Deploy", "March", "Attack", "Scratch"} {
		if checked[k] == 0 {
			t.Errorf("no %s action in 20 games", k)
		}
	}
}

// Case 4: the monitors hold on real games.
func TestDamage_Case4_RealGamesAreClean(t *testing.T) {
	games, stressGames := 500, 200
	if testing.Short() {
		games, stressGames = 50, 20
	}
	cards, err := engine.LoadCards(filepath.Join("..", "..", "..", "data", "cards"))
	if err != nil {
		t.Fatal(err)
	}
	var stress [2]engine.DeckList
	for i, f := range []string{"SOULLESS", "CONCORDIA"} {
		if stress[i], err = FactionDeck(cards, f); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range []struct {
		name   string
		policy func(seed uint64) Policy
	}{{"RandomPolicy", nil}, {"ProgressPolicy", ProgressPolicy}} {
		if _, err := SimulateWith(simStarterDecks(t), games, 1, 400, SimOptions{DeckNames: simDeckNames, Policy: p.policy}); err != nil {
			t.Errorf("%s, starter decks: %v", p.name, err)
		}
		if _, err := SimulateWith(stress, stressGames, 1, 400, SimOptions{Policy: p.policy}); err != nil {
			t.Errorf("%s, stress decks: %v", p.name, err)
		}
	}
}
