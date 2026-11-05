package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

// Tests for T-0178: snapshots and damage events report the target's effective
// keywords, which a rule monitor can read. Each TestEffectiveKeywords_Case<N>
// is Acceptance Case N.

// snapOf returns the asset with id in s's Main Theater.
func snapOf(t *testing.T, s Snapshot, id int) AssetSnap {
	t.Helper()
	for _, a := range s.Theater {
		if a.ID == id {
			return a
		}
	}
	t.Fatalf("no asset %d in the Main Theater of the snapshot", id)
	return AssetSnap{}
}

// Case 1: a face-up asset's snapshot and damage events report its keywords, and
// an asset with no keywords reports none.
func TestEffectiveKeywords_Case1_FaceUp(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Att", 3, 9, Keywords{}))
	def := s.put(1, ZoneMainTheater, unit("Tough Wall", 2, 9, Keywords{Resistant: 1, Stalwart: true}))
	snap := s.g.Snapshot()
	if got, want := snapOf(t, snap, def.ID).Keywords, (Keywords{Resistant: 1, Stalwart: true}); got != want {
		t.Errorf("defender's snapshot keywords = %+v, want %+v", got, want)
	}
	if got := snapOf(t, snap, att.ID).Keywords; got != (Keywords{}) {
		t.Errorf("attacker's snapshot keywords = %+v, want none", got)
	}
	s.do(s.attack(att, def))
	dmg := damageEvents(s.g.Log)
	if len(dmg) != 2 {
		t.Fatalf("got %d damage events, want 2", len(dmg))
	}
	if d := dmg[0]; d.Target != def.ID || d.Resistant != 1 || !d.Stalwart {
		t.Errorf("damage to the defender = %+v, want Resistant 1 and Stalwart", d)
	}
	if d := dmg[1]; d.Target != att.ID || d.Resistant != 0 || d.Stalwart {
		t.Errorf("damage to the attacker = %+v, want no Resistant and no Stalwart", d)
	}
}

// Case 2: effective keywords differ from the card's printed ones when the asset
// is face-down, and the event reports them as the damage lands, not as the last
// snapshot saw them.
func TestEffectiveKeywords_Case2_FaceDownDiffersFromTheCard(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, unit("Src", 1, 9, Keywords{}))
	wall := s.put(1, ZoneMainTheater, unit("Tough Wall", 2, 9, Keywords{Resistant: 1, Stalwart: true}))
	if got := snapOf(t, s.g.Snapshot(), wall.ID).Keywords.Resistant; got != 1 {
		t.Fatalf("face-up snapshot Resistant = %d, want 1", got)
	}
	wall.FaceDown = true
	s.g.dealDamage(src, wall, 3, true)
	dmg := damageEvents(s.g.Log)
	if len(dmg) != 1 {
		t.Fatalf("got %d damage events, want 1", len(dmg))
	}
	if d := dmg[0]; d.Resistant != 0 || d.Stalwart || d.Amount != 3 {
		t.Errorf("damage to the face-down wall = %+v, want no Resistant, no Stalwart, amount 3", d)
	}
	a := snapOf(t, s.g.Snapshot(), wall.ID)
	if a.Card != "Tough Wall" {
		t.Errorf("snapshot Card = %q, want the real name", a.Card)
	}
	if want := ScratchFace().Keywords; a.Keywords != want {
		t.Errorf("face-down snapshot keywords = %+v, want the Scratch Company face's %+v", a.Keywords, want)
	}
}

// Case 3: the JSON names, so written logs keep them. A negative Resistant is
// written, and a target with neither keyword writes neither field.
func TestEffectiveKeywords_Case3_JSON(t *testing.T) {
	s := newScene(t)
	src := s.put(0, ZoneMainTheater, unit("Src", 1, 9, Keywords{}))
	frail := s.put(1, ZoneMainTheater, unit("Frail", 1, 9, Keywords{Resistant: -1, Stalwart: true}))
	s.g.dealDamage(src, frail, 1, false)
	s.g.dealDamage(frail, src, 1, false)
	dmg := damageEvents(s.g.Log)
	if len(dmg) != 2 {
		t.Fatalf("got %d damage events, want 2", len(dmg))
	}
	to, err := json.Marshal(dmg[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"resistant":-1`, `"stalwart":true`} {
		if !strings.Contains(string(to), want) {
			t.Errorf("damage to Frail lacks %s: %s", want, to)
		}
	}
	from, err := json.Marshal(dmg[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, not := range []string{"resistant", "stalwart"} {
		if strings.Contains(string(from), not) {
			t.Errorf("damage to Src writes %s: %s", not, from)
		}
	}
	snap, err := json.Marshal(s.g.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if want := `"keywords":{"resistant":-1,"stalwart":true}`; !strings.Contains(string(snap), want) {
		t.Errorf("snapshot lacks %s: %s", want, snap)
	}
	var back Snapshot
	if err := json.Unmarshal(snap, &back); err != nil {
		t.Fatal(err)
	}
	if got, want := snapOf(t, back, frail.ID).Keywords, (Keywords{Resistant: -1, Stalwart: true}); got != want {
		t.Errorf("snapshot keywords after a round trip = %+v, want %+v", got, want)
	}
}
