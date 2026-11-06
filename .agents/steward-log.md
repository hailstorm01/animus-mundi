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
| 2026-11-06 | **Token review follow-ups.** (1) Deny the three `Artifact*` tools in project settings: 12.5k tokens per VS Code call, no uses in 212 sessions. (2) Narrow Runbook B step 1 to grep the files that only route to a persona, not read them whole. (3) Review shell output in wave sessions, where most of the project's shell tokens are. | User: decide each |
