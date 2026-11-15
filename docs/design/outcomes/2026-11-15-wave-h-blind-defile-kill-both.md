Status: open (lead-architect, 2026-11-15)
Task / Directive: T-0242 (branch: m2/autoloop-wave-h/T-0242-blind-defile, not created yet; run branch autoloop/m2-wave-h)
Cards: Blind Defile

### ❓ Open Ruling Question: do Blind Defile's two kills depend on each other? (2026-11-15)

- **Question:**
  1. *Blind Defile* reads "Whenever an opposing unit attacks, you may kill this structure and that attacking unit." When its trigger resolves and the Defile has already left the battlefield, is the attacking unit still killed?
  2. When its trigger resolves and the attacking unit has already left the battlefield (for example, a second *Blind Defile* of the same player resolved first and killed it), is the Defile still killed?
  3. Should the card read in the `SVY-R14` form, "you may kill this structure. If you do, kill that attacking unit.", so that FRD 01's existing rule answers 1 (no) and 2 (yes, it dies, since "you may" was answered on placement and the kill of the structure is carried out on resolution)?
- **Source:** wave H run (design note Deliverable 6m), the first wave H card, before its card task T-0242 is written. The shape it needs (a trigger on another asset's attack, and "that attacking unit" as a step subject) doesn't depend on the answer and is built first as T-0241.
- **Cards affected:** *Blind Defile* (Path structure, `data/cards/units-and-structures-path.csv:7`: "Whenever an opposing unit attacks, you may kill this structure and that attacking unit."). It is in M2 (wave H, `cards.NotYetExecutable`) and in neither starter deck.
- **Current FRD 01 & engine state:**
  - `FR-TRG-005` (FRD 01 v3.25, line 460) names "*you may kill this structure*" on *Blind Defile* as the optional non-resource action that is "carried out when the trigger resolves (`SVY-R14`)". That treats the self-kill as `SVY-R14`'s X, but the card says "and", not "If you do", so FRD 01 doesn't say whether killing the attacker (the Y) depends on it.
  - `SVY-R14` ("You May X. If You Do, Y", line 391) settles the dependency only for that wording: Y applies only if X was carried out, and a replaced X still counts (`SVY-R15`, line 518).
  - Untargeted selections (line 388): "If nothing matches, only that clause does nothing." Read on its own, that makes each kill independent: each named asset still on the battlefield is killed. "That attacking unit" isn't "chosen", so it is no target (Ward and `FR-TGT-003` don't apply).
  - `SVY-R9` / `FR-MOD-001` §7 (line 431): the "you may" is answered on placement. `SVY-R48` (line 432) doesn't skip it, because neither kill acts on a target.
  - Engine: a kill step kills the assets it names that are still on the battlefield (`stepKill`, [effects.go:260](../../../pkg/engine/effects.go)); `Step.IfDone` makes a step run only if the step before it did something ([ability.go](../../../pkg/engine/ability/ability.go), `IfDone`). The card can be written either way; it is the choice between a plain second kill step and an `IfDone` one.
  - Reach in M2: case 2 happens with two *Blind Defiles* under one player (both trigger on the same attack, `FR-CBT-002` Step 2, and the first to resolve kills the attacker). Case 1 can't happen in M2 except in principle: nothing in Animus or Path removes the Defile between its trigger's placement and its resolution, because its controller's triggers are placed above the active player's (`FR-TRG-003`) and no other M2 card watches an opposing attack.
- **Lead's recommendation:** read it as `SVY-R14` (question 3, and reword the card text to match): the Defile is killed on resolution if it is still there, and the attacker is killed only if the Defile was (a replaced kill counts, `SVY-R15`). That keeps *Blind Defile* a trade, as the card intends, and needs no new rule. A second Defile then dies for nothing, which its controller can avoid by declining its "you may" on placement.
