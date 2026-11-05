---
id: T-0179
title: Rule monitors read effective keywords (directive #116 (b))
kind: harness
tier: flash
wave: D.1 (board directive #116 (b), structural review 8, finding 1)
status: done
branch: m2/autoloop-wave-d.1/T-0179-monitors-read-effective-keywords
depends_on: [T-0178]
claims: [monitor:effective keywords, test:TestMonitorKeywords_Case1_Attack, test:TestMonitorKeywords_Case2_Versatile, test:TestMonitorKeywords_Case3_Ward, test:TestMonitorKeywords_Case4_Damage, test:TestMonitorKeywords_Case5_TargetInNoSnapshot]
cards: []
requirements: []
---

## Goal
The rule monitors read the effective keywords T-0178 exposes, so a keyword that differs from the printed one raises no false violation, and damage that breaks the effective value does raise one.

## Inputs
- Board directive #116 (structural review 8, approved by the user, 2026-11-05). This task is its part (b); T-0178 built part (a). The entry-state monitor (`monitor_entry.go`) keeps reading printed keywords, because `newAsset` does.
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1), and the Rule monitors paragraph, which the lead updates when this merges.
- **Where each keyword is read.** Ethereal and Guardian (`checkAttack`), Versatile (`checkGenerate`) and Ward (`checkTargetDecision`) come from the settled snapshot's `Keywords`. Resistant and Stalwart (`checkDamage`) come from the damage event, because they can change between the last settled snapshot and the damage. `m.card` stays for a card's type and yields, and for the entry-state monitor.
- **The damage check no longer waits for a snapshot.** It wanted a snapshot only to look up the target's card. The event now has everything it reads, so it runs at once, and its `pendingCheck` and its "not in its owner's decklist" failure go: the check doesn't read the decklist any more. The attack, resources and entry-state monitors keep their own decklist failures. A damage event doesn't name its target's card, so the message names the asset ID, like the other messages in that file. One thing gets stricter: a target that was in neither the last snapshot nor the next one used to skip the check, and now it doesn't (Case 5).
- **Coverage is printed by `cardtown sim`,** so it is part of `sim-compare`. The dry run shows the counts don't change: no starter or stress game damages an asset that no snapshot shows.
- **Setup of the existing tests (D2.4).** With the monitor edits and no setup change, these fail, because their hand-built snapshots and events carry no keywords: `TestMonitor_Case1_AttackDeclaration`, `TestMonitor_Case1_Resources`, `TestDamage_Case3_DamageSequence` and `TestDamage_Case3_Targeting`. No test body changes. Two setup helpers change instead, so a hand-built record reports what the engine reports for a face-up asset (edits 2 and 3): `snapAsset` fills `Keywords` from the card's name, through a new `cardKeywords`, and `damageCase` fills a damage event's `Resistant` and `Stalwart` from its target in the snapshot. These are the only edits to existing test code, and they land in the `test(T-0179)` commit.
- Reuse search (D2.3): nothing computes effective keywords in `pkg/sim`, and these edits add no copy. `withKeywords` is one line of test setup.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 5de081f. With only edits 1–3, all five `TestMonitorKeywords` tests fail and `TestMonitor_Case1` and `TestDamage_Case3` pass. With edits 4–7 as well: `gofmt -l pkg/sim` clean, every test in `pkg/sim` and `pkg/eventlog` passes, `go tool task check` exits 0, `go tool task sim` exits 0 with no monitor violation, and `go tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`. Every expected output in the cases below is from that dry run.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are from the files as they stand at 5de081f.

### 1. `pkg/sim/monitor_keywords_test.go`, new, in the `test(T-0179)` commit
Create it with exactly this content, ending in one line break:

```go
package sim

import (
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0179: the monitors read the effective keywords the engine
// reports, not the card's printed ones. Each asset here has keywords that
// differ from its card's. Each TestMonitorKeywords_Case<N> is Acceptance Case N.

// withKeywords is a with effective keywords kw.
func withKeywords(a engine.AssetSnap, kw engine.Keywords) engine.AssetSnap {
	a.Keywords = kw
	return a
}

// Case 1: an attack's defender is legal by the snapshot's Ethereal and Guardian.
func TestMonitorKeywords_Case1_Attack(t *testing.T) {
	att := snapAsset(10, "Vanilla", 0, zMT)
	foe := snapAsset(11, "Vanilla", 1, zMT)

	ethereal := withKeywords(snapAsset(12, "Vanilla", 1, zMT), engine.Keywords{Ethereal: true})
	f, seq := attackFeed(baseSnap(0, 0, att, ethereal), 10, 12)
	wantViolation(t, "defender is Ethereal, its card is not", f.m, "attack-declaration", attackRules, seq)

	guardian := withKeywords(snapAsset(12, "Vanilla", 1, zMT), engine.Keywords{Guardian: true})
	f, seq = attackFeed(baseSnap(0, 0, att, foe, guardian), 10, 11)
	wantViolation(t, "defender beside a Guardian whose card is not one", f.m, "attack-declaration", attackRules, seq)

	ghost := withKeywords(snapAsset(15, "Ghost", 1, zMT), engine.Keywords{})
	f, _ = attackFeed(baseSnap(0, 0, att, ghost), 10, 15)
	wantClean(t, "defender's card is Ethereal, it is not", f.m)

	bodyguard := withKeywords(snapAsset(16, "Bodyguard", 1, zMT), engine.Keywords{})
	f, _ = attackFeed(baseSnap(0, 0, att, foe, bodyguard), 10, 11)
	wantClean(t, "defender beside a Guardian card that is not a Guardian now", f.m)
}

// Case 2: Tactics from Reserves is legal by the snapshot's Versatile.
func TestMonitorKeywords_Case2_Versatile(t *testing.T) {
	generate := func(a engine.AssetSnap) (*feed, int) {
		f := newFeed()
		f.settle(baseSnap(0, 0, a))
		seq := f.event(engine.Event{Kind: "action", Detail: "Generate Tactics", Player: 0, Source: a.ID, Resource: "Tactics"})
		a.Exhausted = true
		f.settle(baseSnap(0, 1, a))
		return f, seq
	}
	f, _ := generate(withKeywords(snapAsset(20, "Vanilla", 0, zRes), engine.Keywords{Versatile: true}))
	wantClean(t, "Versatile, its card is not", f.m)
	wantCounts(t, "Versatile, its card is not", f.m, "resources", "FR-KWD-002i")

	f, seq := generate(withKeywords(snapAsset(20, "Sneak", 0, zRes), engine.Keywords{}))
	wantViolation(t, "its card is Versatile, it is not", f.m, "resources", resourceRules, seq)
}

// Case 3: a Target decision's options are legal by the snapshot's Ward.
func TestMonitorKeywords_Case3_Ward(t *testing.T) {
	target := func(refs ...int) engine.Event {
		return engine.Event{Kind: "decision", Player: 0,
			Decision: &engine.DecisionRecord{Kind: "Target", Player: 0, Refs: refs}}
	}
	warded := withKeywords(snapAsset(11, "Vanilla", 1, zMT), engine.Keywords{Ward: true})
	f, seq := damageCase(baseSnap(0, 0, warded), target(11))
	wantViolation(t, "offered a Ward asset whose card has none", f.m, "targeting", targetingRules, seq)

	unwarded := withKeywords(snapAsset(12, "Warded", 1, zMT), engine.Keywords{})
	f, _ = damageCase(baseSnap(0, 0, unwarded), target(12),
		engine.Event{Kind: "place", Player: 0, Detail: "player 0 places ability on the stack"})
	wantClean(t, "offered an asset whose card has Ward, while it has none", f.m)
}

// Case 4: a damage event is checked by its own Resistant and Stalwart, whatever
// the card prints and whatever the last settled snapshot said.
func TestMonitorKeywords_Case4_Damage(t *testing.T) {
	// feedDamage settles attacker 10 (power 3) and defender 11 of the given card,
	// which the snapshot shows with no keywords, then sends the events as built.
	feedDamage := func(card string, evs ...engine.Event) (*feed, int) {
		f := damageFeed()
		f.settle(baseSnap(0, 0, powered(10, "Vanilla", 0, 3, 0), withKeywords(powered(11, card, 1, 2, 0), engine.Keywords{})))
		seq := 0
		for _, e := range evs {
			seq = f.event(e)
		}
		return f, seq
	}
	resistant := func(e engine.Event) engine.Event { e.Resistant = 1; return e }
	stalwart := func(e engine.Event) engine.Event { e.Stalwart = true; return e }

	f, _ := feedDamage("Vanilla", resistant(dmgEvent(10, 11, 3, false, 0, 0, 2)))
	wantClean(t, "Resistant since the snapshot, its card is not", f.m)
	wantCounts(t, "Resistant since the snapshot", f.m, "damage-sequence", "FR-KWD-002b")
	f, seq := feedDamage("Vanilla", resistant(dmgEvent(10, 11, 3, false, 0, 0, 3)))
	wantViolation(t, "damage ignores the event's Resistant", f.m, "damage-sequence", damageRules, seq)
	f, _ = feedDamage("Tough", dmgEvent(10, 11, 3, false, 0, 0, 3))
	wantClean(t, "its card is Resistant, it is not", f.m)

	f, _ = feedDamage("Vanilla", combatEvent(10, 11), stalwart(dmgEvent(10, 11, 3, true, 0, 0, 1)))
	wantClean(t, "Stalwart since the snapshot, its card is not", f.m)
	wantCounts(t, "Stalwart since the snapshot", f.m, "damage-sequence", "FR-KWD-002c")
	f, seq = feedDamage("Vanilla", combatEvent(10, 11), stalwart(dmgEvent(10, 11, 3, true, 0, 0, 3)))
	wantViolation(t, "combat damage ignores the event's Stalwart", f.m, "damage-sequence", damageRules, seq)
	f, _ = feedDamage("Wall", combatEvent(10, 11), dmgEvent(10, 11, 3, true, 0, 0, 3))
	wantClean(t, "its card is Stalwart, it is not", f.m)
}

// Case 5: the event has all the check reads, so it runs at once, even on a
// target no settled snapshot shows.
func TestMonitorKeywords_Case5_TargetInNoSnapshot(t *testing.T) {
	f, seq := damageCase(baseSnap(0, 0), dmgEvent(10, 99, 3, false, 0, 0, 1))
	wantViolation(t, "wrong amount, target in no snapshot", f.m, "damage-sequence", damageRules, seq)
}
```

### 2. `pkg/sim/monitor_test.go`, lines 44–46, in the `test(T-0179)` commit
Replace `snapAsset` with:
```go
// cardKeywords is the keywords of the card with this name in the test pools.
func cardKeywords(card string) engine.Keywords {
	for _, d := range [][2]engine.DeckList{monitorDecks(), damageDecks()} {
		for _, c := range d[0].Main {
			if c.Name == card {
				return c.Keywords
			}
		}
	}
	return engine.Keywords{}
}

// snapAsset is a face-up asset with its card's keywords, as the engine reports them.
func snapAsset(id int, card string, owner int, zone string) engine.AssetSnap {
	return engine.AssetSnap{ID: id, Inst: id + 1000, Card: card, Owner: owner, Controller: owner, Zone: zone, Keywords: cardKeywords(card)}
}
```

### 3. `pkg/sim/monitor_damage_test.go`, lines 54–62, in the `test(T-0179)` commit
Replace those lines, from the comment on `damageCase` to the end of its `for` loop, with:
```go
// damageCase feeds a snapshot, then events, and returns the feed and the Seq
// of the last event sent. A damage event reports its target's keywords in the
// snapshot, as the engine does.
func damageCase(snap engine.Snapshot, evs ...engine.Event) (*feed, int) {
	f := damageFeed()
	f.settle(snap)
	seq := 0
	for _, e := range evs {
		if t := find(snap, e.Target); t != nil && e.Kind == "damage" {
			e.Resistant, e.Stalwart = t.Keywords.Resistant, t.Keywords.Stalwart
		}
		seq = f.event(e)
	}
```

### 4. `pkg/sim/monitor_attack.go`, in the `feat(sim)` commit
Make the first replacement first, so the line numbers hold.
1. Replace line 82 with:
   ```go
   	if def.Keywords.Ethereal {
   ```
2. Replace lines 72–79, from `oc := m.card(o)` to the line that appends to `guardians`, with:
   ```go
   		if o.Keywords.Ethereal {
   			m.count(mon, "FR-KWD-002g")
   		} else if o.Keywords.Guardian && o.Zone == theater {
   			guardians = append(guardians, o.ID)
   ```

### 5. `pkg/sim/monitor_resources.go`, line 63, in the `feat(sim)` commit
Replace it with:
```go
		case a.Keywords.Versatile:
```

### 6. `pkg/sim/monitor_targeting.go`, lines 33–38, in the `feat(sim)` commit
Replace those lines, the body of the `for` loop over `battlefield(*m.last)`, with:
```go
		if a.Controller != player && a.Keywords.Ward {
			warded[a.ID] = true
		}
```

### 7. `pkg/sim/monitor_damage.go`, in the `feat(sim)` commit
Make the first replacement first, so the line numbers hold.
1. Replace lines 77–112, from `check := func(s engine.Snapshot) bool {` to the `m.pending = append(...)` line, with:
   ```go
   	if e.Resistant != 0 {
   		m.count(mon, "FR-KWD-002b")
   	}
   	if e.Stalwart {
   		m.count(mon, "FR-KWD-002c")
   	}
   	if e.Armor > 0 {
   		m.count(mon, "FR-KWD-002k")
   	}
   	r := max(0, e.Raw-e.Resistant)
   	capped := r
   	if e.Combat && e.Stalwart {
   		capped = min(r, 1)
   	}
   	absorbed := min(capped, e.Armor)
   	if e.Absorbed != absorbed || e.Amount != capped-absorbed {
   		m.fail(mon, e.Seq, "damage %d (combat %t) to %d (Resistant %d, Stalwart %t, Armor %d) absorbed %d and committed %d, want %d and %d",
   			e.Raw, e.Combat, e.Target, e.Resistant, e.Stalwart, e.Armor, e.Absorbed, e.Amount, absorbed, capped-absorbed)
   	}
   ```
2. Replace lines 54–55, the comment on `checkDamage`, with:
   ```go
   // checkDamage checks one damage event against the five-stage sequence
   // (FR-REP-004), using the target's effective Resistant and Stalwart as the
   // event reports them, because they can change after a settled snapshot.
   ```

No other file changes.

## Acceptance Cases (harness tasks only; written by the lead)
Edits 1–3 are the whole of the test commit: commit them as `test(T-0179): the monitors read effective keywords`, run case 1, then make edits 4–7 and commit them as `feat(sim): the monitors read effective keywords (T-0179)`. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **Fails before the edits** (after the test commit, before edit 4): `go -C W test ./pkg/sim -run TestMonitorKeywords` exits 1, and its output has `--- FAIL` for `TestMonitorKeywords_Case1_Attack`, `TestMonitorKeywords_Case2_Versatile`, `TestMonitorKeywords_Case3_Ward`, `TestMonitorKeywords_Case4_Damage` and `TestMonitorKeywords_Case5_TargetInNoSnapshot`. At the same commit, `go -C W test ./pkg/sim -run TestMonitor_Case1` and `go -C W test ./pkg/sim -run TestDamage_Case3` exit 0.
2. **Passes after them** (#116 (b)): `go -C W test ./pkg/sim -run TestMonitorKeywords -v` exits 0, and its output has `--- PASS` for all five.
3. **The existing monitor tests pass, with their bodies unedited:** `go -C W test ./pkg/sim -run TestMonitor_Case1` and `go -C W test ./pkg/sim -run TestDamage_Case3` exit 0.
4. **A search and a full game still replay:** `go -C W test ./pkg/eventlog ./pkg/sim` exits 0.
5. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `3	7	pkg/sim/monitor_attack.go`, `19	35	pkg/sim/monitor_damage.go`, `5	1	pkg/sim/monitor_damage_test.go`, `114	0	pkg/sim/monitor_keywords_test.go`, `1	1	pkg/sim/monitor_resources.go`, `1	4	pkg/sim/monitor_targeting.go` and `14	1	pkg/sim/monitor_test.go`.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those seven files and this task file.
7. **Sims byte-identical** (#116): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
8. `gofmt -l W/pkg/sim` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/sim/monitor_attack.go`
- `pkg/sim/monitor_damage.go`
- `pkg/sim/monitor_resources.go`
- `pkg/sim/monitor_targeting.go`
- `pkg/sim/monitor_test.go` (edit 2 only)
- `pkg/sim/monitor_damage_test.go` (edit 3 only)
- `pkg/sim/monitor_keywords_test.go` (new, edit 1 only)

## Done When
- [x] Cases 1–8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 9): split from #116 under the size rule; T-0178 builds (a). `queued` until T-0178 is `done`.
- 2026-11-05, lead-architect (wave D.1 run, session 10): no `main` commits to sync. Wrote Inputs, Edits, Acceptance Cases and Allowed Files against the run branch at 5de081f. Not `WAVE-COMPLETE`: #116–#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
- 2026-11-05, kernel-engineer (T-0179): implemented rule monitors reading effective keywords.
  - Test commit (e2dcb06): added `pkg/sim/monitor_keywords_test.go` covering Acceptance Cases 1-5; updated `pkg/sim/monitor_test.go` (`cardKeywords`, `snapAsset`) and `pkg/sim/monitor_damage_test.go` (`damageCase`) to populate keywords on hand-built test records.
  - Feat commit (a9cec9e): updated `pkg/sim/monitor_attack.go` (`checkAttack` reads `o.Keywords.Ethereal`/`Guardian` and `def.Keywords.Ethereal`), `pkg/sim/monitor_resources.go` (`checkGenerate` reads `a.Keywords.Versatile`), `pkg/sim/monitor_targeting.go` (`checkTargetDecision` reads `a.Keywords.Ward`), and `pkg/sim/monitor_damage.go` (`checkDamage` reads `e.Resistant` and `e.Stalwart` immediately from event).
  - Reuse: reused existing `Keywords` on `AssetSnap` and `Event` without adding new shapes.
  - Acceptance verification:
    - Case 1: `go test ./pkg/sim -run TestMonitorKeywords` exited 1 before edits 4-7 with failures across all 5 test cases; `TestMonitor_Case1` and `TestDamage_Case3` exited 0.
    - Case 2: `go test ./pkg/sim -run TestMonitorKeywords -v` passed all 5 cases (exit 0).
    - Case 3: `go test ./pkg/sim -run TestMonitor_Case1` and `TestDamage_Case3` passed (exit 0).
    - Case 4: `go test ./pkg/eventlog ./pkg/sim` passed (exit 0).
    - Case 5: `git diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` matched the exact 7 expected file numstats.
    - Case 6: `git diff autoloop/m2-wave-d.1 --stat` changed only the 7 allowed files and this task file.
    - Case 7: `go tool task sim-compare -- autoloop/m2-wave-d.1` exited 0 with `[PASS] starter and stress sims match autoloop/m2-wave-d.1 byte for byte, exit codes included`.
    - Case 8: `gofmt -l pkg/sim` clean; `go tool task check` passed all checks; `go tool task sim` finished cleanly with 0 monitor violations.
  - Status moved to `review`.
- 2026-11-05, lead-architect (wave D.1 run, session 10, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-d.1` as dbc0f5e. Flash run `7ae3e3f7-b8ae-4dd3-8b8b-696839f08559`: exit 0, `status: SUCCESS`, no refusals (no `denied_actions` key; the transcript has no step with `"status":"ERROR"` and no `permission check failed` match). The `.err` file holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-d.1...m2/autoloop-wave-d.1/T-0179-monitors-read-effective-keywords`, not the Log:
  - Scope: only the seven Allowed Files and this file changed, and all seven are byte-identical to the lead's dry run. The three test files changed only in e2dcb06, the `test(T-0179)` commit, and the four monitor files only in a9cec9e. The only edits to existing test code are the two setup helpers this task names; no test body changed.
  - Failed before: at e2dcb06, in a scratch worktree, `go test ./pkg/sim -run TestMonitorKeywords` exits 1 with `--- FAIL` for all five, and `TestMonitor_Case1` and `TestDamage_Case3` pass.
  - Fences and placement: no card name and no `OPEN(` marker. `monitor_entry.go` is untouched, so the entry-state monitor still reads printed keywords. `m.card` stays for type and yields.
  - A limit, known and stated in Inputs: the damage check now takes Resistant and Stalwart from the same `kw` `dealDamage` uses, as #116 asks, so it checks the sequence's arithmetic and not which keywords the engine read. Nothing grants a keyword until #117, whose tests carry the first real grant through a real game.
  - Re-run by the reviewer in a scratch worktree of the task branch, all as expected: cases 2-7; `gofmt -l pkg` clean; `go tool task check` exit 0; `go tool task sim` exit 0, 200 games, P0 84.0% / P1 15.0% / draws 1.0%, 47/50 and 34/50 main cards executable, no monitor violation; `go tool task sim-compare -- autoloop/m2-wave-d.1` `[PASS]`, coverage table included.
  - The merge's `pkg/` and `cmd/` match the task tip's, so the gates hold on the run branch. #116 is fully applied and archived. The lead updated the design note's Rule monitors paragraph in this commit. #117 is next.
