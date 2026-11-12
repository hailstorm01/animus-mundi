package engine

import (
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0216, Veil Render: "You may deploy this card from your discard
// pile by killing an allied unit in addition to paying its deployment costs."
// R3 T3, Reserves, power 5, health 4 (FR-TRN-003, FR-TRN-003a, FR-CST-001,
// FR-MOD-001; SVY-R46). Player 0 is the active player and deploys.

// vrHQ is the Headquarters asset of player p.
func vrHQ(s *scene, p PlayerID) *Asset {
	s.t.Helper()
	for _, a := range s.g.Assets() {
		if a.Controller == p && a.Zone == ZoneHeadquarters {
			return a
		}
	}
	s.t.Fatalf("player %d has no Headquarters asset", p)
	return nil
}

func TestCard_VeilRender_DiscardDeployKillsTheChosenAlliedUnit(t *testing.T) {
	// Covers: FR-TRN-003a
	s := newScene(t)
	main := s.put(0, ZoneMainTheater, unit("Main Unit", 1, 5, Keywords{}))
	res := s.put(0, ZoneReserves, unit("Reserves Unit", 1, 5, Keywords{}))
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	p := s.g.Players[0]
	s.setPool(0, 2, 3)
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("at R2 T3 (one Requisition short) no discard deploy of Veil Render is offered: %d (FR-TRN-003)", n)
	}
	s.setPool(0, 3, 3)
	if n := ddOffers(s.g, inst); n != 1 {
		t.Fatalf("at R3 T3 with allied units, the discard deploy is offered exactly once (Reserves only): %d (FR-TRN-003)", n)
	}
	kcScript(s, nil, main)
	s.do(ddDeploy(inst))
	if s.alive(main) || !inDiscard(p, main.Inst) {
		t.Errorf("the allied unit chosen for the cost is dead and in its owner's discard pile (FR-TRN-003a)")
	}
	if !s.alive(res) {
		t.Errorf("the unchosen allied unit is untouched (FR-TRN-003a)")
	}
	a := dcAssetOf(s.g, inst)
	if a == nil || a.Zone != ZoneReserves || a.Controller != 0 {
		t.Fatalf("Veil Render is an asset in player 0's Reserves (FR-TRN-003)")
	}
	if s.g.Power(a) != 5 || s.g.MaxHealth(a) != 4 {
		t.Errorf("Veil Render has power 5 and Maximum Health 4: got %d/%d", s.g.Power(a), s.g.MaxHealth(a))
	}
	if inDiscard(p, inst) {
		t.Errorf("Veil Render left the discard pile (FR-TRN-003)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the deploy paid R3 T3: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
}

func TestCard_VeilRender_ScratchCompanyCanPay(t *testing.T) {
	// Covers: SVY-R46
	s := newScene(t)
	sc := s.putScratch(0)
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	p := s.g.Players[0]
	s.setPool(0, 3, 3)
	if n := ddOffers(s.g, inst); n != 1 {
		t.Fatalf("a Scratch Company can pay the kill, so the discard deploy is offered: %d (SVY-R46)", n)
	}
	s.script(0)
	s.do(ddDeploy(inst))
	if s.alive(sc) || !inDiscard(p, sc.Inst) {
		t.Errorf("the Scratch Company is killed into player 0's discard pile (SVY-R46)")
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("Veil Render is an asset in Reserves (FR-TRN-003)")
	}
}

func TestCard_VeilRender_NotDeployableWithoutAnAlliedUnit(t *testing.T) {
	// Covers: SVY-R46, SVY-R29
	s := newScene(t)
	st := s.put(0, ZoneMainTheater, structure("Wall", 0, 5, Keywords{}))
	enemy := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	p := s.g.Players[0]
	s.setPool(0, 5, 5)
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("only a Structure and the Headquarters are allied: no discard deploy is offered, got %d (SVY-R46, SVY-R29)", n)
	}
	s.reject(ddDeploy(inst))
	if !inDiscard(p, inst) {
		t.Errorf("Veil Render stays in the discard pile (SVY-R29)")
	}
	if !s.alive(st) || !s.alive(enemy) || !s.alive(vrHQ(s, 0)) || !s.alive(vrHQ(s, 1)) {
		t.Errorf("nothing is killed: the Structure, the enemy unit and both Headquarters are alive (SVY-R46)")
	}
	s.wantDamage(st, 0)
	s.wantDamage(vrHQ(s, 0), 0)
	s.wantDamage(vrHQ(s, 1), 0)
	if p.Req != 5 || p.Tac != 5 {
		t.Errorf("a rejected Deploy spends nothing: pool R%d T%d, want R5 T5 (SVY-R29)", p.Req, p.Tac)
	}
}

func TestCard_VeilRender_TimedOutCostKillsTheFirstAlliedUnit(t *testing.T) {
	// Covers: FR-MOD-001, SVY-R46, SVY-R16
	s := newScene(t)
	main := s.put(0, ZoneMainTheater, unit("Main Unit", 1, 5, Keywords{}))
	a := s.put(0, ZoneReserves, unit("Unit A", 1, 5, Keywords{}))
	b := s.put(0, ZoneReserves, unit("Unit B", 1, 5, Keywords{}))
	st := s.put(0, ZoneMainTheater, structure("Wall", 0, 5, Keywords{}))
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	s.setPool(0, 3, 3)
	s.script(0, Answer{Timeout: true})
	s.do(ddDeploy(inst))
	asks := s.asked(0)
	if len(asks) != 1 {
		t.Fatalf("exactly one decision (the cost's pick) is asked: %d (FR-MOD-001)", len(asks))
	}
	got := ppRefs(asks[0])
	sort.Ints(got)
	want := []int{main.ID, a.ID, b.ID}
	sort.Ints(want)
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("the pick offers exactly the three allied units, never the Structure or the Headquarters: got %v, want %v (SVY-R46)", got, want)
	}
	if s.alive(a) {
		t.Errorf("a timed-out pick kills the first allied unit in fixed order, Reserves before the Main Theater, oldest first: Unit A (FR-MOD-001 6, SVY-R16)")
	}
	if !s.alive(main) || !s.alive(b) || !s.alive(st) || !s.alive(vrHQ(s, 0)) {
		t.Errorf("only Unit A is killed: the Main Theater unit, Unit B, the Structure and the Headquarters are alive (SVY-R46)")
	}
	if d := dcAssetOf(s.g, inst); d == nil || d.Zone != ZoneReserves {
		t.Errorf("Veil Render is an asset in Reserves (FR-TRN-003)")
	}
}

func TestCard_VeilRender_CostDeathTriggerResolvesFirst(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	tide := s.put(0, ZoneMainTheater, s.card("Unceasing Tide"))
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	p := s.g.Players[0]
	s.setPool(0, 3, 3)
	s.script(0)
	s.do(ddDeploy(inst))
	if !inHand(p, tide.Inst) || inDiscard(p, tide.Inst) {
		t.Errorf("Unceasing Tide's death trigger returns it to its owner's hand, not the discard pile (FR-TRN-003a)")
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("Veil Render is an asset in Reserves (FR-TRN-003)")
	}
	var place, trig, deploy int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Veil Render" && place == 0:
			place = e.Seq
		case e.Kind == "place" && e.Card == "Unceasing Tide" && place != 0 && trig == 0:
			trig = e.Seq
		case e.Kind == "deploy" && e.Inst == inst.ID && deploy == 0:
			deploy = e.Seq
		}
	}
	if place == 0 || trig == 0 || deploy == 0 {
		t.Fatalf("missing events: place %d, trigger place %d, deploy %d", place, trig, deploy)
	}
	if trig < place || trig > deploy {
		t.Errorf("the cost's death trigger goes on top of the deploy and resolves first: trigger place seq %d, deploy place seq %d, deploy seq %d (FR-TRN-003a)", trig, place, deploy)
	}
}

func TestCard_VeilRender_ReplacedKillStillPays(t *testing.T) {
	// Covers: SVY-R15, FR-REP-001
	s := newScene(t)
	si := s.put(0, ZoneReserves, s.card("Shifting Impossibility"))
	sc := s.putScratch(0)
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	p := s.g.Players[0]
	s.setPool(0, 3, 3)
	s.scriptRefs(0, nil, sc)
	s.do(ddDeploy(inst))
	if s.alive(sc) || !inHand(p, sc.Inst) || inDiscard(p, sc.Inst) {
		t.Errorf("the Scratch Company chosen for the cost is returned to its owner's hand instead of the discard pile (FR-REP-001)")
	}
	s.wantDamage(si, 1)
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("the replaced kill still pays the cost, so Veil Render is an asset in Reserves (SVY-R15)")
	}
}

func TestCard_VeilRender_HandDeployHasNoKill(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inst := s.give(0, s.card("Veil Render"))
	s.setPool(0, 3, 3)
	s.script(0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a hand Deploy asks no decision: %d (FR-TRN-003)", n)
	}
	if !s.alive(ally) {
		t.Errorf("a hand Deploy kills nothing (FR-TRN-003)")
	}
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Errorf("Veil Render is an asset in Reserves (FR-TRN-003)")
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the hand deploy paid R3 T3: pool R%d T%d, want R0 T0 (FR-TRN-003)", p.Req, p.Tac)
	}
}

func TestCard_VeilRender_FreeDeployWaivesTheKill(t *testing.T) {
	// Covers: FR-TRN-003a, FR-CST-001
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inst := s.inDiscard(0, s.card("Veil Render"))[0]
	man := s.give(0, dfManeuver(0, 0, ability.ZoneAny))
	s.setPool(0, 3, 3)
	s.script(0)
	s.do(enactAction(man))
	if a := dcAssetOf(s.g, inst); a == nil || a.Zone != ZoneReserves {
		t.Fatalf("the free deploy puts Veil Render into Reserves (FR-TRN-003a)")
	}
	if !s.alive(ally) {
		t.Errorf("a free deploy waives the kill cost: the allied unit is alive (FR-TRN-003a)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a free deploy asks no cost decision: %d (FR-CST-001)", n)
	}
	if p := s.g.Players[0]; p.Req != 3 || p.Tac != 3 {
		t.Errorf("a free deploy pays no resources: pool R%d T%d, want R3 T3 (FR-CST-001)", p.Req, p.Tac)
	}
}

func TestCard_VeilRender_NotFromTheOpponentsPile(t *testing.T) {
	// Covers: FR-TRN-003
	s := newScene(t)
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	inst := s.inDiscard(1, s.card("Veil Render"))[0]
	s.setPool(0, 5, 5)
	if n := ddOffers(s.g, inst); n != 0 {
		t.Errorf("a card in the opponent's discard pile is never offered: %d (FR-TRN-003)", n)
	}
	s.reject(ddDeploy(inst))
	if !inDiscard(s.g.Players[1], inst) {
		t.Errorf("Veil Render stays in player 1's discard pile (FR-TRN-003)")
	}
}
