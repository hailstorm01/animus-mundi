# M2 Build Board: Animus vs Path Starter Set

- **Integration branch:** `rework/executable-rules`
- **Mode:** `serial` (at most one task in `tests`, `in-progress` or `changes`)
- **Design note:** [`docs/architecture/m2-ability-model.md`](../../../docs/architecture/m2-ability-model.md) (written 2026-10-19). Shape and card tasks follow its Deliverables order.
- **Scope:** every Animus and Path card plus the starter Headquarters *Animoceps Rager* and *Eternal Watchman*. *Rift* and *Encampment* moved to M3 on 2026-10-19.

Task files are the source of truth. This board is an index.

## Directives for the Lead

Instructions from the user or a readiness review that the lead applies before writing the next task (Runbook D2). When one is applied, the lead changes its status to `applied` and names the task or commit that applied it.

| # | Directive | Source | Status |
| :---: | :--- | :--- | :--- |
| 1 | **Keep the allowlist out of the protected test.** `TestStarterSet_Executable` stays a protected test. Its allowlist of not-yet-executable cards lives outside the test file (for example a variable in `pkg/engine/cards/`), and every card task lists that file under `Allowed Files`. Update the design note's Test Plan to match. | Readiness review, 2026-10-19 | open |
| 2 | **Split the starter-decks deliverable (design note Deliverable 2).** A shape task in which `rules-test-author` writes the `TestFR_SET_001_*` tests, and a harness task with lead-written acceptance cases for the decklist format, `LoadDeck` and the `cardtown sim -deck` option. | Readiness review, 2026-10-19 | open |
| 3 | **Scaffold on Sonnet (Deliverable 1).** Kind `harness`, tier `sonnet`: it fixes the package layout, `Card.Ability` and `LoadCards` that every later task builds on. Fold in T-0001's remaining follow-ups: (a) a missing matrix or survey file returns an error naming the file; (b) the case-2 test asserts exactly one `[ERROR]` line. | Readiness review, 2026-10-19 | open |
| 4 | **Split wave A1 (Deliverable 3)** into 3–4 tasks under the D2 size rule. For example: events, the pending-trigger buffer and APNAP placement; the `Chooser` (scripted and random); last-known effective state; untargeted steps plus the 8 A1 cards. | Readiness review, 2026-10-19 | open |
| 5 | **Show the user T-0002 before dispatching it.** The first real task after the trial gets a user look before any persona runs it. Later tasks dispatch normally unless the user adds a directive saying otherwise. | User, 2026-10-19 | open |

## Tasks

| id | title | status |
| :--- | :--- | :--- |
| T-0001 | Add a fences check to cardtown lint | done |
