---
id: T-0055
title: Make Animoceps Rager executable, the first Headquarters ability (wave B card, flash)
kind: card
tier: flash
wave: B
status: tests
branch: m2/T-0055-animoceps-rager
depends_on: [T-0054]
claims: [card:Animoceps Rager, file:pkg/engine/cards/headquarters.go]
cards: [Animoceps Rager]
requirements: []
---

## Goal
The Animus starter Headquarters *Animoceps Rager* runs in the kernel: "Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn." It is plain `ability` data in a new `pkg/engine/cards/headquarters.go` (design note Boundaries, board directive #45).

## Inputs
- Card (`data/cards/units-and-structures-HQ.csv`): Animus Headquarters, 3/20, yields 1R 2T.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 106, and the §4 note that "is dealt damage" uses the Unbowed threshold.
- Shapes it uses: `HookDamaged` (T-0050), `OpScratch`, and `Trigger.MaxPerTurn` (T-0054).

### Definition (the engineer adds it to a new `headquartersDefs` map in `pkg/engine/cards/headquarters.go`, registered in `init` like `pathDefs`)
- One trigger: `Hook: ability.HookDamaged`, `MaxPerTurn: 1`, no targets.
- One step: `Op: ability.OpScratch`, `Amount: ability.Value{N: 1}`.
- `Text` holds the printed rules text. The CSV cell is one line, so `Text` has one entry with both sentences.

### Also in this task
- `cmd/cardtown/sim.go` prints "HQ abilities are ignored." That stops being true. Replace both lines with a line that reports, for each seat, whether its Headquarters is executable (`Supported()`), in the style of the "main cards executable now" lines.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by rules-test-author in a new `pkg/engine/cards_headquarters_test.go`. Run with `go test ./pkg/engine -run TestCard_AnimocepsRager`. Load the real card with `s.card("Animoceps Rager")`. The lead's guidance on what to cover follows; the author names the tests and lists them here.

**Firing:**
- Combat damage to the Rager (an enemy unit attacks it) Scratches 1 for the Rager's controller. That is the defending player, on the opponent's turn, because the trigger has no "on your turn" gate.
- Effect damage to the Rager Scratches 1. Use the real *Chainbreaker Sappers*, which the Path starter deck plays ("deal 1 damage to chosen structure or Headquarters").
- Retaliation: the Rager attacks (a Headquarters may attack, `FR-CAP-001`) a defender with power above 0. The retaliation damage Scratches 1 for the Rager's controller, on its own turn. The Rager is Exhausted by the attack; the Scratch doesn't need it Readied.
- The Scratch resolves through the stack like any trigger. After it resolves, the stack is empty and the invariants hold.

**The once-per-turn limit (`Trigger.MaxPerTurn`, `FR-TRN-001`):**
- Two damage instances in one turn (two attacks, or an attack and a Sappers) Scratch only once.
- The limit resets at the start of every turn, either player's: damaged on turn N and again on turn N+1, it Scratches both times.

**Not firing:**
- Damage to another asset, including the other player's Headquarters, does not Scratch.
- An Eternal Watchman (the Path Headquarters, no damage trigger) being damaged does not Scratch for anyone.

**Leave out (do not test):**
- Damage reduced to zero by Armor or Resistant. The survey (§4 line 78) records the threshold as "noted but unreachable" with an unruled default: no starter card gives a Headquarters Armor or Resistant. Testing it would pin a guessed ruling.
- Lethal damage to the Rager. The match ends at the sweep, so the trigger's resolution isn't observable.

**Pre-check by the lead (for D4):** the tests cite the card text and `FR-TRN-001` and `FR-REP-004` (Stage 5, damage dealt), not `FR-KWD-002a` as if the Rager had Unbowed. The firing tests fail today, because the card loads as unsupported and nothing fires. The "not firing" tests may pass today.

## Allowed Files
- `pkg/engine/cards/headquarters.go` (new)
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)
- `cmd/cardtown/sim.go`
- `pkg/engine/doc.go` (Headquarters abilities move to the "Built" list; remove "(until T-0055)" and rewrap the paragraph to the surrounding width)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_AnimocepsRager'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-23 lead-architect: Wrote the task, `queued` behind T-0054.
- 2026-10-25 lead-architect: All four directives the user put before this task (#49, #52, #53, #54) are applied (T-0058 to T-0061), and T-0054 is `done`. Re-checked the task against `main` c86efd6, since it was written before those moves. `ability.HookDamaged`, `Trigger.MaxPerTurn` and `ability.OpScratch` exist. Definitions attach to a Headquarters by card name like any other card (`cardload.go`). The HQ asset carries a card instance (`setup.go:44`), and `dealDamage` fires `HookDamaged` on it (`damage.go:60`). `cards.NotYetExecutable` still lists the Rager. `cmd/cardtown/sim.go:154, 156` still print "HQ abilities are ignored", and `doc.go:34-35` still says "(until T-0055)". No change to Goal, Allowed Files or Done When. Wrote the test guidance above. A Headquarters can attack (`FR-CAP-001`), so retaliation onto the Rager is a case. The Armor/Resistant threshold is left out as an unruled, unreachable default (survey §4). Waiting triggers checked: #6, #7, #8, #15, #20, #50 and #51. None has fired. Directive #14 (deck cards first) is satisfied: the Rager is the Animus starter HQ.
  - Next: status `tests`; dispatch `rules-test-author`.
