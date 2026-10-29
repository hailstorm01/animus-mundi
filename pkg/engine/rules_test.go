package engine

import "testing"

// ---- Domain 1: setup, decks, empty-deck loss ----

func TestFR_SET_004_FirstPlayerSkipsTurnOneDraw(t *testing.T) {
	s := newScene(t)
	if n := len(s.g.Players[0].Hand); n != 6 {
		t.Fatalf("first player hand on Turn 1 = %d, want 6", n)
	}
	s.endTurn()
	if n := len(s.g.Players[1].Hand); n != 7 {
		t.Fatalf("second player hand on Turn 2 = %d, want 7", n)
	}
	s.endTurn()
	if n := len(s.g.Players[0].Hand); n != 7 {
		t.Fatalf("first player hand on Turn 3 = %d, want 7", n)
	}
}

func TestFR_SET_002_PartialMulliganKeepsSixCards(t *testing.T) {
	var asideIDs []int
	g, err := NewGame(Config{
		Seed:      7,
		Decks:     [2]DeckList{{HQ: hq(3, 20), Main: filler(20)}, {HQ: hq(3, 20), Main: filler(20)}},
		PlayFirst: func(d PlayerID) bool { return d == 0 },
		Mulligan: func(p PlayerID, hand []*Instance) []int {
			if p != 0 {
				return nil
			}
			for _, inst := range hand[:3] {
				asideIDs = append(asideIDs, inst.ID)
			}
			return asideIDs
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	p := g.Players[0]
	if len(p.Hand) != 6 || len(p.Deck) != 14 {
		t.Fatalf("after mulligan hand=%d deck=%d, want 6 and 14", len(p.Hand), len(p.Deck))
	}
	inDeck := map[int]bool{}
	for _, inst := range p.Deck {
		inDeck[inst.ID] = true
	}
	for _, id := range asideIDs {
		if !inDeck[id] {
			t.Errorf("set-aside card %d was not shuffled back into the deck", id)
		}
	}
}

func TestFR_WIN_001_DrawingFromEmptyDeckLoses(t *testing.T) {
	s := newScene(t)
	s.g.Players[1].Deck = nil
	s.endTurn()
	o := s.g.Outcome
	if !o.Over || o.Draw || o.Winner != 0 {
		t.Fatalf("outcome = %+v, want player 0 wins", o)
	}
}

// ---- Domain 2: zones and Scratch Company ----

func TestFR_ZON_005_ScratchCompanyFromEmptyDeckCreatesNothing(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Deck = nil
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if len(p.Reserves) != 0 || s.g.Outcome.Over {
		t.Fatalf("reserves=%d outcome=%+v, want no unit and no loss", len(p.Reserves), s.g.Outcome)
	}
}

// Covers: FR-ZON-005, FR-KWD-002m
func TestFR_CAP_001_HeadquartersDeploysFaceDownScratchCompany(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	top := p.Deck[0]
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	if !p.HQ.Exhausted {
		t.Error("HQ should exhaust to Deploy Scratch Company")
	}
	if len(p.Reserves) != 1 {
		t.Fatalf("reserves = %d, want 1", len(p.Reserves))
	}
	sc := p.Reserves[0]
	if !sc.FaceDown || sc.Inst != top || s.g.Power(sc) != 1 || s.g.MaxHealth(sc) != 1 || !sc.Printed().HasSubtype("Infantry") {
		t.Fatalf("scratch company = %+v, want face-down 1/1 Infantry from the top of the deck", sc)
	}
}

// Ruling KIM-S01 (2026-10-19).
func TestFR_ZON_005_ScratchCompanyEntersExhausted(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	s.do(Action{Kind: ActScratch, Player: 0, Asset: p.HQ.ID})
	sc := p.Reserves[0]
	if !sc.Exhausted {
		t.Fatal("new Scratch Company entered Readied")
	}
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: sc.ID, Resource: Requisition})
	s.endTurn()
	s.endTurn()
	if sc.Exhausted {
		t.Fatal("Scratch Company should ready in its controller's next Ready Phase")
	}
}

// FR-ZON-005, FR-KWD-002l, KIM-S01 (revised), directive #48.
func TestFR_ZON_005_ScratchFacePrintsDeploysExhausted(t *testing.T) {
	if got, want := ScratchFace().Keywords, (Keywords{DeploysExhausted: true}); got != want {
		t.Errorf("Scratch Company face keywords = %+v, want exactly Deploys Exhausted (FR-ZON-005)", got)
	}
	fd := &Asset{FaceDown: true}
	if !fd.Printed().Keywords.DeploysExhausted {
		t.Errorf("a face-down asset's Printed() lacks Deploys Exhausted, want the Scratch Company face (FR-ZON-005, FR-KWD-002l)")
	}
}

func TestFR_ZON_004_MarchPreservesDamageAndArmor(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 2, 5, Keywords{Armor: 2}))
	u.Damage = 3
	s.do(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	if u.Zone != ZoneMainTheater || !u.Exhausted {
		t.Fatalf("zone=%s exhausted=%v, want Main Theater and exhausted", u.Zone, u.Exhausted)
	}
	if u.Damage != 3 || u.Armor != 2 {
		t.Fatalf("damage=%d armor=%d, want 3 and 2 preserved", u.Damage, u.Armor)
	}
}

func TestFR_ZON_002_StructuresAndHeadquartersCannotMarch(t *testing.T) {
	s := newScene(t)
	st := s.put(0, ZoneMainTheater, structure("Tower", 0, 5, Keywords{}))
	s.reject(Action{Kind: ActMarch, Player: 0, Asset: st.ID})
	s.reject(Action{Kind: ActMarch, Player: 0, Asset: s.g.Players[0].HQ.ID})
}

// ---- Domain 3: economy and turn structure ----

func TestFR_RES_001_GenerateYieldDependsOnZone(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	c := unit("Worker", 1, 1, Keywords{})
	c.ReqYield, c.TacYield = 2, 3
	res := s.put(0, ZoneReserves, c)
	thr := s.put(0, ZoneMainTheater, c)

	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: res.ID, Resource: Tactics})
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: thr.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: res.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: thr.ID, Resource: Tactics})
	if p.Req != 2 || p.Tac != 3 {
		t.Fatalf("pools R%d T%d, want R2 T3", p.Req, p.Tac)
	}
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: res.ID, Resource: Requisition}) // already exhausted
}

func TestFR_KWD_002i_VersatileGeneratesEitherResource(t *testing.T) {
	s := newScene(t)
	v := s.put(0, ZoneReserves, unit("Sneak", 1, 2, Keywords{Versatile: true}))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: v.ID, Resource: Tactics})
	if s.g.Players[0].Tac != 1 {
		t.Fatal("Versatile unit in Reserves should be able to generate Tactics")
	}
}

func TestFR_CAP_001_StructuresAndHeadquartersChooseResource(t *testing.T) {
	s := newScene(t)
	st := s.put(0, ZoneMainTheater, structure("Foundry", 0, 5, Keywords{}))
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: st.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: s.g.Players[0].HQ.ID, Resource: Tactics})
	if p := s.g.Players[0]; p.Req != 2 || p.Tac != 2 {
		t.Fatalf("pools R%d T%d, want R2 T2", p.Req, p.Tac)
	}
}

func TestFR_RES_002_PoolsEmptyAtEndOfTurn(t *testing.T) {
	s := newScene(t)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: s.g.Players[0].HQ.ID, Resource: Tactics})
	s.endTurn()
	if p := s.g.Players[0]; p.Req != 0 || p.Tac != 0 {
		t.Fatalf("pools R%d T%d after End Phase, want 0", p.Req, p.Tac)
	}
}

func TestFR_RES_003_ConvertTwoRequisitionToOneTactics(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	p.Req = 3
	s.do(Action{Kind: ActConvert, Player: 0})
	if p.Req != 1 || p.Tac != 1 {
		t.Fatalf("pools R%d T%d, want R1 T1", p.Req, p.Tac)
	}
	s.reject(Action{Kind: ActConvert, Player: 0})
}

func TestFR_TRN_001_ReadyPhaseReadiesOnlyActivePlayer(t *testing.T) {
	s := newScene(t)
	mine := s.put(0, ZoneReserves, unit("Mine", 1, 1, Keywords{}))
	theirs := s.put(1, ZoneReserves, unit("Theirs", 1, 1, Keywords{}))
	mine.Exhausted, theirs.Exhausted = true, true
	s.endTurn()
	if !mine.Exhausted || theirs.Exhausted {
		t.Fatalf("after handover mine.exhausted=%v theirs.exhausted=%v, want true/false", mine.Exhausted, theirs.Exhausted)
	}
}

func TestFR_TRN_002_NonActivePlayerCannotAct(t *testing.T) {
	s := newScene(t)
	s.reject(Action{Kind: ActGenerate, Player: 1, Asset: s.g.Players[1].HQ.ID, Resource: Requisition})
	s.reject(Action{Kind: ActEndTurn, Player: 1})
}

// ---- Domain 4: deployment ----

func TestFR_TRN_003_DeployPaysCostsAndEntersDeclaredZone(t *testing.T) {
	s := newScene(t)
	p := s.g.Players[0]
	c := unit("Recruit", 2, 2, Keywords{})
	c.ReqCost = 2
	c.DeployZones = []Zone{ZoneReserves}
	inst := s.give(0, c)

	s.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves}) // cannot pay
	p.Req = 2
	s.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater}) // not a printed zone
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if p.Req != 0 || len(p.Reserves) != 1 || p.Reserves[0].Inst != inst || p.Reserves[0].Exhausted {
		t.Fatalf("after deploy R%d reserves=%v, want R0 and one Readied Recruit", p.Req, p.Reserves)
	}
}

func TestFR_KWD_002l_DeploysExhausted(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, unit("Slowpoke", 1, 1, Keywords{DeploysExhausted: true}))
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if a := s.g.Players[0].Reserves[0]; !a.Exhausted {
		t.Fatal("Deploys Exhausted asset entered Readied")
	}
}

func TestFR_KWD_002k_ArmorPoolInitializesOnDeploy(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, unit("Plated", 2, 1, Keywords{Armor: 2}))
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if a := s.g.TheaterOf(0)[0]; a.Armor != 2 {
		t.Fatalf("armor = %d, want 2", a.Armor)
	}
}

// ---- Domain 6: stack and damage pipeline ----

func TestFR_QUE_001_NonEmptyStackBlocksTacticalActionsButNotEconomy(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 1, 1, Keywords{}))
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: u.ID, Resource: Requisition})
}

func TestFR_LAY_001_CombatPowerClampsAtZero(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneMainTheater, unit("Weak", 2, 3, Keywords{}))
	u.TempPower = -5
	if p := s.g.Power(u); p != 0 {
		t.Fatalf("power = %d, want 0", p)
	}
}

// Covers: FR-KWD-002b, FR-KWD-002c
func TestFR_REP_004_ResistantAppliesBeforeStalwart(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Wall", 0, 9, Keywords{Resistant: 1, Stalwart: true}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1) // 5 -> 4 (Resistant) -> 1 (Stalwart)

	s2 := newScene(t)
	att2 := s2.put(0, ZoneMainTheater, unit("Poke", 1, 9, Keywords{}))
	def2 := s2.put(1, ZoneMainTheater, unit("Wall", 0, 9, Keywords{Resistant: 1, Stalwart: true}))
	s2.do(s2.attack(att2, def2))
	s2.wantDamage(def2, 0) // 1 -> 0: Resistant + Stalwart is not 1-damage-minimum
}

// Covers: FR-KWD-002c, FR-LAY-002
func TestFR_REP_004_StalwartCappedDamageIsAbsorbedByArmor(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 5, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Knight", 0, 3, Keywords{Stalwart: true, Armor: 2}))
	s.do(s.attack(att, def))
	if def.Armor != 1 || def.Damage != 0 {
		t.Fatalf("armor=%d damage=%d, want 1 and 0", def.Armor, def.Damage)
	}
}

// ---- Domain 7: combat ----

func TestFR_CBT_001_ReservesCannotBeAttackedOrAttack(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 2, Keywords{}))
	back := s.put(1, ZoneReserves, unit("Safe", 1, 1, Keywords{}))
	s.reject(s.attack(att, back))
	rear := s.put(0, ZoneReserves, unit("Rear", 2, 2, Keywords{}))
	s.reject(s.attack(rear, s.g.Players[1].HQ))
}

func TestFR_CBT_001_StructuresCannotAttack(t *testing.T) {
	s := newScene(t)
	st := s.put(0, ZoneMainTheater, structure("Cannon", 4, 4, Keywords{}))
	s.reject(s.attack(st, s.g.Players[1].HQ))
}

func TestFR_KWD_002j_GuardianRedirectsAttacks(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 2, Keywords{}))
	other := s.put(1, ZoneMainTheater, unit("Other", 1, 1, Keywords{}))
	g := s.put(1, ZoneMainTheater, unit("Bodyguard", 0, 5, Keywords{Guardian: true}))
	s.reject(s.attack(att, s.g.Players[1].HQ))
	s.reject(s.attack(att, other))
	s.do(s.attack(att, g))
}

func TestFR_KWD_002j_GuardianInReservesDoesNotRestrict(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 2, Keywords{}))
	s.put(1, ZoneReserves, unit("Bodyguard", 0, 5, Keywords{Guardian: true}))
	s.do(s.attack(att, s.g.Players[1].HQ))
}

func TestFR_KWD_002g_EtherealCannotBeAttackedAndOverridesGuardian(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 2, 2, Keywords{}))
	ghost := s.put(1, ZoneMainTheater, unit("Ghost Guard", 0, 5, Keywords{Guardian: true, Ethereal: true}))
	s.reject(s.attack(att, ghost))
	s.do(s.attack(att, s.g.Players[1].HQ)) // an Ethereal Guardian does not restrict attackers
}

func TestFR_CBT_002_RetaliationIsSymmetric(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("A", 3, 4, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("D", 2, 5, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
	s.wantDamage(att, 2)

	// Headquarters retaliate too.
	u := s.put(0, ZoneMainTheater, unit("Raider", 2, 2, Keywords{}))
	s.do(s.attack(u, s.g.Players[1].HQ))
	s.wantDamage(s.g.Players[1].HQ, 2)
	if s.alive(u) {
		t.Error("raider should die to 3 HQ retaliation")
	}
}

// Covers: FR-KWD-002b
func TestFR_CBT_001_ZeroPowerAttackIsAmplifiedByNegativeResistant(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Zero", 0, 3, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Brittle", 0, 4, Keywords{Resistant: -1}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 1)
}

func TestFR_CBT_003_HeadquartersDefeat(t *testing.T) {
	s := newScene(t)
	s.g.Players[1].HQ.Damage = 17
	u := s.put(0, ZoneMainTheater, unit("Finisher", 5, 9, Keywords{}))
	s.do(s.attack(u, s.g.Players[1].HQ))
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Fatalf("outcome = %+v, want player 0 wins", o)
	}
	s.reject(Action{Kind: ActEndTurn, Player: 0})
}

// Ruling KIM-S02 (2026-10-19): defeat waits for the next death sweep.
// Covers: FR-LAY-002
func TestFR_CBT_003_HeadquartersDefeatIsCheckedAtTheDeathSweep(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 1, 1, Keywords{}))
	s.g.Players[1].HQ.Damage = 25 // lethal, but nothing has swept yet
	if s.g.Outcome.Over {
		t.Fatal("match ended before any death sweep")
	}
	s.do(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Fatalf("outcome after the sweep = %+v, want player 0 wins", o)
	}
}

func TestFR_CBT_003_BothHeadquartersLethalAtOneSweepIsMatchDraw(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 1, 1, Keywords{}))
	s.g.Players[0].HQ.Damage = 20
	s.g.Players[1].HQ.Damage = 22
	s.do(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	if o := s.g.Outcome; !o.Over || !o.Draw {
		t.Fatalf("outcome = %+v, want Match Draw", o)
	}
}

func TestFR_CBT_003_SimultaneousHeadquartersDefeatIsMatchDraw(t *testing.T) {
	s := newScene(t)
	s.g.Players[0].HQ.Damage = 18
	s.g.Players[1].HQ.Damage = 18
	s.do(s.attack(s.g.Players[0].HQ, s.g.Players[1].HQ))
	if o := s.g.Outcome; !o.Over || !o.Draw {
		t.Fatalf("outcome = %+v, want Match Draw", o)
	}
}

func TestFR_KWD_002f_AggressiveResolvesBeforeDamageAndExpires(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Brawler", 1, 5, Keywords{Aggressive: 2}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
	s.endTurn()
	if p := s.g.Power(att); p != 1 {
		t.Fatalf("power after End Phase = %d, want 1", p)
	}
}

// ---- Rulings KIM-J01, S03, S04 (2026-10-19) ----

// declareThen declares a, runs between (standing in for Step 2 triggers), then
// resolves the stack.
func (s *scene) declareThen(a Action, between func()) {
	s.t.Helper()
	if err := s.g.check(a); err != nil {
		s.t.Fatal(err)
	}
	s.g.declareAttack(a)
	between()
	s.g.resolveStack()
}

func TestFR_CBT_002_CommittedAttackIgnoresLaterGuardianAndEthereal(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Hitter", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Target", 0, 9, Keywords{}))
	s.declareThen(s.attack(att, def), func() {
		s.put(1, ZoneMainTheater, unit("Late Bodyguard", 0, 9, Keywords{Guardian: true}))
		def.Inst.Card.Keywords.Ethereal = true
	})
	s.wantDamage(def, 3)
}

// Covers: FR-RES-001
func TestFR_CAP_001_ZeroYieldGenerationIsLegal(t *testing.T) {
	s := newScene(t)
	c := unit("Tender", 0, 1, Keywords{})
	c.TacYield = 0
	u := s.put(0, ZoneMainTheater, c)
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: u.ID, Resource: Tactics})
	if !u.Exhausted || s.g.Players[0].Tac != 0 {
		t.Fatalf("exhausted=%v tactics=%d, want exhausted for 0", u.Exhausted, s.g.Players[0].Tac)
	}
}

// Covers: FR-LAY-002
func TestFR_WIN_001_EmptyDeckLossWaitsForTheDeathSweep(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 1, 1, Keywords{}))
	s.g.Players[1].Deck = nil
	s.g.draw(1, 1) // as if an effect made the opponent draw mid-resolution
	if s.g.Outcome.Over {
		t.Fatal("empty-deck draw ended the match before a death sweep")
	}
	s.do(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	if o := s.g.Outcome; !o.Over || o.Draw || o.Winner != 0 {
		t.Fatalf("outcome = %+v, want player 0 wins", o)
	}
}

func TestFR_CBT_003_HeadquartersDefeatAndEmptyDeckAtOneSweepIsMatchDraw(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Walker", 1, 1, Keywords{}))
	s.g.Players[0].HQ.Damage = 20
	s.g.Players[1].Deck = nil
	s.g.draw(1, 1)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: u.ID})
	if o := s.g.Outcome; !o.Over || !o.Draw {
		t.Fatalf("outcome = %+v, want Match Draw", o)
	}
}
