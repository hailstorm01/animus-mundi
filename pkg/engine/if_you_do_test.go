package engine

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0072: "you may X. If you do, Y" (SVY-R14), the untargeted pick
// (SVY-R3, SVY-R16) and the kill step (FR-LAY-002). Cards are synthetic. Every
// scene deploys a unit "Source" from player 0's hand into Reserves; its deploy
// trigger runs the steps under test. Player 1 is the opponent.

var (
	iydAlly   = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Other: true}
	iydEnemy  = ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}
	iydOneOpp = []ability.Target{{Filter: iydEnemy}}
)

// iydPick is "[kill] a unit matching f", picked as the step resolves.
func iydPick(f ability.Filter, upTo int) ability.Step {
	return ability.Step{Op: ability.OpKill, Subject: ability.SubjectPick, Filter: f, UpTo: upTo}
}

// iydIfDone gates a step on the step before it.
func iydIfDone(s ability.Step) ability.Step { s.IfDone = true; return s }

// iydTrigger is one deploy ability, optional ("you may") or not.
func iydTrigger(optional bool, targets []ability.Target, steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookDeploy, Optional: optional,
		Effect: ability.Effect{Targets: targets, Steps: steps}}
}

// iydDeploy deploys the Source from player 0's hand and returns its card.
func (s *scene) iydDeploy(trig ability.Trigger) *Card {
	s.t.Helper()
	c := trigUnit("Source", trig)
	s.deployFromHand(0, c)
	return c
}

func (s *scene) iydWantDead(why string, as ...*Asset) {
	s.t.Helper()
	for _, a := range as {
		if s.alive(a) {
			s.t.Errorf("%s should be destroyed: %s", a, why)
		}
	}
}

func (s *scene) iydWantAlive(why string, as ...*Asset) {
	s.t.Helper()
	for _, a := range as {
		if !s.alive(a) {
			s.t.Errorf("%s should still be on the battlefield: %s", a, why)
		}
	}
}

// --- the untargeted pick ---

func TestFR_TGT_001_UntargetedPickIsMadeOnResolutionFromWhatMatchesThen(t *testing.T) {
	// Covers: SVY-R3, SVY-R14
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("A", 0, 5, Keywords{}))
	b := s.put(0, ZoneMainTheater, unit("B", 0, 5, Keywords{}))
	c := s.put(0, ZoneMainTheater, unit("C", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	// Target 0 is Y's (one opposing unit, so no question); target 1 is the
	// ally that a first kill step removes before the pick resolves.
	targets := []ability.Target{{Filter: iydEnemy}, {Filter: iydAlly}}
	kill1 := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget, Target: 1}
	s.scriptPlan(0, nil, []Answer{yes}, pick(a), pick(b))
	s.iydDeploy(iydTrigger(true, targets, kill1, iydPick(iydAlly, 0), iydIfDone(dmgTo(2, 0))))
	asks := s.asked(0)
	want := []DecisionKind{DecideOptional, DecideTarget, DecideTarget}
	if got := kindsOf(asks); !slices.Equal(got, want) {
		t.Fatalf("'you may' and the chosen target are asked on placement, the pick later as X resolves: got %v, want %v (SVY-R14)", got, want)
	}
	if got := refs(asks[1]); !slices.Equal(got, ids(a, b, c)) {
		t.Errorf("on placement all three allied units are legal for the chosen target, got %v", got)
	}
	pk := asks[2]
	if got := refs(pk); !slices.Equal(got, ids(b, c)) {
		t.Errorf("the pick offers what matches when X resolves: %v, want %v; A left earlier in the effect (SVY-R3)", got, ids(b, c))
	}
	if pk.Player != 0 || pk.Min != 1 || pk.Max != 1 {
		t.Errorf("the controller picks exactly one: %+v", pk)
	}
	s.iydWantDead("A was chosen and B was picked", a, b)
	s.iydWantAlive("C was not picked", c)
	s.wantDamage(v, 2)
}

func TestFR_TGT_001_UntargetedPickIgnoresWard(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, unit("Warded", 0, 5, Keywords{Ward: true}))
	p := s.put(1, ZoneMainTheater, unit("Plain", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, nil, pick(w))
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydEnemy, 0)))
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("one pick question, got %d decisions", len(asks))
	}
	if got := refs(asks[0]); !slices.Equal(got, ids(w, p)) {
		t.Errorf("Ward does not apply to an untargeted selection: options %v, want %v", got, ids(w, p))
	}
	s.iydWantDead("the Ward unit was picked and killed", w)
	s.iydWantAlive("only the picked unit dies", p)
}

func TestFR_TGT_001_UntargetedPickMayPickTheSource(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	other := s.put(0, ZoneMainTheater, unit("Other Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	allies := ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit} // Other unset
	// The chooser says yes to "you may" and picks the newest option, which is the source.
	s.asks = map[PlayerID][]Decision{}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind == DecideOptional {
			return yes
		}
		return Answer{Picks: []int{len(d.Options) - 1}} // options are oldest first
	}
	s.iydDeploy(iydTrigger(true, iydOneOpp, iydPick(allies, 0), iydIfDone(dmgTo(2, 0))))
	asks := s.asked(0)
	if len(asks) != 2 || len(asks[1].Options) != 2 {
		t.Fatalf("'you may' then a pick between the other ally and the source: got %v", asks)
	}
	srcID := asks[1].Options[1].Ref
	if srcID == other.ID {
		t.Fatalf("the source is an option when 'other' is unset (SVY-R14): %+v", asks[1].Options)
	}
	if s.g.Asset(srcID) != nil {
		t.Errorf("the source picked itself and is destroyed")
	}
	s.iydWantAlive("the other ally was not picked", other)
	dmg := damageTo(s.g.Log, v)
	if len(dmg) != 1 || dmg[0].Amount != 2 || dmg[0].Source != srcID {
		t.Errorf("Y still resolves after its source was killed, and its damage event names the source (asset %d): %+v (SVY-R14)", srcID, dmg)
	}
}

func TestFR_MOD_001_SinglePickWithOneMatchIsNotAsked(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.script(0)
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydEnemy, 0)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single pick with one match is made without a question (FR-MOD-001 sec. 1): %d decisions", n)
	}
	s.iydWantDead("the only match was picked", v)
}

func TestFR_MOD_001_SinglePickTimeoutPicksFirstInFixedOrder(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	older := s.put(0, ZoneMainTheater, unit("Older Theater", 0, 5, Keywords{}))
	newer := s.put(0, ZoneReserves, unit("Newer Reserves", 0, 5, Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydAlly, 0)))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("two matches are asked once: %d", n)
	}
	s.iydWantDead("on a timeout the first in fixed order is picked: Reserves before the Main Theater (SVY-R16)", newer)
	s.iydWantAlive("the older Main Theater unit comes later in fixed order", older)
}

func TestFR_MOD_001_UpToPickTimeoutPicksNone(t *testing.T) {
	// Covers: SVY-R16
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydEnemy, 1)))
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("an 'up to' pick is asked even with one option: %d decisions", len(asks))
	}
	if asks[0].Min != 0 || asks[0].Max != 1 {
		t.Errorf("'up to 1' may pick none: Min %d Max %d, want 0 and 1", asks[0].Min, asks[0].Max)
	}
	s.iydWantAlive("an 'up to' selection selects none on a timeout (FR-MOD-001 sec. 6)", v)
}

func TestFR_TGT_001_UntargetedPickWithNoMatchDoesNothing(t *testing.T) {
	// Covers: SVY-R3
	s := newScene(t) // player 1 has only a Headquarters
	s.script(0)
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydEnemy, 0), scratchN(1)))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked when nothing matches: %d decisions", n)
	}
	if n := faceDownCount(s.g.Players[0]); n != 1 {
		t.Errorf("only the pick clause does nothing; the ability still resolves, so Scratch happens: %d Scratch Companies, want 1", n)
	}
}

// --- the kill step ---

func TestFR_LAY_002_KillDestroysAtOnceDuringResolution(t *testing.T) {
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	o := s.put(1, ZoneMainTheater, unit("Other", 0, 5, Keywords{}))
	kill := ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget}
	all := ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: 2}, Subject: ability.SubjectAll, Filter: iydEnemy}
	s.scriptPlan(0, nil, nil, pick(v))
	s.iydDeploy(iydTrigger(false, iydOneOpp, kill, all))
	s.iydWantDead("a kill destroys at once", v)
	if !inDiscardPile(s.g.Players[1], v.Inst) {
		t.Errorf("the killed unit is in its owner's discard pile (FR-ZON-002)")
	}
	s.wantDamage(o, 2)
	if n, _ := dealt(s.g, v); n != 0 {
		t.Errorf("the 'all' step right after the kill does not reach the dead unit (FR-LAY-002): %d damage events", n)
	}
	n := 0
	for _, e := range eventsOfKind(s.g.Log, "destroy") {
		if e.Target == v.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("one destroy event names the killed unit, got %d", n)
	}
}

func TestFR_LAY_002_KilledScratchCompanyGoesFaceUpToItsOwnersDiscard(t *testing.T) {
	// Covers: FR-ZON-002
	s := newScene(t)
	sc := s.putScratch(1)
	s.iydDeploy(iydTrigger(false, nil, iydPick(iydEnemy, 0)))
	s.iydWantDead("the face-down Scratch Company is a unit and was the only match", sc)
	if !inDiscardPile(s.g.Players[1], sc.Inst) {
		t.Errorf("a killed Scratch Company is revealed into its owner's discard pile (FR-ZON-002)")
	}
}

func TestFR_LAY_002_KillSkipsAHeadquarters(t *testing.T) {
	// Pins the documented M2 limit, not a rule.
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	killAll := ability.Step{Op: ability.OpKill, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied}}
	c := s.iydDeploy(iydTrigger(false, nil, killAll))
	s.iydWantDead("kill reaches every allied asset except a Headquarters", u)
	if assetOf(s.g, c) != nil {
		t.Errorf("the source is an allied asset and is killed too")
	}
	if hq := s.g.Players[0].HQ; hq == nil || !s.alive(hq) {
		t.Errorf("a kill step leaves the Headquarters on the battlefield")
	}
}

// --- if you do ---

func TestFR_TRG_005_IfYouDoRunsYWhenXWasDone(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes})
	s.iydDeploy(iydTrigger(true, iydOneOpp, iydPick(iydAlly, 0), iydIfDone(dmgTo(2, 0))))
	s.iydWantDead("X killed the ally", a)
	s.wantDamage(v, 2)
}

func TestFR_TRG_005_IfYouDoSkipsYWhenXPickedNothing(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, timeoutPick)
	s.iydDeploy(iydTrigger(true, iydOneOpp, iydPick(iydAlly, 1), iydIfDone(dmgTo(2, 0))))
	s.iydWantAlive("X picked nothing on a timeout", a)
	s.wantDamage(v, 0)
}

func TestFR_TRG_005_IfYouDoSkipsYWhenNothingMatchesX(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes})
	s.iydDeploy(iydTrigger(true, iydOneOpp, iydPick(iydAlly, 0), iydIfDone(dmgTo(2, 0))))
	s.wantDamage(v, 0)
}

func TestFR_TRG_004_DeathTriggersOfAnXKillArePlacedAfterTheTriggerFinishes(t *testing.T) {
	// Covers: SVY-R14
	s := newScene(t)
	hit := ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: 3}, Subject: ability.SubjectAll, Filter: iydEnemy}
	d := s.put(0, ZoneMainTheater, mortal("Doomed", 0, 1, killedTrigger(hit)))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 10, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes})
	s.iydDeploy(iydTrigger(true, iydOneOpp, iydPick(iydAlly, 0), iydIfDone(dmgTo(2, 0))))
	s.iydWantDead("X killed it", d)
	ev := damageTo(s.g.Log, v)
	if len(ev) != 2 || ev[0].Amount != 2 || ev[1].Amount != 3 {
		t.Errorf("Y's 2 damage comes first, then the death trigger's 3: got %+v (FR-TRG-004, SVY-R14)", ev)
	}
}

func TestFR_TRG_005_IfYouDoFollowsADiscardCount(t *testing.T) {
	// Covers: SVY-R14, SVY-R15
	disc := ability.Step{Op: ability.OpDiscard, Amount: ability.Value{N: 1}}
	t.Run("a card in hand", func(t *testing.T) {
		s, insts := discardScene(t, []string{"Card"})
		v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
		s.scriptPlan(0, nil, []Answer{yes})
		s.iydDeploy(iydTrigger(true, iydOneOpp, disc, iydIfDone(dmgTo(2, 0))))
		if !inDiscardPile(s.g.Players[0], insts[0]) {
			t.Fatalf("X discarded the card")
		}
		s.wantDamage(v, 2)
	})
	t.Run("an empty hand", func(t *testing.T) {
		s, _ := discardScene(t, nil)
		v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
		s.scriptPlan(0, nil, []Answer{yes})
		s.iydDeploy(iydTrigger(true, iydOneOpp, disc, iydIfDone(dmgTo(2, 0))))
		s.wantDamage(v, 0)
	})
}
