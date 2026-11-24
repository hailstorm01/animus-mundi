# M2 Build Board: Animus vs Path Starter Set

- **Integration branch:** `main` (was `rework/executable-rules` until 2026-10-24, when main was fast-forwarded to it)
- **Mode:** `serial` (at most one task in `tests`, `in-progress` or `changes`)
- **Design note:** [`docs/architecture/m2-ability-model.md`](../../../docs/architecture/m2-ability-model.md) (written 2026-10-19). Shape and card tasks follow its Deliverables order.
- **Scope:** every Animus and Path card plus the starter Headquarters *Animoceps Rager* and *Eternal Watchman*. *Rift* and *Encampment* moved to M3 on 2026-10-19.
- **Closed:** M2 was done on 2026-11-21. On 2026-11-23 its 8 waiting directives (#6, #7, #15, #71, #91, #106, #120, #142) moved, unchanged, to the [M3 board](../m3/BOARD.md), which continues the numbering from `#179`. This board has no `Last directive` line, so `go tool task next-directive` skips this board and reads M3's.
- **Last task:** `T-0293` (done tasks and their files live in [`archive/tasks/`](archive/tasks/task_archive.md))

Task files are the source of truth. This board is an index of active and waiting work.

## Directives for the Lead

Instructions from the user, a readiness review, a user-approved `game-designer` Design Outcome, or user-approved findings from the lead's structural review (Runbook E) that the lead applies before writing the next task (Runbook D2). When a directive is fully applied, the lead sets its status to `applied` (naming the task or commit) and moves its row to [`archive/directives/directive_archive.md`](archive/directives/directive_archive.md). A `waiting: <trigger>` directive is applied only when its trigger happens.

| # | Directive | Source | Status |
| :---: | :--- | :--- | :--- |

## Tasks

Active and queued tasks only. When a task reaches `done`, the lead moves its file and its row here to [`archive/tasks/`](archive/tasks/task_archive.md).

| id | title | status |
| :--- | :--- | :--- |
