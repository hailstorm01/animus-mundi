package engine

import (
	"slices"
	"testing"
)

// Infested Apostle (T-0073): "When this unit deploys, you may kill an allied
// unit. If you do, deal 2 damage to chosen opposing unit."
// Rulings: SVY-R9 ("you may" asked on placement), SVY-R3 (Y's target locked on
// placement), SVY-R14 (the kill is an untargeted pick on resolution and may
// take the Apostle itself), SVY-R16 (a timeout selects nothing).
// Player 0 deploys the Apostle; player 1 is the opponent.

// deployApostle gives player 0 the real Apostle with enough Requisition and
// declares its deploy into Reserves. It returns the card's instance.
func deployApostle(s *scene) *Instance {
	s.t.Helper()
	inst := s.give(0, s.card("Infested Apostle"))
	s.g.Players[0].Req = 4
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return inst
}

func inDiscard(p *Player, inst *Instance) bool {
	for _, d := range p.Discard {
		if d == inst {
			return true
		}
	}
	return false
}

// Covers: FR-MOD-001, FR-TGT-001, FR-TRG-005
func TestCard_InfestedApostle_KillHappensOnResolutionAfterTargetLocked(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v1 := s.put(1, ZoneMainTheater, unit("Victim One", 0, 5, Keywords{}))
	v2 := s.put(1, ZoneMainTheater, unit("Victim Two", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, pick(v2), pick(ally))
	deployApostle(s)
	asks := s.asked(0)
	want := []DecisionKind{DecideOptional, DecideTarget, DecidePick}
	if got := kindsOf(asks); !slices.Equal(got, want) {
		t.Fatalf("'you may' and the opposing target are asked on placement, the kill pick as the trigger resolves: got %v, want %v (SVY-R9, SVY-R3, SVY-R14)", got, want)
	}
	if got := refs(asks[1]); !slices.Equal(got, ids(v1, v2)) {
		t.Errorf("the chosen target is an opposing unit, so only the opposing units are options: %v, want %v (SVY-R3)", got, ids(v1, v2))
	}
	if pk := asks[2]; !slices.Contains(refs(pk), ally.ID) || len(pk.Options) != 2 {
		t.Errorf("the kill pick offers the allied units on the battlefield, the other ally and the Apostle: %v (SVY-R14)", refs(pk))
	}
	if s.alive(ally) {
		t.Errorf("the picked allied unit is killed (SVY-R14)")
	}
	if !inDiscard(s.g.Players[0], ally.Inst) {
		t.Errorf("the killed allied unit goes to its owner's discard pile")
	}
	s.wantDamage(v2, 2)
	s.wantDamage(v1, 0)
}

// Covers: FR-MOD-001, FR-TGT-001
func TestCard_InfestedApostle_MayKillItself(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	// Say yes to "you may", then pick whichever allied unit is not Ally.
	s.asks = map[PlayerID][]Decision{}
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		if d.Kind == DecideOptional {
			return yes
		}
		for i, o := range d.Options {
			if o.Ref != ally.ID {
				return Answer{Picks: []int{i}}
			}
		}
		return Answer{Timeout: true}
	}
	inst := deployApostle(s)
	if got := len(s.asked(0)); got != 2 {
		t.Fatalf("'you may' and the kill pick are asked, the lone opposing unit is chosen automatically: %d decisions (SVY-R14, FR-MOD-001)", got)
	}
	if !inDiscard(s.g.Players[0], inst) {
		t.Errorf("the Apostle picked itself and is killed: it may select itself (SVY-R14)")
	}
	if !s.alive(ally) {
		t.Errorf("the other ally was not picked and is still on the battlefield")
	}
	s.wantDamage(v, 2)
}

// Covers: FR-MOD-001
func TestCard_InfestedApostle_TimeoutAtKillPickDealsNoDamage(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{yes}, timeoutPick)
	inst := deployApostle(s)
	if got := kindsOf(s.asked(0)); !slices.Equal(got, []DecisionKind{DecideOptional, DecidePick}) {
		t.Fatalf("'you may' then the kill pick are asked: got %v (SVY-R14)", got)
	}
	if !s.alive(ally) || inDiscard(s.g.Players[0], inst) || len(s.g.Players[0].Discard) != 0 {
		t.Errorf("a timeout at the kill pick selects none, so nothing is killed (SVY-R14, SVY-R16)")
	}
	s.wantDamage(v, 0)
	if d := damageTo(s.g.Log, v); len(d) != 0 {
		t.Errorf("X was not carried out, so Y does nothing: %+v (SVY-R14)", d)
	}
}

// Covers: FR-TRG-004, FR-MOD-001
func TestCard_InfestedApostle_DamageResolvesBeforeKilledUnitsDeathTriggers(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	scr := s.put(0, ZoneMainTheater, s.card("Portent Scrivener"))
	v := s.put(1, ZoneMainTheater, unit("Victim", 0, 5, Keywords{}))
	top := s.onTop(0, vanilla("Drawn", "Next")...)
	s.scriptPlan(0, nil, []Answer{yes}, pick(scr))
	start := len(s.g.Log)
	deployApostle(s)
	if s.alive(scr) {
		t.Fatalf("precondition: the Scrivener was picked and killed")
	}
	dmgAt, placeAt := -1, -1
	for i := start; i < len(s.g.Log); i++ {
		e := s.g.Log[i]
		if e.Kind == "damage" && e.Target == v.ID && dmgAt < 0 {
			dmgAt = i
		}
		if e.Kind == "place" && e.Source == scr.ID && placeAt < 0 {
			placeAt = i
		}
	}
	if dmgAt < 0 {
		t.Fatalf("the Apostle deals 2 damage to the chosen unit after the kill (SVY-R14)")
	}
	if placeAt < 0 {
		t.Fatalf("the killed Scrivener's death trigger is placed once the Apostle's trigger has finished (FR-TRG-004)")
	}
	if dmgAt > placeAt {
		t.Errorf("the 2 damage resolves before the killed unit's death trigger is placed (SVY-R14, FR-TRG-004)")
	}
	s.wantDamage(v, 2)
	if !inHand(p, top[0]) {
		t.Errorf("the death trigger then resolves: the Scrivener was killed on its controller's turn, so it draws a card")
	}
}

// Covers: FR-MOD-001
func TestCard_InfestedApostle_DecliningAsksNothingMoreAndDoesNothing(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 0, 5, Keywords{}))
	v1 := s.put(1, ZoneMainTheater, unit("Victim One", 0, 5, Keywords{}))
	v2 := s.put(1, ZoneMainTheater, unit("Victim Two", 0, 5, Keywords{}))
	s.scriptPlan(0, nil, []Answer{no})
	inst := deployApostle(s)
	if got := kindsOf(s.asked(0)); !slices.Equal(got, []DecisionKind{DecideOptional}) {
		t.Errorf("a declined 'you may' asks no target and no pick: got %v (SVY-R9)", got)
	}
	if !s.alive(ally) || len(s.g.Players[0].Discard) != 0 || inDiscard(s.g.Players[0], inst) {
		t.Errorf("declining kills nothing (SVY-R9)")
	}
	s.wantDamage(v1, 0)
	s.wantDamage(v2, 0)
}
