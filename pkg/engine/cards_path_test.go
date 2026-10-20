package engine

import (
	"reflect"
	"testing"
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
		t.Errorf("a timeout picks one random legal target (FR-MOD-001 sec. 6): damage A=%d B=%d, want exactly one with 2", e1.Damage, e2.Damage)
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
		t.Errorf("a timeout picks one random legal target (FR-MOD-001 sec. 6): HQ damage %d and %d, want exactly one with 1", h0.Damage, h1.Damage)
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
