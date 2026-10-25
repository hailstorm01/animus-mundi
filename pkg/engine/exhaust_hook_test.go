package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for the exhaust hook (T-0063): a printed trigger that fires when its
// own asset exhausts. FR-CAP-001 and SVY-R17 (FRD 01, "What Counts as
// Exhausting"): every change from Readied to Exhausted counts, whether by
// Generate, March, Attack or Deploy Scratch Company, and entering the
// battlefield Exhausted does not. Stack-pushed actions place the trigger on top
// of the action, so it resolves before the action does. Generate is off-stack,
// so the trigger is placed after the generation and resolves before the player
// takes another action. FR-TRG-002 makes it a self hook: only the asset's own
// exhausting fires it.
//
// Cards are synthetic: an exhaust trigger that Scratches 1, observed with
// wantScratches and by the order of events in the game log.
//
// This file does not compile until ability has a HookExhausted constant; that
// is expected until T-0063 is implemented.

// exhaustTrigger is one "Whenever this asset exhausts, <steps>" ability.
func exhaustTrigger(steps ...ability.Step) ability.Trigger {
	return ability.Trigger{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: steps}}
}

// exhaustWatcher is a unit printing a "whenever this exhausts, Scratch 1" trigger.
func exhaustWatcher(name string) *Card {
	return striker(name, 1, 5, Keywords{}, exhaustTrigger(scratchN(1)))
}

// logIndex is the position of the first event of the given kind in the log, or -1.
func logIndex(g *Game, kind string) int {
	for i, e := range g.Log {
		if e.Kind == kind {
			return i
		}
	}
	return -1
}

// --- entering Exhausted is not exhausting (FR-CAP-001, SVY-R17) ---

func TestFR_CAP_001_EnteringExhaustedIsNotAnExhaust(t *testing.T) {
	s := newScene(t)
	c := exhaustWatcher("Weary Recruit")
	c.Keywords.DeploysExhausted = true
	s.deployFromHand(0, c)
	var a *Asset
	for _, x := range s.g.Players[0].Reserves {
		if x.Inst.Card == c {
			a = x
		}
	}
	if a == nil || !a.Exhausted {
		t.Fatalf("precondition: a Deploys Exhausted unit enters Reserves Exhausted, got %v", a)
	}
	s.wantScratches(0, 0, "an asset that enters the battlefield Exhausted never was Readied, so it has not exhausted and its exhaust trigger does not fire (SVY-R17)")
	s.wantSettled("deploying a Deploys Exhausted unit")
}

// --- every Readied-to-Exhausted change fires it (FR-CAP-001, SVY-R17) ---

func TestFR_CAP_001_ExhaustTriggerFiresWhenItsAssetGeneratesRequisition(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Requisition})
	s.wantScratches(0, 1, "generating resources exhausts the asset, which fires its exhaust trigger once (SVY-R17)")
	s.wantScratches(1, 0, "the trigger belongs to its asset's controller (FR-TRG-001)")
}

func TestFR_CAP_001_ExhaustTriggerFiresWhenItsAssetGeneratesTactics(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneMainTheater, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Tactics})
	s.wantScratches(0, 1, "generating Tactics in the Main Theater exhausts the asset, which fires its exhaust trigger once (SVY-R17)")
}

func TestFR_CAP_001_ExhaustTriggerFiresWhenItsAssetMarches(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	s.wantScratches(0, 1, "a March exhausts the asset as its cost, which fires its exhaust trigger once (SVY-R17)")
}

func TestFR_CAP_001_ExhaustTriggerFiresWhenItsAssetAttacks(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneMainTheater, exhaustWatcher("Watcher"))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 9, Keywords{}))
	s.do(s.attack(w, def))
	s.wantScratches(0, 1, "an Attack exhausts the attacker as its cost, which fires its exhaust trigger once (SVY-R17)")
	s.wantDamage(def, 1)
}

func TestFR_CAP_001_ExhaustTriggerFiresWhenAHeadquartersDeploysScratchCompany(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.HQ.Inst.Card.Ability = &ability.Def{Triggers: []ability.Trigger{exhaustTrigger(scratchN(1))}}
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	s.wantScratches(0, 2, "Deploy Scratch Company exhausts the Headquarters, which fires its exhaust trigger once (SVY-R17): one Scratch Company from the trigger and one from the action (FR-ZON-005)")
	s.wantSettled("a Headquarters Deploy Scratch Company with an exhaust trigger")
}

func TestFR_CAP_001_HeadquartersExhaustTriggerFiresOnceWhenItGenerates(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.HQ.Inst.Card.Ability = &ability.Def{Triggers: []ability.Trigger{exhaustTrigger(scratchN(1))}}
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: p.HQ.ID, Resource: Tactics})
	s.wantScratches(0, 1, "a Headquarters generating resources exhausts once, so its exhaust trigger fires once (SVY-R17)")
}

// --- placement (SVY-R17) ---

func TestFR_CAP_001_MarchExhaustTriggerResolvesBeforeTheMarch(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	sc, mv := logIndex(s.g, "scratch"), logIndex(s.g, "march")
	if sc < 0 || mv < 0 {
		t.Fatalf("expected a scratch event and a march event in the log, got scratch at %d, march at %d", sc, mv)
	}
	if sc > mv {
		t.Errorf("a March's exhaust trigger is placed on top of the March and resolves first (SVY-R17): scratch event %d came after march event %d", sc, mv)
	}
	if got := s.g.Asset(w.ID); got == nil || got.Zone != ZoneMainTheater {
		t.Errorf("the March still resolves after the trigger: watcher is %v, want in the Main Theater", got)
	}
	s.wantSettled("a March with an exhaust trigger")
}

func TestFR_CAP_001_AttackExhaustTriggerResolvesBeforeCombatDamage(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneMainTheater, exhaustWatcher("Watcher"))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 9, Keywords{}))
	s.do(s.attack(w, def))
	sc, dm := logIndex(s.g, "scratch"), logIndex(s.g, "damage")
	if sc < 0 || dm < 0 {
		t.Fatalf("expected a scratch event and a damage event in the log, got scratch at %d, damage at %d", sc, dm)
	}
	if sc > dm {
		t.Errorf("an Attack's exhaust trigger is placed on top of the Attack and resolves before combat damage (SVY-R17, FR-CBT-002 Step 2): scratch event %d came after damage event %d", sc, dm)
	}
	s.wantSettled("an Attack with an exhaust trigger")
}

func TestFR_CAP_001_GenerateExhaustTriggerResolvesAfterTheGenerationBeforeTheNextAction(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	before := s.g.Players[0].Req
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Requisition})
	gen, sc := logIndex(s.g, "generate"), logIndex(s.g, "scratch")
	if gen < 0 || sc < 0 {
		t.Fatalf("expected a generate event and a scratch event in the log, got generate at %d, scratch at %d", gen, sc)
	}
	if sc < gen {
		t.Errorf("a Generate's exhaust trigger is placed after the generation completes (SVY-R17): scratch event %d came before generate event %d", sc, gen)
	}
	if got := s.g.Players[0].Req; got != before+w.Inst.Card.ReqYield {
		t.Errorf("Requisition = %d, want %d: the generation is not affected by the trigger", got, before+w.Inst.Card.ReqYield)
	}
	s.wantScratches(0, 1, "the trigger has resolved by the time Apply returns, before the player can act again (SVY-R17)")
	s.wantSettled("a Generate with an exhaust trigger")
}

func TestFR_CAP_001_GenerateTriggerResolvesSoALaterMarchIsLegal(t *testing.T) {
	// The Generate's trigger resolves before the next action, so the stack is
	// empty and the March that follows is legal (FR-QUE-001).
	s := newScene(t)
	w := s.put(0, ZoneMainTheater, exhaustWatcher("Generator"))
	m := s.put(0, ZoneReserves, exhaustWatcher("Marcher"))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Tactics})
	s.do(Action{Kind: ActMarch, Player: 0, Asset: m.ID})
	s.wantScratches(0, 2, "each exhaust fires its own asset's trigger once")
	s.wantSettled("a Generate then a March, each with an exhaust trigger")
}

// --- no other change fires it (FR-TRG-002) ---

func TestFR_CAP_001_ReadyingAtTurnStartDoesNotFireTheExhaustTrigger(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Requisition})
	s.wantScratches(0, 1, "precondition: the Generate fired the trigger once")
	s.endTurn()
	s.endTurn()
	if w.Exhausted {
		t.Fatalf("precondition: the watcher readies at the start of its controller's turn")
	}
	s.wantScratches(0, 1, "going from Exhausted to Readied is not an exhaust, so it fires nothing (SVY-R17)")
}

func TestFR_TRG_002_AnotherAssetExhaustingDoesNotFireTheExhaustTrigger(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	other := s.put(0, ZoneReserves, unit("Other", 1, 1, Keywords{}))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: other.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: s.g.Players[0].HQ.ID, Resource: Requisition})
	s.wantScratches(0, 0, "the exhaust hook is a self hook: only the watcher's own exhausting fires its trigger (FR-TRG-002)")
}

func TestFR_TRG_002_AnOpposingAssetExhaustingDoesNotFireTheExhaustTrigger(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneReserves, exhaustWatcher("Watcher"))
	att := s.put(0, ZoneMainTheater, unit("Attacker", 1, 5, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 9, Keywords{}))
	s.do(s.attack(att, def))
	s.wantScratches(1, 0, "an opposing attacker exhausting does not fire the watcher's trigger (FR-TRG-002)")
}

// --- no limit (FR-TRG-002) ---

func TestFR_TRG_002_ExhaustTriggerFiresOnEachExhaustAcrossTurns(t *testing.T) {
	s := newScene(t)
	w := s.put(0, ZoneReserves, exhaustWatcher("Watcher"))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: w.ID, Resource: Requisition})
	s.wantScratches(0, 1, "first exhaust")
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	s.wantScratches(0, 2, "a trigger with no printed limit fires on the next turn's exhaust too")
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: w.ID})
	s.wantScratches(0, 3, "and again on the turn after")
	s.wantSettled("three exhausts over three turns")
}
