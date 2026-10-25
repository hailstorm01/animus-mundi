---
id: T-0064
title: Make Animoceps Shambler executable (wave B card, flash)
kind: card
tier: flash
wave: B
status: queued
branch: m2/T-0064-animoceps-shambler
depends_on: [T-0063]
claims: [card:Animoceps Shambler]
cards: [Animoceps Shambler]
requirements: []
---

## Goal
*Animoceps Shambler* (Animus unit, 2/2, three copies in the Animus starter deck) runs in the kernel: "Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1." It is plain `ability` data in `pkg/engine/cards/animus.go`.

## Inputs
- Card (`data/cards/*.csv`): Animus unit, Fungus, cost 2, 2/2, yields 1R 2T, deploys to Reserves.
- Survey row: `docs/design/m2_starter_set_ability_survey.md` line 107.
- Shapes it uses: `ability.HookExhausted` (T-0063), a target with an allied filter and a Main Theater zone (T-0017, T-0052), and `OpHeal` (T-0019).
- Rulings: `SVY-R17` (placement), `SVY-R3` (targets locked on placement), `SVY-R16` (timeout picks the first legal option in fixed order).

### Definition (the engineer adds it to `animusDefs`)
- One trigger: `Hook: ability.HookExhausted`, one target slot: an allied asset in the Main Theater, any kind.
- One step: `Op: ability.OpHeal`, `Amount: ability.Value{N: 1}`, `Subject: ability.SubjectTarget`, slot 0.
- `Text` holds the printed rules text, one entry.

## Tests (shape and card tasks: written by rules-test-author; engineers never edit these)
To be written by rules-test-author in `pkg/engine/cards_animus_test.go` (or a new file if that is the pattern). Guidance; the author names the rest and lists them here.
- **Named (directive #47 part 4):** `TestCard_AnimocepsShambler_MarchTriggerResolvesBeforeTheMarch` and `TestCard_AnimocepsShambler_GenerateTriggerResolvesBeforeTheNextAction`.
- Its target is locked on placement, so a Shambler marching from Reserves cannot choose itself: it is still in Reserves when the target is chosen. A Shambler attacking from the Main Theater may choose itself.
- Only allied assets in the Main Theater are offered: not Reserves, not the Headquarters, not enemy assets.
- No legal target: nothing is asked and nothing happens.
- Heal 1 removes 1 damage from the chosen asset, and does nothing to an undamaged one.
- An attack's trigger resolves before combat damage. A Shambler in the Main Theater with 1 damage attacks a defender with power 1 and chooses itself. It heals to full before the retaliation, so it survives with 1 damage. If the heal came after combat it would have died.

## Allowed Files
- `pkg/engine/cards/animus.go`
- `pkg/engine/cards/cards.go` (remove the name from `NotYetExecutable`)

## Done When
- [ ] The card's tests pass: `go test ./pkg/engine -run 'TestCard_AnimocepsShambler'`
- [ ] `TestDefs_MatchRulesText` and `TestStarterSet_Executable` pass unchanged
- [ ] `go tool task check` passes
- [ ] `go tool task sim` runs clean; report the before and after summary lines in the Log

## Log
- 2026-10-25 lead-architect: Wrote the task, `queued` behind T-0063 (the exhaust hook).
