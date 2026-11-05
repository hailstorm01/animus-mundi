package engine

import (
	"reflect"
	"sort"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0158 (FR-ACT-001, FR-ACT-002, FR-TRN-003a, FR-QUE-001,
// FR-TGT-001, FR-TGT-003, FR-KWD-002h; SVY-R2, SVY-R15, SVY-R17): activating a
// printed "[Cost]: [Effect]." ability of a battlefield asset. Its costs are
// paid and its targets chosen on declaration, and it resolves as a stack item.
// Cards are synthetic. Every exactly-one target slot below has a legal entity
// on declaration, except in the SVY-R36 cases (T-0163), which name it. Player 0
// is the active player.

var (
	actExhaust = ability.Cost{Kind: ability.CostExhaustSelf}
	actKill    = ability.Cost{Kind: ability.CostKillSelf}
)

func actRes(req, tac int) ability.Cost {
	return ability.Cost{Kind: ability.CostResources, Req: req, Tac: tac}
}

// actAbility is one activated ability paying costs for the steps.
func actAbility(costs []ability.Cost, ss ...ability.Step) ability.Activated {
	return ability.Activated{Costs: costs, Effect: ability.Effect{Steps: ss}}
}

// actUnit is a 1/5 unit printing the activated abilities, and the triggers if any.
func actUnit(name string, trigs []ability.Trigger, acts ...ability.Activated) *Card {
	c := unit(name, 1, 5, Keywords{})
	c.Ability = &ability.Def{Triggers: trigs, Activated: acts}
	return c
}

func activate(a *Asset, i int) Action {
	return Action{Kind: ActActivate, Player: a.Controller, Asset: a.ID, Ability: i}
}

func offersActivate(g *Game, a *Asset, i int) bool {
	for _, x := range g.LegalActions() {
		if x.Kind == ActActivate && x.Asset == a.ID && x.Ability == i {
			return true
		}
	}
	return false
}

func actHQDamage(n int) ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Amount: ability.Value{N: n}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindHeadquarters}}
}

func actArmorSelf(n int) ability.Step {
	return ability.Step{Op: ability.OpAddArmor, Amount: ability.Value{N: n}, Subject: ability.SubjectSelf}
}

func (s *scene) setPool(p PlayerID, req, tac int) {
	s.g.Players[p].Req, s.g.Players[p].Tac = req, tac
}

func sortedCopy(xs []int) []int {
	out := append([]int{}, xs...)
	sort.Ints(out)
	return out
}

func TestFR_ACT_001_CostsPaidAndEffectResolves(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 1), actExhaust}, drawN(1))))
	s.setPool(0, 1, 1)
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.do(activate(a, 0))
	if p.Req != 0 || p.Tac != 0 {
		t.Errorf("activating pays R1 T1: pool R%d T%d, want R0 T0 (FR-ACT-001)", p.Req, p.Tac)
	}
	if !a.Exhausted {
		t.Errorf("the exhaust cost exhausts the source (FR-ACT-001)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the ability's effect resolved: drew %d, want 1 (FR-ACT-001)", got)
	}
	s.wantSettled("an activated ability")
}

func TestFR_QUE_001_ActivatedAbilityIsPlacedOnTheStack(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actExhaust}, actHQDamage(1))))
	s.do(activate(a, 0))
	hqID := s.g.Players[1].HQ.ID
	place, dmg := 0, 0
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == "Engine" && place == 0:
			place = e.Seq
			if e.Amount != 1 {
				t.Errorf("the activated ability is placed at stack depth 1, got %d (FR-QUE-001)", e.Amount)
			}
		case e.Kind == "damage" && e.Target == hqID && dmg == 0:
			dmg = e.Seq
		}
	}
	if place == 0 || dmg == 0 {
		t.Fatalf("missing events: place %d, damage %d", place, dmg)
	}
	if place > dmg {
		t.Errorf("the ability is placed (seq %d) before its steps' events (seq %d) (FR-QUE-001)", place, dmg)
	}
}

func TestFR_ACT_001_NeedsAnEmptyStack(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(nil, drawN(1)), actAbility([]ability.Cost{actRes(1, 0), actExhaust}, drawN(1))))
	s.setPool(0, 1, 0)
	s.g.Stack = append(s.g.Stack, &StackItem{Kind: StackScratch, Controller: 0})
	s.reject(activate(a, 0))
	s.reject(activate(a, 1))
	if offersActivate(s.g, a, 0) || offersActivate(s.g, a, 1) {
		t.Errorf("no Activate is offered with an item on the stack (FR-ACT-001)")
	}
	if s.g.Players[0].Req != 1 || a.Exhausted {
		t.Errorf("a rejected Activate spends nothing: R%d exhausted %v, want R1 false (FR-ACT-001)", s.g.Players[0].Req, a.Exhausted)
	}
}

func TestFR_ACT_001_OnlyTheActiveControllerActivates(t *testing.T) {
	s := newScene(t)
	mine := s.put(0, ZoneMainTheater, actUnit("Mine", nil, actAbility(nil, drawN(1))))
	theirs := s.put(1, ZoneMainTheater, actUnit("Theirs", nil, actAbility(nil, drawN(1))))
	s.reject(activate(theirs, 0))
	s.reject(Action{Kind: ActActivate, Player: 0, Asset: theirs.ID, Ability: 0})
	s.reject(Action{Kind: ActActivate, Player: 1, Asset: mine.ID, Ability: 0})
	if offersActivate(s.g, theirs, 0) {
		t.Errorf("the active player is not offered an opponent's asset's ability (FR-ACT-001, FR-ZON-003)")
	}
	s.endTurn()
	if s.g.Active != 1 {
		t.Fatalf("precondition: it is player 1's turn")
	}
	s.reject(activate(mine, 0))
	s.do(activate(theirs, 0))

	s2 := newScene(t)
	stolen := s2.put(0, ZoneMainTheater, actUnit("Stolen", nil, actAbility(nil, drawN(1))))
	stolen.Inst.Owner = 1
	deck := len(s2.g.Players[0].Deck)
	s2.do(activate(stolen, 0))
	if got := deck - len(s2.g.Players[0].Deck); got != 1 {
		t.Errorf("the controller, not the owner, activates and draws: drew %d, want 1 (FR-ZON-003)", got)
	}
}

func TestFR_ACT_001_RejectsWhatHasNoSuchAbility(t *testing.T) {
	s := newScene(t)
	one := s.put(0, ZoneMainTheater, actUnit("One", nil, actAbility(nil, drawN(1))))
	s.reject(activate(one, 1))
	s.reject(activate(one, -1))
	bare := s.put(0, ZoneMainTheater, unit("Bare", 1, 1, Keywords{}))
	s.reject(activate(bare, 0))
	trigOnly := s.put(0, ZoneMainTheater, mortal("Trigger Only", 1, 1, killedTrigger(drawN(1))))
	s.reject(activate(trigOnly, 0))
	down := s.put(0, ZoneReserves, actUnit("Hidden", nil, actAbility(nil, drawN(1))))
	down.FaceDown = true
	s.reject(activate(down, 0))
	if offersActivate(s.g, down, 0) || offersActivate(s.g, bare, 0) || offersActivate(s.g, one, 1) {
		t.Errorf("only a printed ability of a face-up card is offered (FR-ACT-001)")
	}
	if !offersActivate(s.g, one, 0) {
		t.Errorf("the one printed ability is offered (FR-ACT-001)")
	}
}

func TestFR_ACT_001_UnpayableIsNotOffered(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 1), actExhaust}, drawN(1))))
	p := s.g.Players[0]
	hand, deck := len(p.Hand), len(p.Deck)
	check := func(why string, req, tac int, exhausted bool) {
		t.Helper()
		s.setPool(0, req, tac)
		a.Exhausted = exhausted
		if offersActivate(s.g, a, 0) {
			t.Errorf("%s: the ability is not offered (FR-ACT-001)", why)
		}
		s.reject(activate(a, 0))
		if p.Req != req || p.Tac != tac || a.Exhausted != exhausted || len(p.Hand) != hand || len(p.Deck) != deck {
			t.Errorf("%s: a rejected Activate spends nothing (FR-ACT-001)", why)
		}
	}
	check("too little Requisition", 0, 1, false)
	check("too little Tactics", 1, 0, false)
	check("an Exhausted source with an exhaust cost", 1, 1, true)
	s.setPool(0, 1, 1)
	a.Exhausted = false
	if !offersActivate(s.g, a, 0) {
		t.Errorf("once every cost can be paid the ability is offered (FR-ACT-001)")
	}
	s.do(activate(a, 0))
}

func TestFR_ACT_001_OfferedOncePerPrintedAbility(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Twin", nil, actAbility(nil, actArmorSelf(2)), actAbility(nil, drawN(1))))
	var idx []int
	for _, x := range s.g.LegalActions() {
		if x.Kind == ActActivate && x.Asset == a.ID {
			idx = append(idx, x.Ability)
		}
	}
	if !reflect.DeepEqual(sortedCopy(idx), []int{0, 1}) {
		t.Fatalf("each printed ability is offered once: got indexes %v, want [0 1] (FR-ACT-001)", idx)
	}
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.do(activate(a, 1))
	if a.Armor != 0 || deck-len(p.Deck) != 1 {
		t.Errorf("index 1 runs the second ability only: armor %d, drew %d, want 0, 1 (FR-ACT-001)", a.Armor, deck-len(p.Deck))
	}
	s.do(activate(a, 0))
	if a.Armor != 2 || deck-len(p.Deck) != 1 {
		t.Errorf("index 0 runs the first ability only: armor %d, drew %d, want 2, 1 (FR-ACT-001)", a.Armor, deck-len(p.Deck))
	}
}

func TestFR_ACT_002_NoExhaustCostWorksWhileExhausted(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 0)}, actArmorSelf(1))))
	a.Exhausted = true
	s.setPool(0, 1, 0)
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("an Exhausted asset's ability with no exhaust cost is offered (FR-ACT-002)")
	}
	s.do(activate(a, 0))
	if a.Armor != 1 || !a.Exhausted || s.g.Players[0].Req != 0 {
		t.Errorf("it activates and stays Exhausted: armor %d, exhausted %v, R%d, want 1, true, 0 (FR-ACT-002)", a.Armor, a.Exhausted, s.g.Players[0].Req)
	}
}

func TestFR_ACT_002_NoExhaustCostRepeats(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility([]ability.Cost{actRes(1, 0)}, actArmorSelf(1))))
	s.setPool(0, 2, 0)
	s.do(activate(a, 0))
	s.do(activate(a, 0))
	if a.Armor != 2 {
		t.Errorf("with no exhaust cost it activates twice in a turn: armor %d, want 2 (FR-ACT-002)", a.Armor)
	}
	s.reject(activate(a, 0))
	if a.Armor != 2 {
		t.Errorf("a third activation lacks Requisition and does nothing: armor %d, want 2 (FR-ACT-001)", a.Armor)
	}

	s2 := newScene(t)
	b := s2.put(0, ZoneMainTheater, actUnit("Tapper", nil, actAbility([]ability.Cost{actExhaust}, actArmorSelf(1))))
	s2.do(activate(b, 0))
	s2.reject(activate(b, 0))
	if b.Armor != 1 {
		t.Errorf("an ability with an exhaust cost cannot activate again while Exhausted: armor %d, want 1 (FR-ACT-001)", b.Armor)
	}
}

func TestFR_ACT_001_KillSelfCostPaidOnDeclaration(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Martyr", nil, actAbility([]ability.Cost{actKill}, drawN(1))))
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.do(activate(a, 0))
	if s.alive(a) || !inDiscardPile(p, a.Inst) {
		t.Errorf("the kill-self cost puts the source's card in its owner's discard pile (FR-ACT-001)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("the ability still resolves after its source left: drew %d, want 1 (FR-ACT-001)", got)
	}
	s.wantSettled("a kill-self activation")
}

func TestFR_ACT_001_SacrificedCardIsNotAChoice(t *testing.T) {
	// Covers: SVY-R2, FR-TGT-001
	pile := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	recall := func() *Card {
		return actUnit("Recaller", nil, ability.Activated{
			Costs:  []ability.Cost{actKill},
			Effect: ability.Effect{Targets: []ability.Target{pile}, Steps: []ability.Step{ptStep(ability.OpReturnToHand, 0)}},
		})
	}
	t.Run("one other card", func(t *testing.T) {
		s := newScene(t)
		other := s.inDiscard(0, unit("Other", 1, 1, Keywords{}))[0]
		a := s.put(0, ZoneMainTheater, recall())
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		asks := s.asked(0)
		if len(asks) > 1 || (len(asks) == 1 && !reflect.DeepEqual(refs(asks[0]), []int{other.ID})) {
			t.Errorf("the only option is the other card, never the sacrificed one: asked %d (SVY-R2)", len(asks))
		}
		p := s.g.Players[0]
		if !inHand(p, other) || inDiscardPile(p, other) {
			t.Errorf("the other card ends in hand (SVY-R2)")
		}
		if !inDiscardPile(p, a.Inst) || inHand(p, a.Inst) {
			t.Errorf("the sacrificed card stays in the discard pile (SVY-R2)")
		}
	})
	t.Run("two other cards", func(t *testing.T) {
		s := newScene(t)
		others := s.inDiscard(0, unit("Other One", 1, 1, Keywords{}), unit("Other Two", 1, 1, Keywords{}))
		a := s.put(0, ZoneMainTheater, recall())
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		asks := s.asked(0)
		if len(asks) != 1 || asks[0].Kind != DecideTarget {
			t.Fatalf("two options are asked once as a target decision, got %d decisions (SVY-R2)", len(asks))
		}
		if got, want := sortedCopy(refs(asks[0])), sortedCopy([]int{others[0].ID, others[1].ID}); !reflect.DeepEqual(got, want) {
			t.Errorf("the options are the two other cards, not the sacrificed one: got %v, want %v (SVY-R2)", got, want)
		}
		if !inDiscardPile(s.g.Players[0], a.Inst) {
			t.Errorf("the sacrificed card stays in the discard pile (SVY-R2)")
		}
	})
}

// actOrder checks the ability sits at stack depth 1 and the cost's trigger at
// depth 2, and that the trigger's 2 damage to the opposing Headquarters comes
// before the ability's 1.
func actOrder(t *testing.T, s *scene, name string) {
	t.Helper()
	hqID := s.g.Players[1].HQ.ID
	var abPlace, trigPlace, trigDmg, abDmg int
	for _, e := range s.g.Log {
		switch {
		case e.Kind == "place" && e.Card == name && e.Amount == 1:
			abPlace = e.Seq
		case e.Kind == "place" && e.Card == name && e.Amount == 2:
			trigPlace = e.Seq
		case e.Kind == "damage" && e.Target == hqID && e.Amount == 2 && trigDmg == 0:
			trigDmg = e.Seq
		case e.Kind == "damage" && e.Target == hqID && e.Amount == 1 && abDmg == 0:
			abDmg = e.Seq
		}
	}
	if abPlace == 0 || trigPlace == 0 || trigDmg == 0 || abDmg == 0 {
		t.Fatalf("missing events: ability place at depth 1 %d, trigger place at depth 2 %d, trigger damage %d, ability damage %d (FR-TRN-003a)", abPlace, trigPlace, trigDmg, abDmg)
	}
	if trigPlace < abPlace {
		t.Errorf("the cost's trigger is placed above the declared ability (seq %d vs %d) (FR-TRN-003a)", trigPlace, abPlace)
	}
	if trigDmg > abDmg {
		t.Errorf("the cost's trigger resolves (seq %d) before the ability (seq %d) (FR-TRN-003a)", trigDmg, abDmg)
	}
}

func TestFR_TRN_003a_ActivatedCostTriggersResolveFirst(t *testing.T) {
	t.Run("kill this unit", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Martyr", []ability.Trigger{killedTrigger(actHQDamage(2))},
			actAbility([]ability.Cost{actKill}, actHQDamage(1))))
		s.do(activate(a, 0))
		actOrder(t, s, "Martyr")
	})
	t.Run("exhaust", func(t *testing.T) {
		// Covers: SVY-R17
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Tapper",
			[]ability.Trigger{{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: []ability.Step{actHQDamage(2)}}}},
			actAbility([]ability.Cost{actExhaust}, actHQDamage(1))))
		s.do(activate(a, 0))
		actOrder(t, s, "Tapper")
	})
}

func TestFR_TRN_003a_ReplacedKillSelfCostStaysPaid(t *testing.T) {
	// Covers: SVY-R15, FR-REP-001
	s := newScene(t)
	guardian := structure("Guardian", 1, 9, Keywords{})
	guardian.Ability = &ability.Def{Replacements: []ability.Replacement{{
		Event:  ability.WouldBeDestroyed,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit},
		Steps:  []ability.Step{{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced}},
	}}}
	s.put(0, ZoneMainTheater, guardian)
	a := s.put(0, ZoneMainTheater, actUnit("Martyr", []ability.Trigger{killedTrigger(actHQDamage(2))},
		actAbility([]ability.Cost{actKill}, drawN(1))))
	p := s.g.Players[0]
	deck := len(p.Deck)
	s.do(activate(a, 0))
	if s.alive(a) || !inHand(p, a.Inst) || inDiscardPile(p, a.Inst) {
		t.Errorf("the replacement returns the source to its owner's hand instead of destroying it (FR-REP-001)")
	}
	if got := deck - len(p.Deck); got != 1 {
		t.Errorf("a replaced kill still pays the cost, so the ability resolves: drew %d, want 1 (SVY-R15)", got)
	}
	if d := s.g.Players[1].HQ.Damage; d != 0 {
		t.Errorf("a replaced death fires no death trigger: opposing HQ damage %d, want 0 (FR-REP-001)", d)
	}
	for _, e := range s.g.Log {
		if e.Kind == "place" && e.Card == "Martyr" && e.Amount != 1 {
			t.Errorf("only the ability is placed, not a death trigger: %+v (FR-REP-001)", e)
		}
	}
}

func killChosenOpposing() ability.Effect { return killChosenUnit(ability.AllegianceOpposing) }

func TestFR_TGT_001_ActivatedTargetChosenOnDeclaration(t *testing.T) {
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Assassin", nil, ability.Activated{
		Costs: []ability.Cost{actRes(0, 1)}, Effect: killChosenOpposing()}))
	u1 := s.put(1, ZoneMainTheater, unit("Foe One", 1, 1, Keywords{}))
	u2 := s.put(1, ZoneMainTheater, unit("Foe Two", 1, 1, Keywords{}))
	s.setPool(0, 0, 1)
	s.scriptRefs(0, nil, u2)
	s.do(activate(a, 0))
	asks := s.asked(0)
	if len(asks) != 1 || asks[0].Kind != DecideTarget {
		t.Fatalf("one target decision is asked, got %d (FR-TGT-001)", len(asks))
	}
	if got, want := sortedCopy(refs(asks[0])), sortedCopy(ids(u1, u2)); !reflect.DeepEqual(got, want) {
		t.Errorf("both opposing units are options: got %v, want %v (FR-TGT-001)", got, want)
	}
	if s.alive(u2) || !s.alive(u1) {
		t.Errorf("the chosen unit dies and the other survives (FR-TGT-001)")
	}
	if s.g.Players[0].Tac != 0 {
		t.Errorf("the cost was paid: Tactics %d, want 0", s.g.Players[0].Tac)
	}
}

func TestFR_KWD_002h_WardBlocksActivatedTarget(t *testing.T) {
	card := func() *Card {
		return actUnit("Assassin", nil, ability.Activated{Effect: killChosenOpposing()})
	}
	t.Run("another unit is offered", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, card())
		p1 := s.put(1, ZoneMainTheater, unit("Plain", 1, 1, Keywords{}))
		p2 := s.put(1, ZoneMainTheater, unit("Plain Two", 1, 1, Keywords{}))
		w := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		asks := s.asked(0)
		if len(asks) != 1 {
			t.Fatalf("two non-Warded options are asked once, got %d decisions", len(asks))
		}
		if got, want := sortedCopy(refs(asks[0])), sortedCopy(ids(p1, p2)); !reflect.DeepEqual(got, want) {
			t.Errorf("an opposing Ward unit is not an option for an activated ability: got %v, want %v (FR-KWD-002h)", got, want)
		}
		if !s.alive(w) {
			t.Errorf("the Ward unit is untouched (FR-KWD-002h)")
		}
	})
	t.Run("one plain unit and a Ward unit", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, card())
		p1 := s.put(1, ZoneMainTheater, unit("Plain", 1, 1, Keywords{}))
		w := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		for _, d := range s.asked(0) {
			if !reflect.DeepEqual(refs(d), []int{p1.ID}) {
				t.Errorf("only the plain unit is an option: got %v (FR-KWD-002h)", refs(d))
			}
		}
		if s.alive(p1) || !s.alive(w) {
			t.Errorf("the plain unit dies and the Ward unit survives (FR-KWD-002h)")
		}
	})
}

func TestFR_TGT_003_FizzledAbilityLeavesSourceOnBattlefield(t *testing.T) {
	s := newScene(t)
	sweepFoes := ability.Step{Op: ability.OpKill, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}}
	a := s.put(0, ZoneMainTheater, actUnit("Tapper",
		[]ability.Trigger{{Hook: ability.HookExhausted, Effect: ability.Effect{Steps: []ability.Step{sweepFoes}}}},
		ability.Activated{Costs: []ability.Cost{actExhaust}, Effect: killChosenOpposing()}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 1, Keywords{}))
	s.script(0, Answer{Picks: []int{0}})
	s.do(activate(a, 0))
	if s.alive(foe) {
		t.Fatalf("precondition: the exhaust trigger removed the chosen unit first")
	}
	if !s.alive(a) {
		t.Errorf("when an ability on a battlefield asset fizzles the asset remains on the battlefield (FR-TGT-003)")
	}
	s.wantSettled("a fizzled activated ability")
}

// --- T-0163: an activated ability needs its required targets (SVY-R36) ---

// actPileFilter is "chosen allied Fungus unit card in your discard pile".
var actPileFilter = ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit,
	Subtypes: []string{"Fungus"}, Zone: ability.ZoneDiscard}

// actRecaller is a Fungus unit whose ability is "Kill this unit: return chosen
// allied Fungus unit card in your discard pile to your hand". Its own card
// carries the filter's subtype, so it would match if it counted (SVY-R2).
func actRecaller() *Card {
	c := actUnit("Recaller", nil, ability.Activated{
		Costs:  []ability.Cost{actKill},
		Effect: ability.Effect{Targets: []ability.Target{{Filter: actPileFilter}}, Steps: []ability.Step{ptStep(ability.OpReturnToHand, 0)}},
	})
	c.Subtypes = []string{"Fungus"}
	return c
}

// actNotOffered checks ability i of a is neither offered nor accepted, and that
// the rejection changes nothing the cost would have changed.
func actNotOffered(t *testing.T, s *scene, a *Asset, i int, why string) {
	t.Helper()
	p := s.g.Players[a.Controller]
	req, tac, pile, hand := p.Req, p.Tac, len(p.Discard), len(p.Hand)
	if offersActivate(s.g, a, i) {
		t.Errorf("%s: the ability is not offered (SVY-R36, FR-ACT-001)", why)
	}
	s.reject(activate(a, i))
	if !s.alive(a) || a.Exhausted || p.Req != req || p.Tac != tac || len(p.Discard) != pile || len(p.Hand) != hand {
		t.Errorf("%s: a rejected Activate pays nothing and moves nothing: alive %v exhausted %v R%d T%d pile %d hand %d (SVY-R36)",
			why, s.alive(a), a.Exhausted, p.Req, p.Tac, len(p.Discard), len(p.Hand))
	}
	if n := len(s.asked(a.Controller)); n != 0 {
		t.Errorf("%s: no decision is asked, got %d (SVY-R36)", why, n)
	}
}

func TestFR_ACT_001_NoLegalTargetCannotActivate(t *testing.T) {
	// Covers: SVY-R36, SVY-R2, FR-TGT-001
	t.Run("empty pile", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actRecaller())
		actNotOffered(t, s, a, 0, "an empty discard pile, the source's own card not counting")
	})
	t.Run("only non-matching cards", func(t *testing.T) {
		s := newScene(t)
		s.inDiscard(0, unit("Plain", 1, 1, Keywords{}))
		a := s.put(0, ZoneMainTheater, actRecaller())
		actNotOffered(t, s, a, 0, "a pile holding only a card that is not a Fungus")
	})
	t.Run("a cost that would otherwise be paid", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actUnit("Assassin", nil, ability.Activated{
			Costs: []ability.Cost{actRes(1, 1), actExhaust}, Effect: killChosenOpposing()}))
		s.setPool(0, 1, 1)
		actNotOffered(t, s, a, 0, "no opposing unit on the battlefield")
		if p := s.g.Players[0]; p.Req != 1 || p.Tac != 1 {
			t.Errorf("the pool stays R1 T1, got R%d T%d (SVY-R36)", p.Req, p.Tac)
		}
	})
	t.Run("a legal entity appears", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, actRecaller())
		actNotOffered(t, s, a, 0, "an empty discard pile")
		fungus := s.inDiscard(0, fsSub(unit("Spore", 1, 1, Keywords{}), "Fungus"))[0]
		if !offersActivate(s.g, a, 0) {
			t.Fatalf("with a matching card in the pile the ability is offered (SVY-R36)")
		}
		s.script(0, Answer{Picks: []int{0}})
		s.do(activate(a, 0))
		p := s.g.Players[0]
		if !inHand(p, fungus) || inDiscardPile(p, fungus) {
			t.Errorf("the matching card returns to hand (FR-ACT-001)")
		}
		if s.alive(a) || !inDiscardPile(p, a.Inst) {
			t.Errorf("the source is in the discard pile (FR-ACT-001)")
		}
	})
}

func TestFR_ACT_001_UpToSlotStillActivates(t *testing.T) {
	// Covers: SVY-R36, SVY-R11
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Tapper", nil, ability.Activated{
		Costs: []ability.Cost{actRes(1, 0), actExhaust},
		Effect: ability.Effect{
			Targets: []ability.Target{{Filter: ability.Filter{Allegiance: ability.AllegianceOpposing, Kinds: ability.KindUnit}, UpTo: 2}},
			Steps:   []ability.Step{{Op: ability.OpKill, Subject: ability.SubjectTarget}},
		}}))
	s.setPool(0, 1, 0)
	if !offersActivate(s.g, a, 0) {
		t.Fatalf("an \"up to 2\" slot with no legal entity does not stop the ability (SVY-R36)")
	}
	s.do(activate(a, 0))
	if p := s.g.Players[0]; p.Req != 0 || !a.Exhausted {
		t.Errorf("its costs are paid: R%d exhausted %v, want R0 true (FR-ACT-001)", p.Req, a.Exhausted)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("nothing is asked with no legal entity, got %d decisions (SVY-R36)", n)
	}
	if !s.alive(a) {
		t.Errorf("the ability resolved with no target affected and the source stays (FR-TGT-003)")
	}
	s.wantSettled("an up-to activation with no target")
}

func TestFR_ACT_001_WardedOnlyTargetCannotActivate(t *testing.T) {
	// Covers: SVY-R36, FR-KWD-002h, FR-TGT-001
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Assassin", nil, ability.Activated{
		Costs: []ability.Cost{actRes(0, 1)}, Effect: killChosenOpposing()}))
	w := s.put(1, ZoneMainTheater, unit("Warded", 1, 5, Keywords{Ward: true}))
	s.setPool(0, 0, 1)
	actNotOffered(t, s, a, 0, "the only opposing unit is Warded")
	if !s.alive(w) {
		t.Errorf("the Warded unit is untouched (FR-KWD-002h)")
	}
}

func TestFR_ACT_001_AnotherSlotNeedsTwoEntities(t *testing.T) {
	// Covers: SVY-R36, SVY-R34, FR-TGT-002
	card := func() *Card {
		anyOpp := tgt(ability.AllegianceOpposing, ability.KindUnit, false)
		another := anyOpp
		another.Another = 1
		return actUnit("Ambusher", nil, ability.Activated{
			Costs:  []ability.Cost{actExhaust},
			Effect: ability.Effect{Targets: []ability.Target{anyOpp, another}, Steps: []ability.Step{dealerStep(1, 1)}},
		})
	}
	t.Run("one legal entity", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, card())
		s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
		actNotOffered(t, s, a, 0, "one unit cannot fill two slots that must differ")
	})
	t.Run("two legal entities", func(t *testing.T) {
		s := newScene(t)
		a := s.put(0, ZoneMainTheater, card())
		x := s.put(1, ZoneMainTheater, unit("X", 4, 30, Keywords{}))
		y := s.put(1, ZoneMainTheater, unit("Y", 1, 30, Keywords{}))
		if !offersActivate(s.g, a, 0) {
			t.Fatalf("two units can fill the two slots, so the ability is offered (SVY-R34)")
		}
		s.scriptModal(0, nil, nil, nil, pick(x), pick(y))
		s.do(activate(a, 0))
		if y.Damage != 4 || x.Damage != 0 {
			t.Errorf("the two slots took different units, X dealing its power 4 to Y: X damage %d, Y damage %d, want 0, 4 (SVY-R34)", x.Damage, y.Damage)
		}
	})
}

func TestFR_ACT_001_ChooseNNeedsEnoughLegalModes(t *testing.T) {
	// Covers: SVY-R36, FR-MOD-001, SVY-R28
	// Mode 0 is "kill chosen opposing unit", mode 1 "kill chosen other allied unit".
	killOpp := mode([]ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)},
		ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget})
	otherAlly := tgt(ability.AllegianceAllied, ability.KindUnit, false)
	otherAlly.Filter.Other = true
	killAlly := mode([]ability.Target{otherAlly}, ability.Step{Op: ability.OpKill, Subject: ability.SubjectTarget})
	tests := []struct {
		name      string
		n         int
		foe, ally bool
		offered   bool
		modes     []int
	}{
		{"choose one, no legal mode", 1, false, false, false, nil},
		{"choose one, one legal mode", 1, true, false, true, []int{0}},
		{"choose two, one legal mode", 2, true, false, false, nil},
		{"choose two, two legal modes", 2, true, true, true, []int{0, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newScene(t)
			a := s.put(0, ZoneMainTheater, actUnit("Chooser", nil, ability.Activated{
				Costs:  []ability.Cost{actExhaust},
				Effect: ability.Effect{Modes: []ability.Effect{killOpp, killAlly}, ChooseN: tc.n},
			}))
			var foe, ally *Asset
			if tc.foe {
				foe = s.put(1, ZoneMainTheater, unit("Foe", 1, 1, Keywords{}))
			}
			if tc.ally {
				ally = s.put(0, ZoneMainTheater, unit("Friend", 1, 1, Keywords{}))
			}
			if !tc.offered {
				actNotOffered(t, s, a, 0, tc.name)
				return
			}
			if !offersActivate(s.g, a, 0) {
				t.Fatalf("%s: enough legal modes, so the ability is offered (SVY-R36)", tc.name)
			}
			var picks []tpick
			for _, m := range tc.modes {
				if m == 0 {
					picks = append(picks, pick(foe))
				} else {
					picks = append(picks, pick(ally))
				}
			}
			s.scriptModal(0, nil, nil, []Answer{modeAnswer(tc.modes...)}, picks...)
			s.do(activate(a, 0))
			if foe != nil && s.alive(foe) != !actHasMode(tc.modes, 0) {
				t.Errorf("%s: the opposing unit dies exactly when mode 0 was chosen (FR-MOD-001)", tc.name)
			}
			if ally != nil && s.alive(ally) != !actHasMode(tc.modes, 1) {
				t.Errorf("%s: the allied unit dies exactly when mode 1 was chosen (FR-MOD-001)", tc.name)
			}
			if !a.Exhausted {
				t.Errorf("%s: the cost was paid (FR-ACT-001)", tc.name)
			}
		})
	}
}

func actHasMode(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

func TestFR_ACT_001_LegalTargetDoesNotExcuseACost(t *testing.T) {
	// Covers: SVY-R36, FR-TGT-001
	s := newScene(t)
	a := s.put(0, ZoneMainTheater, actUnit("Assassin", nil, ability.Activated{
		Costs: []ability.Cost{actRes(2, 0)}, Effect: killChosenOpposing()}))
	foe := s.put(1, ZoneMainTheater, unit("Foe", 1, 1, Keywords{}))
	s.setPool(0, 1, 0)
	actNotOffered(t, s, a, 0, "a legal target exists but R2 cannot be paid")
	if !s.alive(foe) {
		t.Errorf("the legal target is untouched (SVY-R36)")
	}
}
