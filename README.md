# Animus Mundi

Animus Mundi (working title Card Town) is a strategy card game. A physical prototype is in playtesting. This repository holds its rules, its card data and an executable rules engine written in Go. A Godot client comes later (M4).

## Status

M2 is in progress: the **Animus vs Path starter set**. It covers construction-legal starter decks for both factions (`data/decks/`), every Animus and Path card made executable in the engine wave by wave, and a text-mode `cardtown play` for human playtests. Rift and Encampment move to M3. Milestones M1–M6 are in [`docs/PRODUCT_SUMMARY.md`](docs/PRODUCT_SUMMARY.md).

## Developer Setup

The only prerequisite is Go (see `go.mod` for the version). [Task](https://taskfile.dev) is pinned as a Go tool in `go.mod`, so it needs no separate install:

```sh
go tool task              # list every command
go tool task build        # build the cardtown CLI into bin/
go tool task check        # gofmt, vet, tests and every linter (the gate for every change)
go tool task sim          # play the Animus vs Path starter decks
go tool task sim-log      # play one starter game and print its events
go tool task sim-events -- out.jsonl   # write the starter games' events as JSON Lines
go tool task sim-compare  # pure-move check: both sims match main byte for byte (or pass another ref after --)
go tool task hook         # install the pre-commit hook (once per clone)
```

If Task is installed globally, `task build` and the rest work too. Commands are defined in `Taskfile.yml`. The engine and CLI use only the Go standard library; Task is the one tool dependency. The pre-commit hook lints staged files and rebuilds a stale `bin/cardtown`.

## The `cardtown` CLI

```sh
cardtown sim -deck starter -p0 ANIMUS -p1 PATH -games 500      # playouts on real card data
cardtown sim -deck starter -p0 ANIMUS -p1 PATH -log -games 1 -seed 7   # one game, readable events
cardtown sim -deck starter -p0 ANIMUS -p1 PATH -events out.jsonl      # every game's events as JSON Lines
cardtown replay -events out.jsonl -game 3 -at 120              # rebuild game 3 and print its state at event 120
cardtown play -deck starter -p0 ANIMUS -p1 PATH                # play a match against the bot (go tool task play)
cardtown lint --all                                            # every linter
```

- **`sim`** plays matches between two seat policies: `-policy progress` (the default) or `random`. Without `-deck` it builds stress decks from each faction's executable cards. Every game is checked for termination, engine invariants and **rule monitors**. A monitor checks one FRD 01 rule against every event and prints a coverage table of how often each rule was exercised. The same seed always gives byte-identical output.
- **`-events`** writes the structured event stream: a match-start record (seed, decks, Headquarters, mulligans, engine build), every event and player decision, a full snapshot at each turn start, and a match-end record. Match metrics will be built as queries over these files.
- **`replay`** rebuilds a recorded game deterministically to any event. It refuses a file recorded at a different git revision, warns when either side has uncommitted changes, and reports the first event where the replay diverges from the record.
- **`play`** is a text-mode match against a sim policy. The screen and the event log come only from one seat's view (`pkg/view`), so the human never sees the bot's hand, its deck or a face-down card. Actions and decision options are numbered picks; `q` quits. Cards the kernel cannot run yet are marked `[not executable yet]`.
- **`lint`** checks FRD cross-references, rules traceability (`TestFR_*` tests against FRD 01 requirements), TDDs, card CSVs, rules-text terminology and package fences (no card names in package code, and import direction between packages).
- `hook`, `audit`, `tdd` and `whats-up` support the docs workflow. Run `cardtown <command> -h` for flags.

## How the Rules Are Built

- **One rules authority.** [`docs/frd/01_Game_Rules_FRD.md`](docs/frd/01_Game_Rules_FRD.md) is the source of truth. `pkg/engine` is its executable form. Scenario tests are named after requirement IDs (`TestFR_CBT_002_*`), and card tests are named `TestCard_*`. FRDs 02–05 are frozen design intent until their milestones.
- **Abilities are data.** Card abilities are Go values built from the shapes in `pkg/engine/ability`, keyed by card name in `pkg/engine/cards`. Raw stats and rules text live only in `data/cards/*.csv`. Cards that cannot execute yet are listed in `cards.NotYetExecutable`, which shrinks as each wave lands.
- **Player choices** go through a synchronous `Chooser` with rules-defined timeout fallbacks (ADR-0003). Decisions are always picks from options the engine offers.
- **Decisions** are recorded one page each in [`docs/architecture/decisions/`](docs/architecture/decisions/). The M2 design note is [`docs/architecture/m2-ability-model.md`](docs/architecture/m2-ability-model.md).

## How Work Gets Done

M2 is built by an agent task loop. The lead architect (Opus) writes tasks in `.agents/tasks/m2/` and reviews every change. A rules test author writes the failing tests from FRD 01, without reading the engine code. A kernel engineer (Sonnet or Gemini Flash) makes them pass. Each task merges into the integration branch once `go tool task check` passes. [`BOARD.md`](.agents/tasks/m2/BOARD.md) holds the task list and the owner's directives for the lead. The loop runs **one task per human prompt**. Persona instructions are in `.agents/skills/`, and routing and guard rails are in `.agents/rules/AGENTS.md`.

## Repository Structure

- `cmd/cardtown/` - The developer CLI, one file per command.
- `pkg/engine/` - The rules engine (FRD 01 made executable): game state, turns, actions, the stack, combat and damage, triggers, targeting, choices, events and invariants.
  - `pkg/engine/ability/` - Ability shapes as plain data: hooks, ops, filters, targets.
  - `pkg/engine/cards/` - Card ability definitions by name, and the not-yet-executable list.
- `pkg/sim/` - Match driving: seat policies, playouts, `Simulate`, stress decks and rule monitors. Imports `pkg/engine` only.
- `pkg/eventlog/` - Event formatting, JSON Lines output and replay. Imports `pkg/engine` only.
- `pkg/view/` - One seat's view of a game and its event log, for `cardtown play`. Imports `pkg/engine` and `pkg/eventlog`.
- `pkg/toolchain/` - The linters, pre-commit hook, audit registry, TDD slicing and docs status behind the CLI.
- `data/cards/` - Card data CSVs, one per faction and card type.
- `data/decks/` - Decklists (`<faction>-<name>.txt`), currently the Animus and Path starter decks.
- `docs/` - Product summary, FRDs (`frd/`), architecture notes and decision records (`architecture/`), the M2 ability survey (`design/`), TDDs (`tdd/`) and the audit registry (`audit/`). [`docs/AGENTS.md`](docs/AGENTS.md) maps them.
- `.agents/` - Agent personas (`skills/`), routing rules (`rules/`) and the build-task queue (`tasks/`). `.claude/agents/` pins models for the Claude Code subagents.
- `tools/sheets/` - Publishes the `data/cards` CSVs to the Card Town Google Sheets (one tab per faction).
- `tools/linter/` - The card standardization audit data and the pre-commit hook script.
- `src/` - Godot project root (`res://`), empty until M4.
