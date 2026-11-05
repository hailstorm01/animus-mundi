---
id: T-0178
title: Snapshots and damage events report effective keywords (directive #116 (a), harness, flash)
kind: harness
tier: flash
wave: D.1 (board directive #116 (a), structural review 8, finding 1)
status: ready
branch: m2/autoloop-wave-d.1/T-0178-effective-keywords-in-records
depends_on: [T-0177]
claims: [field:AssetSnap.Keywords, field:Event.Resistant, field:Event.Stalwart, json:Keywords, test:TestEffectiveKeywords_Case1_FaceUp, test:TestEffectiveKeywords_Case2_FaceDownDiffersFromTheCard, test:TestEffectiveKeywords_Case3_JSON]
cards: []
requirements: []
---

## Goal
A rule monitor can read an asset's effective keywords without working them out from its card. `AssetSnap` carries the asset's effective `Keywords`, and a damage event carries the target's effective Resistant and Stalwart as the damage lands. No monitor reads them yet: T-0179 (#116 (b)) does that.

## Inputs
- Board directive #116 (structural review 8, approved by the user, 2026-11-05). This task is its part (a); T-0179, `queued`, is its part (b).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1). ADR-0005: "Adding a field is additive", and "Events have JSON tags".
- Reuse search (D2.3): `g.Keywords` (`layers.go`) is the one read of effective keywords, and `dealDamage` already holds its result in `kw`, so nothing computes them a second time. `EffectiveState` (`state.go`) is a leave-record for last-known information and carries no IDs, so it isn't a snapshot. No field on `AssetSnap` or `Event` says this today.
- **What a test can vary today.** `g.Keywords` returns `a.Printed().Keywords`, so nothing yet makes an asset's effective keywords differ from its current face. #116 allows "a test-only card or state", and the state is a face-down asset: its card prints Resistant and Stalwart, and its effective keywords are the Scratch Company face's. Case 2 flips an asset face-down after a snapshot, which is also why the event carries the values: they can change between the last settled snapshot and the damage. #117 adds the first real grant, with its own tests.
- **JSON.** `Keywords` has no tags today because nothing writes it. Snapshots go to written logs, so it gets lower-camel `omitempty` tags like the rest of the schema, and an asset with no keywords writes `"keywords":{}`. `Replay` compares recorded and replayed snapshots after a JSON round trip, so the round trip has to hold; the existing replay tests and Case 3 check that. A negative Resistant is non-zero, so `omitempty` keeps it.
- Kept on purpose: `eventlog`'s divergence check doesn't compare `Raw`, `Combat`, `Armor` or `Absorbed` field by field, because `Detail` already prints the stages. The two new fields follow them, so `replay.go` doesn't change.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 26043a1. With only the test file, `go test ./pkg/engine -run TestEffectiveKeywords` exits 1 on `has no field or method Keywords`. With the edits: `gofmt -l pkg` clean, the three tests pass, `go tool task check` exits 0, `go tool task sim` exits 0, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are from the files as they stand at 26043a1.

### 1. `pkg/engine/effective_keywords_test.go`, new, in the `test(T-0178)` commit
Create it with exactly this content, ending in one line break:

```go
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
```

### 2. `pkg/engine/card.go`, lines 30–38, in the `feat(engine)` commit
Replace the body of `Keywords` with:
```go
	Resistant        int  `json:"resistant,omitempty"`
	Stalwart         bool `json:"stalwart,omitempty"`
	Aggressive       int  `json:"aggressive,omitempty"`
	Ethereal         bool `json:"ethereal,omitempty"`
	Ward             bool `json:"ward,omitempty"`
	Versatile        bool `json:"versatile,omitempty"`
	Guardian         bool `json:"guardian,omitempty"`
	Armor            int  `json:"armor,omitempty"`
	DeploysExhausted bool `json:"deploysExhausted,omitempty"`
```

### 3. `pkg/engine/snapshot.go`, in the `feat(engine)` commit
1. Replace line 14 with:
   ```go
   // when the asset is face-down; Power, MaxHealth and Keywords are effective.
   ```
2. Replace lines 16–28, the body of `AssetSnap`, with:
   ```go
   	ID         int      `json:"id"`
   	Inst       int      `json:"inst"`
   	Card       string   `json:"card"`
   	Owner      int      `json:"owner"`
   	Controller int      `json:"controller"`
   	Zone       string   `json:"zone"`
   	Exhausted  bool     `json:"exhausted"`
   	FaceDown   bool     `json:"faceDown"`
   	Damage     int      `json:"damage"`
   	Armor      int      `json:"armor"`
   	TempPower  int      `json:"tempPower"`
   	Power      int      `json:"power"`
   	MaxHealth  int      `json:"maxHealth"`
   	Keywords   Keywords `json:"keywords"`
   ```
3. Replace line 63 with:
   ```go
   		TempPower: a.TempPower, Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: g.Keywords(a),
   ```

### 4. `pkg/engine/events.go`, lines 222–225, in the `feat(engine)` commit
Replace those four lines, the fields under `// Damage events only.`, with:
```go
	Raw       int  `json:"raw,omitempty"`       // the incoming damage before Stage 2
	Combat    bool `json:"combat,omitempty"`    // combat damage (Stalwart applies)
	Armor     int  `json:"armor,omitempty"`     // the target's Armor pool before this damage
	Absorbed  int  `json:"absorbed,omitempty"`  // Stage 4: damage absorbed by Armor
	Resistant int  `json:"resistant,omitempty"` // the target's effective Resistant as the damage lands
	Stalwart  bool `json:"stalwart,omitempty"`  // the target's effective Stalwart as the damage lands
```

### 5. `pkg/engine/damage.go`, line 58, in the `feat(engine)` commit
Replace it with:
```go
		Raw: raw, Combat: combat, Armor: armorBefore, Absorbed: r.Absorbed, Resistant: kw.Resistant, Stalwart: kw.Stalwart},
```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
Edit 1 is the whole of the test commit: commit it as `test(T-0178): snapshots and damage events report effective keywords`, run case 1, then make edits 2–5. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Fails before the edits** (after the test commit, before edit 2): `go -C W test ./pkg/engine -run TestEffectiveKeywords` exits 1, and its output has `has no field or method Keywords`.
2. **Passes after them** (#116 (a)): `go -C W test ./pkg/engine -run TestEffectiveKeywords -v` exits 0, and its output has `--- PASS` for `TestEffectiveKeywords_Case1_FaceUp`, `TestEffectiveKeywords_Case2_FaceDownDiffersFromTheCard` and `TestEffectiveKeywords_Case3_JSON`.
3. **A search and a full game still replay:** `go -C W test ./pkg/eventlog ./pkg/sim` exits 0, with no test edited.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `9	9	pkg/engine/card.go`, `1	1	pkg/engine/damage.go`, `123	0	pkg/engine/effective_keywords_test.go`, `6	4	pkg/engine/events.go` and `16	15	pkg/engine/snapshot.go`.
5. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those five files and this task file.
6. **Sims byte-identical** (#116): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
7. `gofmt -l W/pkg/engine` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/card.go`
- `pkg/engine/snapshot.go`
- `pkg/engine/events.go`
- `pkg/engine/damage.go`
- `pkg/engine/effective_keywords_test.go` (new, edit 1 only)

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 9): no `main` commits to sync. Wrote this task from board directive #116 (a), next in board order, and T-0179 `queued` for (b) (D2.4). Not `WAVE-COMPLETE`: #116–#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: the finds are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
