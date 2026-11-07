package engine

import (
	"fmt"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0191 (directive #129): an effect logs one event per asset or card
// it acts on. Cards are synthetic. Player 0 is the active player.

// effectEvents returns the events of kind logged from Log index from on.
func effectEvents(g *Game, from int, kind string) []Event {
	var out []Event
	for _, e := range g.Log[from:] {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// onAlliedKinds is a step acting on every allied asset of the given kinds.
func onAlliedKinds(op ability.Op, n int, kinds ability.Kinds) ability.Step {
	return ability.Step{Op: op, Amount: ability.Value{N: n}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: kinds}}
}

// effectSource is a unit in the Main Theater whose free ability runs step.
func (s *scene) effectSource(step ability.Step) *Asset {
	return s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(nil, step)))
}

func (s *scene) wall() *Asset {
	return s.put(0, ZoneMainTheater, structure("Wall", 0, 9, Keywords{}))
}

// wantAssetEvents checks one event per asset, in order: its target, its amount
// and its line, with no acting player and no card.
func wantAssetEvents(t *testing.T, got []Event, as []*Asset, amounts []int, format string) {
	t.Helper()
	if len(got) != len(as) {
		t.Fatalf("%d events, want one per asset (%d): %+v", len(got), len(as), got)
	}
	for i, a := range as {
		e := got[i]
		if e.Target != a.ID || e.Amount != amounts[i] || e.Player != -1 || e.Card != "" {
			t.Errorf("event %d: target %d, amount %d, player %d, card %q; want target %d, amount %d, player -1, no card",
				i, e.Target, e.Amount, e.Player, e.Card, a.ID, amounts[i])
		}
		if want := fmt.Sprintf(format, a, amounts[i]); e.Detail != want {
			t.Errorf("event %d: detail %q, want %q", i, e.Detail, want)
		}
	}
}

func TestEffectEvents_HealLogsTheDamageRemoved(t *testing.T) {
	s := newScene(t)
	src := s.effectSource(onAlliedKinds(ability.OpHeal, 2, ability.KindStructure))
	hurt, scratched, whole := s.wall(), s.wall(), s.wall()
	hurt.Damage, scratched.Damage = 3, 1
	from := len(s.g.Log)
	s.do(activate(src, 0))
	wantAssetEvents(t, effectEvents(s.g, from, EventHeal), []*Asset{hurt, scratched, whole}, []int{2, 1, 0}, "%s heals %d")
	if hurt.Damage != 1 || scratched.Damage != 0 || whole.Damage != 0 {
		t.Errorf("damage %d, %d, %d after Heal 2, want 1, 0, 0", hurt.Damage, scratched.Damage, whole.Damage)
	}
}

func TestEffectEvents_ArmorLogsTheArmorAdded(t *testing.T) {
	s := newScene(t)
	src := s.effectSource(onAlliedKinds(ability.OpAddArmor, 2, ability.KindStructure))
	bare, plated := s.wall(), s.wall()
	plated.Armor = 3
	from := len(s.g.Log)
	s.do(activate(src, 0))
	wantAssetEvents(t, effectEvents(s.g, from, EventArmor), []*Asset{bare, plated}, []int{2, 2}, "%s gains %d Armor")
	if bare.Armor != 2 || plated.Armor != 5 {
		t.Errorf("Armor %d, %d, want 2, 5", bare.Armor, plated.Armor)
	}
}

func TestEffectEvents_PowerLogsThePowerChange(t *testing.T) {
	for _, n := range []int{2, -1} {
		t.Run(fmt.Sprintf("%+d", n), func(t *testing.T) {
			s := newScene(t)
			src := s.effectSource(onAlliedKinds(ability.OpModifyStat, n, ability.KindStructure))
			a, b := s.wall(), s.wall()
			from := len(s.g.Log)
			s.do(activate(src, 0))
			wantAssetEvents(t, effectEvents(s.g, from, EventPower), []*Asset{a, b}, []int{n, n}, "%s gets %+d power until end of turn")
			if a.TempPower != n || b.TempPower != n {
				t.Errorf("TempPower %d, %d, want %d", a.TempPower, b.TempPower, n)
			}
		})
	}
}

func TestEffectEvents_FaceDownAssetStaysHidden(t *testing.T) {
	for _, c := range []struct {
		op   ability.Op
		kind string
	}{
		{ability.OpHeal, EventHeal},
		{ability.OpAddArmor, EventArmor},
		{ability.OpModifyStat, EventPower},
	} {
		t.Run(c.kind, func(t *testing.T) {
			s := newScene(t)
			src := s.effectSource(onAlliedKinds(c.op, 1, ability.KindUnit))
			fd := s.putScratch(0)
			from := len(s.g.Log)
			s.do(activate(src, 0))
			found := false
			for _, e := range effectEvents(s.g, from, c.kind) {
				if e.Target != fd.ID {
					continue
				}
				found = true
				if strings.Contains(e.Detail, "Filler") || e.Card != "" || !strings.HasPrefix(e.Detail, fd.String()) {
					t.Errorf("the event names the hidden card: card %q, detail %q", e.Card, e.Detail)
				}
			}
			if !found {
				t.Fatalf("no %s event for the face-down asset", c.kind)
			}
		})
	}
}

func TestEffectEvents_CardMoveNamesTheCardWhileItIsPublic(t *testing.T) {
	pile := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	for _, c := range []struct {
		op   ability.Op
		zone Zone
		text string
	}{
		{ability.OpReturnToHand, ZoneHand, "Relic returns to its owner's hand"},
		{ability.OpPutOnDeck, ZoneDeck, "Relic is put into its owner's deck"},
	} {
		t.Run(c.zone.String(), func(t *testing.T) {
			s := newScene(t)
			inst := s.inDiscard(0, unit("Relic", 1, 1, Keywords{}))[0]
			a := s.put(0, ZoneMainTheater, actUnit("Recaller", nil, ability.Activated{
				Effect: ability.Effect{Targets: []ability.Target{pile}, Steps: []ability.Step{ptStep(c.op, 0)}},
			}))
			s.script(0, Answer{Picks: []int{0}})
			inPileAtEvent := false
			s.g.obs = &recorder{onEvent: func(e Event) {
				if e.Kind == EventCardMove {
					inPileAtEvent = inDiscardPile(s.g.Players[0], inst)
				}
			}}
			from := len(s.g.Log)
			s.do(activate(a, 0))
			got := effectEvents(s.g, from, EventCardMove)
			if len(got) != 1 {
				t.Fatalf("%d card-move events, want 1: %+v", len(got), got)
			}
			e := got[0]
			if e.Card != "Relic" || e.Inst != inst.ID || e.Zone != c.zone.String() || e.Player != -1 || e.Detail != c.text {
				t.Errorf("event %+v, want card Relic, inst %d, zone %s, player -1, detail %q", e, inst.ID, c.zone, c.text)
			}
			if !inPileAtEvent {
				t.Errorf("the event was logged after the card left the discard pile")
			}
			if inDiscardPile(s.g.Players[0], inst) {
				t.Errorf("the card is still in the discard pile")
			}
		})
	}
}

// A card an earlier step moved out of the discard pile is gone (FR-TGT-003): the
// step doesn't move it again, doesn't count it and logs nothing, so a later move
// never names a card that went into a hand.
func TestEffectEvents_CardNoLongerInTheDiscardPileLogsNothing(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, unit("Relic", 1, 1, Keywords{}))
	r := &resolution{it: &StackItem{Controller: 0}, cards: [][]*Instance{{inst}}}
	from := len(s.g.Log)
	if n := s.g.runStep(r, ptStep(ability.OpPutOnDeck, 0)); n != 0 {
		t.Errorf("the step did %d things, want 0", n)
	}
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the card left the hand")
	}
	if got := effectEvents(s.g, from, EventCardMove); len(got) != 0 {
		t.Errorf("card-move events %+v, want none", got)
	}
}
