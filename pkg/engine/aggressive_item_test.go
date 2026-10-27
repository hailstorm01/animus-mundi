package engine

import (
	"reflect"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0074 (board directive #60 part 1): Aggressive N is an ability.Trigger
// placed as an ordinary StackAbility item marked with the keyword it came from.

// recordPlaced makes s record a copy of every item at the moment it is placed.
func recordPlaced(s *scene) *[]StackItem {
	var items []StackItem
	r := &recorder{}
	r.onEvent = func(e Event) {
		if e.Kind == "place" {
			items = append(items, *s.g.Stack[len(s.g.Stack)-1])
		}
	}
	s.g.obs = r
	return &items
}

func placeDepths(log []Event) []int {
	var out []int
	for _, e := range eventsOfKind(log, "place") {
		out = append(out, e.Amount)
	}
	return out
}

func raiderWithPrinted(s *scene) (att, def *Asset) {
	att = s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{Aggressive: 2}, attackTrigger(scratchN(1))))
	def = s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	return att, def
}

func TestFR_CBT_002_AggressiveIsListedFirstAndPlacedFirstByDefault(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	att, def := raiderWithPrinted(s)
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	d := s.asked(0)
	if len(d) != 1 || d[0].Kind != DecideOrder || len(d[0].Options) != 2 || d[0].Options[0].Label != "Raider: Aggressive 2" {
		t.Fatalf("want one DecideOrder with option 0 %q: %+v", "Raider: Aggressive 2", d)
	}
	if got := placeDepths(s.g.Log); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("place depths = %v, want [1 2 3]", got)
	}
	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].Keyword != ability.KeywordAggressive || (*items)[2].Keyword != ability.KeywordNone {
		t.Errorf("placed items = %+v, want Attack, Aggressive, printed", *items)
	}
}

func TestFR_TRG_003_AggressiveCanBePlacedLast(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{1, 0}})
	att, def := raiderWithPrinted(s)
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	if len(*items) != 3 || (*items)[1].Keyword != ability.KeywordNone || (*items)[2].Keyword != ability.KeywordAggressive {
		t.Errorf("placed items = %+v, want Attack, printed, Aggressive on top", *items)
	}
}

func TestFR_CBT_002_AggressiveItemIsAnAbilityItemWithTheKeyword(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	att, def := raiderWithPrinted(s)
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	ag, printed := (*items)[1], (*items)[2]
	if ag.Kind != StackAbility || ag.Keyword != ability.KeywordAggressive || ag.Source != att.ID || ag.Controller != 0 ||
		!reflect.DeepEqual(ag.Effect, ability.Aggressive(2).Effect) {
		t.Errorf("Aggressive item = %+v", ag)
	}
	if printed.Keyword != ability.KeywordNone {
		t.Errorf("printed item has keyword %v", printed.Keyword)
	}
	pl := eventsOfKind(s.g.Log, "place")
	if want := "player 0 places Aggressive trigger of Raider on the stack"; pl[1].Detail != want {
		t.Errorf("Aggressive place Detail = %q, want %q", pl[1].Detail, want)
	}
	if want := "player 0 places ability of Raider on the stack"; pl[2].Detail != want {
		t.Errorf("printed place Detail = %q, want %q", pl[2].Detail, want)
	}
}

func TestFR_KWD_002f_AggressiveConstructor(t *testing.T) {
	tr := ability.Aggressive(3)
	want := ability.Trigger{Hook: ability.HookAttacks, Keyword: ability.KeywordAggressive,
		Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpModifyStat, Subject: ability.SubjectSelf, Amount: ability.Value{N: 3}}}}}
	if !reflect.DeepEqual(tr, want) {
		t.Errorf("Aggressive(3) = %+v, want %+v", tr, want)
	}
	if tr.Optional || tr.Cornered || len(tr.Effect.Targets) != 0 {
		t.Errorf("Aggressive(3) must not be Optional, Cornered or targeted: %+v", tr)
	}
	if ability.KeywordAggressive.String() != "Aggressive" || ability.KeywordNone.String() != "" {
		t.Errorf("Keyword names = %q, %q", ability.KeywordAggressive, ability.KeywordNone)
	}
}

func TestFR_KWD_002f_AggressiveIsReadWhenTheAttackHappens(t *testing.T) {
	// (i) printed Aggressive 2, removed before the attack.
	s := newScene(t)
	c := unit("Brawler", 1, 9, Keywords{Aggressive: 2})
	att := s.put(0, ZoneMainTheater, c)
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	c.Keywords.Aggressive = 0
	items := recordPlaced(s)
	s.do(s.attack(att, def))
	if len(*items) != 1 {
		t.Errorf("no Aggressive item expected: placed %+v", *items)
	}
	s.wantDamage(def, 1)

	// (ii) none printed, granted before the attack.
	s = newScene(t)
	c = unit("Brawler", 1, 9, Keywords{})
	att = s.put(0, ZoneMainTheater, c)
	def = s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	c.Keywords.Aggressive = 2
	items = recordPlaced(s)
	s.do(s.attack(att, def))
	if len(*items) != 2 || (*items)[1].Keyword != ability.KeywordAggressive {
		t.Errorf("one Aggressive item expected: placed %+v", *items)
	}
	s.wantDamage(def, 3)
}

func TestFR_KWD_002f_AggressiveResolvesAndExpires(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Brawler", 1, 9, Keywords{Aggressive: 2}))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 10, Keywords{}))
	s.do(s.attack(att, def))
	s.wantDamage(def, 3)
	s.endTurn()
	if p := s.g.Power(att); p != 1 {
		t.Errorf("power after end of turn = %d, want 1", p)
	}
}

func TestFR_ZON_004_AggressiveFizzlesWhenTheAttackerLeaves(t *testing.T) {
	// Picks [0 1]: Aggressive is placed first (below), the printed self-damage
	// trigger resolves first and kills the attacker.
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0, 1}})
	suicide := attackTrigger(ability.Step{Op: ability.OpDealDamage, Subject: ability.SubjectSelf, Amount: ability.Value{N: 20}})
	att := s.put(0, ZoneMainTheater, striker("Raider", 1, 9, Keywords{Aggressive: 2}, suicide))
	def := s.put(1, ZoneMainTheater, unit("Dummy", 0, 20, Keywords{}))
	s.do(s.attack(att, def))
	if s.alive(att) {
		t.Fatalf("precondition: the attacker dies to its own trigger")
	}
	if len(s.g.Stack) != 0 {
		t.Errorf("stack not empty: %+v", s.g.Stack)
	}
}
