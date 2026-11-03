# Card Town Engineering & Design Guidelines

This directory houses the specifications, decision records and card balance data for Card Town. Status and precedence are in [`frd/README.md`](frd/README.md); read it before editing anything under `frd/`.

## Documentation Map
- **Game Rules (active, source of truth):** `frd/01_Game_Rules_FRD.md` (pure game rules, card & asset taxonomy, zone topology, turn phases, 6-layer rules, 5-stage damage mitigation, combat, keyword dictionary). Its executable form is `../pkg/engine`, with scenario tests named `TestFR_<DOM>_<NNN>_*`.
- **Keyword Interaction Matrix (active):** `frd/keyword_interaction_matrix.md`, the bounded register of keyword-pair and keyword-vs-core-rule edge cases, with open items awaiting rulings.
- **Frozen design intent (not synced; FRD 01 wins any conflict):**
  - `frd/02_Card_Data_Pipeline_FRD.md` and `frd/03_Card_Compiler_FRD.md`: rewritten at the start of M3.
  - `frd/04_Client_Engine_Protocol_Networking_FRD.md`: rewritten at the start of M5.
  - `frd/05_Headless_Game_Engine_FRD.md`: design reference while extending `pkg/engine`.
- **M2 Starter Set Ability Survey:** `design/m2_starter_set_ability_survey.md`, which maps every Animus and Path card onto FRD 02's ability model and lists the M2 build waves, FRD 02 gaps and open ruling questions.
- **Design Outcomes:** `design/outcomes/<YYYY-MM-DD>-<slug>.md`, one per `game-designer` discussion (also written as a `Status: open` stub by `lead-architect` when stopping on a rules question, so `@game-designer` picks it up automatically). Each records the user's decisions, the FRD 01 and register wording for `pm-frd-drafter`, and the board directive. The body is frozen once approved. `game-designer` dispatches the `pm-frd-drafter` subagent to apply the outcome and set its `Status:` line to `applied: …`.
- **Frozen Technical Design Document:** `tdd/03_Card_Compiler_TDD.md`, design intent only, frozen with FRD 03. The TDD path is retired: M3 plans with a design note, as M2 did.
- **Design Notes:** `architecture/<milestone>-<topic>.md`, one per milestone topic. `architecture/m2-ability-model.md` fixes the ability shape types, where card definitions live, and the M2 deliverable order.
- **Architecture Decision Records:** `architecture/decisions/*.md`, one page per technical decision, written when the code that needs it is built.
- **Card Data Sheets:** CSVs in `../data/cards/` across Concordia, Soulless, Path, Animus, HQ, and Common.

> [!TIP]
> The FRDs are large (40-90 KB each). Read the section you need by line range rather than whole files.

## Linting & Verification
Run `go tool task check` for every gate at once (gofmt, `go vet`, all tests and `cardtown lint --all`; `Taskfile.yml` lists the other commands). Run `cardtown lint` (auto-detects changed files; `-all` forces everything). `cardtown lint rules` traces engine scenario tests to FRD 01. `go test ./pkg/engine` runs the scenario tests, `go test ./pkg/sim` the seeded random playouts, `go tool task test` every Go test, and `cardtown sim` runs playouts on real card data. Details live in the `cardtown-linter` skill. The pre-commit hook lints staged FRDs, CSVs and engine files automatically.

## Product Management & Rules
Product vision, milestones and `PRODUCT_SUMMARY.md` are owned by `pm-lead`. FRD authoring is driven by `pm-frd-drafter`, for FRD 01 now and for other FRDs when their milestone starts. Technical design and decision records are owned by `lead-architect`, which also owns the build-task queue in `../.agents/tasks/` and reviews all build work (Runbook D). A wave can also be built unattended on an `autoloop/` branch: the user starts `go tool task wave-run` (`../.agents/scripts/waverun`, Runbook D7) and merges the branch afterwards. `rules-test-author` writes each task's failing tests from FRD 01, and `kernel-engineer` makes them pass. `game-designer` leads design discussions with the user, including picking up a `Status: open` ruling question from `lead-architect`, commits the approved outcome and a build-board directive for `lead-architect`, dispatches the `pm-frd-drafter` subagent, and emits a prompt for `pm-lead` only when milestone scope changes. `workflow-steward` looks after the persona system itself: it checks personas' work, edits their skills and routing, and logs open threads and boundary exceptions in `../.agents/steward-log.md`. Persona behavior lives only in `.agents/skills/` (`workflow-steward`, `pm-lead`, `pm-frd-drafter`, `game-designer`, `lead-architect`, `rules-test-author`, `kernel-engineer`). The Claude Code subagents in `../.claude/agents/` (`rules-test-author`, `kernel-engineer`, `pm-frd-drafter`) only pin a model and point at those files.
