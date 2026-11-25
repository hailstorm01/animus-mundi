package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0304: Stormbolt Ray (structure, R3 T0, 0/7, Main Theater): "When
// this structure deploys, deal 3 damage to chosen asset." and "T2, exhaust:
// Return this structure to its owner's hand." The real card is loaded from
// data/cards. Player 0 controls the Ray and is the active player.

// stormChoose makes seat 0 pick the option that refers to the asset pick
// returns when the question is asked (so the Ray itself can be chosen before its
// asset exists). It records every decision in s.asked(0).
func stormChoose(s *scene, pick func() *Asset) {
	if s.asks == nil {
		s.asks = map[PlayerID][]Decision{}
	}
	s.asks[0] = nil
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		s.asks[0] = append(s.asks[0], d)
		want := pick()
		for i, o := range d.Options {
			if want != nil && o.Ref == want.ID {
				return Answer{Picks: []int{i}}
			}
		}
		s.t.Errorf("the chosen asset is not among the options %v", refs(d))
		return Answer{Timeout: true}
	}
}

// stormDeploy gives player 0 the real Ray and the stated pool, and deploys it
// into zone. It returns the card; find the asset with assetOf.
func stormDeploy(s *scene, req, tac int, zone Zone) (*Card, *Instance) {
	s.t.Helper()
	return stormDeployCard(s, s.card("Stormbolt Ray"), req, tac, zone)
}

// stormDeployCard is stormDeploy for a card the test already made, so the Ray
// can be found by that card while a question is being asked.
func stormDeployCard(s *scene, c *Card, req, tac int, zone Zone) (*Card, *Instance) {
	s.t.Helper()
	inst := s.give(0, c)
	s.setPool(0, req, tac)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: zone})
	return c, inst
}

// stormRay is a scene with a Readied Ray put in player 0's Main Theater, pool R0 T2.
func stormRay(t *testing.T) (*scene, *Asset) {
	t.Helper()
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, s.card("Stormbolt Ray"))
	s.setPool(0, 0, 2)
	return s, r
}

// stormTapCard is a copy of the real Ray (the shared definition is not changed)
// whose triggers also hold one synthetic HookExhausted trigger with step. Paying
// the ability's exhaust cost fires it, so it is placed on top of the ability and
// resolves first (FR-TRN-003a, SVY-R17).
func stormTapCard(s *scene, step ability.Step) *Card {
	s.t.Helper()
	c := *s.card("Stormbolt Ray")
	if c.Ability == nil {
		s.t.Fatalf("the Ray has no executable definition")
	}
	d := *c.Ability
	d.Triggers = append(append([]ability.Trigger(nil), d.Triggers...),
		ability.Trigger{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: []ability.Step{step}}})
	c.Ability = &d
	return &c
}

// stormKill removes a from the battlefield into its owner's discard pile.
func stormKill(s *scene, a *Asset) {
	s.g.removeFromBattlefield(a)
	s.g.putCard(a.Inst, ZoneDiscard, false)
}

// --- Deploying ---

func TestCard_StormboltRay_EntersMainTheaterAsZeroSeven(t *testing.T) {
	// Covers: FR-TRN-003, FR-SET-004
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	c, _ := stormDeploy(s, 3, 0, ZoneMainTheater)
	a := assetOf(s.g, c)
	if a == nil || a.Zone != ZoneMainTheater {
		t.Fatalf("the Ray is on the battlefield in the Main Theater after deploying (FR-TRN-003)")
	}
	if p := a.Printed(); p.Power != 0 || p.Health != 7 {
		t.Errorf("the Ray is a 0/7: got %d/%d", p.Power, p.Health)
	}
	if a.Exhausted {
		t.Errorf("a newly deployed structure enters Readied (FR-SET-004)")
	}
	if s.g.Players[0].Req != 0 {
		t.Errorf("the Ray costs 3 Requisition: %d left", s.g.Players[0].Req)
	}

	s2 := newScene(t)
	c2 := s2.card("Stormbolt Ray")
	inst := s2.give(0, c2)
	s2.setPool(0, 3, 0)
	s2.reject(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneReserves})
	if !inHand(s2.g.Players[0], inst) || assetOf(s2.g, c2) != nil {
		t.Errorf("deploying the Ray to Reserves is rejected and the card stays in hand (FR-TRN-003)")
	}

	s3 := newScene(t)
	c3 := s3.card("Stormbolt Ray")
	inst3 := s3.give(0, c3)
	s3.setPool(0, 2, 0)
	s3.reject(Action{Kind: ActDeploy, Player: 0, Card: inst3.ID, Zone: ZoneMainTheater})
	if !inHand(s3.g.Players[0], inst3) || assetOf(s3.g, c3) != nil {
		t.Errorf("with 2 Requisition the Ray (cost 3) cannot be deployed (FR-TRN-003)")
	}
}

func TestCard_StormboltRay_AnyAssetOffered(t *testing.T) {
	// Covers: FR-TGT-001, SVY-R57
	s := newScene(t)
	b := s.targetBoard()
	allyRes := s.put(0, ZoneReserves, unit("AllyRes", 1, 5, ability.Keywords{}))
	foeRes := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return s.g.Players[1].HQ })
	c, _ := stormDeploy(s, 3, 0, ZoneMainTheater)
	ray := assetOf(s.g, c)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one chosen target is one question: asked %d times (FR-TGT-001)", len(asked))
	}
	d := asked[0]
	if d.Kind != DecideTarget || d.Player != 0 || d.Min != 1 || d.Max != 1 {
		t.Errorf("the question is a mandatory single DecideTarget for the deploying player: %+v", d)
	}
	wantOptions(t, d, "every asset on the battlefield: units in the Main Theater and Reserves, a Scratch Company, a structure, both Headquarters and the Ray itself, of either side (SVY-R57)",
		b.u0, b.u1, allyRes, foeRes, b.c0, b.c1, b.s0, b.s1, s.g.Players[0].HQ, s.g.Players[1].HQ, ray)
}

func TestCard_StormboltRay_DealsThreeToEachKind(t *testing.T) {
	// Covers: FR-TGT-001
	cases := []struct {
		name string
		pick func(s *scene, b tboard, res *Asset) *Asset
	}{
		{"allied unit", func(s *scene, b tboard, res *Asset) *Asset { return b.u0 }},
		{"opposing unit in Reserves", func(s *scene, b tboard, res *Asset) *Asset { return res }},
		{"allied structure", func(s *scene, b tboard, res *Asset) *Asset { return b.s0 }},
		{"opposing structure", func(s *scene, b tboard, res *Asset) *Asset { return b.s1 }},
		{"allied Headquarters", func(s *scene, b tboard, res *Asset) *Asset { return s.g.Players[0].HQ }},
		{"opposing Headquarters", func(s *scene, b tboard, res *Asset) *Asset { return s.g.Players[1].HQ }},
		{"the Ray itself", func(s *scene, b tboard, res *Asset) *Asset { return nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			b := s.targetBoard()
			res := s.put(1, ZoneReserves, unit("FoeRes", 1, 5, ability.Keywords{}))
			c := s.card("Stormbolt Ray")
			stormChoose(s, func() *Asset {
				if w := tc.pick(s, b, res); w != nil {
					return w
				}
				return assetOf(s.g, c)
			})
			stormDeployCard(s, c, 3, 0, ZoneMainTheater)
			want := tc.pick(s, b, res)
			if want == nil {
				want = assetOf(s.g, c)
			}
			s.wantDamage(want, 3)
			if got := totalDamage(s.g); got != 3 {
				t.Errorf("only the chosen %s takes the 3 damage: total %d, want 3 (FR-TGT-001)", tc.name, got)
			}
		})
	}
}

func TestCard_StormboltRay_DamageIsNotCombat(t *testing.T) {
	// Covers: FR-REP-004
	t.Run("event", func(t *testing.T) {
		s := newScene(t)
		b := s.targetBoard()
		stormChoose(s, func() *Asset { return b.u1 })
		c, _ := stormDeploy(s, 3, 0, ZoneMainTheater)
		evs := damageTo(s.g.Log, b.u1)
		if len(evs) != 1 {
			t.Fatalf("one non-combat damage event against the chosen unit, got %d (FR-REP-004)", len(evs))
		}
		e, ray := evs[0], assetOf(s.g, c)
		if e.Source != ray.ID || e.Raw != 3 || e.Amount != 3 || e.Combat {
			t.Errorf("3 non-combat damage whose Source is the Ray's asset: %+v (Ray asset %d)", e, ray.ID)
		}
	})
	t.Run("Stalwart takes the full 3", func(t *testing.T) {
		s := newScene(t)
		w := s.put(1, ZoneMainTheater, unit("Stalwart Wall", 0, 9, ability.Keywords{Stalwart: true}))
		stormChoose(s, func() *Asset { return w })
		stormDeploy(s, 3, 0, ZoneMainTheater)
		s.wantDamage(w, 3)
	})
	t.Run("Resistant and Armor reduce it", func(t *testing.T) {
		s := newScene(t)
		w := s.put(1, ZoneMainTheater, unit("Hardened", 0, 9, ability.Keywords{Resistant: 1, Armor: 1}))
		stormChoose(s, func() *Asset { return w })
		stormDeploy(s, 3, 0, ZoneMainTheater)
		s.wantDamage(w, 1)
		if w.Armor != 0 {
			t.Errorf("Resistant 1 cut 3 to 2 and the 1 Armor absorbed 1 and is spent: Armor %d, want 0 (FR-REP-004)", w.Armor)
		}
	})
}

func TestCard_StormboltRay_LethalDamageKills(t *testing.T) {
	// Covers: FR-LAY-002
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Frail", 1, 3, ability.Keywords{}))
	inst := foe.Inst
	stormChoose(s, func() *Asset { return foe })
	stormDeploy(s, 3, 0, ZoneMainTheater)
	if s.alive(foe) || !inDiscardPile(s.g.Players[1], inst) {
		t.Errorf("a Health 3 unit taking 3 dies and its card is in its owner's discard pile (FR-LAY-002)")
	}

	s2 := newScene(t)
	tough := s2.put(1, ZoneMainTheater, unit("Tough", 1, 4, ability.Keywords{}))
	stormChoose(s2, func() *Asset { return tough })
	stormDeploy(s2, 3, 0, ZoneMainTheater)
	if !s2.alive(tough) {
		t.Errorf("a Health 4 unit survives 3 damage (FR-LAY-002)")
	}
	s2.wantDamage(tough, 3)
}

func TestCard_StormboltRay_OpposingHeadquartersToZeroWins(t *testing.T) {
	// Covers: FR-CBT-003
	s := newScene(t)
	s.g.Players[1].HQ.Damage = 17
	stormChoose(s, func() *Asset { return s.g.Players[1].HQ })
	stormDeploy(s, 3, 0, ZoneMainTheater)
	if !s.g.Outcome.Over || s.g.Outcome.Draw || s.g.Outcome.Winner != 0 {
		t.Errorf("an opposing Headquarters at 0 Current Health loses the match: outcome %+v", s.g.Outcome)
	}
}

func TestCard_StormboltRay_WardedOpposingAssetNotOffered(t *testing.T) {
	// Covers: FR-KWD-002h
	s := newScene(t)
	foeWard := s.put(1, ZoneMainTheater, unit("FoeWard", 1, 5, ability.Keywords{Ward: true}))
	allyWard := s.put(0, ZoneMainTheater, unit("AllyWard", 1, 5, ability.Keywords{Ward: true}))
	stormChoose(s, func() *Asset { return allyWard })
	c, _ := stormDeploy(s, 3, 0, ZoneMainTheater)
	asked := s.asked(0)
	if len(asked) != 1 {
		t.Fatalf("one question, got %d", len(asked))
	}
	wantOptions(t, asked[0], "an opposing Ward asset is not offered; the allied Ward unit is (FR-KWD-002h)",
		allyWard, s.g.Players[0].HQ, s.g.Players[1].HQ, assetOf(s.g, c))
	s.wantDamage(foeWard, 0)
	s.wantDamage(allyWard, 3)
}

func TestCard_StormboltRay_TimeoutTakesTheFirstOption(t *testing.T) {
	// Covers: FR-MOD-001
	s := newScene(t)
	b := s.targetBoard()
	s.script(0, Answer{Timeout: true})
	stormDeploy(s, 3, 0, ZoneMainTheater)
	if n := len(s.asked(0)); n != 1 {
		t.Fatalf("asked %d times, want 1", n)
	}
	s.wantDamage(s.g.Players[0].HQ, 3)
	s.wantDamage(s.g.Players[1].HQ, 0)
	s.wantDamage(b.u0, 0)
	if got := totalDamage(s.g); got != 3 {
		t.Errorf("a timeout deals exactly 3 to exactly one target, the first in fixed order, the choosing player's own Headquarters (FR-MOD-001 sec. 6, SVY-R16): total %d", got)
	}
}

func TestCard_StormboltRay_TargetGoneDealsNothing(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Fleeting", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	s.ptPlaceHook(func(int, StackItem) {
		if s.alive(foe) {
			s.g.removeFromBattlefield(foe)
		}
	})
	c, _ := stormDeploy(s, 3, 0, ZoneMainTheater)
	if n := len(damageEvents(s.g.Log)); n != 0 {
		t.Errorf("a target that left before resolution is skipped and no damage event is logged: %d (FR-TGT-003)", n)
	}
	if assetOf(s.g, c) == nil {
		t.Errorf("the Ray stays on the battlefield")
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("nothing takes damage: total %d (FR-TGT-003)", got)
	}
}

func TestCard_StormboltRay_SourceGoneStillDeals(t *testing.T) {
	// Covers: FR-TGT-003
	s := newScene(t)
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 5, ability.Keywords{}))
	stormChoose(s, func() *Asset { return foe })
	c := s.card("Stormbolt Ray")
	s.ptPlaceHook(func(int, StackItem) {
		if ray := assetOf(s.g, c); ray != nil {
			stormKill(s, ray)
		}
	})
	inst := s.give(0, c)
	s.setPool(0, 3, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	if assetOf(s.g, c) != nil {
		t.Fatalf("setup: the Ray should have left before resolution")
	}
	s.wantDamage(foe, 3)
}

// --- Activating ---

func TestCard_StormboltRay_ReturnsToOwnersHand(t *testing.T) {
	// Covers: FR-ACT-001, FR-ZON-004
	s, r := stormRay(t)
	b := s.targetBoard()
	s.script(0)
	s.do(activate(r, 0))
	if s.alive(r) || inAssets(s.g.Theater, r) {
		t.Errorf("the Ray leaves the battlefield (FR-ZON-004)")
	}
	p := s.g.Players[0]
	if !inHand(p, r.Inst) {
		t.Errorf("the Ray's card is in its owner's hand (FR-ZON-004)")
	}
	if inDiscardPile(p, r.Inst) {
		t.Errorf("the Ray's card is not in a discard pile (FR-ZON-004)")
	}
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("the 2 Tactics are spent: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked: %d decisions (SVY-R36)", n)
	}
	for _, a := range []*Asset{b.u0, b.s0, b.c0, b.u1, b.s1, b.c1} {
		if !s.alive(a) || a.Damage != 0 {
			t.Errorf("no other asset changes: %s alive %v damage %d", a, s.alive(a), a.Damage)
		}
	}
	if got := totalDamage(s.g); got != 0 {
		t.Errorf("no damage is dealt: total %d", got)
	}
}

func TestCard_StormboltRay_CostsPaidBeforeItResolves(t *testing.T) {
	// Covers: FR-ACT-001
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, stormTapCard(s, actHQDamage(1)))
	s.setPool(0, 0, 2)
	places := 0
	var onField, exhausted, returnBelow bool
	var tac int
	s.ptPlaceHook(func(int, StackItem) {
		places++
		if places != 2 {
			return
		}
		onField, exhausted, tac = s.alive(r), r.Exhausted, s.g.Players[0].Tac
		if n := len(s.g.Stack); n >= 2 {
			below := s.g.Stack[n-2]
			returnBelow = below.Kind == StackAbility && len(below.Effect.Steps) == 1 && below.Effect.Steps[0].Op == ability.OpReturnToHand
		}
	})
	s.do(activate(r, 0))
	if places != 2 {
		t.Fatalf("the ability and the exhaust trigger are placed, got %d items (SVY-R17)", places)
	}
	if !returnBelow {
		t.Errorf("when the trigger is placed the return is still on the stack below it (FR-ACT-001)")
	}
	if !onField || !exhausted {
		t.Errorf("when the trigger is placed the Ray is still on the battlefield and Exhausted: on battlefield %v, exhausted %v (FR-ACT-001)", onField, exhausted)
	}
	if tac != 0 {
		t.Errorf("the 2 Tactics are paid on declaration: %d left when the trigger is placed (FR-ACT-001)", tac)
	}
	if s.alive(r) || !inHand(s.g.Players[0], r.Inst) {
		t.Errorf("the return then resolves: the Ray is in its owner's hand (FR-ZON-004)")
	}
}

func TestCard_StormboltRay_GoesToItsOwnersHand(t *testing.T) {
	// Covers: FR-ZON-003
	s := newScene(t)
	r := s.putControlled(1, 0, s.card("Stormbolt Ray"))
	s.setPool(0, 0, 2)
	s.do(activate(r, 0))
	if s.alive(r) {
		t.Fatalf("the Ray leaves the battlefield (FR-ZON-004)")
	}
	if !inHand(s.g.Players[1], r.Inst) {
		t.Errorf("the card returns to its owner's (player 1's) hand (FR-ZON-003)")
	}
	if inHand(s.g.Players[0], r.Inst) {
		t.Errorf("the card does not go to the controller's hand (FR-ZON-003)")
	}
}

func TestCard_StormboltRay_RedeployedIsNew(t *testing.T) {
	// Covers: FR-ZON-004, FR-SET-004
	s, r := stormRay(t)
	r.Damage = 2
	r.Armor = 1
	c, inst := r.Inst.Card, r.Inst
	s.do(activate(r, 0))
	if !inHand(s.g.Players[0], inst) {
		t.Fatalf("setup: the Ray is back in hand")
	}
	hq1 := s.g.Players[1].HQ
	stormChoose(s, func() *Asset { return hq1 })
	s.setPool(0, 3, 0)
	s.do(Action{Kind: ActDeploy, Player: 0, Card: inst.ID, Zone: ZoneMainTheater})
	n := assetOf(s.g, c)
	if n == nil {
		t.Fatalf("the Ray is deployed again")
	}
	if n.Damage != 0 || n.Armor != 0 || n.Exhausted {
		t.Errorf("a redeployed Ray is a new Readied asset with no damage or Armor: damage %d, armor %d, exhausted %v (FR-ZON-004, FR-SET-004)", n.Damage, n.Armor, n.Exhausted)
	}
	if p := n.Printed(); p.Power != 0 || p.Health != 7 {
		t.Errorf("a redeployed Ray is a 0/7: got %d/%d", p.Power, p.Health)
	}
	s.wantDamage(hq1, 3)
}

func TestCard_StormboltRay_MayActivateTheTurnItDeploys(t *testing.T) {
	// Covers: FR-SET-004
	s := newScene(t)
	stormChoose(s, func() *Asset { return s.g.Players[1].HQ })
	c, _ := stormDeploy(s, 3, 2, ZoneMainTheater)
	ray := assetOf(s.g, c)
	if ray == nil {
		t.Fatalf("the Ray is deployed")
	}
	if !offersActivate(s.g, ray, 0) {
		t.Errorf("a Ray deployed this turn may activate right away: there is no summoning sickness (FR-SET-004)")
	}
}

func TestCard_StormboltRay_NotOfferedWithoutTwoTactics(t *testing.T) {
	// Covers: FR-ACT-001
	s, r := stormRay(t)
	s.setPool(0, 5, 1)
	if offersActivate(s.g, r, 0) {
		t.Errorf("with R5 T1 the Ray's T2 cost cannot be paid, so Activate is not offered (FR-ACT-001)")
	}
	s.reject(activate(r, 0))
	p := s.g.Players[0]
	if !s.alive(r) || r.Exhausted || inHand(p, r.Inst) || p.Req != 5 || p.Tac != 1 {
		t.Errorf("a rejected Activate moves nothing: alive %v, exhausted %v, pool R%d T%d (FR-ACT-001)", s.alive(r), r.Exhausted, p.Req, p.Tac)
	}
}

func TestCard_StormboltRay_NotOfferedWhileExhausted(t *testing.T) {
	// Covers: FR-ACT-002
	s, r := stormRay(t)
	s.setPool(0, 0, 4)
	r.Exhausted = true
	if offersActivate(s.g, r, 0) {
		t.Errorf("an Exhausted Ray cannot pay its exhaust cost, so Activate is not offered (FR-ACT-002)")
	}
	s.reject(activate(r, 0))
	p := s.g.Players[0]
	if !s.alive(r) || inHand(p, r.Inst) || p.Tac != 4 {
		t.Errorf("a rejected Activate moves nothing: alive %v, Tactics %d (FR-ACT-002)", s.alive(r), p.Tac)
	}
}

func TestCard_StormboltRay_GoneBeforeResolutionStaysGone(t *testing.T) {
	// Covers: FR-ZON-004, FR-TGT-003
	s := newScene(t)
	r := s.put(0, ZoneMainTheater, stormTapCard(s, ability.Step{Op: ability.OpKill, Subject: ability.SubjectSelf}))
	s.setPool(0, 0, 2)
	s.do(activate(r, 0))
	p := s.g.Players[0]
	if s.alive(r) {
		t.Fatalf("setup: the Ray should be dead")
	}
	if !inDiscardPile(p, r.Inst) {
		t.Errorf("the Ray's card stays in its owner's discard pile (FR-ZON-004)")
	}
	if inHand(p, r.Inst) {
		t.Errorf("a Ray that left the battlefield is not returned to hand: the last-known source is gone (FR-ZON-004, FR-TGT-003)")
	}
}
