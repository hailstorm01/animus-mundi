package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0214 (FR-TRN-003, FR-TRN-003a, FR-CST-001; SVY-R15, SVY-R29,
// FR-REP-001, FR-MOD-001, FR-QUE-001): a Unit that prints a discard-active
// deploy permission ("You may deploy this card from your discard pile by ...")
// can be deployed by its owner from their own discard pile as a manual Deploy,
// with Action.From = ZoneDiscard. The stack must be empty, the card moves from
// the pile onto the stack, its effective cost is paid and the permission's
// additional costs are paid after the push, their triggers going on top.
// Cards are synthetic. Player 0 is the active player and deploys.

// ddPermission is the printed "deploy from your discard pile by killing an
// allied unit" permission.
func ddPermission() ability.Static {
	return ability.Static{
		Kind:       ability.StaticDeployPermission,
		ActiveZone: ability.ZoneDiscard,
		Costs:      []ability.Cost{{Kind: ability.CostKill, Filter: alliedUnitFilter}},
	}
}

// ddUnit is a Reserves-only Unit costing req Requisition and tac Tactics that
// prints only the discard deploy permission.
func ddUnit(name string, req, tac int) *Card {
	c := unit(name, 1, 1, Keywords{})
	c.ReqCost, c.TacCost = req, tac
	c.DeployZones = []Zone{ZoneReserves}
	c.Ability = &ability.Def{Statics: []ability.Static{ddPermission()}}
	return c
}

// ddDeploy is player 0 deploying inst to Reserves from their discard pile.
func ddDeploy(inst *Instance) Action {
	return Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves, From: ZoneDiscard}
}

// ddOffers counts the discard deploys of inst in LegalActions.
func ddOffers(g *Game, inst *Instance) int {
	n := 0
	for _, a := range g.LegalActions() {
		if a.Kind == ActDeploy && a.Card == inst.ID && a.From == ZoneDiscard {
			n++
		}
	}
	return n
}

func TestFR_TRN_003_DiscardDeployPaysItsKillCostAfterThePush(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	pile := ability.Filter{Allegiance: ability.AllegianceAllied, Zone: ability.ZoneDiscard}
	martyr := s.put(0, ZoneMainTheater, mortal("Martyr", 1, 1, ability.Trigger{
		Hook: ability.HookKilled,
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: pile, UpTo: 1}},
			Steps:   []ability.Step{drawN(1)},
		},
	}))
	bystander := s.put(0, ZoneMainTheater, unit("Bystander", 1, 5, Keywords{}))
	inst := s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]
	p := s.g.Players[0]
	s.setPool(0, 2, 1)
	kcScript(s, nil, martyr)
	s.do(ddDeploy(inst))
	if s.alive(martyr) || !inDiscard(p, martyr.Inst) {
		t.Errorf("the allied unit chosen for the cost is killed into its owner's discard pile (FR-TRN-003a)")
	}
	if !s.alive(bystander) {
		t.Errorf("only the chosen allied unit is killed; the other is untouched (FR-TRN-003a)")
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Errorf("the deployed card is an asset in player 0's Reserves once the Deploy resolves (FR-TRN-003)")
	}
	if inDiscard(p, inst) {
		t.Errorf("the deployed card left the discard pile (FR-TRN-003)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Deploy paid the card's R2 T1: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
	var place, destroy, trig, deploy int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Phoenix" && place == 0:
			place = e.Seq
		case e.Kind == "destroy" && e.Target == martyr.ID && destroy == 0:
			destroy = e.Seq
		case e.Kind == "place" && e.Card == "Martyr" && place != 0 && trig == 0:
			trig = e.Seq
		case e.Kind == "deploy" && e.Inst == inst.ID && deploy == 0:
			deploy = e.Seq
		}
	}
	if place == 0 || destroy == 0 || trig == 0 || deploy == 0 {
		t.Fatalf("missing events: place %d, destroy %d, trigger place %d, deploy %d", place, destroy, trig, deploy)
	}
	if place > destroy {
		t.Errorf("the deploy is placed on the stack before the cost's kill (seq %d vs %d) (FR-TRN-003a)", place, destroy)
	}
	if trig < place || trig > deploy {
		t.Errorf("the cost's death trigger goes on top of the deploy and resolves first: trigger place seq %d, deploy place seq %d, deploy seq %d (FR-TRN-003a)", trig, place, deploy)
	}
	var targetAsk *Decision
	for _, d := range s.asked(0) {
		if d.Kind == DecideTarget {
			d := d
			targetAsk = &d
		}
	}
	if targetAsk == nil {
		t.Fatalf("the death trigger's up-to-1 target slot is asked as a DecideTarget (FR-TRN-003)")
	}
	var sawMartyr, sawDeployed bool
	for _, r := range ppRefs(*targetAsk) {
		sawMartyr = sawMartyr || r == martyr.Inst.ID
		sawDeployed = sawDeployed || r == inst.ID
	}
	if !sawMartyr {
		t.Errorf("the killed unit's card is in the discard pile and so a legal target (FR-TRN-003)")
	}
	if sawDeployed {
		t.Errorf("the deployed card is already on the stack, not in its origin zone, when the cost's triggers choose targets (FR-TRN-003)")
	}
	acts := eventsOfKind(s.g.Log, "action")
	if len(acts) != 1 || acts[0].Card != "Phoenix" || acts[0].Inst != inst.ID || acts[0].From != "Discard" {
		t.Errorf("the action event names the card and its origin, From \"Discard\": got %+v (FR-TRN-003)", acts)
	}
}

func TestFR_TRN_003a_DiscardDeployReplacedKillStaysPaid(t *testing.T) {
	// Covers: SVY-R15, FR-REP-001
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	inst := s.inDiscard(0, ddUnit("Phoenix", 0, 0))[0]
	p := s.g.Players[0]
	s.scriptRefs(0, nil, sc)
	s.do(ddDeploy(inst))
	if s.alive(sc) || !inHand(p, sc.Inst) || inDiscard(p, sc.Inst) {
		t.Errorf("the Scratch Company killed to pay the cost is returned to its owner's hand instead (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("the replaced kill still pays the cost, so the deploy resolves: the card is an asset in Reserves (SVY-R15)")
	}
}

func TestFR_TRN_003_DiscardDeployNeedsThePermission(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.setPool(0, 5, 5)
	p := s.g.Players[0]
	plain := unit("Plain", 1, 1, Keywords{})
	plain.DeployZones = []Zone{ZoneReserves}
	wrong := ddUnit("WrongZone", 1, 1)
	wrong.Ability.Statics[0].ActiveZone = ability.ZoneReserves
	for _, c := range []*Card{plain, wrong} {
		inst := s.inDiscard(0, c)[0]
		if n := ddOffers(s.g, inst); n != 0 {
			t.Errorf("%s has no discard-active deploy permission: %d discard deploys offered, want 0 (FR-TRN-003)", c.Name, n)
		}
		s.reject(ddDeploy(inst))
		if !inDiscard(p, inst) {
			t.Errorf("%s stays in the discard pile after a rejected Deploy (FR-TRN-003)", c.Name)
		}
	}
	if p.Req != 5 || p.Tac != 5 {
		t.Errorf("rejected Deploys spend nothing: pool R%d T%d, want R5 T5 (FR-TRN-003)", p.Req, p.Tac)
	}
}

func TestFR_TRN_003_DiscardDeployOnlyFromYourOwnPile(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.setPool(0, 5, 5)
	inst := s.inDiscard(1, ddUnit("Phoenix", 1, 1))[0]
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("a card in the opponent's discard pile is never offered: %d offers (FR-TRN-003)", n)
	}
	s.reject(ddDeploy(inst))
	if !inDiscard(s.g.Players[1], inst) {
		t.Errorf("the card stays in player 1's discard pile (FR-TRN-003)")
	}
}

func TestFR_TRN_003_DiscardDeployNeedsAnEmptyStack(t *testing.T) {
	// Covers: FR-QUE-001
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.setPool(0, 5, 5)
	inst := s.inDiscard(0, ddUnit("Phoenix", 1, 1))[0]
	if n := ddOffers(s.g, inst); n != 1 {
		t.Fatalf("precondition: with an empty stack the discard deploy is offered once, got %d", n)
	}
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(ddDeploy(inst))
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("no discard deploy is offered with an item on the stack: %d (FR-TRN-003, FR-QUE-001)", n)
	}
	if !inDiscard(s.g.Players[0], inst) {
		t.Errorf("the card stays in the discard pile (FR-TRN-003)")
	}
}

func TestFR_TRN_003a_DiscardDeployCostsMustBePayable(t *testing.T) {
	// Covers: SVY-R29
	s := newScene(t)
	inst := s.inDiscard(0, ddUnit("Phoenix", 2, 0))[0]
	p := s.g.Players[0]
	s.setPool(0, 5, 5)
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("with no allied unit to kill, the discard deploy is not offered: %d (SVY-R29, FR-TRN-003a)", n)
	}
	s.reject(ddDeploy(inst))
	if !inDiscard(p, inst) || p.Req != 5 || p.Tac != 5 {
		t.Errorf("a rejected Deploy changes nothing: in discard %v, pool R%d T%d, want true, R5 T5 (SVY-R29)", inDiscard(p, inst), p.Req, p.Tac)
	}
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.setPool(0, 1, 0)
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("one Requisition short of R2, the discard deploy is not offered: %d (FR-TRN-003)", n)
	}
	s.reject(ddDeploy(inst))
	if !s.alive(ally) || !inDiscard(p, inst) || p.Req != 1 {
		t.Errorf("a rejected Deploy kills nothing and spends nothing (FR-TRN-003)")
	}
	s.setPool(0, 2, 0)
	if n := ddOffers(s.g, inst); n != 1 {
		t.Errorf("with an allied unit and R2 the discard deploy is offered: %d (FR-TRN-003a)", n)
	}
}

func TestFR_CST_001_HandReductionDoesNotApplyInThePile(t *testing.T) {
	c := ddUnit("Amalgam", 2, 1)
	c.Ability.CostReduction = ability.CostReduction{Req: 1, Per: ability.Filter{
		Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	s := newScene(t)
	s.inDiscard(0, unit("Fodder A", 1, 1, Keywords{}), unit("Fodder B", 1, 1, Keywords{}))
	inst := s.inDiscard(0, c)[0]
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	p := s.g.Players[0]
	if req, tac := s.g.EffectiveCost(inst); req != 2 || tac != 1 {
		t.Errorf("a hand-active reduction does not apply in the discard pile: effective cost R%d T%d, want R2 T1 (FR-CST-001)", req, tac)
	}
	s.setPool(0, 1, 1)
	s.reject(ddDeploy(inst))
	if !inDiscard(p, inst) || p.Req != 1 || p.Tac != 1 {
		t.Errorf("a pile deploy with one Requisition less than printed is illegal and changes nothing (FR-CST-001)")
	}
	s.setPool(0, 2, 1)
	s.script(0)
	s.do(ddDeploy(inst))
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the pile deploy pays the printed cost: pool R%d T%d, want R0 T0 (FR-CST-001)", p.Req, p.Tac)
	}
}

func TestFR_TRN_003_HandDeployIgnoresThePermission(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inst := s.give(0, ddUnit("Phoenix", 2, 1))
	s.setPool(0, 2, 1)
	s.script(0) // any decision asked is an error
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a hand Deploy asks for no permission cost, got %d decisions (FR-TRN-003)", n)
	}
	if !s.alive(ally) {
		t.Errorf("a hand Deploy kills nothing (FR-TRN-003)")
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("the card is an asset in Reserves (FR-TRN-003)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("only R2 T1 was paid: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
	acts := eventsOfKind(s.g.Log, "action")
	if len(acts) != 1 || acts[0].From != "" {
		t.Errorf("a hand Deploy's action event has an empty From: got %+v (FR-TRN-003)", acts)
	}
}

func TestFR_TRN_003_DiscardDeployOfferedForEachZone(t *testing.T) {
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.setPool(0, 5, 5)
	c := ddUnit("Phoenix", 1, 1)
	c.DeployZones = []Zone{ZoneReserves, ZoneMainTheater}
	inst := s.inDiscard(0, c)[0]
	zones := map[Zone]int{}
	for _, a := range s.g.LegalActions() {
		if a.Kind == ActDeploy && a.Card == inst.ID && a.From == ZoneDiscard {
			zones[a.Zone]++
		}
	}
	if len(zones) != 2 || zones[ZoneReserves] != 1 || zones[ZoneMainTheater] != 1 {
		t.Errorf("one discard deploy per printed zone, Reserves and the Main Theater: got %v (FR-TRN-003)", zones)
	}
	s.script(0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater, From: ZoneDiscard})
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneMainTheater {
		t.Errorf("deploying to the Main Theater puts the card there (FR-TRN-003)")
	}
}

func TestFR_TRN_003_DeployFromAnotherZoneIsIllegal(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, unit("Plain", 1, 1, Keywords{}))
	s.setPool(0, 5, 5)
	s.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves, From: ZoneMainTheater})
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("a Deploy from a zone other than the hand or the discard pile is illegal and the card stays (FR-TRN-003)")
	}
}

func TestFR_CST_001_FreeDeployWaivesThePermissionCost(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Ally One", 1, 5, Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("Ally Two", 1, 5, Keywords{}))
	inst := s.inDiscard(0, ddUnit("Phoenix", 2, 1))[0]
	man := s.give(0, dfManeuver(0, 0, ability.ZoneAny))
	s.setPool(0, 3, 3)
	s.script(0) // any decision asked is an error
	s.do(enactAction(man))
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Fatalf("the free deploy puts the card into Reserves (FR-TRN-003a)")
	}
	if !s.alive(a1) || !s.alive(a2) {
		t.Errorf("a free deploy excludes the permission's kill cost: every allied unit survives (FR-TRN-003a, FR-CST-001)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a free deploy asks for no cost: %d decisions (FR-CST-001)", n)
	}
	if p := s.g.Players[0]; p.Req != 3 || p.Tac != 3 {
		t.Errorf("a free deploy pays no resources: pool R%d T%d, want R3 T3 (FR-CST-001)", p.Req, p.Tac)
	}
}
