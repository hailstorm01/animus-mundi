package engine

import (
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0053: an Attack's combatants are re-verified when it resolves
// (SVY-R8, FRD 01 v3.8: FR-CBT-002 Step 3 "Combatant Target Check", FR-TGT-003).
//
// No starter card changes control, so these scenes use a synthetic attacker with
// an optional "whenever this attacks, you may draw a card" trigger. Its "you may"
// question is put to the attacker's controller at Step 2, after the attack is
// declared and before it resolves, and the scene's chooser changes the board
// (a side or Ward) as it is asked. The question is always answered yes.

// midStack installs a chooser for seat 0 that runs mutate when it is asked the
// "you may" question, then says yes. It orders simultaneous triggers as listed,
// and picks target (if set) for any target question.
func (s *scene) midStack(mutate func(), target *Asset) {
	s.t.Helper()
	s.g.choosers[0] = func(_ *Game, d Decision) Answer {
		switch d.Kind {
		case DecideOptional:
			mutate()
			return yes
		case DecideOrder:
			picks := make([]int, len(d.Options))
			for i := range picks {
				picks[i] = i
			}
			return Answer{Picks: picks}
		case DecideTarget:
			if target != nil {
				for i, o := range d.Options {
					if o.Ref == target.ID {
						return Answer{Picks: []int{i}}
					}
				}
			}
		}
		s.t.Errorf("unexpected %v decision %q", d.Kind, d.Prompt)
		return Answer{Timeout: true}
	}
}

func mayDraw() ability.Trigger { return attackTrig(true, nil, drawN(1)) }

// sides is the board of the side-change scenes: player 0's attacker (power 3)
// attacks Defender (power 2).
type sides struct {
	s        *scene
	att, def *Asset
}

func newSides(t *testing.T) *sides {
	t.Helper()
	s := newScene(t)
	b := &sides{s: s}
	b.att = s.put(0, ZoneMainTheater, striker("Attacker", 3, 9, Keywords{}, mayDraw()))
	b.def = s.put(1, ZoneMainTheater, unit("Defender", 2, 9, Keywords{}))
	return b
}

func (b *sides) attack() { b.s.do(b.s.attack(b.att, b.def)) }

func (b *sides) wantFizzle() {
	b.s.t.Helper()
	if !b.s.alive(b.att) || !b.s.alive(b.def) {
		b.s.t.Fatalf("no combatant should have been destroyed")
	}
	if b.att.Damage != 0 {
		b.s.t.Errorf("the attack fizzles, so there is no retaliation damage to the attacker (FR-CBT-002 SVY-R8): attacker damage %d", b.att.Damage)
	}
	if b.def.Damage != 0 {
		b.s.t.Errorf("the attack fizzles, so there is no combat damage to the defender (FR-CBT-002 SVY-R8): defender damage %d", b.def.Damage)
	}
}

// Covers: FR-TGT-003
func TestFR_CBT_002_DefenderChangesSidesFizzles(t *testing.T) {
	b := newSides(t)
	b.s.midStack(func() { b.def.Controller = 0 }, nil)
	b.attack()
	b.wantFizzle()
}

// Covers: FR-TGT-003
func TestFR_CBT_002_AttackerChangesSidesFizzles(t *testing.T) {
	b := newSides(t)
	b.s.midStack(func() { b.att.Controller = 1 }, nil)
	b.attack()
	b.wantFizzle()
}

// An opposing target of a chosen-target trigger that gains Ward before the
// trigger resolves is no longer legal, so that clause does nothing to it.
func TestFR_TGT_003_TargetThatGainsWardBeforeResolutionIsIllegal(t *testing.T) {
	// Hero has two Step 2 triggers: "deal 3 damage to chosen opposing unit"
	// (target chosen as it is placed) and a "you may" whose answer grants Ward.
	// Covers: FR-KWD-002h
	s := newScene(t)
	hit := attackTrig(false, []ability.Target{tgt(ability.AllegianceOpposing, ability.KindUnit, false)}, dmgTo(3, 0))
	hero := s.put(0, ZoneMainTheater, striker("Hero", 0, 9, Keywords{}, hit, mayDraw()))
	dummy := s.put(1, ZoneMainTheater, unit("Dummy", 0, 9, Keywords{}))
	victim := s.put(1, ZoneMainTheater, unit("Victim", 0, 9, Keywords{}))
	picked := false
	grantAfterPick := func() {
		if !picked {
			t.Fatalf("test setup: Ward must be granted after the target was chosen")
		}
		victim.Inst.Card.Keywords.Ward = true
	}
	s.midStack(grantAfterPick, victim)
	inner := s.g.choosers[0]
	s.g.choosers[0] = func(g *Game, d Decision) Answer {
		if d.Kind == DecideTarget {
			picked = true
		}
		return inner(g, d)
	}
	s.do(s.attack(hero, dummy))
	s.wantDamage(victim, 0)
}

// Both combatants change sides: they are still opposing assets, so the defender
// check passes and the attack resolves for the attacker's current controller.
// Covers: FR-TGT-003
func TestFR_CBT_002_CombatantsSwapSidesStillResolves(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, striker("Hero", 3, 9, Keywords{}, mayDraw()))
	def := s.put(1, ZoneMainTheater, unit("Defender", 2, 9, Keywords{}))
	s.midStack(func() { att.Controller, def.Controller = 1, 0 }, nil)
	s.do(s.attack(att, def))
	s.wantDamage(def, 3) // the defender is still an opposing asset of the attacker's current controller
	s.wantDamage(att, 2)
}
