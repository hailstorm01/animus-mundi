---
id: T-0117
title: A Maneuver can be enacted only if each exactly-one target slot has a legal entity, and a "Choose N" one only with N legal modes (FR-MNV-001, SVY-R29, directive #88 part 2, wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: review
branch: m2/T-0117-enact-needs-targets
depends_on: [T-0116]
claims: [legality:enact-targets]
cards: []
requirements: [FR-MNV-001, FR-MOD-001, FR-TGT-001]
---

## Goal
Enacting a Maneuver is legal only when its effect could be placed with every required target: each exactly-one target slot has a legal entity on declaration, and a "Choose N" Maneuver has at least N legal modes (`SVY-R29`). An illegal Maneuver is absent from `LegalActions`, and applying it is rejected with the card still in hand. This is directive #88's enacting bullet, part 2. T-0116 built the enact action without it.

## Inputs
- FRD 01:
  - `FR-MNV-001` (line 334, `SVY-R29`): a Maneuver can be enacted only if each of its exactly-one target slots has a legal entity when it is declared. A "Choose N" Maneuver also needs at least N legal modes. An "up to N" slot never stops it. A Maneuver that can't be enacted stays in hand. Examples: "Kill chosen unit" with no unit on the battlefield; *Reaver's Rush* without a unit card of power 3 or less in its controller's discard pile.
  - `FR-MOD-001` §5 (line 384, `SVY-R28`, `SVY-R29`, `SVY-R30`): a mode is legal only if each of its exactly-one target slots has a legal entity. A triggered ability is still placed with fewer than N legal modes. That stays as it is.
  - `FR-TGT-001`: the legal entities of a slot.
  - Line 373: a deployment does not need targets. Deploy is not touched.
- Rulings that apply: `SVY-R28`, `SVY-R29`, `SVY-R30` (triggers only).
- Board directive #88, first bullet ("it is a legal action only if each exactly-one target slot of its effect has a legal entity on declaration, and for a `ChooseN` effect only if `legalModes` returns at least N"), tests (a) and (b).
- Kernel today:
  - `legalModes(pl, e)` (`placement.go:36`) holds the exactly-one slot check inline: a slot with `UpTo == 0` and no `targetDecision(pl, i, t).Options` makes the mode illegal.
  - `check`'s `ActEnact` case (`actions.go:156`) runs `handPlay` and the `TypeManeuver` check only.
  - `Apply`'s `ActEnact` case (`actions.go:83-88`) builds `placement{controller: a.Player, inst: inst, label: inst.Card.Name}` inline.
  - `chooseOnPlacement` (`placement.go:26`): an effect with `Modes` resolves only its chosen modes, so a modal effect's own `Targets` are never read (`ability.Effect`: "an effect with Modes has no Targets or Steps of its own").

### Shape (lead decisions, binding)
- **One slot check (reuse, D2.3).** In `placement.go`, move `legalModes`' inner loop into `(g *Game) slotsLegal(pl placement, ts []ability.Target) bool`: true when every slot with `UpTo == 0` has at least one legal entity (`targetDecision`'s `Options`). `legalModes` calls it for each mode, with no change in behavior. There is no second copy of the check.
- **`(g *Game) placeable(pl placement, e ability.Effect) bool`** (`placement.go`): whether e can be placed with everything it requires (`SVY-R29`). With `Modes`, it is `len(g.legalModes(pl, e)) >= e.ChooseN`. Otherwise it is `g.slotsLegal(pl, e.Targets)`. Its doc comment says it applies to enacting a Maneuver only: a trigger is placed whatever it finds (`SVY-R30`, `FR-TGT-003`).
- **One Maneuver placement.** Add `maneuverPlacement(pid PlayerID, inst *Instance) placement` (`placement.go`), giving today's `{controller: pid, inst: inst, label: inst.Card.Name}` with `source` 0. `Apply` and `check` both use it, so legality is checked against the placement the targets are then chosen for. Update `placement`'s doc comment: "a played Maneuver will too" becomes the present tense, naming `maneuverPlacement`.
- **`check`, `ActEnact`** (`actions.go`): after the `TypeManeuver` check, `if !g.placeable(maneuverPlacement(a.Player, inst), inst.Card.Ability.Effect)`, return `illegal("%s has no legal choice for a required target or enough legal modes (SVY-R29)", name)`, or two messages if that reads better: one for a modal effect, one for a target. `LegalActions` already filters through `check`, so it needs no edit.
- **Order of checks:** the card is checked while it is still in hand, before payment. Paying and taking the card out of hand do not change which entities are legal for any filter the kernel has (battlefield assets and discard piles). The Maneuver is never in a discard pile when declared, so it can't target itself.
- **Triggers unchanged.** `chooseModes`' fewer-than-N and no-mode results stay, for triggers (`SVY-R30`). An enacted Maneuver no longer reaches them. Add one sentence to `chooseModes`' doc comment saying so.
- **Out of scope:** distinctness between two slots ("another chosen unit" in *Ambush Tactics*' mode 3) is the delegated dealer's task (#88 tests (c), (d)). No filter compares one slot with another today. Costs locked with a mode (#61, fifth question) wait for their ruling.
- **Sims:** no card gains a `Def.Effect`, so `go tool task sim-compare -- autoloop/m2-wave-c` should pass byte for byte. If it differs, don't change code to make it match: report the first differing game in the Log and hand off.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
Synthetic Maneuvers only, added to `pkg/engine/maneuver_test.go` (from T-0116), reusing its `maneuver`, `steps`, `enactAction` and `offersEnact` helpers. Pile filters can reuse the ones in `pile_target_test.go`. Every existing test passes unchanged.
- `TestFR_MNV_001_CannotEnactWithoutRequiredTarget` (#88 (a)): with no unit on the battlefield, a "Kill chosen unit" Maneuver (any allegiance) is not among the legal actions, and applying it is rejected with `ErrIllegal`: the card stays in hand and the pool is unchanged. Once one unit is on the battlefield, it is a legal action.
- `TestFR_MOD_001_ChooseTwoNeedsTwoLegalModes` (#88 (b)): a "Choose two" Maneuver with two modes, "draw a card" and "kill chosen opposing unit", with no opposing unit, is not a legal action. When the second mode's slot is "up to 1" instead, the Maneuver is a legal action.
- `TestFR_MOD_001_ChooseOneNeedsOneLegalMode`: a "Choose one" Maneuver whose two modes each need a chosen opposing unit is not a legal action with no opposing unit, and is with one.
- `TestFR_MNV_001_UpToSlotNeverStopsEnact`: a "Return up to 1 chosen unit to its owner's hand" Maneuver, with no unit on the battlefield, is a legal action. Enacting it asks no target decision, and the card ends in its owner's discard pile.
- `TestFR_MNV_001_PileTargetNeededToEnact`: a Maneuver with one exactly-one slot for "a unit card in your discard pile" (filter `Zone: ZoneDiscard`, allied, `KindUnit`) and a step from an existing pile-target shape (`pile_target_test.go` shows which steps take one; if none fits a Maneuver, assert legality only). With the enacting player's discard pile empty, or holding only the opponent's unit cards, it is not a legal action. With one of their own unit cards there, it is.
- `TestFR_MNV_001_SecondRequiredSlotAlsoNeeded`: a Maneuver with two exactly-one slots, "chosen allied unit" and "chosen opposing unit", and one step per slot (for example return each to its owner's hand). With only an allied unit, it is not a legal action. With one of each, it is.

## Helpers Needed (engineer adds these to pkg/engine/scenario_test.go)
- None expected.

## Allowed Files
- `pkg/engine/placement.go` (`slotsLegal`, `placeable`, `maneuverPlacement`, `legalModes`, the `placement` and `chooseModes` doc comments)
- `pkg/engine/actions.go` (`check`'s and `Apply`'s `ActEnact` cases)

## Done When
- [ ] The listed tests pass: `go test ./pkg/engine -run 'TestFR_MNV_001|TestFR_MOD_001|TestFR_TGT_001|TestFR_QUE_001_Maneuver|TestCard_SupportedManeuver'`
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks
- [ ] `go tool task sim-compare -- autoloop/m2-wave-c` passes (no card gains a `Def.Effect`; see Shape, Sims)
- [ ] The exactly-one slot check is written once, in `slotsLegal`, which `legalModes` and `placeable` both call.

## Log
- 2026-10-30 lead-architect: wave C run, session 2. No `main` commits to sync. Directives read on the run branch: #88 is `open`, and its enacting bullet's second part is this task (#88 tests (a), (b)). It edits two non-test files, `placement.go` and `actions.go`. Waiting triggers checked: #58 (no exhaust in `claims:`), #59 (2) (no repeated targets), #61's fifth question (no card with modes in `cards:`), #71 (no each-player effect), #15 and #8 (no hook or `deck.go` change). None has fired. Duplicates searched: the exactly-one slot check exists once, inline in `legalModes`. This task moves it into `slotsLegal` so that `placeable` shares it, and puts the Maneuver placement in one place (`maneuverPlacement`) for `check` and `Apply`. Status `tests`; dispatching `rules-test-author`. After this task come the delegated dealer (#88 (c), (d), which also corrects `resolveAbility`'s "triggered ability" comment in `effects.go`), `MOVE_ZONE` (#88 (e)), the client harness task (T-0116, Shape, Clients) and the card tasks, deck cards first (#14 (a)).
- 2026-10-30 rules-test-author: tests written in `pkg/engine/maneuver_test.go` (plus a small `killChosenUnit` helper there): `TestFR_MNV_001_CannotEnactWithoutRequiredTarget`, `TestFR_MOD_001_ChooseTwoNeedsTwoLegalModes`, `TestFR_MOD_001_ChooseOneNeedsOneLegalMode`, `TestFR_MNV_001_UpToSlotNeverStopsEnact`, `TestFR_MNV_001_PileTargetNeededToEnact`, `TestFR_MNV_001_SecondRequiredSlotAlsoNeeded`. Helpers requested: none. `go vet` clean, `cardtown lint rules` traces all. Five fail as expected; `UpToSlotNeverStopsEnact` passes already (it pins existing behavior against regression). Nothing left out. Status `tests-review`.
- 2026-10-30 lead-architect: tests reviewed (D4), approved. The six tests check `SVY-R29` as `FR-MNV-001` (line 334) and `FR-MOD-001` §5 state it, through `LegalActions` and `Apply`, not through implementation details: a required slot with no legal entity (on the battlefield, in a discard pile, or in a second slot) and too few legal modes for "Choose one" and "Choose two" make the Maneuver illegal, and the rejected card stays in hand with its cost unpaid. An "up to 1" slot never does. `TestFR_MNV_001_UpToSlotNeverStopsEnact` already passes, which is right: it pins the "up to N" side against a too-strict check. Re-ran in a review worktree: the other five fail, and only on the missing check. Status `ready`; dispatching `kernel-engineer` (sonnet).
- 2026-10-30 kernel-engineer: implemented. `placement.go`: `slotsLegal` (the exactly-one slot check, moved out of `legalModes`, which now calls it), `placeable` (modal: `len(legalModes) >= ChooseN`; else `slotsLegal`), `maneuverPlacement`; `placement` and `chooseModes` doc comments updated. `actions.go`: `check`'s `ActEnact` rejects with an SVY-R29 message when `!placeable`, and `Apply` uses `maneuverPlacement`. Reuse: `legalModes`, `targetDecision`, `chooseOnPlacement` unchanged; no second copy of the slot check. Ran: the Done When test pattern (ok), `go tool task check` (pass), `go tool task sim` (clean), `go tool task sim-compare -- autoloop/m2-wave-c` (byte for byte match). Nothing unusual for the reviewer; the one message covers both modal and target cases.
