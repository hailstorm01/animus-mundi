package engine

import (
	"reflect"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Card tests for Path cards. Each is written from the card's CSV rules text in
// data/cards/units-and-structures-path.csv and the FRD 01 lines the task cites.
// Cards are the real ones loaded with s.card and deployed from hand through
// ActDeploy. Targets are chosen by asset (scriptRefs, never by option index)
// and options are read from the recorded Decisions.

// deployPath gives player 0 the real card with req Requisition and declares its
// deploy into Reserves. It returns the card; find the asset with assetOf.
func deployPath(s *scene, name string, req int) *Card {
	s.t.Helper()
	c := s.card(name)
	inst := s.give(0, c)
	s.g.Players[0].Req = req
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	return c
}

func deploySaboteur(s *scene) *Card { return deployPath(s, "Saboteur", 3) }
func deploySappers(s *scene) *Card  { return deployPath(s, "Chainbreaker Sappers", 2) }

func wantOptions(t *testing.T, d Decision, what string, want ...*Asset) {
	t.Helper()
	if got, w := refs(d), ids(want...); !reflect.DeepEqual(got, w) {
		t.Errorf("%s: options got %v, want %v", what, got, w)
	}
}

// --- Saboteur ---
// "When this unit deploys, deal 2 damage to chosen opposing non-Scratch unit
// and 1 damage to chosen allied unit."  Unit, 3 Requisition, Reserves, 3/3.

// Control: the unit itself is a 3/3 that enters Reserves.
func TestCard_Saboteur_EntersReservesAsThreeThree(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	s.script(0)
	c := deploySaboteur(s)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Saboteur is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 3 || p.Health != 3 {
		t.Errorf("Saboteur is a 3/3: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Saboteur costs 3 Requisition: %d left", s.g.Players[0].Req)
	}
}

// Covers: FR-TGT-001
func TestCard_Saboteur_BothClausesHitTheirOwnChosenTarget(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("Ally B", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e2, a1)
	deploySaboteur(s)
	s.wantDamage(e2, 2)
	s.wantDamage(a1, 1)
	s.wantDamage(e1, 0)
	s.wantDamage(a2, 0)
	if got := totalDamage(s.g); got != 3 {
		t.Errorf("2 damage to the chosen opposing unit and 1 to the chosen allied unit, nothing else: total %d, want 3", got)
	}
}

// Covers: FR-TGT-001, FR-MOD-001
func TestCard_Saboteur_AsksOpposingClauseFirstThenAlliedClause(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e1, a1)
	c := deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 2 {
		t.Fatalf("two chosen targets with several options each are two questions: asked %d times", len(asked))
	}
	for i, d := range asked {
		if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
			t.Errorf("question %d is a mandatory single DecideTarget for the deploying player: %+v", i, d)
		}
	}
	wantOptions(t, asked[0], "first question (opposing non-Scratch unit)", e1, e2)
	wantOptions(t, asked[1], "second question (allied unit, which includes the Saboteur)", a1, assetOf(s.g, c))
}

// Covers: FR-TGT-001
func TestCard_Saboteur_ScratchCompanyIsNotAnOpposingOptionButIsAnAlliedOne(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	s.putScratch(1)
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	sc0 := s.putScratch(0)
	s.scriptRefs(0, nil, e1, sc0)
	c := deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 2 {
		t.Fatalf("asked %d times, want 2", len(asked))
	}
	wantOptions(t, asked[0], "the opposing clause says non-Scratch, so the opposing Scratch Company is excluded", e1, e2)
	wantOptions(t, asked[1], "the allied clause has no Scratch exclusion, so the allied Scratch Company is an option", a1, sc0, assetOf(s.g, c))
	if s.alive(sc0) {
		t.Errorf("the chosen allied Scratch Company (a 1/1) takes 1 damage and is destroyed")
	}
}

// Covers: FR-TGT-001
func TestCard_Saboteur_StructuresAndHeadquartersAreNotOptions(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, structure("Enemy Wall", 0, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally A", 1, 5, Keywords{}))
	s.put(0, ZoneMainTheater, structure("Ally Wall", 0, 5, Keywords{}))
	s.scriptRefs(0, nil, e1, a1)
	c := deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 2 {
		t.Fatalf("asked %d times, want 2", len(asked))
	}
	wantOptions(t, asked[0], "opposing clause is units only", e1, e2)
	wantOptions(t, asked[1], "allied clause is units only", a1, assetOf(s.g, c))
}

// The allied clause is "chosen allied unit" with no "other", so the Saboteur,
// already on the battlefield when its trigger is placed, is a legal choice.
// The chooser picks it by matching the option whose asset is the Saboteur.
// Covers: FR-TGT-001
func TestCard_Saboteur_ItselfIsALegalAlliedTarget(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	other := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	sab := s.card("Saboteur")
	s.asks = map[PlayerID][]Decision{}
	s.g.choosers[0] = func(g *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		for i, o := range d.Options {
			if x := g.Asset(o.Ref); x != nil && x.Inst.Card == sab {
				return Answer{Picks: []int{i}}
			}
		}
		return Answer{Picks: []int{0}}
	}
	inst := s.give(0, sab)
	s.g.Players[0].Req = 3
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	a := assetOf(s.g, sab)
	if a == nil {
		t.Fatalf("Saboteur is on the battlefield")
	}
	s.wantDamage(a, 1)
	s.wantDamage(other, 0)
	s.wantDamage(e, 2)
}

// FR-TGT-003 "Deployment Triggers Without Legal Targets" names Saboteur: with
// no opposing non-Scratch unit, the opposing clause does nothing, the Saboteur
// still deploys, and the allied clause still resolves (on the only allied
// unit, the Saboteur itself, chosen without a question).
// Covers: FR-TGT-001, FR-MOD-001
func TestFR_TGT_003_SaboteurWithNoOpposingNonScratchUnitStillDeploysAndAlliedClauseResolves(t *testing.T) {
	s := newScene(t)
	sc := s.putScratch(1) // an opposing Scratch Company is not a legal target of the opposing clause
	s.script(0)
	c := deploySaboteur(s)
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("Saboteur still enters the battlefield (FR-TGT-003)")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal opposing target and one legal allied target: no question (FR-MOD-001), asked %d times", n)
	}
	s.wantDamage(a, 1)
	s.wantDamage(sc, 0)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("only the allied clause resolves: total damage %d, want 1 (FR-TGT-003)", got)
	}
}

// Covers: FR-TGT-003
func TestCard_Saboteur_NoOpposingUnitStillResolvesAlliedClauseOnChosenUnit(t *testing.T) {
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, a1)
	c := deploySaboteur(s)
	if assetOf(s.g, c) == nil {
		t.Fatalf("Saboteur still enters the battlefield (FR-TGT-003)")
	}
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("only the allied clause has choices: asked %d times, want 1", len(asked))
	}
	wantOptions(t, asked[0], "the only question is the allied clause", a1, assetOf(s.g, c))
	s.wantDamage(a1, 1)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("the opposing clause does nothing: total damage %d, want 1", got)
	}
}

// Covers: FR-KWD-002h, FR-TGT-001
func TestCard_Saboteur_OpposingWardUnitCannotBeChosen(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e1, a1)
	deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) < 1 {
		t.Fatalf("no question asked")
	}
	wantOptions(t, asked[0], "an opposing Ward unit cannot be chosen by a triggered ability (FR-KWD-002h)", e1, e2)
	s.wantDamage(ward, 0)
}

// Covers: FR-KWD-002h, FR-TGT-003
func TestCard_Saboteur_OnlyOpposingUnitHasWardSoOpposingClauseDoesNothing(t *testing.T) {
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.script(0)
	c := deploySaboteur(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal opposing target and only the Saboteur as allied target: no question, asked %d times", n)
	}
	s.wantDamage(ward, 0)
	a := assetOf(s.g, c)
	if a == nil {
		t.Fatalf("Saboteur still deploys (FR-TGT-003)")
	}
	s.wantDamage(a, 1)
}

// Covers: FR-KWD-002h
func TestCard_Saboteur_AlliedWardUnitIsAnAlliedOption(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	ward := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, ward)
	c := deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("the single opposing unit is chosen without a question and the allied clause asks: asked %d times, want 1", len(asked))
	}
	wantOptions(t, asked[0], "Ward restricts only opponents, so the controller's Ward unit is an option (FR-KWD-002h)", ward, assetOf(s.g, c))
	s.wantDamage(ward, 1)
	s.wantDamage(e, 2)
}

// Covers: FR-ZON-002
func TestCard_Saboteur_LethalDamageDestroysBothChosenUnits(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Fragile Enemy", 1, 2, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Other Enemy", 1, 5, Keywords{}))
	a := s.put(0, ZoneMainTheater, unit("Fragile Ally", 1, 1, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Other Ally", 1, 5, Keywords{}))
	eInst, aInst := e.Inst, a.Inst
	s.scriptRefs(0, nil, e, a)
	deploySaboteur(s)
	if s.alive(e) || !inZone(s.g.Players[1].Discard, eInst) {
		t.Errorf("the 1/2 opposing unit takes 2 and is destroyed into its owner's Discard")
	}
	if s.alive(a) || !inZone(s.g.Players[0].Discard, aInst) {
		t.Errorf("the 1/1 allied unit takes 1 and is destroyed into its owner's Discard")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after resolution: %d items", len(s.g.Stack))
	}
}

// Covers: FR-MOD-001
func TestCard_Saboteur_SoleOpposingTargetIsNotAsked(t *testing.T) {
	s := newScene(t)
	s.putScratch(1)
	e := s.put(1, ZoneMainTheater, unit("Only Enemy", 1, 5, Keywords{}))
	a1 := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, a1)
	c := deploySaboteur(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("a single legal opposing target is selected automatically (FR-MOD-001 sec. 1) and only the allied clause asks: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "the one question is the allied clause", a1, assetOf(s.g, c))
	s.wantDamage(e, 2)
	s.wantDamage(a1, 1)
}

// Covers: FR-MOD-001
func TestCard_Saboteur_TimeoutOnOpposingChoiceDamagesExactlyOneLegalUnit(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	sc := s.putScratch(1)
	s.script(0, Answer{Timeout: true})
	c := deploySaboteur(s)
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("the opposing clause asks once and the allied clause (only the Saboteur) is automatic: asked %d times", n)
	}
	if (e1.Damage == 2) == (e2.Damage == 2) || e1.Damage+e2.Damage != 2 {
		t.Errorf("a timeout picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16): damage A=%d B=%d, want exactly one with 2", e1.Damage, e2.Damage)
	}
	s.wantDamage(sc, 0)
	s.wantDamage(assetOf(s.g, c), 1)
}

// --- Chainbreaker Sappers ---
// "When this unit deploys, deal 1 damage to chosen structure or Headquarters."
// Unit, 2 Requisition, Reserves, 3/2.

// Control: the unit itself is a 3/2 that enters Reserves.
func TestCard_ChainbreakerSappers_EntersReservesAsThreeTwo(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	c := deploySappers(s)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Sappers is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 3 || p.Health != 2 {
		t.Errorf("Sappers is a 3/2: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Sappers costs 2 Requisition: %d left", s.g.Players[0].Req)
	}
}

// Covers: FR-TGT-001
func TestCard_ChainbreakerSappers_OptionsAreStructuresAndHeadquartersOfEitherSideOnly(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	s.scriptRefs(0, nil, b.s1)
	deploySappers(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one chosen target is one question: asked %d times", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for the deploying player: %+v", d)
	}
	wantOptions(t, d, "structures and Headquarters of both players; units, Scratch Companies and the Sappers itself are excluded",
		b.s0, b.s1, s.g.Players[0].HQ, s.g.Players[1].HQ)
	s.wantDamage(b.s1, 1)
	if got := totalDamage(s.g); got != 1 {
		t.Errorf("exactly the chosen entity takes 1 damage: total %d, want 1", got)
	}
}

// Covers: FR-TGT-001
func TestCard_ChainbreakerSappers_CanDamageEachKindOfLegalTarget(t *testing.T) {
	cases := []struct {
		name string
		pick func(s *scene, b tboard) *Asset
	}{
		{"allied structure", func(s *scene, b tboard) *Asset { return b.s0 }},
		{"opposing structure", func(s *scene, b tboard) *Asset { return b.s1 }},
		{"allied Headquarters", func(s *scene, b tboard) *Asset { return s.g.Players[0].HQ }},
		{"opposing Headquarters", func(s *scene, b tboard) *Asset { return s.g.Players[1].HQ }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			b := s.targetBoard()
			want := tc.pick(s, b)
			s.scriptRefs(0, nil, want)
			deploySappers(s)
			s.wantDamage(want, 1)
			if got := totalDamage(s.g); got != 1 {
				t.Errorf("only the chosen %s takes the 1 damage: total %d", tc.name, got)
			}
		})
	}
}

// Covers: FR-KWD-002h
func TestCard_ChainbreakerSappers_OpposingWardStructureCannotBeChosenButAlliedCan(t *testing.T) {
	s := newScene(t)
	wardE := s.put(1, ZoneMainTheater, structure("Warded Enemy Wall", 0, 5, Keywords{Ward: true}))
	wall := s.put(1, ZoneMainTheater, structure("Enemy Wall", 0, 5, Keywords{}))
	wardA := s.put(0, ZoneMainTheater, structure("Warded Ally Wall", 0, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, wall)
	deploySappers(s)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("asked %d times, want 1", len(asked))
	}
	wantOptions(t, asked[0], "Ward removes only the opposing Ward structure (FR-KWD-002h)",
		wall, wardA, s.g.Players[0].HQ, s.g.Players[1].HQ)
	s.wantDamage(wardE, 0)
}

// Covers: FR-ZON-002
func TestCard_ChainbreakerSappers_LethalDamageDestroysAOneHealthStructure(t *testing.T) {
	s := newScene(t)
	w := s.put(1, ZoneMainTheater, structure("Brittle Wall", 0, 1, Keywords{}))
	inst := w.Inst
	s.scriptRefs(0, nil, w)
	deploySappers(s)
	if s.alive(w) || !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("a structure taking lethal damage from the trigger is destroyed into its owner's Discard")
	}
}

// Covers: FR-TGT-001
func TestCard_ChainbreakerSappers_DamagesHeadquartersWithoutEndingTheMatch(t *testing.T) {
	s := newScene(t)
	hq1 := s.g.Players[1].HQ
	s.scriptRefs(0, nil, hq1)
	deploySappers(s)
	s.wantDamage(hq1, 1)
	if s.g.Outcome.Over {
		t.Errorf("1 damage to a 20-health Headquarters does not end the match")
	}
}

// Covers: FR-LAY-002, FR-CBT-003
func TestCard_ChainbreakerSappers_DamageToOpposingHeadquartersToZeroWinsTheMatch(t *testing.T) {
	s := newScene(t)
	s.g.Players[1].HQ.Damage = 19
	s.scriptRefs(0, nil, s.g.Players[1].HQ)
	deploySappers(s)
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("an opposing HQ at 0 Current Health loses the match: outcome %+v", s.g.Outcome)
	}
}

// With no structures on the board the two Headquarters are the only options.
// Covers: FR-MOD-001
func TestCard_ChainbreakerSappers_TimeoutDamagesExactlyOneLegalTarget(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	deploySappers(s)
	h0, h1 := s.g.Players[0].HQ, s.g.Players[1].HQ
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("two Headquarters are two options: asked %d times, want 1", n)
	}
	if (h0.Damage == 1) == (h1.Damage == 1) || h0.Damage+h1.Damage != 1 {
		t.Errorf("a timeout picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16): HQ damage %d and %d, want exactly one with 1", h0.Damage, h1.Damage)
	}
}

// With the opposing Headquarters unable to be chosen (Ward) and no structures,
// the allied Headquarters is the sole legal target.
// Covers: FR-MOD-001, FR-KWD-002h
func TestCard_ChainbreakerSappers_SoleLegalTargetIsNotAsked(t *testing.T) {
	s := newScene(t)
	s.g.Players[1].HQ.Inst.Card.Keywords.Ward = true
	s.script(0)
	deploySappers(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option is chosen without a question (FR-MOD-001 sec. 1): asked %d times", n)
	}
	s.wantDamage(s.g.Players[0].HQ, 1)
	s.wantDamage(s.g.Players[1].HQ, 0)
}

// colossusCycle marches a freshly deployed Reserves unit to the Main Theater
// and passes a full round, so it is Readied there on its controller's next
// turn. Deploying is not moving (FR-ZON-002), so the march is a separate action.
func colossusCycle(s *scene, a *Asset) {
	s.t.Helper()
	s.do(Action{Kind: ActMarch, Player: a.Controller, Asset: a.ID})
	if a.Zone != ZoneMainTheater {
		s.t.Fatalf("precondition: %s marched to the Main Theater, zone %s", a, a.Zone)
	}
	s.endTurn()
	s.endTurn()
}

// --- Devilhorn Guerrillas ---
// "When this unit deploys, if it was deployed to the Main Theater, deal 2
// damage to chosen opposing unit."  Unit, 4 Requisition, Main Theater or
// Reserves, 3/5.

// deployGuerrillas gives player 0 the real card with 4 Requisition and declares
// its deploy into zone z. It returns the card; find the asset with assetOf.
func deployGuerrillas(s *scene, z Zone) *Card {
	s.t.Helper()
	c := s.card("Devilhorn Guerrillas")
	inst := s.give(0, c)
	s.g.Players[0].Req = 4
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: z})
	return c
}

// Control: the unit is a 3/5 that enters the zone it was deployed to.
func TestCard_DevilhornGuerrillas_EntersDeclaredZoneAsThreeFive(t *testing.T) {
	for _, z := range []Zone{ZoneMainTheater, ZoneReserves} {
		s := newScene(t)
		s.put(1, ZoneMainTheater, unit("Enemy", 1, 9, Keywords{}))
		s.script(0)
		c := deployGuerrillas(s, z)
		a := assetOf(s.g, c)
		if a == nil || a.Zone != z {
			t.Fatalf("Guerrillas deployed to %s are on the battlefield in %s", z, z)
		}
		if p := a.Printed(); p.Power != 3 || p.Health != 5 {
			t.Errorf("Guerrillas are a 3/5: got %d/%d", p.Power, p.Health)
		}
		if s.g.Players[0].Req != 0 {
			t.Errorf("Guerrillas cost 4 Requisition: %d left", s.g.Players[0].Req)
		}
	}
}

// Covers: FR-TGT-001
func TestCard_DevilhornGuerrillas_MainTheaterOffersOnlyOpposingUnitsAndDeals2(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	e2 := s.put(1, ZoneReserves, unit("Reserve Enemy", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e2)
	deployGuerrillas(s, ZoneMainTheater)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one opposing-unit question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "chosen opposing unit: opposing Units only, Scratch Companies included, no Structures or Headquarters, no allied units", b.u1, b.c1, e2)
	s.wantDamage(e2, 2)
	s.wantDamage(b.u1, 0)
	s.wantDamage(b.c1, 0)
}

// The text has no "non-Scratch", so a Scratch Company is a legal target.
func TestCard_DevilhornGuerrillas_ScratchCompanyIsALegalTargetAndDies(t *testing.T) {
	s := newScene(t)
	sc := s.putScratch(1)
	inst := sc.Inst
	s.script(0)
	deployGuerrillas(s, ZoneMainTheater)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a Scratch Company is the sole legal target and is selected automatically: asked %d times", n)
	}
	if s.alive(sc) || !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("2 damage destroys the 1/1 Scratch Company into its owner's Discard (FR-ZON-002)")
	}
}

// Covers: FR-TRN-003
func TestCard_DevilhornGuerrillas_ReservesAsksNothingAndDamagesNothing(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	s.script(0)
	c := deployGuerrillas(s, ZoneReserves)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("deployed to Reserves the gate fails and no trigger is placed, so no target is asked: asked %d times (FR-TRN-003)", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("deployed to Reserves, nothing happens: total damage %d, want 0", got)
	}
	s.wantDamage(e1, 0)
	if a := assetOf(s.g, c); a == nil || a.Zone != ZoneReserves {
		t.Errorf("the unit still deploys to Reserves")
	}
}

// Covers: FR-TRN-003
func TestCard_DevilhornGuerrillas_ReservesWithSoleOpposingUnitDoesNothing(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Only Enemy", 1, 5, Keywords{}))
	s.script(0)
	deployGuerrillas(s, ZoneReserves)
	s.wantDamage(e, 0)
}

// Covers: FR-TGT-003
func TestCard_DevilhornGuerrillas_NoOpposingUnitStillDeploys(t *testing.T) {
	s := newScene(t)
	s.put(1, ZoneMainTheater, structure("Enemy Structure", 0, 5, Keywords{}))
	s.script(0)
	c := deployGuerrillas(s, ZoneMainTheater)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target: no question, asked %d times", n)
	}
	if a := assetOf(s.g, c); a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the unit still deploys to the Main Theater when the trigger has no legal target (FR-TGT-003)")
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("an opposing Structure and HQ are not units: total damage %d, want 0", got)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty: %d items", len(s.g.Stack))
	}
}

// Covers: FR-KWD-002h, FR-TGT-001
func TestCard_DevilhornGuerrillas_OpposingWardUnitCannotBeChosen(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, e1)
	deployGuerrillas(s, ZoneMainTheater)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "an opposing Ward unit cannot be chosen by a triggered ability (FR-KWD-002h)", e1, e2)
	s.wantDamage(ward, 0)
	s.wantDamage(e1, 2)
}

// Covers: FR-KWD-002h, FR-TGT-003
func TestCard_DevilhornGuerrillas_OnlyOpposingUnitHasWardSoNothingHappens(t *testing.T) {
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.script(0)
	c := deployGuerrillas(s, ZoneMainTheater)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target: no question, asked %d times", n)
	}
	s.wantDamage(ward, 0)
	if assetOf(s.g, c) == nil {
		t.Errorf("the unit still deploys (FR-TGT-003)")
	}
}

// Covers: FR-MOD-001
func TestCard_DevilhornGuerrillas_SoleOpposingTargetIsNotAsked(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Only Enemy", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, structure("Enemy Structure", 0, 5, Keywords{}))
	s.script(0)
	deployGuerrillas(s, ZoneMainTheater)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal target is selected automatically (FR-MOD-001 sec. 1): asked %d times", n)
	}
	s.wantDamage(e, 2)
}

// Covers: FR-MOD-001
func TestCard_DevilhornGuerrillas_TimeoutDamagesExactlyOneLegalUnit(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	st := s.put(1, ZoneMainTheater, structure("Enemy Structure", 0, 5, Keywords{}))
	s.script(0, Answer{Timeout: true})
	c := deployGuerrillas(s, ZoneMainTheater)
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("asked %d times, want 1", n)
	}
	if (e1.Damage == 2) == (e2.Damage == 2) || e1.Damage+e2.Damage != 2 {
		t.Errorf("a timeout picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16): damage A=%d B=%d, want exactly one with 2", e1.Damage, e2.Damage)
	}
	s.wantDamage(st, 0)
	s.wantDamage(assetOf(s.g, c), 0)
}

// Covers: FR-ZON-002
func TestCard_DevilhornGuerrillas_LethalDamageDestroysTheUnit(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Fragile Enemy", 1, 2, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Other Enemy", 1, 5, Keywords{}))
	inst := e.Inst
	s.scriptRefs(0, nil, e)
	deployGuerrillas(s, ZoneMainTheater)
	if s.alive(e) || !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("the 1/2 opposing unit takes 2 and is destroyed into its owner's Discard")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty after resolution: %d items", len(s.g.Stack))
	}
}

// --- Ironclad Colossus ---
// "Whenever this unit deploys or attacks, chosen unit gains 2 armor."  Unit,
// 6 Requisition, Reserves, 6/7.

func TestCard_IroncladColossus_EntersReservesAsSixSeven(t *testing.T) {
	s := newScene(t)
	s.script(0)
	c := deployPath(s, "Ironclad Colossus", 6)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Ironclad Colossus is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 6 || p.Health != 7 {
		t.Errorf("Ironclad Colossus is a 6/7: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Ironclad Colossus costs 6 Requisition: %d left", s.g.Players[0].Req)
	}
}

// Covers: FR-TGT-001, FR-KWD-002h
func TestCard_IroncladColossus_DeployOffersAnyUnitButNotStructuresHeadquartersOrOpposingWard(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.scriptRefs(0, nil, b.u0)
	c := deployPath(s, "Ironclad Colossus", 6)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one question: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "chosen unit: every Unit of either side (Scratch Companies and the Colossus itself included), no Structures, Headquarters or opposing Ward units", b.u0, b.c0, b.u1, b.c1, assetOf(s.g, c))
	if b.u0.Armor != 2 {
		t.Errorf("the chosen unit gains 2 Armor: got %d", b.u0.Armor)
	}
}

// Covers: FR-KWD-002h
func TestCard_IroncladColossus_AlliedWardUnitIsAnOption(t *testing.T) {
	s := newScene(t)
	ward := s.put(0, ZoneMainTheater, unit("Warded Ally", 1, 5, Keywords{Ward: true}))
	s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, ward)
	deployPath(s, "Ironclad Colossus", 6)
	if ward.Armor != 2 {
		t.Errorf("Ward restricts only opponents, so the controller's Ward unit can be chosen (FR-KWD-002h): Armor %d, want 2", ward.Armor)
	}
}

func TestCard_IroncladColossus_CanArmorAnOpposingUnit(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 5, Keywords{}))
	s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, e)
	deployPath(s, "Ironclad Colossus", 6)
	if e.Armor != 2 {
		t.Errorf("the chosen unit may be opposing: Armor %d, want 2", e.Armor)
	}
}

// Covers: FR-MOD-001
func TestCard_IroncladColossus_OnlyItselfOnTheBoardIsGivenArmorWithoutAsking(t *testing.T) {
	s := newScene(t)
	s.script(0)
	c := deployPath(s, "Ironclad Colossus", 6)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the Colossus is the sole legal target and is selected automatically (FR-MOD-001 sec. 1): asked %d times", n)
	}
	if a := assetOf(s.g, c); a == nil || a.Armor != 2 {
		t.Errorf("the Colossus gains 2 Armor")
	}
}

// Covers: FR-KWD-002h, FR-TGT-003
func TestCard_IroncladColossus_OpposingWardIsTheOnlyOtherUnitSoOnlyItselfIsChosen(t *testing.T) {
	s := newScene(t)
	ward := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.script(0)
	c := deployPath(s, "Ironclad Colossus", 6)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("asked %d times, want 0", n)
	}
	if ward.Armor != 0 {
		t.Errorf("an opposing Ward unit cannot be chosen: Armor %d", ward.Armor)
	}
	if a := assetOf(s.g, c); a == nil || a.Armor != 2 {
		t.Errorf("the Colossus gains 2 Armor")
	}
}

// Covers: FR-KWD-002k, FR-LAY-002
func TestCard_IroncladColossus_GrantAddsToPrintedArmor(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Armored Ally", 1, 5, Keywords{Armor: 1}))
	s.scriptRefs(0, nil, ally)
	deployPath(s, "Ironclad Colossus", 6)
	if ally.Armor != 3 {
		t.Errorf("printed Armor 1 plus a grant of 2 is one pool of 3: got %d (FR-KWD-002k)", ally.Armor)
	}
}

// Covers: FR-LAY-002
func TestCard_IroncladColossus_GrantPersistsPastEndOfTurn(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, ally)
	deployPath(s, "Ironclad Colossus", 6)
	s.endTurn()
	s.endTurn()
	if ally.Armor != 2 {
		t.Errorf("granted Armor never expires at end of turn: got %d after a full round, want 2 (FR-KWD-002k)", ally.Armor)
	}
}

// Covers: FR-CBT-002
func TestCard_IroncladColossus_EachAttackGrantsAgainAndGrantsStack(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 20, Keywords{}))
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 30, Keywords{}))
	s.scriptRefs(0, nil, ally, ally)
	c := deployPath(s, "Ironclad Colossus", 6)
	a := assetOf(s.g, c)
	colossusCycle(s, a)
	if ally.Armor != 2 {
		t.Fatalf("precondition: the deploy grant landed: Armor %d", ally.Armor)
	}
	s.do(s.attack(a, dummy))
	if ally.Armor != 4 {
		t.Errorf("attacking grants 2 more Armor, stacking with the deploy grant: got %d, want 4", ally.Armor)
	}
	if n := len(s.asked(0)); n != 2 {
		t.Errorf("one question for the deploy and one for the attack: asked %d times, want 2", n)
	}
}

// Covers: FR-CBT-002, FR-REP-004
func TestCard_IroncladColossus_AttackGrantResolvesBeforeCombatDamage(t *testing.T) {
	// The Colossus chooses itself at Step 2 and gains 2 Armor. Retaliation of 3
	// then loses 2 to Armor 1-for-1 and 1 reaches health. Without a Step 2
	// grant, 3 would reach health.
	s := newScene(t)
	s.script(0)
	c := deployPath(s, "Ironclad Colossus", 6)
	a := assetOf(s.g, c)
	colossusCycle(s, a)
	a.Armor = 0
	def := s.put(1, ZoneMainTheater, unit("Defender", 3, 30, Keywords{}))
	s.scriptRefs(0, nil, a)
	s.do(s.attack(a, def))
	s.wantDamage(a, 1)
	if a.Armor != 0 {
		t.Errorf("the 2 granted Armor was consumed by retaliation: Armor %d, want 0 (FR-REP-004 Stage 4)", a.Armor)
	}
}

// Covers: FR-CBT-002
func TestCard_IroncladColossus_DefendingDoesNotTrigger(t *testing.T) {
	s := newScene(t)
	s.script(0)
	c := deployPath(s, "Ironclad Colossus", 6)
	a := assetOf(s.g, c)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	s.endTurn()
	a.Armor = 0
	hitter := s.put(1, ZoneMainTheater, unit("Hitter", 1, 30, Keywords{}))
	s.do(s.attack(hitter, a))
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only attacking triggers the Colossus: asked %d times while defending", n)
	}
	if a.Armor != 0 {
		t.Errorf("no Armor is granted when defending: got %d", a.Armor)
	}
	s.wantDamage(a, 1)
}

// --- Necropolis Hunt ---
// "When this unit is killed, deal damage equal to its power to chosen
// non-Headquarters asset."  Unit, 3 Requisition and 3 Tactics, Main Theater
// only, 5/2.

// deployHunt gives player 0 the real card with 3 Requisition and 3 Tactics and
// declares its deploy into the Main Theater. It returns the asset.
func deployHunt(s *scene) *Asset {
	s.t.Helper()
	c := s.card("Necropolis Hunt")
	inst := s.give(0, c)
	s.g.Players[0].Req = 3
	s.g.Players[0].Tac = 3
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	a := assetOf(s.g, c)
	if a == nil {
		s.t.Fatalf("precondition: the Hunt is on the battlefield after deploying")
	}
	return a
}

// huntAttacks deploys the Hunt, readies it over a full round and has it attack
// a fresh 2/5 opposing Defender, so both die. Player 0 answers its death
// question by choosing want in order. It returns the Hunt and the Defender.
func huntAttacks(s *scene, want ...*Asset) (hunt, def *Asset) {
	s.t.Helper()
	hunt = deployHunt(s)
	s.endTurn()
	s.endTurn()
	def = s.put(1, ZoneMainTheater, unit("Defender", 2, 5, Keywords{}))
	s.scriptRefs(0, nil, want...)
	s.do(s.attack(hunt, def))
	if s.alive(hunt) {
		s.t.Fatalf("precondition: the Hunt is destroyed by the Defender's retaliation")
	}
	return hunt, def
}

// huntDefends deploys the Hunt, passes the turn and has an opposing 2/20
// Hitter attack it. Player 0 answers its death question by choosing want in
// order; player 1 is given a script that records any question. It returns the
// Hunt and the Hitter.
func huntDefends(s *scene, want ...*Asset) (hunt, hit *Asset) {
	s.t.Helper()
	hunt = deployHunt(s)
	s.endTurn()
	hit = s.put(1, ZoneMainTheater, unit("Hitter", 2, 20, Keywords{}))
	s.scriptRefs(0, nil, want...)
	s.script(1)
	s.do(s.attack(hit, hunt))
	if s.alive(hunt) {
		s.t.Fatalf("precondition: the Hunt is destroyed by the Hitter")
	}
	return hunt, hit
}

func TestCard_NecropolisHunt_EntersMainTheaterAsFiveTwoForThreeRequisitionThreeTactics(t *testing.T) {
	s := newScene(t)
	a := deployHunt(s)
	if a.Zone != ZoneMainTheater {
		t.Errorf("the Hunt deploys to the Main Theater: %s", a.Zone)
	}
	if p := a.Printed(); p.Power != 5 || p.Health != 2 {
		t.Errorf("the Hunt is a 5/2: got %d/%d", p.Power, p.Health)
	}
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Errorf("the Hunt costs 3 Requisition and 3 Tactics: %d Requisition and %d Tactics left", p.Req, p.Tac)
	}
}

// Covers: FR-TGT-001, FR-KWD-002h, FR-ZON-004
func TestCard_NecropolisHunt_KilledAttackingOffersEveryUnitAndStructureButNotHeadquartersItselfOrOpposingWard(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	s.put(1, ZoneMainTheater, unit("Warded Unit", 1, 5, Keywords{Ward: true}))
	s.put(1, ZoneMainTheater, structure("Warded Structure", 0, 5, Keywords{Ward: true}))
	reserve := s.put(1, ZoneReserves, unit("Reserve Enemy", 1, 5, Keywords{}))
	_, def := huntAttacks(s, b.s1)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one death question for the Hunt's controller: asked %d times", len(asked))
	}
	if d := asked[0]; d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("a mandatory single DecideTarget for the Hunt's controller: %+v", d)
	}
	wantOptions(t, asked[0], "any unit or structure of either side, Scratch Companies included; no Headquarters, no dead Hunt, no opposing Ward asset",
		b.u0, b.s0, b.c0, b.u1, b.s1, b.c1, reserve)
	for _, o := range asked[0].Options {
		if o.Ref == def.ID {
			t.Errorf("the Defender died with the Hunt in the same sweep, so it is not on the battlefield when the trigger is placed and is not an option (FR-TGT-001)")
		}
	}
	s.wantDamage(b.s1, 5)
}

// Covers: FR-ZON-004, FR-REP-004
func TestCard_NecropolisHunt_DamageEventNamesTheDeadHuntAsSourceAndIsNonCombat(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	hunt, _ := huntAttacks(s, b.u1)
	var found *Event
	for _, e := range damageEvents(s.g.Log) {
		if e.Target == b.u1.ID {
			e := e
			found = &e
		}
	}
	if found == nil {
		t.Fatalf("a damage event against the chosen unit")
	}
	if found.Source != hunt.ID || found.Raw != 5 || found.Amount != 5 || found.Combat {
		t.Errorf("5 non-combat damage whose Source is the dead Hunt's asset: %+v (Hunt asset %d)", *found, hunt.ID)
	}
	s.wantDamage(b.u1, 5)
}

// Covers: FR-ZON-004
func TestCard_NecropolisHunt_KilledOnOpponentsTurnStillTriggersAndItsControllerChooses(t *testing.T) {
	s := newScene(t)
	ally := s.put(0, ZoneMainTheater, unit("Ally", 1, 9, Keywords{}))
	e := s.put(1, ZoneMainTheater, unit("Enemy", 1, 9, Keywords{}))
	_, hit := huntDefends(s, e)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("no turn gate: the Hunt's controller is asked once on the opponent's turn: asked %d times", len(asked))
	}
	wantOptions(t, asked[0], "every other unit and structure of either side", ally, e, hit)
	s.wantDamage(e, 5)
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the Hunt's controller chooses, not the killer's: opponent asked %d times", n)
	}
}

// Covers: FR-REP-004
func TestCard_NecropolisHunt_StalwartTargetTakesTheFullFive(t *testing.T) {
	s := newScene(t)
	wall := s.put(1, ZoneMainTheater, unit("Stalwart Wall", 0, 9, Keywords{Stalwart: true}))
	huntDefends(s, wall)
	s.wantDamage(wall, 5)
}

// Covers: FR-REP-004
func TestCard_NecropolisHunt_ResistantTargetReducesTheFive(t *testing.T) {
	s := newScene(t)
	worm := s.put(1, ZoneMainTheater, unit("Resistant Unit", 0, 9, Keywords{Resistant: 2}))
	huntDefends(s, worm)
	s.wantDamage(worm, 3)
}

// Covers: FR-REP-004
func TestCard_NecropolisHunt_ArmorAbsorbsPartOfTheFive(t *testing.T) {
	s := newScene(t)
	armored := s.put(1, ZoneMainTheater, unit("Armored Unit", 0, 9, Keywords{Armor: 2}))
	huntDefends(s, armored)
	s.wantDamage(armored, 3)
	if armored.Armor != 0 {
		t.Errorf("the 2 Armor absorbed 2 of the 5 damage and is spent: Armor %d, want 0", armored.Armor)
	}
}

// Covers: FR-ZON-002
func TestCard_NecropolisHunt_FiveDamageDestroysAFiveHealthUnit(t *testing.T) {
	s := newScene(t)
	frail := s.put(1, ZoneMainTheater, unit("Frail", 0, 5, Keywords{}))
	inst := frail.Inst
	huntDefends(s, frail)
	if s.alive(frail) || !inZone(s.g.Players[1].Discard, inst) {
		t.Errorf("5 damage destroys a 5-health unit into its owner's Discard")
	}
}

// Covers: FR-TGT-003, FR-MOD-001
func TestCard_NecropolisHunt_NoOtherUnitOrStructureAsksNothingAndDoesNothing(t *testing.T) {
	s := newScene(t)
	huntAttacks(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("no legal target: no question, asked %d times", n)
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no legal target: nothing is damaged, total damage %d", got)
	}
	if s.g.Outcome.Over {
		t.Errorf("Headquarters are not legal targets and the match goes on: %+v", s.g.Outcome)
	}
}

// Covers: FR-MOD-001
func TestCard_NecropolisHunt_SoleLegalTargetIsNotAsked(t *testing.T) {
	s := newScene(t)
	wall := s.put(1, ZoneMainTheater, structure("Only Wall", 0, 9, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Warded Unit", 1, 5, Keywords{Ward: true}))
	huntAttacks(s)
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a single legal option is chosen without a question (FR-MOD-001 sec. 1): asked %d times", n)
	}
	s.wantDamage(wall, 5)
}

// Covers: FR-MOD-001
func TestCard_NecropolisHunt_TimeoutDamagesExactlyOneLegalTarget(t *testing.T) {
	s := newScene(t)
	u := s.put(1, ZoneMainTheater, unit("Enemy A", 0, 9, Keywords{}))
	w := s.put(1, ZoneMainTheater, structure("Enemy Wall", 0, 9, Keywords{}))
	hunt := deployHunt(s)
	s.endTurn()
	s.endTurn()
	def := s.put(1, ZoneMainTheater, unit("Defender", 2, 5, Keywords{}))
	s.script(0, Answer{Timeout: true})
	s.do(s.attack(hunt, def))
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("one death question: asked %d times", n)
	}
	if (u.Damage == 5) == (w.Damage == 5) || u.Damage+w.Damage != 5 {
		t.Errorf("a timeout picks the first legal target in fixed order (FR-MOD-001 sec. 6, SVY-R16): damage %d and %d, want exactly one with 5", u.Damage, w.Damage)
	}
}

// Covers: FR-ZON-004
func TestCard_NecropolisHunt_FaceDownCopyFiresNothingWhenKilled(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, s.card("Necropolis Hunt"))
	e := s.put(1, ZoneMainTheater, unit("Enemy A", 0, 9, Keywords{}))
	e2 := s.put(1, ZoneMainTheater, unit("Enemy B", 0, 9, Keywords{}))
	s.script(0)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if !scratched(p, top[0]) {
		t.Fatalf("precondition: the Hunt on top of the deck became a face-down Scratch Company")
	}
	var sc *Asset
	for _, a := range p.Reserves {
		if a.Inst == top[0] {
			sc = a
		}
	}
	// Ready the Scratch Company over a round, march it to the Main Theater (it
	// cannot be attacked in Reserves), then have the opponent kill it.
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sc.ID})
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(hit, sc))
	if s.alive(sc) {
		t.Fatalf("precondition: the face-down Scratch Company is destroyed")
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a face-down Scratch Company has no printed abilities, so no death trigger: asked %d times (FR-ZON-004)", n)
	}
	s.wantDamage(e, 0)
	s.wantDamage(e2, 0)
	s.wantDamage(hit, 1) // only the Scratch Company's own 1 retaliation; no death damage
}

func TestCard_NecropolisHunt_NoEventDetailHasAFormatError(t *testing.T) {
	s := newScene(t)
	b := s.targetBoard()
	huntAttacks(s, b.u1)
	for _, e := range s.g.Log {
		if strings.Contains(e.Detail, "%!") {
			t.Errorf("event %d (%s) Detail has a formatting error: %q", e.Seq, e.Kind, e.Detail)
		}
	}
}

// SVY-R10: "its power" of a source that has left is its last-known effective
// power, so a temporary bonus the Hunt had when it died still counts.
// Covers: FR-ZON-004
func TestCard_NecropolisHunt_DealsItsLastKnownEffectivePowerNotAFixedFive(t *testing.T) {
	s := newScene(t)
	e := s.put(1, ZoneMainTheater, unit("Enemy", 0, 20, Keywords{}))
	hunt := deployHunt(s)
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 2, 20, Keywords{}))
	hunt.TempPower = 2 // stands in for an until-end-of-turn bonus, still active this turn
	s.scriptRefs(0, nil, e)
	s.script(1)
	s.do(s.attack(hit, hunt))
	if s.alive(hunt) {
		t.Fatalf("precondition: the Hunt is destroyed by the Hitter")
	}
	s.wantDamage(e, 7)
}

// --- Hammerhand Brigands ---
// "Unbowed: Scratch 1."  Unit, 1 Requisition, Reserves, 2/1. Unbowed fires when
// the asset takes damage (FR-KWD-002a, FR-REP-004 Stage 5): Armor absorbed plus
// Health committed is more than 0. Its Scratch does not act on the dead asset,
// so a destroyed Brigands still resolves it (FR-KWD-002a second bullet).

// putBrigands places the real card in zone z for player p with the given
// current Armor (0 for none).
func (s *scene) putBrigands(p PlayerID, z Zone, armor int) *Asset {
	s.t.Helper()
	a := s.put(p, z, s.card("Hammerhand Brigands"))
	a.Armor = armor
	return a
}

// Control: a 2/1 that enters Reserves for 1 Requisition and does nothing on deploy.
func TestCard_HammerhandBrigands_EntersReservesAsTwoOneForOneRequisitionAndDoesNothing(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1)
	c := deployPath(s, "Hammerhand Brigands", 1)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneReserves {
		t.Fatalf("Hammerhand Brigands is on the battlefield in Reserves after deploying")
	}
	if p := a.Printed(); p.Power != 2 || p.Health != 1 {
		t.Errorf("Hammerhand Brigands is a 2/1: got %d/%d", p.Power, p.Health)
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("Hammerhand Brigands costs 1 Requisition: %d left", s.g.Players[0].Req)
	}
	s.wantScratches(0, 0, "Unbowed is not a deploy trigger: nothing happens on deploy")
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("no question is ever asked: asked %d times", n)
	}
}

// Covers: FR-ZON-004
func TestCard_HammerhandBrigands_LethalCombatDamageAsDefenderScratchesFromDiscard(t *testing.T) {
	s := newScene(t)
	s.script(0)
	s.script(1)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.putBrigands(1, ZoneMainTheater, 0)
	inst := def.Inst
	s.do(s.attack(att, def))
	if s.alive(def) || !inZone(s.g.Players[1].Discard, inst) {
		t.Fatalf("precondition: the 2/1 Brigands is destroyed into its owner's Discard")
	}
	s.wantScratches(1, 1, "lethal damage still fires Unbowed and Scratch 1 resolves in full from the discard pile (FR-KWD-002a)")
	s.wantScratches(0, 0, "the Scratch goes to the Brigands' controller, not the opponent (FR-TRG-001)")
	if n := len(s.asked(0)) + len(s.asked(1)); n != 0 {
		t.Errorf("no question is ever asked: asked %d times", n)
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once the trigger resolves: %d items", len(s.g.Stack))
	}
	if err := s.g.CheckInvariants(); err != nil {
		t.Errorf("state after the Scratch: %v", err)
	}
}

// Covers: FR-ZON-004, FR-CBT-002
func TestCard_HammerhandBrigands_LethalRetaliationAsAttackerScratchesFromDiscard(t *testing.T) {
	s := newScene(t)
	att := s.putBrigands(0, ZoneMainTheater, 0)
	def := s.put(1, ZoneMainTheater, unit("Wall", 2, 9, Keywords{}))
	inst := att.Inst
	s.do(s.attack(att, def))
	if s.alive(att) || !inZone(s.g.Players[0].Discard, inst) {
		t.Fatalf("precondition: the attacking Brigands takes 2 retaliation on 1 Health and is destroyed")
	}
	s.wantScratches(0, 1, "retaliation is damage taken: Unbowed fires for the attacker and Scratches 1 (FR-REP-004 Stage 5)")
	s.wantScratches(1, 0, "the defender's controller does not Scratch")
}

// Covers: FR-REP-004
func TestCard_HammerhandBrigands_ArmorAbsorbingAllDamageStillScratches(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.putBrigands(1, ZoneMainTheater, 3)
	s.do(s.attack(att, def))
	if !s.alive(def) {
		t.Fatalf("precondition: Armor 3 absorbs all 2 damage and the Brigands survives")
	}
	s.wantDamage(def, 0)
	s.wantScratches(1, 1, "damage absorbed entirely by Armor still counts as taken (FR-REP-004 Stage 5)")
}

// Covers: FR-TGT-001, FR-REP-004
func TestCard_HammerhandBrigands_SaboteurDamageToAlliedBrigandsInReservesScratches(t *testing.T) {
	s := newScene(t)
	e1 := s.put(1, ZoneMainTheater, unit("Enemy A", 1, 5, Keywords{}))
	s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	b := s.putBrigands(0, ZoneReserves, 0)
	inst := b.Inst
	s.scriptRefs(0, nil, e1, b)
	deploySaboteur(s)
	if s.alive(b) || !inZone(s.g.Players[0].Discard, inst) {
		t.Fatalf("precondition: Saboteur's 1 damage destroys the allied 2/1 Brigands in Reserves")
	}
	s.wantScratches(0, 1, "non-combat damage to a Brigands in Reserves fires Unbowed for its controller (FR-KWD-002a)")
	s.wantScratches(1, 0, "the opponent does not Scratch")
}

// Covers: FR-TGT-001, FR-REP-004
func TestCard_HammerhandBrigands_OpposingGuerrillasDamageScratchesForTheBrigandsController(t *testing.T) {
	s := newScene(t)
	b := s.putBrigands(1, ZoneMainTheater, 0)
	s.put(1, ZoneMainTheater, unit("Enemy B", 1, 5, Keywords{}))
	s.scriptRefs(0, nil, b)
	deployGuerrillas(s, ZoneMainTheater)
	if s.alive(b) {
		t.Fatalf("precondition: 2 damage destroys the 2/1 Brigands")
	}
	s.wantScratches(1, 1, "the opponent's deploy damage fires the Brigands' Unbowed for its controller (FR-KWD-002a)")
	s.wantScratches(0, 0, "the deploying player does not Scratch")
}

// Covers: FR-REP-004, FR-TRG-004
func TestCard_HammerhandBrigands_TwoBrigandsHitByOneAllAssetsEffectScratchTwice(t *testing.T) {
	s := newScene(t)
	s.putBrigands(1, ZoneMainTheater, 0)
	s.putBrigands(1, ZoneReserves, 0)
	s.deployFromHand(0, trigUnit("Cataclysm", targetTrigger(nil, allDamage(1, ability.Filter{}))))
	s.wantScratches(1, 2, "each Brigands took damage in the one effect, so each fires its own Unbowed: 2 Scratches (FR-KWD-002a)")
	s.wantScratches(0, 0, "player 0 has no Unbowed asset")
	if len(s.g.Stack) != 0 {
		t.Errorf("the stack is empty once both triggers resolve: %d items", len(s.g.Stack))
	}
}

// Covers: FR-REP-004
func TestCard_HammerhandBrigands_OneBrigandsHitByTwoSeparateInstancesScratchesTwice(t *testing.T) {
	s := newScene(t)
	a1 := s.put(0, ZoneMainTheater, unit("Hitter One", 1, 9, Keywords{}))
	a2 := s.put(0, ZoneMainTheater, unit("Hitter Two", 1, 9, Keywords{}))
	def := s.putBrigands(1, ZoneMainTheater, 5)
	s.do(s.attack(a1, def))
	s.wantScratches(1, 1, "the first damage instance fires Unbowed once")
	s.do(s.attack(a2, def))
	if !s.alive(def) {
		t.Fatalf("precondition: Armor keeps the Brigands alive through both hits")
	}
	s.wantScratches(1, 2, "a second, separate damage instance fires Unbowed again (FR-KWD-002a)")
}

// Covers: FR-ZON-004
func TestCard_HammerhandBrigands_ScratchGoesToTheControllerNotTheOwner(t *testing.T) {
	s := newScene(t)
	s.endTurn() // player 1 is active
	att := s.put(1, ZoneMainTheater, unit("Hitter", 1, 9, Keywords{}))
	// Owned by player 1, controlled by player 0, with Armor so it survives.
	def := s.putControlled(1, 0, s.card("Hammerhand Brigands"))
	def.Armor = 2
	s.do(s.attack(att, def))
	if !s.alive(def) {
		t.Fatalf("precondition: Armor 2 absorbs the 1 damage and the captured Brigands survives")
	}
	s.wantScratches(0, 1, "the trigger belongs to the asset's controller at the time of the damage, player 0 (FR-ZON-004)")
	s.wantScratches(1, 0, "the card's owner is not the trigger's controller")
}

// Covers: FR-REP-004
func TestCard_HammerhandBrigands_ZeroPowerAttackerDoesNotScratch(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Pacifist", 0, 9, Keywords{}))
	def := s.putBrigands(1, ZoneMainTheater, 0)
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	s.wantScratches(1, 0, "0 damage commits and absorbs nothing, so nothing is taken (FR-REP-004 Stage 5)")
}

// Covers: FR-REP-004, FR-KWD-002b
func TestCard_HammerhandBrigands_DamageReducedToZeroByResistantDoesNotScratch(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	c := s.card("Hammerhand Brigands")
	c.Keywords.Resistant = 2
	def := s.put(1, ZoneMainTheater, c)
	s.do(s.attack(att, def))
	s.wantDamage(def, 0)
	s.wantScratches(1, 0, "2 damage less Resistant 2 is 0 at Stage 2, so no damage is taken (FR-REP-004 Stage 5)")
}

func TestCard_HammerhandBrigands_DamageToAnotherAssetDoesNotScratch(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Plain", 1, 5, Keywords{}))
	s.putBrigands(1, ZoneMainTheater, 0)
	s.do(s.attack(att, def))
	s.wantScratches(1, 0, "Unbowed fires for damage to the Brigands itself only (FR-KWD-002a)")
}

// Covers: FR-ZON-004
func TestCard_HammerhandBrigands_FaceDownCopyFiresNothingWhenDamaged(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := s.onTop(0, s.card("Hammerhand Brigands"))
	s.script(0)
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	var sc *Asset
	for _, a := range p.Reserves {
		if a.Inst == top[0] {
			sc = a
		}
	}
	if sc == nil || !sc.FaceDown {
		t.Fatalf("precondition: the Brigands on top of the deck became a face-down Scratch Company")
	}
	// Ready it over a round, march it to the Main Theater, then have the opponent kill it.
	s.endTurn()
	s.endTurn()
	s.do(Action{Kind: ActMarch, Player: 0, Asset: sc.ID})
	s.endTurn()
	hit := s.put(1, ZoneMainTheater, unit("Hitter", 9, 9, Keywords{}))
	s.do(s.attack(hit, sc))
	if s.alive(sc) {
		t.Fatalf("precondition: the face-down Scratch Company is destroyed")
	}
	s.wantScratches(0, 0, "a face-down Scratch Company has no printed abilities, so Unbowed does not fire (FR-ZON-004)")
}
