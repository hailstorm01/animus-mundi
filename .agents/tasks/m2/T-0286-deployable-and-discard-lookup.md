---
id: T-0286
title: "`Card.Deployable`, and `DiscardCard` finds a card through `pileCard` (directive #173 (a), (b); wave H.2 harness, flash)"
kind: harness
tier: flash
wave: H.2 (board directive #173 (a), (b))
status: in-progress
branch: m2/autoloop-wave-h.2/T-0286-deployable-and-discard-lookup
depends_on: [T-0285]
claims: [func:engine.Card.Deployable]
cards: []
requirements: [FR-TRN-003]
---

## Goal
Two small repeats in `pkg/engine` are written once each, as pure moves (directive #173 (a) and (b)):
- **(a)** "Only Unit and Structure cards are deployed" (`FR-TRN-003`) is written in `check` ([actions.go:142](../../../pkg/engine/actions.go)) and in `freeDeployZones` ([freedeploy.go:56](../../../pkg/engine/freedeploy.go)). After this task, `Card.Deployable()` in `card.go` says it, and both call it.
- **(b)** Finding a card by instance ID in a discard pile is written in `DiscardCard` ([zones.go:168-178](../../../pkg/engine/zones.go)) and in `pileCard` ([zones.go:182-196](../../../pkg/engine/zones.go)). After this task, `DiscardCard` calls `pileCard(pl, id, ZoneDiscard)` for each player and has no ID loop of its own.

#173 (c) and (d) are T-0287 (`queued`). #173 is `applied` when both merge.

## Inputs
- Board directive #173 (structural review 14, finding 5), approved by the user 2026-11-19, `open: before the next Deliverables task`. It allows (a)–(d) to split in two under the D2 size rule, and all four would touch six non-test files (`card.go`, `actions.go`, `freedeploy.go`, `zones.go`, `ability/ability.go`, `cardload.go`). This task is (a) and (b), four files; T-0287 is (c) and (d), two files.
- Where it is today (at `autoloop/m2-wave-h.2` d5fa2c2): as the Goal says.
- Design (lead decisions):
  - `Deployable` sits in `card.go` beside `IsScratchCompany` and `Supported`, the other readers of a `Card`'s printed type. It reads `Card.Type` only. `freeDeployZones` keeps its `Supported()` check beside it.
  - `DiscardCard` keeps its doc and signature, and loops over `g.Players` (`[2]*Player`, seat order, as today), so the first match is the same card.
  - Nothing else in `pkg/engine` compares a card's type with both `TypeUnit` and `TypeStructure` (`git grep -n TypeStructure -- pkg/engine ":!*_test.go"`: the declaration, the CSV loader, `filter.go`'s type match and the new `Deployable`). `pkg/sim/monitor_attack.go:66` tests which assets may be attack targets, a different rule, and stays.
- Reuse search (D2.3): no `Deployable` or equivalent exists (`git grep -n Deployable -- pkg cmd` finds nothing). `pileCard` is the only other ID lookup in a pile.
- Size (D2.4): four non-test files, within the limit.
- Docs: `pkg/engine/doc.go`, the design note and `engine-structs.md` name `Card.Deployable` with #176 (b), after the wave's last code task.
- **Expected output difference:** none. `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at d5fa2c2:
  - `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`.
  - Blobs: `card.go` 9bb2fc9, `actions.go` 46396bd, `freedeploy.go` 4fcebd3, `zones.go` 0b61b4d.

## Commits
1. `refactor(engine): Card.Deployable, and DiscardCard finds a card through pileCard (T-0286)` with Edits 1-4 (no double quotes in the message: the allow-list refuses an escaped one). A pure move adds no test, so there is no `test(T-0286)` commit.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/card.go` (`Deployable`)
```diff
@@ -63,6 +63,12 @@ func (c *Card) IsScratchCompany() bool {
 	return c.HasSubtype("Scratch Company")
 }
 
+// Deployable reports whether the card is deployed rather than played: only
+// Unit and Structure cards are (FR-TRN-003; a Maneuver is played, FR-MNV-001).
+func (c *Card) Deployable() bool {
+	return c.Type == TypeUnit || c.Type == TypeStructure
+}
+
 // Supported reports whether the kernel can fully execute this card.
 func (c *Card) Supported() bool {
 	if len(c.Unsupported) != 0 {
```

### 2. `pkg/engine/actions.go` (`check` calls it)
```diff
@@ -139,7 +139,7 @@ func (g *Game) check(a Action) error {
 		if a.pile() == ZoneDiscard && deployPermission(c) == nil {
 			return illegal("%s has no permission to be deployed from the discard pile (FR-TRN-003)", c.Name)
 		}
-		if c.Type != TypeUnit && c.Type != TypeStructure {
+		if !c.Deployable() {
 			return illegal("%s cannot be deployed", c.Name)
 		}
 		if a.Zone == ZoneHeadquarters || !slices.Contains(c.DeployZones, a.Zone) {
```

### 3. `pkg/engine/freedeploy.go` (`freeDeployZones` calls it)
```diff
@@ -53,7 +53,7 @@ func (g *Game) stepDeployFree(r *resolution, st ability.Step) int {
 // Reserves, then the Main Theater, whichever c.DeployZones contains
 // (FR-TRN-003a 1).
 func freeDeployZones(c *Card, to ability.Zone) []Zone {
-	if (c.Type != TypeUnit && c.Type != TypeStructure) || !c.Supported() {
+	if !c.Deployable() || !c.Supported() {
 		return nil
 	}
 	if zone, ok := battlefieldZone(to); ok {
```

### 4. `pkg/engine/zones.go` (`DiscardCard` calls `pileCard`)
```diff
@@ -167,10 +167,8 @@ func (g *Game) putCard(inst *Instance, to Zone, bottom bool) {
 // uses it to show a chosen card that is still there.
 func (g *Game) DiscardCard(id int) *Instance {
 	for _, pl := range g.Players {
-		for _, inst := range pl.Discard {
-			if inst.ID == id {
-				return inst
-			}
+		if inst := pileCard(pl, id, ZoneDiscard); inst != nil {
+			return inst
 		}
 	}
 	return nil
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **`check` and `freeDeployZones` compare no card type for it:** `git -C W grep -n TypeStructure -- pkg/engine/actions.go pkg/engine/freedeploy.go` prints nothing.
2. **Both call `Deployable`:** `git -C W grep -n "Deployable()" -- pkg/engine` prints exactly three lines: `pkg/engine/actions.go:142`, `pkg/engine/card.go:68` (the declaration) and `pkg/engine/freedeploy.go:56`.
3. **One ID loop over a pile:** `git -C W grep -n "inst.ID == id" -- pkg/engine ":!*_test.go"` prints exactly one line, `pkg/engine/zones.go:189` (in `pileCard`), and `git -C W grep -n "pileCard(pl, id, ZoneDiscard)" -- pkg/engine` prints exactly one line, `pkg/engine/zones.go:170`.
4. **The same tests pass, unedited:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints nothing.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly four lines: `1	1	pkg/engine/actions.go`, `6	0	pkg/engine/card.go`, `1	1	pkg/engine/freedeploy.go` and `2	4	pkg/engine/zones.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the four files of case 6 and this task file.
8. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/card.go` (Edit 1)
- `pkg/engine/actions.go` (Edit 2)
- `pkg/engine/freedeploy.go` (Edit 3)
- `pkg/engine/zones.go` (Edit 4)

## Tests
- None added or edited: a pure move. Every existing test passes unedited.

## Done When
- [ ] Cases 1-8 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 6): written against d5fa2c2 for #173 (a) and (b), next in wave H.2's order (#172 is `applied`).
  - `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #173 and #174 are `open: before the next Deliverables task`.
  - #173 split in two as it allows (six non-test files together): (c) and (d) are T-0287, `queued` with board row and task file (D2.4).
  - This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run.
  - Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted.
  - Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards). None has fired.
  - Next: `kernel-engineer` on Flash.
