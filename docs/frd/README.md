# Functional Requirements Documents

## What is active

| Document | Status | Checked by |
| :-- | :-- | :-- |
| [01 Game Rules](01_Game_Rules_FRD.md) | **Active, single source of truth for rules** | `pkg/engine` scenario tests, `cardtown lint rules`, `pm-devils-advocate` |
| [Keyword Interaction Matrix](keyword_interaction_matrix.md) | **Active, bounded edge-case register for FRD 01** | Closes when every cell has a ruling |
| [02 Card Data Pipeline](02_Card_Data_Pipeline_FRD.md) | Frozen as design intent | Rewritten at the start of M3 |
| [03 Card Compiler](03_Card_Compiler_FRD.md) | Frozen as design intent | Rewritten at the start of M3 |
| [04 Client-Engine Protocol & Networking](04_Client_Engine_Protocol_Networking_FRD.md) | Frozen as design intent | Rewritten at the start of M5 |
| [05 Headless Game Engine](05_Headless_Game_Engine_FRD.md) | Frozen as design reference | `pkg/engine` is the engine's source of truth |

Milestones (M1–M6) are defined in [`PRODUCT_SUMMARY.md`](../PRODUCT_SUMMARY.md). The reasoning behind this split is in [ADR-0001](../architecture/decisions/0001-executable-rules-and-frozen-frds.md).

## Precedence

1. **FRD 01** says what the rules are.
2. **`pkg/engine` scenario tests** say what the rules do. A test named `TestFR_CBT_002_...` pins `FR-CBT-002`. If a test and FRD 01 disagree, one of them is a bug: fix it in the same change.
3. **Every other document defers.** Frozen FRDs, TDDs and decision records may cite an FRD 01 requirement but must not restate its behavior in their own words. A restated rule is how FRD 01 and FRD 05 ended up disagreeing about Berserk.

## Executable rules

`pkg/engine` is FRD 01 in code. M1 covers setup, turns, the economy, deploy/march/attack/scratch, the stack, combat, the damage pipeline, defeat conditions, and the ten standalone keywords. Its package doc lists what is out of scope until M2.

```powershell
go test ./pkg/engine          # scenario tests + 300 random-policy fuzz playouts
cardtown lint rules           # which active FRD 01 requirements have a scenario test
cardtown sim -p0 SOULLESS -p1 CONCORDIA -games 500   # playouts on real card data
```

**Turning a question into a ruling:**

1. Someone asks "what happens if …" (a DA finding, a playtest dispute, a matrix cell).
2. If FRD 01 answers it, write the scenario test. A failing test is a kernel bug.
3. If FRD 01 does not answer it, record it in the [matrix](keyword_interaction_matrix.md) (or as a DA finding for non-keyword rules), get a ruling, put it in FRD 01, then write the test.
4. A physical-prototype ruling follows the same path: FRD 01 text plus a scenario test.

Where FRD 01 is silent and the kernel has to pick something to run at all, the code carries an `OPEN(KIM-...)` comment naming the matrix item.

## Writing a new FRD

New FRDs (for example 06 Server, 07 Client, 08 Deckbuilder) are drafted when their milestone starts, not before. They describe what the player or operator needs. Technology choices, wire formats, and performance budgets belong in decision records written alongside the code.
