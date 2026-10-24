package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0030: optional ("you may") triggers (SVY-R9), "up to N" chosen
// targets (SVY-R11, FR-TGT-002), partial invalidation (FR-TGT-003) and "other"
// (FR-TGT-001). Cards are synthetic. SVY-R9 and SVY-R11 are ruled but not yet in
// FRD 01, so tests that pin them carry a `// Covers: SVY-R9` or
// `// Covers: SVY-R11` comment and a TestFR_ name for the closest FRD 01
// requirement. The choice machinery (DecideOptional, the Min fallback of an
// "up to" DecideTarget, distinct picks) is T-0006's tested FR-MOD-001 behavior.
//
// Every scene has Hero (player 0, Main Theater, 30 health) attack Dummy with the
// given attack triggers. Hero has no Aggressive, so the attack triggers are its
// only Step 2 triggers. Dummy has 0 power, so combat adds no damage to anyone but
// Dummy; the tests count only non-combat damage.

// arena is the board of every scene. Assets that can be chosen, besides the two
// Headquarters: Hero, Dummy, E1 (health 5), E2 and A1 (Hero's ally).
type arena struct {
	s                       *scene
	hero, dummy, e1, e2, a1 *Asset
}

func newArena(t *testing.T, heroPower int, trigs ...ability.Trigger) *arena {
	t.Helper()
	s := newScene(t)
	a := &arena{s: s}
	a.hero = s.put(0, ZoneMainTheater, striker("Hero", heroPower, 30, Keywords{}, trigs...))
	a.dummy = s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, Keywords{}))
	a.e1 = s.put(1, ZoneMainTheater, unit("E1", 0, 5, Keywords{}))
	a.e2 = s.put(1, ZoneMainTheater, unit("E2", 0, 30, Keywords{}))
	a.a1 = s.put(0, ZoneMainTheater, unit("A1", 0, 30, Keywords{}))
	return a
}

func (a *arena) hq(p PlayerID) *Asset { return a.s.g.Players[p].HQ }

func (a *arena) attack() { a.s.do(a.s.attack(a.hero, a.dummy)) }

// attackTrig is one "Whenever this unit attacks, [you may] <steps>" ability.
func attackTrig(optional bool, targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookAttacks, Optional: optional,
		Effect: ability.Effect{Targets: targets, Steps: steps}}
}

func upTo(n int, f ability.Filter) ability.Target { return ability.Target{Filter: f, UpTo: n} }

var (
	opposingUnit = ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
	otherAsset   = ability.Filter{Other: true}
	yes          = Answer{Picks: []int{0}}
	no           = Answer{Picks: []int{}}
)

// tpick is a scripted answer to one DecideTarget: the assets to pick, by asset,
// or a timeout. The same asset listed twice is a repeated pick.
type tpick struct {
	assets  []*Asset
	timeout bool
}

func pick(assets ...*Asset) tpick { return tpick{assets: assets} }

var timeoutPick = tpick{timeout: true}

// scriptPlan installs a chooser for seat p that answers every DecideOrder with
// order, the DecideOptionals in turn with opts and the DecideTargets in turn
// with tgts (by asset). It records every Decision and reports a decision asked
// past the end of its plan.
func (s *scene) scriptPlan(p PlayerID, order []int, opts []Answer, tgts ...tpick) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	nOpt, nTgt := 0, 0
	s.g.choosers[p] = func(_ *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		switch d.Kind {
		case DecideOrder:
			return Answer{Picks: order}
		case DecideOptional:
			if nOpt >= len(opts) {
				s.t.Errorf("player %d is asked a further 'you may' question past the end of its plan: %q", p, d.Prompt)
				return Answer{Timeout: true}
			}
			nOpt++
			return opts[nOpt-1]
		case DecideTarget:
			if nTgt >= len(tgts) {
				s.t.Errorf("player %d is asked a further target question past the end of its plan: %q", p, d.Prompt)
				return Answer{Timeout: true}
			}
			tp := tgts[nTgt]
			nTgt++
			if tp.timeout {
				return Answer{Timeout: true}
			}
			picks := []int{}
			for _, w := range tp.assets {
				found := false
				for i, o := range d.Options {
					if o.Ref == w.ID {
						picks = append(picks, i)
						found = true
						break
					}
				}
				if !found {
					s.t.Errorf("wanted asset %s is not among the options of %q", w, d.Prompt)
					return Answer{Timeout: true}
				}
			}
			return Answer{Picks: picks}
		}
		s.t.Errorf("unexpected %v decision %q", d.Kind, d.Prompt)
		return Answer{Timeout: true}
	}
}

func kindsOf(ds []Decision) []DecisionKind {
	out := make([]DecisionKind, len(ds))
	for i, d := range ds {
		out[i] = d.Kind
	}
	return out
}

// placesFrom counts the place events whose Source is a. The attack itself is one.
func placesFrom(g *Game, a *Asset) int {
	n := 0
	for _, e := range eventsOfKind(g.Log, "place") {
		if e.Source == a.ID {
			n++
		}
	}
	return n
}

// dealt is the damage committed to a by non-combat damage events.
func dealt(g *Game, a *Asset) (events, total int) {
	for _, e := range damageTo(g.Log, a) {
		events++
		total += e.Amount
	}
	return
}

func wantDealt(t *testing.T, g *Game, a *Asset, wantEvents, wantTotal int, why string) {
	t.Helper()
	if n, sum := dealt(g, a); n != wantEvents || sum != wantTotal {
		t.Errorf("%s: %d non-combat damage events totalling %d, want %d events totalling %d (%s)", a, n, sum, wantEvents, wantTotal, why)
	}
}

func wantNoNonCombatDamage(t *testing.T, g *Game, why string) {
	t.Helper()
	if ev := nonCombatDamage(g.Log); len(ev) != 0 {
		t.Errorf("no damage is dealt, got %d non-combat damage events (%s)", len(ev), why)
	}
}

func wantNoInvalid(t *testing.T, g *Game) {
	t.Helper()
	if n := len(invalidLogs(g)); n != 0 {
		t.Errorf("no invalid-answer event is logged, got %d", n)
	}
}

func targetRecords(g *Game) []DecisionRecord {
	var out []DecisionRecord
	for _, e := range eventsOfKind(g.Log, "decision") {
		if e.Decision != nil && e.Decision.Kind == "Target" {
			out = append(out, *e.Decision)
		}
	}
	return out
}

// --- SVY-R9: optional triggers ---

func TestFR_MOD_001_OptionalTriggerYesAsksOptionalThenTargetsThenResolves(t *testing.T) {
	// Covers: SVY-R9
	a := newArena(t, 1, attackTrig(true, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0)))
	a.s.scriptPlan(0, nil, []Answer{yes}, pick(a.e2))
	a.s.scriptPlan(1, nil, nil)
	a.attack()
	asked := a.s.asked(0)
	if want := []DecisionKind{DecideOptional, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("the controller is asked 'you may' first and its targets second: got %v, want %v (SVY-R9)", kindsOf(asked), want)
	}
	if d := asked[0]; d.Player != 0 || d.Min != 0 || d.Max != 1 || len(d.Options) != 1 || d.Prompt == "" {
		t.Errorf("the 'you may' decision goes to the controller with one option, Min 0, Max 1 and a prompt: %+v", d)
	}
	if n := len(a.s.asked(1)); n != 0 {
		t.Errorf("the opponent is asked nothing: %d decisions", n)
	}
	wantDealt(t, a.s.g, a.e2, 1, 3, "yes: the trigger is placed with its chosen target and resolves")
	wantDealt(t, a.s.g, a.e1, 0, 0, "only the chosen target takes damage")
	if got := placesFrom(a.s.g, a.hero); got != 2 {
		t.Errorf("yes places the trigger: the attack and the trigger are 2 place events from Hero, got %d", got)
	}
	noFormatErrors(t, a.s.g)
}

func TestFR_MOD_001_OptionalTriggerNoIsNotPlacedAndAsksNoTargets(t *testing.T) {
	// Covers: SVY-R9
	// Three opposing units are legal targets, so a mandatory trigger would be asked.
	a := newArena(t, 1, attackTrig(true, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0)))
	a.s.scriptPlan(0, nil, []Answer{no})
	a.s.scriptPlan(1, nil, nil)
	a.attack()
	if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
		t.Errorf("a declined trigger chooses no targets: asked %v, want %v (SVY-R9)", kindsOf(a.s.asked(0)), want)
	}
	wantNoNonCombatDamage(t, a.s.g, "a declined trigger does nothing")
	if got := placesFrom(a.s.g, a.hero); got != 1 {
		t.Errorf("a declined trigger is not placed: only the attack is a place event from Hero, got %d (SVY-R9)", got)
	}
	if len(a.s.g.Stack) != 0 {
		t.Errorf("the stack is empty: %d items", len(a.s.g.Stack))
	}
	if n := len(targetRecords(a.s.g)); n != 0 {
		t.Errorf("a declined trigger leaves no Target decision in the log: %d", n)
	}
	wantNoInvalid(t, a.s.g)
}

func TestFR_MOD_001_OptionalTriggerTimeoutNoChooserAndInvalidAnswerDecline(t *testing.T) {
	// Covers: SVY-R9
	for _, tc := range []struct {
		name        string
		install     func(a *arena)
		wantAsked   bool
		wantInvalid int
	}{
		{"timeout", func(a *arena) { a.s.scriptPlan(0, nil, []Answer{{Timeout: true}}) }, true, 0},
		{"no chooser", func(a *arena) { a.s.g.choosers[0] = nil }, false, 0},
		{"out-of-range answer", func(a *arena) { a.s.scriptPlan(0, nil, []Answer{{Picks: []int{5}}}) }, true, 1},
	} {
		a := newArena(t, 1, attackTrig(true, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0)))
		tc.install(a)
		a.attack()
		if tc.wantAsked {
			if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
				t.Errorf("%s: asked %v, want only the 'you may' question", tc.name, kindsOf(a.s.asked(0)))
			}
		}
		wantNoNonCombatDamage(t, a.s.g, tc.name+": declines (FR-MOD-001 §6, SVY-R9)")
		if got := placesFrom(a.s.g, a.hero); got != 1 {
			t.Errorf("%s: a declined trigger is not placed, got %d place events from Hero, want 1 (the attack)", tc.name, got)
		}
		if n := len(invalidLogs(a.s.g)); n != tc.wantInvalid {
			t.Errorf("%s: invalid-answer events = %d, want %d", tc.name, n, tc.wantInvalid)
		}
		if n := len(targetRecords(a.s.g)); n != 0 {
			t.Errorf("%s: no Target decision is recorded: %d", tc.name, n)
		}
	}
}

func TestFR_MOD_001_OptionalTriggerWithNoTargetsIsStillAskedBothWays(t *testing.T) {
	// Covers: SVY-R9
	// "Whenever this unit attacks, you may draw a card."
	for _, tc := range []struct {
		name      string
		ans       Answer
		wantDrawn bool
	}{{"yes", yes, true}, {"no", no, false}, {"timeout", Answer{Timeout: true}, false}} {
		a := newArena(t, 1, attackTrig(true, nil, drawN(1)))
		top := a.s.onTop(0, vanilla("Top")...)
		a.s.scriptPlan(0, nil, []Answer{tc.ans})
		a.attack()
		if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
			t.Errorf("%s: the controller is asked 'you may' even though the trigger has no targets: %v", tc.name, kindsOf(a.s.asked(0)))
		}
		if got := inHand(a.s.g.Players[0], top[0]); got != tc.wantDrawn {
			t.Errorf("%s: the top card is drawn = %v, want %v", tc.name, got, tc.wantDrawn)
		}
		want := 1
		if tc.wantDrawn {
			want = 2
		}
		if got := placesFrom(a.s.g, a.hero); got != want {
			t.Errorf("%s: place events from Hero = %d, want %d (a declined trigger is not placed)", tc.name, got, want)
		}
	}
}

func TestFR_MOD_001_OptionalQuestionComesAfterTheOrderDecisionInPlacementOrder(t *testing.T) {
	// Covers: SVY-R9, FR-TRG-003
	// Two optional triggers: T1 draws 1, T2 draws 2. Options are listed in printed
	// order, so picks [1 0] place T2 first. The first 'you may' question is T2's
	// turn: answering yes then no accepts T2 (draws 2) and declines T1.
	a := newArena(t, 1, attackTrig(true, nil, drawN(1)), attackTrig(true, nil, drawN(2)))
	top := a.s.onTop(0, vanilla("C1", "C2", "C3")...)
	a.s.scriptPlan(0, []int{1, 0}, []Answer{yes, no})
	a.attack()
	want := []DecisionKind{DecideOrder, DecideOptional, DecideOptional}
	if !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
		t.Fatalf("the Order decision comes first, then one 'you may' per optional trigger: got %v, want %v", kindsOf(a.s.asked(0)), want)
	}
	if !inHand(a.s.g.Players[0], top[0]) || !inHand(a.s.g.Players[0], top[1]) || inHand(a.s.g.Players[0], top[2]) {
		t.Errorf("the first question belongs to the first trigger placed (T2, draw 2): exactly the top 2 cards are drawn")
	}
	if got := placesFrom(a.s.g, a.hero); got != 2 {
		t.Errorf("the attack and the accepted trigger are placed: %d place events from Hero, want 2", got)
	}
}

func TestFR_MOD_001_OptionalAndMandatoryTriggersOfOnePlayerShareAnOrderDecision(t *testing.T) {
	// Covers: SVY-R9, FR-TRG-003
	// T1 is mandatory (draw 1), T2 optional (draw 2). The Order decision lists both
	// before any 'you may' question; declining T2 still resolves T1.
	a := newArena(t, 1, attackTrig(false, nil, drawN(1)), attackTrig(true, nil, drawN(2)))
	top := a.s.onTop(0, vanilla("C1", "C2", "C3")...)
	a.s.scriptPlan(0, []int{0, 1}, []Answer{no})
	a.attack()
	asked := a.s.asked(0)
	if want := []DecisionKind{DecideOrder, DecideOptional}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("Order first (both triggers), then the optional question: got %v, want %v", kindsOf(asked), want)
	}
	if len(asked[0].Options) != 2 {
		t.Errorf("both triggers, the optional one included, are options of the Order decision: %d", len(asked[0].Options))
	}
	if !inHand(a.s.g.Players[0], top[0]) || inHand(a.s.g.Players[0], top[1]) {
		t.Errorf("T1 resolves (draws 1) and the declined T2 does not (no second card)")
	}
}

// --- SVY-R11: "up to N" ---

// upToArena is Hero with the mandatory trigger "deal 3 damage to each of up to 2
// chosen opposing units" (or the given target slot). With the default slot the
// legal options are Dummy, E1 and E2.
func upToArena(t *testing.T, ts ...ability.Target) *arena {
	t.Helper()
	if len(ts) == 0 {
		ts = []ability.Target{upTo(2, opposingUnit)}
	}
	return newArena(t, 1, attackTrig(false, ts, dmgTo(3, 0)))
}

func TestFR_TGT_002_UpToTwoAcceptsZeroOneOrTwoDistinctTargets(t *testing.T) {
	// Covers: SVY-R11
	for _, tc := range []struct {
		name string
		pick func(a *arena) tpick
		hit  func(a *arena) []*Asset
	}{
		{"none", func(a *arena) tpick { return pick() }, func(a *arena) []*Asset { return nil }},
		{"one", func(a *arena) tpick { return pick(a.e2) }, func(a *arena) []*Asset { return []*Asset{a.e2} }},
		{"two", func(a *arena) tpick { return pick(a.e2, a.dummy) }, func(a *arena) []*Asset { return []*Asset{a.e2, a.dummy} }},
	} {
		a := upToArena(t)
		a.s.scriptPlan(0, nil, nil, tc.pick(a))
		a.s.scriptPlan(1, nil, nil)
		a.attack()
		asked := a.s.asked(0)
		if len(asked) != 1 || asked[0].Kind != DecideTarget {
			t.Fatalf("%s: one target question for the one slot: %v", tc.name, kindsOf(asked))
		}
		if d := asked[0]; d.Min != 0 || d.Max != 2 {
			t.Errorf("%s: 'up to 2' of 3 legal options is Min 0, Max 2: got Min %d, Max %d (SVY-R11)", tc.name, d.Min, d.Max)
		}
		hit := tc.hit(a)
		for _, x := range []*Asset{a.dummy, a.e1, a.e2} {
			n, sum := 0, 0
			for _, h := range hit {
				if h == x {
					n, sum = 1, 3
				}
			}
			wantDealt(t, a.s.g, x, n, sum, tc.name+": each chosen entity gets its own damage instance, and no other entity is hit")
		}
		for _, e := range nonCombatDamage(a.s.g.Log) {
			if e.Source != a.hero.ID {
				t.Errorf("%s: the damage source is Hero %d, got %d", tc.name, a.hero.ID, e.Source)
			}
		}
		if got := placesFrom(a.s.g, a.hero); got != 2 {
			t.Errorf("%s: choosing %d targets still places the trigger: %d place events from Hero, want 2", tc.name, len(hit), got)
		}
		wantNoInvalid(t, a.s.g)
		noFormatErrors(t, a.s.g)
	}
}

func TestFR_TGT_002_UpToWithOneLegalOptionIsStillAsked(t *testing.T) {
	// Covers: SVY-R11
	// Max is min(N, options): with one legal option "up to 2" is Min 0, Max 1,
	// and Min != Max means no fast path (T-0006).
	oppStructure := ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure}
	a := upToArena(t, upTo(2, oppStructure))
	st := a.s.put(1, ZoneMainTheater, structure("Enemy Keep", 0, 9, Keywords{}))
	a.s.scriptPlan(0, nil, nil, pick(st))
	a.attack()
	asked := a.s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("with exactly one legal option the player is still asked: %d questions", len(asked))
	}
	if d := asked[0]; d.Min != 0 || d.Max != 1 || !reflect.DeepEqual(refs(d), ids(st)) {
		t.Errorf("options = %v, Min %d, Max %d, want only the Enemy Keep, Min 0, Max 1", refs(d), d.Min, d.Max)
	}
	wantDealt(t, a.s.g, st, 1, 3, "the picked sole option takes damage")

	// Declining the one option is valid and deals nothing.
	b := upToArena(t, upTo(2, oppStructure))
	b.s.put(1, ZoneMainTheater, structure("Enemy Keep", 0, 9, Keywords{}))
	b.s.scriptPlan(0, nil, nil, pick())
	b.attack()
	wantNoNonCombatDamage(t, b.s.g, "picking none of one option")
	wantNoInvalid(t, b.s.g)
}

func TestFR_TGT_002_UpToOptionsAreOldestFirst(t *testing.T) {
	// Covers: SVY-R11
	a := upToArena(t)
	a.s.scriptPlan(0, nil, nil, pick())
	a.attack()
	d := a.s.asked(0)[0]
	got := make([]int, len(d.Options))
	for i, o := range d.Options {
		got[i] = o.Ref
	}
	if want := []int{a.dummy.ID, a.e1.ID, a.e2.ID}; !reflect.DeepEqual(got, want) {
		t.Errorf("'up to' offers the same legal options as a single choice, oldest first: got %v, want %v", got, want)
	}
}

func TestFR_TGT_002_UpToRepeatedPickIsInvalidAndFallsBackToNone(t *testing.T) {
	// Covers: SVY-R11
	a := upToArena(t)
	a.s.scriptPlan(0, nil, nil, pick(a.e2, a.e2))
	a.attack()
	if n := len(invalidLogs(a.s.g)); n != 1 {
		t.Errorf("picking one entity twice is an invalid answer and is logged: %d invalid-answer events, want 1 (SVY-R11)", n)
	}
	wantNoNonCombatDamage(t, a.s.g, "the repeated pick falls back to none (not 'E2 twice', not 'E2 once')")
	if recs := targetRecords(a.s.g); len(recs) != 1 || recs[0].How != "invalid" {
		t.Errorf("the Target decision is recorded as an invalid answer: %+v", recs)
	}
}

func TestFR_TGT_002_UpToNoLegalOptionAsksNothingAndDoesNothing(t *testing.T) {
	// Covers: SVY-R11
	// No opposing structure exists.
	a := upToArena(t, upTo(2, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure}))
	a.s.scriptPlan(0, nil, nil)
	a.s.scriptPlan(1, nil, nil)
	a.attack()
	if n := len(a.s.asked(0)) + len(a.s.asked(1)); n != 0 {
		t.Errorf("with no legal option nothing is asked: %d decisions", n)
	}
	if n := len(targetRecords(a.s.g)); n != 0 {
		t.Errorf("with no legal option no decision is recorded: %d", n)
	}
	wantNoNonCombatDamage(t, a.s.g, "the slot stays empty")
	if len(a.s.g.Stack) != 0 {
		t.Errorf("the stack is empty: %d items", len(a.s.g.Stack))
	}
	noFormatErrors(t, a.s.g)
}

func TestFR_MOD_001_UpToTimeoutAndNoChooserPickNone(t *testing.T) {
	// Covers: SVY-R11
	// A mandatory single target's timeout picks the first legal target in fixed order; an
	// "up to" timeout picks Min = 0 of them.
	for _, tc := range []struct {
		name    string
		install func(a *arena)
		how     string
	}{
		{"timeout", func(a *arena) { a.s.scriptPlan(0, nil, nil, timeoutPick) }, "timeout"},
		{"no chooser", func(a *arena) { a.s.g.choosers[0] = nil }, ""},
	} {
		a := upToArena(t)
		tc.install(a)
		a.attack()
		wantNoNonCombatDamage(t, a.s.g, tc.name+": an 'up to' timeout picks none (FR-MOD-001 §6, SVY-R11)")
		wantNoInvalid(t, a.s.g)
		if got := placesFrom(a.s.g, a.hero); got != 2 {
			t.Errorf("%s: the trigger is still placed with no targets: %d place events from Hero, want 2", tc.name, got)
		}
		if recs := targetRecords(a.s.g); tc.how != "" && (len(recs) != 1 || recs[0].How != tc.how) {
			t.Errorf("%s: the Target decision is recorded as %q: %+v", tc.name, tc.how, recs)
		}
	}
}

func TestFR_KWD_002h_UpToOptionsExcludeOpposingWardAssets(t *testing.T) {
	// Covers: SVY-R11
	a := upToArena(t)
	warded := a.s.put(1, ZoneMainTheater, unit("Warded", 0, 30, Keywords{Ward: true}))
	a.s.scriptPlan(0, nil, nil, pick(a.e1, a.e2))
	a.attack()
	d := a.s.asked(0)[0]
	if want := ids(a.dummy, a.e1, a.e2); !reflect.DeepEqual(refs(d), want) {
		t.Errorf("an opposing Ward unit is not an option of an 'up to' clause: options %v, want %v (FR-KWD-002h)", refs(d), want)
	}
	if d.Max != 2 {
		t.Errorf("Max is min(2, 3 legal options) = 2, got %d", d.Max)
	}
	wantDealt(t, a.s.g, warded, 0, 0, "a Ward unit is not chosen")
}

// --- FR-TGT-003: partial invalidation of a multi-entity slot ---

func TestFR_TGT_003_UpToResolvesForTheSurvivingChosenEntity(t *testing.T) {
	// Hero has two mandatory attack triggers. T1: 2 damage to each of up to 2
	// chosen opposing units. T2: 9 damage to a chosen opposing unit. Picks [0 1]
	// place T1 first (resolves last) and T2 last (resolves first). T1 chooses E1
	// (health 5) and E2, T2 chooses E1: E1 is destroyed before T1 resolves.
	a := newArena(t, 1,
		attackTrig(false, []ability.Target{upTo(2, opposingUnit)}, dmgTo(2, 0)),
		attackTrig(false, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(9, 0)))
	a.s.scriptPlan(0, []int{0, 1}, nil, pick(a.e1, a.e2), pick(a.e1))
	a.attack()
	if a.s.alive(a.e1) {
		t.Fatalf("precondition: T2's 9 damage destroys E1 (health 5) before T1 resolves")
	}
	if want := []DecisionKind{DecideOrder, DecideTarget, DecideTarget}; !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
		t.Fatalf("precondition: Order, then T1's targets, then T2's: %v", kindsOf(a.s.asked(0)))
	}
	wantDealt(t, a.s.g, a.e2, 1, 2, "the surviving chosen entity is still legal at resolution and takes T1's damage (FR-TGT-003 partial invalidation)")
	for _, e := range damageTo(a.s.g.Log, a.e1) {
		if e.Raw == 2 {
			t.Errorf("the destroyed entity takes none of T1's damage: %+v", e)
		}
	}
	if len(a.s.g.Stack) != 0 {
		t.Errorf("both triggers resolved and the stack is empty: %d items", len(a.s.g.Stack))
	}
	noFormatErrors(t, a.s.g)
}

// --- FR-TGT-001: "other" ---

func TestFR_TGT_001_OtherExcludesTheSourceFromChosenOptionsAndOnlyThat(t *testing.T) {
	for _, tc := range []struct {
		name  string
		slot  func(f ability.Filter) ability.Target
		other bool
	}{
		{"single, Other", func(f ability.Filter) ability.Target { return ability.Target{Filter: f} }, true},
		{"single, no Other", func(f ability.Filter) ability.Target { return ability.Target{Filter: f} }, false},
		{"up to 2, Other", func(f ability.Filter) ability.Target { return upTo(2, f) }, true},
		{"up to 2, no Other", func(f ability.Filter) ability.Target { return upTo(2, f) }, false},
	} {
		slot := tc.slot(ability.Filter{Other: tc.other})
		a := newArena(t, 1, attackTrig(false, []ability.Target{slot}, dmgTo(1, 0)))
		a.s.scriptPlan(0, nil, nil, pick(a.a1))
		a.attack()
		asked := a.s.asked(0)
		if len(asked) != 1 {
			t.Fatalf("%s: one target question: %d", tc.name, len(asked))
		}
		want := ids(a.hero, a.dummy, a.e1, a.e2, a.a1, a.hq(0), a.hq(1))
		if tc.other {
			want = ids(a.dummy, a.e1, a.e2, a.a1, a.hq(0), a.hq(1))
		}
		if got := refs(asked[0]); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: options = %v, want %v ('other' removes the source and nothing else; without it the source is an option, FR-TGT-001)", tc.name, got, want)
		}
	}
}

func TestFR_TGT_002_AllWithOtherSkipsTheSourceAndHitsTheRest(t *testing.T) {
	for _, tc := range []struct {
		name       string
		f          ability.Filter
		wantHero   int
		wantHeroEv int
		wantA1     int
		wantHQ0    int
		wantEnemy  int
	}{
		{"any asset, Other", ability.Filter{Other: true}, 0, 0, 2, 2, 2},
		{"any asset, no Other", ability.Filter{}, 2, 1, 2, 2, 2},
		{"allied units, Other", ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true}, 0, 0, 2, 0, 0},
		{"allied units, no Other", ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}, 2, 1, 2, 0, 0},
	} {
		a := newArena(t, 1, attackTrig(false, nil, allDamage(2, tc.f)))
		a.s.scriptPlan(0, nil, nil)
		a.attack()
		wantDealt(t, a.s.g, a.hero, tc.wantHeroEv, tc.wantHero, tc.name+": the source is skipped only with Other")
		_, a1 := dealt(a.s.g, a.a1)
		_, hq0 := dealt(a.s.g, a.hq(0))
		_, e2 := dealt(a.s.g, a.e2)
		_, hq1 := dealt(a.s.g, a.hq(1))
		if a1 != tc.wantA1 || hq0 != tc.wantHQ0 || e2 != tc.wantEnemy || hq1 != tc.wantEnemy {
			t.Errorf("%s: damage A1 %d, own HQ %d, E2 %d, enemy HQ %d; want %d, %d, %d, %d", tc.name, a1, hq0, e2, hq1, tc.wantA1, tc.wantHQ0, tc.wantEnemy, tc.wantEnemy)
		}
		if n := len(a.s.asked(0)); n != 0 {
			t.Errorf("%s: 'all' asks nothing: %d decisions", tc.name, n)
		}
	}
}

// --- Composite: "you may deal damage equal to its power to each of up to 2 other chosen assets" ---

func craterfallShape(t *testing.T) *arena {
	t.Helper()
	return newArena(t, 4, attackTrig(true, []ability.Target{upTo(2, otherAsset)}, powerStep(0)))
}

func TestFR_TGT_002_CompositeYesWithTwoAssetsDealsPowerToEach(t *testing.T) {
	// Covers: SVY-R9, SVY-R11
	a := craterfallShape(t)
	a.s.scriptPlan(0, nil, []Answer{yes}, pick(a.hq(0), a.e2))
	a.attack()
	asked := a.s.asked(0)
	if want := []DecisionKind{DecideOptional, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("'you may' first, then the target question: %v", kindsOf(asked))
	}
	if d := asked[1]; d.Min != 0 || d.Max != 2 {
		t.Errorf("up to 2 of the 6 other assets: Min 0, Max 2, got Min %d, Max %d", d.Min, d.Max)
	}
	wantOpts := ids(a.hq(0), a.hq(1), a.dummy, a.e1, a.e2, a.a1)
	if got := refs(asked[1]); !reflect.DeepEqual(got, wantOpts) {
		t.Errorf("options = %v, want %v: both Headquarters are legal (any asset), Hero itself is not", got, wantOpts)
	}
	wantDealt(t, a.s.g, a.hq(0), 1, 4, "Hero's own Headquarters took damage equal to Hero's power")
	wantDealt(t, a.s.g, a.e2, 1, 4, "the second chosen asset took damage equal to Hero's power")
	wantDealt(t, a.s.g, a.hero, 0, 0, "Hero is never offered")
	wantDealt(t, a.s.g, a.e1, 0, 0, "unchosen assets take none")
	noFormatErrors(t, a.s.g)
}

func TestFR_TGT_002_CompositeYesWithOneAssetDealsPowerToIt(t *testing.T) {
	// Covers: SVY-R9, SVY-R11
	a := craterfallShape(t)
	a.s.scriptPlan(0, nil, []Answer{yes}, pick(a.e2))
	a.attack()
	wantDealt(t, a.s.g, a.e2, 1, 4, "one chosen asset takes damage equal to power")
	if got := len(nonCombatDamage(a.s.g.Log)); got != 1 {
		t.Errorf("exactly one damage instance: %d (up to 2 is not exactly 2)", got)
	}
}

func TestFR_TGT_002_CompositeNoDealsNothing(t *testing.T) {
	// Covers: SVY-R9
	a := craterfallShape(t)
	a.s.scriptPlan(0, nil, []Answer{no})
	a.attack()
	wantNoNonCombatDamage(t, a.s.g, "declined")
	if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(a.s.asked(0)), want) {
		t.Errorf("a declined trigger asks no targets: %v", kindsOf(a.s.asked(0)))
	}
	if got := placesFrom(a.s.g, a.hero); got != 1 {
		t.Errorf("a declined trigger is not placed: %d place events from Hero, want 1", got)
	}
}
