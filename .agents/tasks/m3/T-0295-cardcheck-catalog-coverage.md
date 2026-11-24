---
id: T-0295
title: "Coverage checkers move to pkg/cardcheck and cover the whole catalog; NotYetExecutable holds the 67 cards; every starter deck validates (M3 Deliverable 1a; harness, flash)"
kind: harness
tier: flash
wave: none (M3 design note Deliverable 1, part a)
status: review
branch: m3/T-0295-cardcheck-catalog-coverage
depends_on: [T-0294]
claims: [package:pkg/cardcheck, func:cardcheck.checkExecutable, func:cardcheck.checkDefs, var:cards.NotYetExecutable, test:TestCards_Executable, test:TestDecks_StarterDecksValidate]
cards: []
requirements: []
---

## Goal
The coverage checkers leave `pkg/engine/defs_test.go` for a non-test file in the new `pkg/cardcheck` package, and the executable check covers every card in `data/cards/`, not just the M2 starter set. `cards.NotYetExecutable` lists the 67 cards that can't run yet, so the wave runner and every card task count down from it. A new test validates every `data/decks/*-starter.txt` under `FR-SET-001`.

## Inputs
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/cardcheck`, approved by the user 2026-11-23), Test Plan (Coverage), Deliverable 1 (a) (34368abb).
- Board: no open directive applies. Directive #184, written with this task, holds the `needsRework` exemption until the first card is flagged.
- Where it is today (at `main` 34368abb):
  - `pkg/engine/defs_test.go` (T-0002) holds:
    - the checkers `checkStarterSet` and `checkDefs`;
    - the scope `inStarterScope`: Animus and Path, except *Rift* and *Encampment*;
    - the helpers `synthCard` and `wantOneProblem`, which nothing else uses (`git grep`);
    - tests 12-19: `TestStarterSet_Executable`, five `TestCheckStarterSet_*`, `TestDefs_MatchRulesText` and `TestCheckDefs_Synthetic`.
  - `cards.NotYetExecutable` ([cards.go:13](../../../pkg/engine/cards/cards.go)) is empty. Loading `data/cards/` gives 169 cards, and 67 of them are not `Supported()`: 33 Concordia, 32 Soulless, *Rift* and *Encampment* (the lead counted them 2026-11-23).
  - The fences linter forbids card names in non-test files under `pkg/`, except in `pkg/engine/cards` and `pkg/toolchain`. Its `allowedImports` ([fences.go:43-51](../../../pkg/toolchain/linter/fences/fences.go)) lists the in-module imports each fenced package may use.
- Design (lead decisions):
  - **Move and widen in one task.** The checker can't move as it is: `inStarterScope` names *Rift* and *Encampment*, and the fences linter rejects card names in a non-test file under `pkg/cardcheck`. The catalog-wide check has no scope, so the move drops it.
  - **Renames.** The old names leave the repository (ADR-0007):

    | Old | New |
    | :--- | :--- |
    | `checkStarterSet` | `checkExecutable` (no scope) |
    | `TestStarterSet_Executable` | `TestCards_Executable` |
    | `TestCheckStarterSet_*` | `TestCheckExecutable_*` |

    `TestCheckStarterSet_OutOfScopeEntry` keeps only its unknown-name case, as `TestCheckExecutable_UnknownEntry`. `TestCheckStarterSet_OtherFactionIgnored` is inverted as `TestCheckExecutable_EveryFactionCounts`: with no scope, every faction's card needs an entry. At merge, the lead updates line 74 of the M2 design note and directive #7's mention of `defs_test.go`.
  - `checkDefs` and its two tests move with no change in behavior, only `engine.` qualifiers added.
  - The package's functions stay unexported. Only its own tests call them, and nothing imports `pkg/cardcheck`.
  - `fences.allowedImports` gets `pkg/cardcheck` → `pkg/engine`, `pkg/engine/ability`, `pkg/engine/cards`.
  - `pkg/engine/doc.go` stops saying `cards.NotYetExecutable` is empty.
  - Empty `concordia.go` and `soulless.go` are not created here (design note Deliverable 1 (a)).
  - **Expected output difference:** none. The sims run on `Supported()`, not on the allowlist, and `go tool task sim-compare` passes.
- Size (D2.4): four non-test files: `coverage.go` (new), `cards.go`, `fences.go`, and `doc.go`, which changes only a comment that names the list.
- The lead dry-ran these exact files in a throwaway worktree of `main` at 9ef7fb77:
  - All 9 `pkg/cardcheck` tests passed, and `gofmt -l pkg cmd` was clean.
  - `go vet` was clean, and `go tool task check` exited 0.
  - `go tool task sim-compare` printed `[PASS]`.

  Blobs: `coverage.go` 8d79ba1, `coverage_test.go` 72c4a40, `decks_test.go` 760d58a, `cards.go` 8554c96, `fences.go` 80a6990, `doc.go` de67bdf.

## Acceptance Cases (harness tasks only; written by the lead)
1. Real data, real allowlist → `TestCards_Executable` passes: every card in `data/cards/` that can't run is listed, and nothing else is.
2. A card that can't run and isn't listed → one problem naming it and `NotYetExecutable`.
3. A listed card that runs → one problem naming it and saying to remove it.
4. A listed name that matches no card → one problem saying it matches no card.
5. A name listed twice, or out of order → one problem each. A sorted list → none.
6. A Soulless unit, a Concordia Maneuver with no ability, and a Path Headquarters that can't run → one problem each when not listed, none when listed (no starter-set scope).
7. `checkDefs`: a def with no card, or with empty `Text` → one problem each. A matching def → none.
8. Every `data/decks/*-starter.txt` loads and passes `engine.ValidateDeck` → `TestDecks_StarterDecksValidate` passes. It fails, naming the file, on an illegal starter deck or a missing Animus, Concordia, Path or Soulless one.
9. `pkg/engine/defs_test.go` is gone. `git grep` finds `checkStarterSet`, `inStarterScope`, `TestStarterSet_Executable` and `TestCheckStarterSet` only in frozen M2 history (archived tasks, run files, outcomes), in the M3 design note's description of this change, and in the two references the lead updates at merge.

## Tests (copy exactly; tabs as shown)
Delete `pkg/engine/defs_test.go` (`git rm pkg/engine/defs_test.go`). Add:

`pkg/cardcheck/coverage_test.go`:
```go
package cardcheck

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/engine/cards"
)

func synthCard(name, faction string, typ engine.CardType, unsupported ...string) *engine.Card {
	return &engine.Card{Name: name, Faction: faction, Type: typ, Unsupported: unsupported}
}

func wantOneProblem(t *testing.T, problems []string, parts ...string) {
	t.Helper()
	if len(problems) != 1 {
		t.Fatalf("got %d problems %q, want 1", len(problems), problems)
	}
	for _, p := range parts {
		if !strings.Contains(problems[0], p) {
			t.Fatalf("problem %q should contain %q", problems[0], p)
		}
	}
}

// The real allowlist matches the real data: every card in data/cards that
// cannot run is listed, and nothing else is.
func TestCards_Executable(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkExecutable(all, cards.NotYetExecutable) {
		t.Error(p)
	}
}

// A card that cannot run and is missing from the allowlist.
func TestCheckExecutable_UnlistedUnsupported(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text")}
	wantOneProblem(t, checkExecutable(all, nil), "Alpha", "NotYetExecutable")
}

// An allowlisted card that is now supported.
func TestCheckExecutable_ListedButSupported(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "ANIMUS", engine.TypeUnit)}
	wantOneProblem(t, checkExecutable(all, []string{"Alpha"}), "Alpha", "remove")
}

// An allowlisted name that matches no card.
func TestCheckExecutable_UnknownEntry(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit)}
	wantOneProblem(t, checkExecutable(all, []string{"No Such Card"}), "No Such Card", "matches no card")
}

// Duplicate or unsorted allowlists.
func TestCheckExecutable_DuplicateOrUnsorted(t *testing.T) {
	all := map[string]*engine.Card{
		"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text"),
		"Beta":  synthCard("Beta", "PATH", engine.TypeUnit, "text"),
	}
	wantOneProblem(t, checkExecutable(all, []string{"Alpha", "Alpha", "Beta"}))
	wantOneProblem(t, checkExecutable(all, []string{"Beta", "Alpha"}))
	if p := checkExecutable(all, []string{"Alpha", "Beta"}); len(p) != 0 {
		t.Fatalf("sorted list should pass, got %q", p)
	}
}

// Every faction counts, Headquarters included: there is no starter-set scope.
func TestCheckExecutable_EveryFactionCounts(t *testing.T) {
	for _, c := range []*engine.Card{
		synthCard("Hollow", "SOULLESS", engine.TypeUnit, "text"),
		synthCard("Bastion", "CONCORDIA", engine.TypeManeuver),
		synthCard("Outpost", "PATH", engine.TypeHeadquarters, "text"),
	} {
		all := map[string]*engine.Card{c.Name: c}
		wantOneProblem(t, checkExecutable(all, nil), c.Name)
		if p := checkExecutable(all, []string{c.Name}); len(p) != 0 {
			t.Fatalf("listed %q should pass, got %q", c.Name, p)
		}
	}
}

// The real defs match the real data.
func TestDefs_MatchRulesText(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkDefs(all, cards.Defs) {
		t.Error(p)
	}
}

// checkDefs on synthetic input.
func TestCheckDefs_Synthetic(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text")}
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Ghost": {Text: []string{"x"}}}), "Ghost")
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Alpha": {}}), "Alpha")
	if p := checkDefs(all, map[string]*ability.Def{"Alpha": {Text: []string{"text"}}}); len(p) != 0 {
		t.Fatalf("valid def should pass, got %q", p)
	}
}
```

`pkg/cardcheck/decks_test.go`:
```go
package cardcheck

import (
	"path/filepath"
	"testing"

	"animus-mundi/pkg/engine"
)

// Every starter deck in data/decks is construction-legal (FR-SET-001), and each
// of the four Headquarters factions has one.
func TestDecks_StarterDecksValidate(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("..", "..", "data", "decks")
	paths, err := filepath.Glob(filepath.Join(dir, "*-starter.txt"))
	if err != nil {
		t.Fatal(err)
	}
	found := make(map[string]bool)
	for _, p := range paths {
		found[filepath.Base(p)] = true
		d, err := engine.LoadDeck(p, all)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
			continue
		}
		if err := engine.ValidateDeck(d); err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
		}
	}
	for _, f := range []string{"animus", "concordia", "path", "soulless"} {
		if !found[f+"-starter.txt"] {
			t.Errorf("data/decks has no %s-starter.txt", f)
		}
	}
}
```

## Code (copy exactly)
New file:

`pkg/cardcheck/coverage.go`:
```go
// Package cardcheck holds checks over the whole card set that are not game
// rules (M3 design note, Boundaries): the executable-coverage checker and the
// definition-to-text checker, which this package's tests run against
// data/cards in every go tool task check. It imports engine, ability and
// cards, and nothing imports it.
package cardcheck

import (
	"fmt"
	"sort"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
)

// checkExecutable compares every card the kernel cannot execute against allow
// (cards.NotYetExecutable). Each such card must be listed, each listed name
// must be a card that still cannot run, and allow must be sorted with no name
// twice. Card tasks change only cards.NotYetExecutable, never this checker.
func checkExecutable(all map[string]*engine.Card, allow []string) []string {
	var problems []string
	listed := make(map[string]bool)
	for i, name := range allow {
		if listed[name] {
			problems = append(problems, fmt.Sprintf("NotYetExecutable lists %q more than once", name))
		} else if i > 0 && name < allow[i-1] {
			problems = append(problems, fmt.Sprintf("NotYetExecutable is not sorted at %q", name))
		}
		listed[name] = true
		c, ok := all[name]
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("NotYetExecutable names %q, which matches no card", name))
		case c.Supported():
			problems = append(problems, fmt.Sprintf("%q is executable now: remove it from NotYetExecutable", name))
		}
	}
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !all[name].Supported() && !listed[name] {
			problems = append(problems, fmt.Sprintf("%q cannot run: it must become executable or go on NotYetExecutable", name))
		}
	}
	return problems
}

// checkDefs reports each definition whose card is not loaded or whose Text is
// empty.
func checkDefs(all map[string]*engine.Card, defs map[string]*ability.Def) []string {
	var problems []string
	names := make([]string, 0, len(defs))
	for name := range defs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := all[name]; !ok {
			problems = append(problems, fmt.Sprintf("def %q matches no loaded card", name))
		} else if len(defs[name].Text) == 0 {
			problems = append(problems, fmt.Sprintf("def %q has empty Text", name))
		}
	}
	return problems
}
```

Replace the whole of `pkg/engine/cards/cards.go`:

`pkg/engine/cards/cards.go`:
```go
// Package cards holds the card definitions built from ability shapes
// (ADR-0004) and the allowlist of cards that cannot run yet.
// It imports only package ability.
package cards

import "animus-mundi/pkg/engine/ability"

// Defs maps a card name to its ability definition.
var Defs = map[string]*ability.Def{}

// NotYetExecutable lists, sorted, every card in data/cards the kernel cannot
// execute yet (pkg/cardcheck's TestCards_Executable). Card tasks remove names
// from it; M3 is done when it is empty.
var NotYetExecutable = []string{
	"All-Out War",
	"Armored Charge",
	"Awaken Nightmare",
	"Bring It Down",
	"Cathedral of Self",
	"Colossus Platform",
	"Command Barge",
	"Comms Relay",
	"Conscript Rifles",
	"Consuming Earth",
	"Culverin Brigade",
	"Dolus Infusion",
	"Dolus Macrocannon",
	"Dolus Refiner",
	"Dominator Warsuit",
	"Dread Manifestation",
	"Earthshaker Colossus",
	"Echoing Nightmare",
	"Encampment",
	"Enclave",
	"Field Mechanic",
	"Fire for Effect",
	"Flickering Ember",
	"Forgemaster's Resilience",
	"Frostfire Breath",
	"Gemstone Titan",
	"Growing Legion",
	"Hellfire Armorers",
	"Hourglass Diviner",
	"Hurricane Outrigger",
	"Icewind Regent",
	"Inspired Salvage",
	"Land Fortress",
	"Leyline Regrowth",
	"Light of the World Soul",
	"MAX-8 Autonomous Warsuit",
	"Macrocannon Barrage",
	"Mage Seeker's Inquisition",
	"March of Progress",
	"Mobile Armory",
	"NO-AH-5",
	"Natural Attunement",
	"Piston Fist",
	"Piston Fist Prototype",
	"Popup Bunker",
	"Rage's Wake",
	"Railway Reinforcements",
	"Redline Infusion",
	"Redlined Brawler",
	"Reverberating Terror",
	"Rift",
	"Samsara, Living Tempest",
	"Sandstorm Lightbringer",
	"Screaming Palisade",
	"Seismic Instability",
	"Soul Whip",
	"Stardust Colossus",
	"Stormbolt Ray",
	"Swirling Dustdevil",
	"Tempest City Ray",
	"Tempest Runner",
	"Tempest Surge",
	"Tidal Leviathan",
	"Visionary's Eye",
	"Visionary's Insight",
	"Will Breaker",
	"World Soul Messenger",
}
```

In `pkg/toolchain/linter/fences/fences.go`, add one line to `allowedImports` right after the `"pkg/engine/cards"` line (line 46). gofmt leaves it aligned like this:
```go
	"pkg/cardcheck":      {"pkg/engine", "pkg/engine/ability", "pkg/engine/cards"},
```

In `pkg/engine/doc.go`, replace lines 108-109:
```go
//     set and cannot be deployed or enacted. No in-scope M2 card is in that
//     state, so cards.NotYetExecutable is empty.
```
with:
```go
//     set and cannot be deployed or enacted. cards.NotYetExecutable lists every
//     such card (pkg/cardcheck).
```

## Commits
1. `test(T-0295): catalog-wide coverage and starter-deck tests in pkg/cardcheck` holds only the deletion of `pkg/engine/defs_test.go` and the two new test files. On this commit `go test ./pkg/cardcheck` fails, because the package has no non-test file yet.
2. `feat(cardcheck): coverage checkers cover the whole catalog, NotYetExecutable lists the 67 cards that can't run (T-0295)` holds `coverage.go`, `cards.go`, `fences.go` and `doc.go`.

No double quotes in either message: the allow-list refuses an escaped one.

## Allowed Files
- `pkg/cardcheck/coverage.go` (new)
- `pkg/engine/cards/cards.go`
- `pkg/toolchain/linter/fences/fences.go`
- `pkg/engine/doc.go` (one comment, a reference update)

## Done When
- [x] `go test ./pkg/cardcheck -count=1` passes
- [x] `gofmt -l pkg cmd` prints nothing
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim-compare` passes (starter and stress sims byte-identical to `main`)

## Log
- 2026-11-23 lead-architect: written from M3 design note Deliverable 1 (a). The same commit queues T-0296 (the sheet checks, Deliverable 1 (b)) and adds directive #184 (the `needsRework` exemption). Status `ready` (a harness task skips `tests`), tier `flash`.
- 2026-11-23 kernel-engineer: implemented T-0295.
  - Test commit: deleted `pkg/engine/defs_test.go` and added `pkg/cardcheck/coverage_test.go` and `pkg/cardcheck/decks_test.go`. Verified `go test ./pkg/cardcheck` failed with undefined functions (`test(T-0295): catalog-wide coverage and starter-deck tests in pkg/cardcheck`, 2f3095c8).
  - Implementation commit: added `pkg/cardcheck/coverage.go` (`checkExecutable` and `checkDefs`), updated `pkg/engine/cards/cards.go` with 67 sorted unexecutable card entries, updated `pkg/toolchain/linter/fences/fences.go` with `pkg/cardcheck` allowed imports, and updated `pkg/engine/doc.go` (`feat(cardcheck): coverage checkers cover the whole catalog, NotYetExecutable lists the 67 cards that can't run (T-0295)`, 6d7adb12).
  - Verification: `go test ./pkg/cardcheck -count=1` passed; `gofmt -l pkg cmd` clean; `go tool task check` passed; `go tool task sim-compare` passed with identical byte output against main.
  - Status moved to `review`.
