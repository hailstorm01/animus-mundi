package sim

import (
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0179: the monitors read the effective keywords the engine
// reports, not the card's printed ones. Each asset here has keywords that
// differ from its card's. Each TestMonitorKeywords_Case<N> is Acceptance Case N.

// withKeywords is a with effective keywords kw.
func withKeywords(a engine.AssetSnap, kw engine.Keywords) engine.AssetSnap {
	a.Keywords = kw
	return a
}

// Case 1: an attack's defender is legal by the snapshot's Ethereal and Guardian.
func TestMonitorKeywords_Case1_Attack(t *testing.T) {
	att := snapAsset(10, "Vanilla", 0, zMT)
	foe := snapAsset(11, "Vanilla", 1, zMT)

	ethereal := withKeywords(snapAsset(12, "Vanilla", 1, zMT), engine.Keywords{Ethereal: true})
	f, seq := attackFeed(baseSnap(0, 0, att, ethereal), 10, 12)
	wantViolation(t, "defender is Ethereal, its card is not", f.m, "attack-declaration", attackRules, seq)

	guardian := withKeywords(snapAsset(12, "Vanilla", 1, zMT), engine.Keywords{Guardian: true})
	f, seq = attackFeed(baseSnap(0, 0, att, foe, guardian), 10, 11)
	wantViolation(t, "defender beside a Guardian whose card is not one", f.m, "attack-declaration", attackRules, seq)

	ghost := withKeywords(snapAsset(15, "Ghost", 1, zMT), engine.Keywords{})
	f, _ = attackFeed(baseSnap(0, 0, att, ghost), 10, 15)
	wantClean(t, "defender's card is Ethereal, it is not", f.m)

	bodyguard := withKeywords(snapAsset(16, "Bodyguard", 1, zMT), engine.Keywords{})
	f, _ = attackFeed(baseSnap(0, 0, att, foe, bodyguard), 10, 11)
	wantClean(t, "defender beside a Guardian card that is not a Guardian now", f.m)
}

// Case 2: Tactics from Reserves is legal by the snapshot's Versatile.
func TestMonitorKeywords_Case2_Versatile(t *testing.T) {
	generate := func(a engine.AssetSnap) (*feed, int) {
		f := newFeed()
		f.settle(baseSnap(0, 0, a))
		seq := f.event(engine.Event{Kind: "action", Detail: "Generate Tactics", Player: 0, Source: a.ID, Resource: "Tactics"})
		a.Exhausted = true
		f.settle(baseSnap(0, 1, a))
		return f, seq
	}
	f, _ := generate(withKeywords(snapAsset(20, "Vanilla", 0, zRes), engine.Keywords{Versatile: true}))
	wantClean(t, "Versatile, its card is not", f.m)
	wantCounts(t, "Versatile, its card is not", f.m, "resources", "FR-KWD-002i")

	f, seq := generate(withKeywords(snapAsset(20, "Sneak", 0, zRes), engine.Keywords{}))
	wantViolation(t, "its card is Versatile, it is not", f.m, "resources", resourceRules, seq)
}

// Case 3: a Target decision's options are legal by the snapshot's Ward.
func TestMonitorKeywords_Case3_Ward(t *testing.T) {
	target := func(refs ...int) engine.Event {
		return engine.Event{Kind: "decision", Player: 0,
			Decision: &engine.DecisionRecord{Kind: "Target", Player: 0, Refs: refs}}
	}
	warded := withKeywords(snapAsset(11, "Vanilla", 1, zMT), engine.Keywords{Ward: true})
	f, seq := damageCase(baseSnap(0, 0, warded), target(11))
	wantViolation(t, "offered a Ward asset whose card has none", f.m, "targeting", targetingRules, seq)

	unwarded := withKeywords(snapAsset(12, "Warded", 1, zMT), engine.Keywords{})
	f, _ = damageCase(baseSnap(0, 0, unwarded), target(12),
		engine.Event{Kind: "place", Player: 0, Detail: "player 0 places ability on the stack"})
	wantClean(t, "offered an asset whose card has Ward, while it has none", f.m)
}

// Case 4: a damage event is checked by its own Resistant and Stalwart, whatever
// the card prints and whatever the last settled snapshot said.
func TestMonitorKeywords_Case4_Damage(t *testing.T) {
	// feedDamage settles attacker 10 (power 3) and defender 11 of the given card,
	// which the snapshot shows with no keywords, then sends the events as built.
	feedDamage := func(card string, evs ...engine.Event) (*feed, int) {
		f := damageFeed()
		f.settle(baseSnap(0, 0, powered(10, "Vanilla", 0, 3, 0), withKeywords(powered(11, card, 1, 2, 0), engine.Keywords{})))
		seq := 0
		for _, e := range evs {
			seq = f.event(e)
		}
		return f, seq
	}
	resistant := func(e engine.Event) engine.Event { e.Resistant = 1; return e }
	stalwart := func(e engine.Event) engine.Event { e.Stalwart = true; return e }

	f, _ := feedDamage("Vanilla", resistant(dmgEvent(10, 11, 3, false, 0, 0, 2)))
	wantClean(t, "Resistant since the snapshot, its card is not", f.m)
	wantCounts(t, "Resistant since the snapshot", f.m, "damage-sequence", "FR-KWD-002b")
	f, seq := feedDamage("Vanilla", resistant(dmgEvent(10, 11, 3, false, 0, 0, 3)))
	wantViolation(t, "damage ignores the event's Resistant", f.m, "damage-sequence", damageRules, seq)
	f, _ = feedDamage("Tough", dmgEvent(10, 11, 3, false, 0, 0, 3))
	wantClean(t, "its card is Resistant, it is not", f.m)

	f, _ = feedDamage("Vanilla", combatEvent(10, 11), stalwart(dmgEvent(10, 11, 3, true, 0, 0, 1)))
	wantClean(t, "Stalwart since the snapshot, its card is not", f.m)
	wantCounts(t, "Stalwart since the snapshot", f.m, "damage-sequence", "FR-KWD-002c")
	f, seq = feedDamage("Vanilla", combatEvent(10, 11), stalwart(dmgEvent(10, 11, 3, true, 0, 0, 3)))
	wantViolation(t, "combat damage ignores the event's Stalwart", f.m, "damage-sequence", damageRules, seq)
	f, _ = feedDamage("Wall", combatEvent(10, 11), dmgEvent(10, 11, 3, true, 0, 0, 3))
	wantClean(t, "its card is Stalwart, it is not", f.m)
}

// Case 5: the event has all the check reads, so it runs at once, even on a
// target no settled snapshot shows.
func TestMonitorKeywords_Case5_TargetInNoSnapshot(t *testing.T) {
	f, seq := damageCase(baseSnap(0, 0), dmgEvent(10, 99, 3, false, 0, 0, 1))
	wantViolation(t, "wrong amount, target in no snapshot", f.m, "damage-sequence", damageRules, seq)
}

// T-0181 (directive #117 (c)): real games with a keyword behind Cornered break
// no monitor. Its card prints no Resistant and no other card has it, so each
// FR-KWD-002b check is a damage event that reports the grant.
func TestMonitorKeywords_CorneredResistantInRealGames(t *testing.T) {
	hide := unit("Hide", 2, 4, engine.Keywords{})
	hide.Ability = &ability.Def{Cornered: ability.CorneredKeywords{Resistant: 1}}
	var main []*engine.Card
	for range 10 {
		main = append(main, hide, unit("Vanilla", 3, 2, engine.Keywords{}))
	}
	decks := [2]engine.DeckList{{HQ: hq(3, 20), Main: main}, {HQ: hq(3, 20), Main: main}}
	for _, p := range []struct {
		name   string
		policy func(seed uint64) Policy
	}{{"RandomPolicy", nil}, {"ProgressPolicy", ProgressPolicy}} {
		st, err := SimulateWith(decks, 50, 1, 400, SimOptions{Policy: p.policy})
		if err != nil {
			t.Errorf("%s: %v", p.name, err)
			continue
		}
		if st.Coverage["damage-sequence"]["FR-KWD-002b"] == 0 {
			t.Errorf("%s: no damage event reported the Resistant behind Cornered", p.name)
		}
	}
}
