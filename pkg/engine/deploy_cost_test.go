package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0213 (FR-TRN-003, FR-TRN-003a; SVY-R15, SVY-R29, FR-REP-001,
// FR-MOD-001): a Deploy goes through the declare sequence, so the deployed
// card's additional costs (Def.ExtraCosts) are paid after it is pushed, the
// triggers they fire go on top of it, and a Deploy whose additional cost cannot
// be paid is not legal. Cards are synthetic; no printed Unit or Structure has
// an additional cost. Player 0 is the active player and deploys.

// killCostUnit is a Unit costing req Requisition and tac Tactics whose printed
// additional cost is "kill an allied unit".
func killCostUnit(name string, req, tac int) *Card {
	c := unit(name, 1, 1, Keywords{})
	c.ReqCost, c.TacCost = req, tac
	c.Ability = &ability.Def{ExtraCosts: []ability.Cost{{Kind: ability.CostKill, Filter: alliedUnitFilter}}}
	return c
}

// dcDeployToReserves is player 0 deploying inst to Reserves.
func dcDeployToReserves(inst *Instance) Action {
	return Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves}
}

func dcOffersDeploy(g *Game, inst *Instance) bool {
	for _, a := range g.LegalActions() {
		if a.Kind == ActDeploy && a.Card == inst.ID && a.Zone == ZoneReserves {
			return true
		}
	}
	return false
}

// dcAssetOf returns the battlefield asset of inst, or nil.
func dcAssetOf(g *Game, inst *Instance) *Asset {
	for _, a := range g.Assets() {
		if a.Inst == inst {
			return a
		}
	}
	return nil
}

func TestFR_TRN_003a_DeployPaysItsKillCostAfterThePush(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	martyr := s.put(0, ZoneMainTheater, mortal("Martyr", 1, 1, killedTrigger(drawN(1))))
	bystander := s.put(0, ZoneMainTheater, unit("Bystander", 1, 5, Keywords{}))
	inst := s.give(0, killCostUnit("Sacrificer", 2, 1))
	p := s.g.Players[0]
	p.Req, p.Tac = 2, 1
	kcScript(s, nil, martyr)
	hand := len(p.Hand) // includes the card being deployed
	s.do(dcDeployToReserves(inst))
	if s.alive(martyr) || !inDiscard(p, martyr.Inst) {
		t.Errorf("the allied unit chosen for the cost is killed into its owner's discard pile (FR-TRN-003a)")
	}
	if !s.alive(bystander) {
		t.Errorf("only the chosen allied unit is killed; the other is untouched (FR-TRN-003a)")
	}
	a := dcAssetOf(s.g, inst)
	if a == nil || a.Zone != ZoneReserves {
		t.Errorf("the deployed card is an asset in player 0's Reserves once the Deploy resolves (FR-TRN-003a)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Deploy paid the card's Requisition and Tactics: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecidePick {
		t.Fatalf("the kill cost is asked as one DecidePick, got %d decisions %v (FR-TRN-003a)", len(asks), kindsOf(asks))
	}
	got, want := ppRefs(asks[0]), []int{martyr.ID, bystander.ID}
	if len(got) != 2 || !((got[0] == want[0] && got[1] == want[1]) || (got[0] == want[1] && got[1] == want[0])) {
		t.Errorf("the pick lists both allied units: got %v, want %v (FR-TRN-003a)", got, want)
	}
	// A successful draw logs no event, so the trigger's resolution is pinned by its own
	// place event (it went on top of the deploy) and by player 0's hand growing by one.
	var place, destroy, trig, deploy int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Sacrificer" && place == 0:
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
	if len(p.Hand) != hand {
		t.Errorf("the death trigger's draw resolved: hand is %d, want %d (the deployed card left, one card drawn) (FR-TRN-003a)", len(p.Hand), hand)
	}
}

func TestFR_TRN_003a_DeployKillCostNeedsAnAlliedUnit(t *testing.T) {
	// Covers: SVY-R29, FR-TRN-003
	s := newScene(t)
	inst := s.give(0, killCostUnit("Sacrificer", 1, 1))
	p := s.g.Players[0]
	p.Req, p.Tac = 1, 1
	if dcOffersDeploy(s.g, inst) {
		t.Errorf("a Deploy whose kill cost has no allied unit to pay it is not offered (SVY-R29, FR-TRN-003a)")
	}
	s.reject(dcDeployToReserves(inst))
	if !inHand(p, inst) || p.Req != 1 || p.Tac != 1 {
		t.Errorf("a rejected Deploy changes nothing: in hand %v, Req %d, Tac %d, want true, 1, 1 (SVY-R29)", inHand(p, inst), p.Req, p.Tac)
	}
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	if !dcOffersDeploy(s.g, inst) {
		t.Fatalf("with an allied unit to kill, the Deploy is offered (FR-TRN-003a)")
	}
	s.script(0) // any decision asked is an error
	s.do(dcDeployToReserves(inst))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option for the cost is taken without a decision, got %d (FR-MOD-001 sec. 1)", n)
	}
	if s.alive(ally) || !inDiscard(p, ally.Inst) {
		t.Errorf("the only allied unit is killed as the cost (FR-TRN-003a)")
	}
}

func TestFR_TRN_003a_DeployReplacedKillCostStaysPaid(t *testing.T) {
	// Covers: SVY-R15, FR-REP-001
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	inst := s.give(0, killCostUnit("Sacrificer", 0, 0))
	p := s.g.Players[0]
	s.scriptRefs(0, nil, sc)
	s.do(dcDeployToReserves(inst))
	if s.alive(sc) || !inHand(p, sc.Inst) || inDiscard(p, sc.Inst) {
		t.Errorf("the Scratch Company killed to pay the cost is returned to its owner's hand instead (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("the replaced kill still pays the cost, so the deploy resolves: the card is an asset in Reserves (SVY-R15)")
	}
}

func TestFR_TRN_003_DeployWithoutAdditionalCostsAsksNothing(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	inst := s.give(0, unit("Plain", 1, 1, Keywords{}))
	s.script(0) // any decision asked is an error
	placed := recordPlaced(s)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a Deploy with no additional costs asks no decision, got %d (FR-TRN-003)", n)
	}
	if len(*placed) == 0 || (*placed)[0].Kind != StackDeploy || (*placed)[0].Zone != ZoneMainTheater {
		t.Errorf("the placed deploy item carries the declared zone, the Main Theater: got %+v (FR-TRN-003)", *placed)
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneMainTheater {
		t.Errorf("the asset enters the declared zone, the Main Theater (FR-TRN-003)")
	}
}
