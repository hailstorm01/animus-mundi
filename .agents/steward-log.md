# Workflow Steward Log

Kept by `workflow-steward` (`.agents/skills/workflow-steward/SKILL.md`). The steward reads this at the start of every session, so it holds only what is still open. Remove a thread when it is resolved or declined. Finished changes and boundary exceptions are in `.agents/steward-history/<YYYY-MM-DD>.md`, which isn't read at session start.

## Open Threads

| Opened | Thread | Waiting on |
| :--- | :--- | :--- |
| 2026-10-21 | **Explainer persona (deferred).** Structural judgment went to `lead-architect` (Runbook E, placement checks in D2/D5). What remains is an optional read-only persona that explains how parts of the code work. | User: only if wanted later |
| 2026-10-24 | **Product summary's build status is behind.** `docs/PRODUCT_SUMMARY.md` line 195 says 155 tasks are done, waves D, E and F are next, and #103 is open. At 2026-11-06: T-0186 is the latest task, D and D.1 are merged, #103 is applied, and E is next. The retired branch name it opened with is gone. Prompt handed to the user for `@pm-lead`. | User to run pm-lead |
| 2026-10-26 | **Wave runner notes (from lead-architect's review of wave B run 1, `.agents/tasks/m2/runs/autoloop-m2-wave-b-review.md` §3).** (1) done: the Go runner times each session itself (see history, 2026-10-27). (2) Flash usage isn't in the spend budget: `--max-budget-usd` sees only Claude cost. (3) done: D7.3's `Kept` line. | User: take up (2), or decline |
| 2026-10-27 | **agy's `go` rule fences the shell, not Go.** The allow-list still permits `go run` and `go generate`, which run arbitrary code. | User: tighten, or accept |
| 2026-10-27 | **Empty folder `C:\tmp\la` left behind** by wave B session 11. The safety check refuses the steward's `rmdir`. | User to delete |
| 2026-11-06 | **Check the runner's token changes on the next wave run** (see history, 2026-11-06). Sessions should start near 18.8k tokens and read the skill without searching first. A `STOP` for a missing tool means the flags need another look. | Next wave run |
| 2026-11-11 | **Check the wave F rerun.** Session 19 should resolve the `BOARD.md` sync under D7.2 step 2 (`main`'s #143 next to the run branch's #142, `Last directive` set to #143), and any directive it adds should take its number from `go tool task next-directive`. | Next wave run |
| 2026-11-23 | **M6 has two names in `docs/PRODUCT_SUMMARY.md` §4.** The mermaid chart says "Metagame, Deckbuilder, Card Studio & Authoring Pipeline" and the `### M6` heading says "… Card Studio & Telemetry". pm-lead's skill now uses the chart's name (it reflects the 2026-11-21 move). Prompt for `@pm-lead` handed to the user. | User to run pm-lead |
| 2026-11-23 | **`docs/AGENTS.md` line 16 still lists "the catalog export"** among the M3 design note's contents. Drop it once `lead-architect` removes Deliverable 7 from `docs/architecture/m3-full-catalog.md`, not before. | lead-architect: design note edit |
| 2026-11-23 | **Check the first sync-handoff chain** (`.agents/rules/sync-handoffs.md`). Did the coordinator dispatch rather than hand out prompts, did each receiver work from the named commits and end with `Ripples:`, did it stop on decisions, and did the chain report list every hop? | Next pm-lead, drafter or architect change with knock-on effects |
| 2026-11-23 | **`docs/PRODUCT_SUMMARY.md` §2.1 intro (line 75) still calls FRD 03 active** ("FRD 01, FRD 02 and FRD 03 are active specifications … FRDs 02 and 03 were rewritten short at the M3 start"), while item 3 and §4.1 say it is frozen until M5. Prompt for `@pm-lead` handed to the user. | User to run pm-lead |
