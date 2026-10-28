package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0083: modal triggered abilities ("choose N" modes, FR-MOD-001 §5).
// The controller chooses the modes as the trigger is placed, before any target
// (FR-TGT-001, FR-MOD-001 §7, SVY-R3, SVY-R9), then only the chosen modes'
// targets. The item resolves only the chosen modes, whose targets are verified
// again at resolution (FR-TGT-003). Cards are synthetic.
//
// The board of every scene: opposing units E1 (health 5) and E2, allied units A1
// and A2. A modal unit deploys from hand into Reserves. Its modes, unless a test
// says otherwise, are:
//   mode 1: deal 3 damage to chosen opposing unit (E1 or E2)
//   mode 2: chosen allied non-Scratch unit (A1, A2 or the source) gains 2 armor

type modalBoard struct {
	s              *scene
	e1, e2, a1, a2 *Asset
}

func newModalBoard(t *testing.T) *modalBoard {
	t.Helper()
	s := newScene(t)
	return &modalBoard{
		s:  s,
		e1: s.put(1, ZoneMainTheater, unit("E1", 0, 5, Keywords{})),
		e2: s.put(1, ZoneMainTheater, unit("E2", 0, 30, Keywords{})),
		a1: s.put(0, ZoneMainTheater, unit("A1", 0, 30, Keywords{})),
		a2: s.put(0, ZoneMainTheater, unit("A2", 0, 30, Keywords{})),
	}
}

// mode is one mode: its own chosen targets and steps. The slices have spare
// capacity so that appending onto them would show (see the Def test).
func mode(targets []ability.Target, steps ...ability.Step) ability.Effect {
	return ability.Effect{
		Targets: append(make([]ability.Target, 0, 4), targets...),
		Steps:   append(make([]ability.Step, 0, 4), steps...),
	}
}

func modalTrigger(optional bool, n int, modes ...ability.Effect) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDeploy, Optional: optional,
		Effect: ability.Effect{Modes: modes, ChooseN: n}}
}

func hitMode() ability.Effect {
	return mode([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
}

func armorMode() ability.Effect {
	return mode([]ability.Target{tgt(ability.AllegianceAllied, ability.KindUnit, true)}, armorStep(2, 0))
}

// hitArmorTrigger is "When this deploys, choose one: hit mode, armor mode".
func hitArmorTrigger() ability.Trigger { return modalTrigger(false, 1, hitMode(), armorMode()) }

func totalArmor(g *Game) int {
	n := 0
	for _, a := range g.Assets() {
		n += a.Armor
	}
	return n
}

func modeRecords(g *Game) []DecisionRecord {
	var out []DecisionRecord
	for _, e := range eventsOfKind(g.Log, "decision") {
		if e.Decision != nil && e.Decision.Kind == DecideMode.String() {
			out = append(out, *e.Decision)
		}
	}
	return out
}

func placeCount(g *Game) int { return len(eventsOfKind(g.Log, "place")) }

// scriptModal installs a chooser for seat p: DecideOrder gets order, the
// DecideOptionals in turn get opts, the DecideModes in turn get modes (picks are
// option indexes) and the DecideTargets in turn get tgts (by asset). It records
// every Decision, and returns the number of place events logged when each was
// asked.
func (s *scene) scriptModal(p PlayerID, order []int, opts, modes []Answer, tgts ...tpick) *[]int {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[p] = nil
	placesAtAsk := []int{}
	nOpt, nMode, nTgt := 0, 0, 0
	s.g.choosers[p] = func(g *Game, d Decision) Answer {
		s.asks[p] = append(s.asks[p], d)
		placesAtAsk = append(placesAtAsk, placeCount(g))
		switch d.Kind {
		case DecideOrder:
			return Answer{Picks: order}
		case DecideOptional:
			if nOpt >= len(opts) {
				s.t.Errorf("asked a further 'you may' question past the end of the plan: %q", d.Prompt)
				return Answer{Timeout: true}
			}
			nOpt++
			return opts[nOpt-1]
		case DecideMode:
			if nMode >= len(modes) {
				s.t.Errorf("asked a further mode question past the end of the plan: %q", d.Prompt)
				return Answer{Timeout: true}
			}
			nMode++
			return modes[nMode-1]
		case DecideTarget:
			if nTgt >= len(tgts) {
				s.t.Errorf("asked a further target question past the end of the plan: %q", d.Prompt)
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
	return &placesAtAsk
}

func labels(d Decision) []string {
	out := make([]string, len(d.Options))
	for i, o := range d.Options {
		out[i] = o.Label
	}
	return out
}

func modeAnswer(picks ...int) Answer { return Answer{Picks: picks} }

// --- asked on placement, before targets ---

func TestFR_MOD_001_ModalTriggerAsksForItsModeBeforeItsTargets(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R3, SVY-R9
	b := newModalBoard(t)
	places := b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(b.a2))
	c := trigUnit("Modal Unit", hitArmorTrigger())
	b.s.deployFromHand(0, c)
	src := assetOf(b.s.g, c)
	asked := b.s.asked(0)
	if want := []DecisionKind{DecideMode, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("the mode is asked first, then only the chosen mode's target, nothing for the other mode: got %v, want %v (FR-MOD-001 §5, FR-TGT-001)", kindsOf(asked), want)
	}
	m := asked[0]
	if m.Player != 0 || m.Min != 1 || m.Max != 1 || m.Repeats {
		t.Errorf("'choose one' of two modes is Min 1, Max 1, no repeats, asked of the controller: %+v (SVY-R6)", m)
	}
	if want := []string{"mode 1", "mode 2"}; !reflect.DeepEqual(labels(m), want) {
		t.Errorf("mode options = %v, want %v, in printed order", labels(m), want)
	}
	for _, o := range m.Options {
		if o.Ref != src.ID {
			t.Errorf("each mode option refers to the trigger's source asset %d, got %d", src.ID, o.Ref)
		}
	}
	if want := "Modal Unit: trigger 1: choose one"; m.Prompt != want {
		t.Errorf("mode prompt = %q, want %q", m.Prompt, want)
	}
	if want := ids(b.a1, b.a2, src); !reflect.DeepEqual(refs(asked[1]), want) {
		t.Errorf("the target question is mode 2's (allied non-Scratch units): options %v, want %v", refs(asked[1]), want)
	}
	for i, n := range *places {
		if n != 1 {
			t.Errorf("question %d is asked before the trigger's place event: %d place events had been logged, want 1 (the deploy) (FR-TGT-001, FR-MOD-001 §7)", i+1, n)
		}
	}
	if got := placeCount(b.s.g); got != 2 {
		t.Errorf("the deploy and the trigger are 2 place events, got %d", got)
	}
}

// --- only the chosen mode resolves ---

func TestFR_MOD_001_ModalTriggerResolvesOnlyTheChosenMode(t *testing.T) {
	b := newModalBoard(t)
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(1)}, pick(b.a2))
	b.s.deployFromHand(0, trigUnit("Modal Unit", hitArmorTrigger()))
	if b.a2.Armor != 2 {
		t.Errorf("mode 2 resolves: A2 armor = %d, want 2", b.a2.Armor)
	}
	if got := totalArmor(b.s.g); got != 2 {
		t.Errorf("only the chosen target gains armor: total armor %d, want 2", got)
	}
	wantNoNonCombatDamage(t, b.s.g, "mode 1 was not chosen, so its damage step does not run")
	wantNoInvalid(t, b.s.g)
	noFormatErrors(t, b.s.g)

	c := newModalBoard(t)
	c.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)}, pick(c.e2))
	c.s.deployFromHand(0, trigUnit("Modal Unit", hitArmorTrigger()))
	wantDealt(t, c.s.g, c.e2, 1, 3, "mode 1 resolves on the entity chosen for it")
	wantDealt(t, c.s.g, c.e1, 0, 0, "unchosen entities take none")
	if got := totalArmor(c.s.g); got != 0 {
		t.Errorf("mode 2 was not chosen, so no armor is gained: total armor %d", got)
	}
}

// --- timeout ---

func TestFR_MOD_001_TimedOutModeIsTheFirstListed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		install func(b *modalBoard)
	}{
		{"timeout", func(b *modalBoard) {
			b.s.g.choosers[0] = func(_ *Game, d Decision) Answer { return Answer{Timeout: true} }
		}},
		{"no chooser", func(b *modalBoard) { b.s.g.choosers[0] = nil }},
	} {
		b := newModalBoard(t)
		tc.install(b)
		b.s.deployFromHand(0, trigUnit("Modal Unit", hitArmorTrigger()))
		// Mode 1's target is chosen by its own timeout fallback: the first legal
		// target in fixed order, the oldest opposing unit.
		wantDealt(t, b.s.g, b.e1, 1, 3, tc.name+": mode 1 (the first listed) is chosen (FR-MOD-001 §6)")
		if got := totalArmor(b.s.g); got != 0 {
			t.Errorf("%s: mode 2 is not chosen: total armor %d, want 0", tc.name, got)
		}
		if recs := modeRecords(b.s.g); tc.name == "timeout" && (len(recs) != 1 || recs[0].How != "timeout") {
			t.Errorf("%s: the Mode decision is recorded as a timeout: %+v", tc.name, recs)
		}
	}
}

// --- illegal first mode ---

func TestFR_MOD_001_ModeWithoutALegalTargetIsNotOffered(t *testing.T) {
	// Mode 1 needs exactly one opposing structure, and there is none.
	noStructure := mode([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindStructure, false)}, dmgTo(3, 0))
	trig := func() ability.Trigger { return modalTrigger(false, 1, noStructure, armorMode()) }

	b := newModalBoard(t)
	b.s.scriptModal(0, nil, nil, nil, pick(b.a1))
	b.s.deployFromHand(0, trigUnit("Modal Unit", trig()))
	for _, d := range b.s.asked(0) {
		if d.Kind == DecideMode {
			t.Errorf("a decision with the single legal mode is answered automatically, not asked (FR-MOD-001 §1): %+v", d)
		}
	}
	if recs := modeRecords(b.s.g); len(recs) != 1 || !reflect.DeepEqual(recs[0].Options, []string{"mode 2"}) || recs[0].How != "auto" {
		t.Errorf("the mode decision lists only the legal mode 2 and is answered automatically: %+v", recs)
	}
	if b.a1.Armor != 2 || totalArmor(b.s.g) != 2 {
		t.Errorf("mode 2 resolves on the chosen target: A1 armor %d, total %d, want 2 and 2", b.a1.Armor, totalArmor(b.s.g))
	}
	wantNoNonCombatDamage(t, b.s.g, "mode 1 is not legal")

	// A timeout (or no chooser) takes the first listed LEGAL mode (FR-MOD-001 §6).
	for _, install := range []func(c *modalBoard){
		func(c *modalBoard) {
			c.s.g.choosers[0] = func(_ *Game, d Decision) Answer { return Answer{Timeout: true} }
		},
		func(c *modalBoard) { c.s.g.choosers[0] = nil },
	} {
		c := newModalBoard(t)
		install(c)
		c.s.deployFromHand(0, trigUnit("Modal Unit", trig()))
		if got := totalArmor(c.s.g); got != 2 {
			t.Errorf("on timeout the first listed legal mode, mode 2, resolves: total armor %d, want 2 (FR-MOD-001 §6)", got)
		}
		wantNoNonCombatDamage(t, c.s.g, "the illegal mode 1 is never chosen")
	}
}

// --- up to N ---

func TestFR_MOD_001_UpToTargetDoesNotMakeAModeIllegal(t *testing.T) {
	// Covers: SVY-R11
	upToStructure := mode([]ability.Target{upTo(1, ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindStructure})}, dmgTo(3, 0))
	b := newModalBoard(t)
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0)})
	b.s.deployFromHand(0, trigUnit("Modal Unit", modalTrigger(false, 1, upToStructure, armorMode())))
	asked := b.s.asked(0)
	if len(asked) == 0 || asked[0].Kind != DecideMode {
		t.Fatalf("a modal choice with two legal modes is asked: %v", kindsOf(asked))
	}
	if want := []string{"mode 1", "mode 2"}; !reflect.DeepEqual(labels(asked[0]), want) {
		t.Errorf("an 'up to 1' target with no legal entity does not make mode 1 illegal: options %v, want %v", labels(asked[0]), want)
	}
	if len(asked) != 1 {
		t.Errorf("mode 1's empty 'up to' slot asks nothing further: %v", kindsOf(asked))
	}
	wantNoNonCombatDamage(t, b.s.g, "the chosen mode has no target, so it does nothing")
	if got := totalArmor(b.s.g); got != 0 {
		t.Errorf("mode 2 was not chosen: total armor %d", got)
	}
}

// --- choose two ---

func TestFR_MOD_001_ChooseTwoResolvesTwoDifferentModes(t *testing.T) {
	b := newModalBoard(t)
	top := b.s.onTop(0, vanilla("Top")...)
	drawMode := mode(nil, drawN(1))
	trig := modalTrigger(false, 2, hitMode(), drawMode, armorMode())
	// Modes 3 and 1 are chosen, in that pick order. Their targets are asked in
	// printed order (mode 1's, then mode 3's).
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(2, 0)}, pick(b.e2), pick(b.a2))
	b.s.deployFromHand(0, trigUnit("Modal Unit", trig))
	asked := b.s.asked(0)
	if want := []DecisionKind{DecideMode, DecideTarget, DecideTarget}; !reflect.DeepEqual(kindsOf(asked), want) {
		t.Fatalf("one mode question, then the two chosen modes' targets: got %v, want %v", kindsOf(asked), want)
	}
	if m := asked[0]; m.Min != 2 || m.Max != 2 || m.Repeats || len(m.Options) != 3 || m.Prompt != "Modal Unit: trigger 1: choose 2" {
		t.Errorf("'choose two' of three modes is Min 2, Max 2, no repeats, 3 options, prompt 'Modal Unit: trigger 1: choose 2': %+v (SVY-R6)", m)
	}
	wantDealt(t, b.s.g, b.e2, 1, 3, "mode 1's damage acts on the entity chosen for mode 1")
	wantDealt(t, b.s.g, b.e1, 0, 0, "mode 1's other candidate takes none")
	if b.a2.Armor != 2 || totalArmor(b.s.g) != 2 {
		t.Errorf("mode 3's armor acts on the entity chosen for mode 3: A2 armor %d, total %d, want 2 and 2", b.a2.Armor, totalArmor(b.s.g))
	}
	if inHand(b.s.g.Players[0], top[0]) {
		t.Errorf("mode 2 was not chosen, so its draw does not happen")
	}
	wantNoInvalid(t, b.s.g)
	noFormatErrors(t, b.s.g)
}

// --- second verification ---

func TestFR_TGT_003_ChosenModesTargetIsVerifiedAgainAtResolution(t *testing.T) {
	// The card has two deploy triggers. T1 is modal (hit mode, armor mode); T2
	// deals 9 damage to a chosen opposing unit. Order picks [0 1] place T1 first
	// (it resolves last) and T2 second (it resolves first). T1 chooses mode 1 and
	// E1 (health 5), and T2 destroys E1 before T1 resolves.
	b := newModalBoard(t)
	t2 := targetTrigger([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(9, 0))
	b.s.scriptModal(0, []int{0, 1}, nil, []Answer{modeAnswer(0)}, pick(b.e1), pick(b.e1))
	b.s.deployFromHand(0, trigUnit("Modal Unit", hitArmorTrigger(), t2))
	if want := []DecisionKind{DecideOrder, DecideMode, DecideTarget, DecideTarget}; !reflect.DeepEqual(kindsOf(b.s.asked(0)), want) {
		t.Fatalf("precondition: Order, T1's mode, T1's target, T2's target: got %v", kindsOf(b.s.asked(0)))
	}
	if b.s.alive(b.e1) {
		t.Fatalf("precondition: T2's 9 damage destroys E1 before T1 resolves")
	}
	if n := len(nonCombatDamage(b.s.g.Log)); n != 1 {
		t.Errorf("only T2's damage happens; the chosen mode's targeted step does nothing for a target that left: %d damage events, want 1 (FR-TGT-003)", n)
	}
	wantDealt(t, b.s.g, b.e2, 0, 0, "the step does not retarget to another entity")
	if got := totalArmor(b.s.g); got != 0 {
		t.Errorf("no step of the unchosen mode 2 runs: total armor %d, want 0", got)
	}
	if len(b.s.g.Stack) != 0 {
		t.Errorf("both items resolved: %d on the stack", len(b.s.g.Stack))
	}
	noFormatErrors(t, b.s.g)
}

// --- optional and modal ---

func TestFR_MOD_001_DeclinedModalTriggerAsksNoMode(t *testing.T) {
	// Covers: SVY-R9
	base := newScene(t)
	base.deployFromHand(0, trigUnit("Modal Unit"))
	baseline := placeCount(base.g)

	b := newModalBoard(t)
	b.s.scriptModal(0, nil, []Answer{no}, nil)
	b.s.deployFromHand(0, trigUnit("Modal Unit", modalTrigger(true, 1, hitMode(), armorMode())))
	if want := []DecisionKind{DecideOptional}; !reflect.DeepEqual(kindsOf(b.s.asked(0)), want) {
		t.Errorf("a declined modal trigger asks only 'you may', no mode and no targets: got %v, want %v (SVY-R9)", kindsOf(b.s.asked(0)), want)
	}
	if n := len(modeRecords(b.s.g)); n != 0 {
		t.Errorf("no Mode decision is recorded: %d", n)
	}
	if got := placeCount(b.s.g); got != baseline {
		t.Errorf("a declined trigger is not placed: %d place events, want %d (the deploy alone)", got, baseline)
	}
	wantNoNonCombatDamage(t, b.s.g, "declined")
	if got := totalArmor(b.s.g); got != 0 {
		t.Errorf("declined: no armor %d", got)
	}
}

// --- the card definition is never changed ---

func TestFR_MOD_001_ModalTriggerDoesNotChangeTheCardDefinition(t *testing.T) {
	b := newModalBoard(t)
	c := trigUnit("Modal Unit", hitArmorTrigger())
	b.s.scriptModal(0, nil, nil, []Answer{modeAnswer(0), modeAnswer(1)}, pick(b.e2), pick(b.a1))
	b.s.deployFromHand(0, c)
	b.s.deployFromHand(0, c)
	wantDealt(t, b.s.g, b.e2, 1, 3, "first resolution took mode 1")
	if b.a1.Armor != 2 {
		t.Errorf("second resolution took mode 2: A1 armor %d, want 2", b.a1.Armor)
	}
	got := c.Ability.Triggers[0].Effect
	want := hitArmorTrigger().Effect
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the card's Def is unchanged by placing and resolving a modal trigger:\n got %+v\nwant %+v", got, want)
	}
	for i, m := range got.Modes {
		for _, tg := range m.Targets[len(m.Targets):cap(m.Targets)] {
			if !reflect.DeepEqual(tg, ability.Target{}) {
				t.Errorf("mode %d: a target was appended onto the card's own Targets slice: %+v", i+1, tg)
			}
		}
		for _, st := range m.Steps[len(m.Steps):cap(m.Steps)] {
			if !reflect.DeepEqual(st, ability.Step{}) {
				t.Errorf("mode %d: a step was appended onto the card's own Steps slice: %+v", i+1, st)
			}
		}
	}
}
