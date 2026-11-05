---
id: T-0175
title: Filters get their own file, filter.go (directive #113, harness, flash, pure move)
kind: harness
tier: flash
wave: D.1 (board directive #113, structural review 8, finding 10)
status: in-progress
branch: m2/autoloop-wave-d.1/T-0175-filter-file
depends_on: [T-0174]
claims: [file:pkg/engine/filter.go]
cards: []
requirements: []
---

## Goal
`pkg/engine/targeting.go` stops having three jobs. `kindOf`, `matchesCard`, `legalTarget`, `matchesPrinted` and `matchesFilter` move unchanged to a new `pkg/engine/filter.go`, with a header line (#113). `pileTarget` stays in `targeting.go`, because it is target verification. A pure move: same declarations, same test names, same sim output.

## Inputs
- Board directive #113 (structural review 8, approved by the user, 2026-11-05).
- Design note: `docs/architecture/m2-ability-model.md`, Deliverable 6g (wave D.1).
- Reuse search (D2.3): no `filter.go` exists in `pkg/engine` (only the test file `filter_subtypes_test.go`, which stays), and none of the five functions has a second copy. Costs, replacements and "all" steps call them, and they stay in package `engine`, so no caller changes.
- Why it is a pure move: the five functions are the same text in the same package, and Go doesn't care which file holds them. `targeting.go` still uses `fmt`, `slices` and `ability`, so its imports don't change. The moved code uses only `ability`.
- Not in this task: `doc.go`'s file map still says `targeting.go: target choice and filters`. #121 (c) fixes it last, with `resources.go` from #114, so `doc.go` is not an Allowed File here. `resources.go` is #114.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 27a83dd: `gofmt -l` clean, `go tool task check` passes, `go tool task sim` runs clean, and `go tool task sim-compare -- autoloop/m2-wave-d.1` printed `[PASS]`. Every expected output in the cases below is from that dry run.

### Where it is today (at `autoloop/m2-wave-d.1`, 27a83dd)
| Where | Today |
| :--- | :--- |
| `pkg/engine/targeting.go:1` | the header names three jobs |
| `pkg/engine/targeting.go:199-208` | `pileTarget` (it stays) |
| `pkg/engine/targeting.go:210-296` | `kindOf`, `matchesCard`, `legalTarget`, `matchesPrinted`, `matchesFilter` (the end of the file) |

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)

### 1. `pkg/engine/filter.go`, a new file
Its whole content, ending in one line break after the last `}`. Everything from `// kindOf maps` on is `targeting.go`'s lines 210-296, unchanged, including the long comment line on `matchesFilter`; don't rewrap it:
```go
// Filters: checking assets and cards in discard piles against an ability's filter (FR-TGT-001, FR-ZON-004, FR-KWD-002h).

package engine

import "animus-mundi/pkg/engine/ability"

// kindOf maps a printed card type to its filter kind.
func kindOf(t CardType) ability.Kinds {
	switch t {
	case TypeUnit:
		return ability.KindUnit
	case TypeStructure:
		return ability.KindStructure
	case TypeHeadquarters:
		return ability.KindHeadquarters
	case TypeManeuver:
		return ability.KindManeuver
	}
	return 0
}

// matchesCard reports whether card inst, in a discard pile, meets every
// criterion of the pile filter f for an ability controlled by ctl whose own
// card is src: allegiance by the card's owner, Kinds, Subtypes and Scratch by its
// printed characteristics (FR-ZON-004), Other not src. f must be InPile.
func (g *Game) matchesCard(ctl PlayerID, src *Instance, f ability.Filter, inst *Instance) bool {
	if !f.InPile() || (f.Other && inst == src) {
		return false
	}
	return matchesPrinted(ctl, f, inst.Owner, inst.Card)
}

// legalTarget reports whether battlefield asset a meets every criterion of f
// for an ability controlled by ctl whose source is asset src. An opponent's
// Ward asset is never legal (FR-KWD-002h).
func (g *Game) legalTarget(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	if a.Controller != ctl && g.Keywords(a).Ward {
		return false
	}
	return g.matchesFilter(ctl, src, f, a)
}

// matchesPrinted is the card-level part of a filter, shared by assets and
// cards in piles: Allegiance by side, the player the entity belongs to (relative
// to ctl), then Kinds, Subtypes and Scratch by the card's printed characteristics.
func matchesPrinted(ctl PlayerID, f ability.Filter, side PlayerID, c *Card) bool {
	switch f.Allegiance {
	case ability.AllegianceAllied:
		if side != ctl {
			return false
		}
	case ability.AllegianceOpposing:
		if side == ctl {
			return false
		}
	}
	if f.Kinds != 0 && f.Kinds&kindOf(c.Type) == 0 {
		return false
	}
	if len(f.Subtypes) > 0 {
		found := false
		for _, st := range f.Subtypes {
			if c.HasSubtype(st) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	switch f.Scratch {
	case ability.ScratchOnly:
		return c.IsScratchCompany()
	case ability.ScratchExcluded:
		return !c.IsScratchCompany()
	}
	return true
}

// matchesFilter reports whether battlefield asset a meets every criterion of f
// (allegiance relative to ctl, kinds, subtypes, Scratch, Other: not the source asset
// src, Zone: currently in that battlefield zone). Targeting adds Ward on top (legalTarget); an untargeted "all" step uses
// this alone.
func (g *Game) matchesFilter(ctl PlayerID, src int, f ability.Filter, a *Asset) bool {
	if f.Other && a.ID == src {
		return false
	}
	if f.InPile() || !deployedTo(a, f.Zone) {
		return false
	}
	return matchesPrinted(ctl, f, a.Controller, a.Printed())
}
```

### 2. `pkg/engine/targeting.go`, the end of the file
Delete lines 209-296: the blank line after `pileTarget`'s closing brace, through the end of the file. The file then ends with `pileTarget`'s closing brace (line 208) and one line break.

### 3. `pkg/engine/targeting.go`, its header
Replace line 1 with:
```go
// Targeting: choosing targets on placement and untargeted picks on resolution (FR-TGT-001, FR-TGT-003, SVY-R14). Filters are filter.go's.
```

No other file changes. The import block of `targeting.go` stays as it is.

## Acceptance Cases (harness tasks only; written by the lead)
No new tests: a pure move, so no `test(T-0175)` commit. W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them. Stage the new file (`git -C W add pkg/engine/filter.go`) before running cases 3 and 4, because `git grep` and `git diff` don't see an untracked file.
1. **Existing tests unchanged** (#113): `git -C W diff autoloop/m2-wave-d.1 -- '*_test.go'` prints nothing, and `go -C W test ./...` passes.
2. **Sims byte-identical** (#113): `go -C W tool task sim-compare -- autoloop/m2-wave-d.1` prints `[PASS]`.
3. **The same declarations, in their new places** (#113): `git -C W grep -n -e 'func kindOf' -e ') matchesCard' -e ') legalTarget' -e 'func matchesPrinted' -e ') matchesFilter' -e ') pileTarget' -- pkg/engine` prints exactly six lines: `pkg/engine/filter.go:8`, `:26`, `:36`, `:46` and `:85`, and `pkg/engine/targeting.go:203`.
4. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.1 --numstat -- pkg cmd` prints exactly these lines: `93	0	pkg/engine/filter.go` and `1	89	pkg/engine/targeting.go`.
5. **Fences:** `go -C W run ./cmd/cardtown lint fences` exits 0.
6. **Nothing else changed:** `git -C W diff autoloop/m2-wave-d.1 --stat` lists only those two files and this task file.
7. `gofmt -l W/pkg/engine` prints nothing, and `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/filter.go`
- `pkg/engine/targeting.go`

## Done When
- [ ] Cases 1–7 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.1` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-05, lead-architect (wave D.1 run, session 6): synced `main` into the run branch first (27a83dd, bringing 17fbf28 and directive #122; the only conflict was `BOARD.md`'s counters, resolved to `#122` and `T-0174`). Wrote this task from board directive #113, next in board order. Not `WAVE-COMPLETE`: #113–#119 and #121 are `open: before the next Deliverables task`. #122 is `open: after Wave D.1 (#121), before Wave E`, so it doesn't apply yet. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Searched for existing equivalents: the finds are in Inputs. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #8: no `deck.go`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120: no such card). Next: `kernel-engineer` on Flash.
