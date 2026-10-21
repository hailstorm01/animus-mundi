---
id: T-0025
title: Rule monitors in every sim game, with a coverage table (framework and three monitors)
kind: harness
tier: sonnet
wave: none (board directive #23, part 1 of 2)
status: in-progress
branch: m2/T-0025-rule-monitors
depends_on: [T-0024]
claims: [type:Violation, type:Monitors, type:Coverage, func:NewMonitors, func:ReplayCommand, field:SimOptions.NoMonitors, field:SimStats.Coverage, cli:sim-monitors, monitor:attack-declaration, monitor:entry-state, monitor:resources]
cards: []
requirements: [FR-CBT-001, FR-CAP-001, FR-KWD-002g, FR-KWD-002j, FR-SET-004, FR-ZON-005, FR-KWD-002l, FR-KWD-002k, FR-KWD-002i, FR-RES-002, FR-RES-003, KIM-S03]
---

## Goal
Every `cardtown sim` game is checked against specific FRD 01 rules, not only termination and `CheckInvariants`. A *rule monitor* is an `engine.Observer` in `pkg/engine/sim`. It watches every event and settled snapshot of a game and fails the run when a rule it names is broken, with a command that replays the game to that event. `cardtown sim` prints a coverage table, so it shows which rules the sim reaches and which it never reaches. This task builds the framework and three monitors that need nothing the events don't already carry. T-0026 (queued) adds the damage-sequence and target-timing monitors, together with the event facts they need.

## Inputs
- Board directive #23 (`.agents/tasks/m2/BOARD.md`).
- ADR-0005 (`docs/architecture/decisions/0005-structured-game-events-and-observer.md`): the observer contract. An observer gets copies only, and `Settled` comes after setup, after each resolved stack item and after each `Apply`. So the most recent settled snapshot before an `action` event is the state in which that action was declared.
- `pkg/engine/events.go`: `Observer`, `Snapshot`, `AssetSnap`. `AssetSnap.Card` is the real name even when `FaceDown`. Events and their fields:
  - `action` events: `Detail` begins with the action kind (`Attack`, `Generate`, `Convert`, …). Attack has `Source`, `Target` and `Secondary`; Generate has `Source` and `Resource`.
  - `deploy` (`Source` is the new asset, plus `Zone` and `Card`) and `scratch` (`Source` is the new asset).
  - `turn-start` (`Snapshot`) and `match-start` (`Start.Seed`).
- `pkg/engine/sim/sim.go`: `SimOptions`, `SimulateWith`, `SimStats`. `cmd/cardtown/main.go`: `runSim`.
- `cardtown replay -events <file> -game N -at <seq>` (T-0023). It is exact on the same build.
- The rules, in FRD 01:
  - `FR-CBT-001`: eligible attackers and defenders, Reserves protection, Ethereal, and Guardian priority.
  - `FR-CAP-001`: Attack and Generate.
  - `FR-KWD-002g`, `FR-KWD-002i`, `FR-KWD-002j`, `FR-KWD-002k`, `FR-KWD-002l`.
  - `FR-SET-004`: Immediate Readiness.
  - `FR-ZON-005`: a Scratch Company enters face-down, Exhausted, in Reserves, as a 1/1 with yields 1/1 and no abilities.
  - `FR-RES-002` and `FR-RES-003`.
  - `KIM-S03`: zero-yield generation is legal.
- Rulings that apply: none beyond those cited. Monitors check rules; they change no rule.

### Shapes
```go
// pkg/engine/sim/monitor.go

// Violation is the first broken rule a game's monitors found.
type Violation struct {
	Monitor string   // e.g. "attack-declaration"
	Rules   []string // every ID the monitor names
	Game    int      // counts from 1, as in -events
	Seed    uint64   // the game's seed, from its match-start
	Seq     int      // the event at which the rule broke
	Detail  string   // what was wrong, in words
}
func (v *Violation) Error() string // one line naming all six fields

// Coverage counts checks: monitor name -> rule ID -> number of checks.
type Coverage map[string]map[string]int

// Monitors is one game's set of rule monitors, as one engine.Observer.
// decks gives each seat's card characteristics, looked up by the asset's
// owner and card name.
func NewMonitors(decks [2]engine.DeckList, game int) *Monitors
func (m *Monitors) Event(e engine.Event)
func (m *Monitors) Settled(s engine.Snapshot)
func (m *Monitors) Violation() *Violation // nil while every rule held
func (m *Monitors) Coverage() Coverage    // every monitor and every rule ID it names, zero counts included

// ReplayCommand is the two commands, one per line, that reach v. simArgs are
// the sim flags the run used, minus -seed, -games, -log and -events. File
// is monitor-<seed>.jsonl:
//   cardtown sim <simArgs> -seed <Seed> -games 1 -events <file>
//   cardtown replay -events <file> -game 1 -at <Seq>
func ReplayCommand(simArgs []string, v *Violation) string

// pkg/engine/sim/sim.go
// SimOptions gains:
	NoMonitors bool // true runs without rule monitors; nothing else changes
// SimStats gains:
	Coverage Coverage // summed over games; nil when NoMonitors
```

### Lead decisions (technical)
- **What a monitor sees:** only `Event` and `Settled` calls, plus the decklists' card data. It never reads a `*Game`. It uses printed characteristics from the decklist cards. A face-down asset uses the Scratch Company face instead: a Unit, 1/1, yields 1/1, no keywords (`FR-ZON-005`). Printed keywords equal effective ones today because nothing grants or removes keywords. When the first ability or keyword grant lands, directive #13 switches the monitors to effective keywords carried in the snapshot.
- **Card lookup:** by (owner, card name) in that owner's decklist, HQ included. Test decks build their cards in code, so the two seats can hold different cards under one name, and name alone would be ambiguous.
- **"Before" and "after":** "before" an action is the most recent settled snapshot before its `action` event. "After" is the next settled snapshot. A check that is still waiting when the game ends without another settled snapshot is dropped and not counted.
- **Monitors and checks.** Each check adds 1 to the count of every rule ID it exercised. A violation reports every ID the monitor names.
  1. **`attack-declaration`** (`FR-CBT-001`, `FR-CAP-001`, `FR-KWD-002g`, `FR-KWD-002j`). On every Attack `action` event, using "before":
     - The phase is `main`, and the stack is empty.
     - The actor is the active player and controls the attacker.
     - The attacker is Readied. It is either a Unit in the Main Theater or the actor's HQ in the Headquarters Zone.
     - The defender is controlled by the opponent. It is either a Unit or Structure in the Main Theater, or the opponent's HQ in its Headquarters Zone.
     - The defender is not Ethereal. `FR-KWD-002g` is counted when the opponent has any Ethereal asset in the Main Theater or Headquarters Zone.
     - If the opponent controls a non-Ethereal Guardian asset in the Main Theater, the defender is one of those. `FR-KWD-002j` is counted only when such a Guardian exists.
     - `FR-CBT-001` and `FR-CAP-001` are counted on every attack.
     - Berserk's secondary target is not checked in this task.
  2. **`entry-state`** (`FR-SET-004`, `FR-ZON-005`, `FR-KWD-002l`, `FR-KWD-002k`). On every `deploy` and `scratch` event, using the next settled snapshot. If the asset is no longer on the battlefield by then, skip it and count nothing.
     - A `scratch` asset is in its controller's Reserves, face-down and Exhausted (`FR-ZON-005`).
     - A `deploy` asset is face-up, in the event's `Zone`, and Exhausted exactly when its card has Deploys Exhausted. That is `FR-KWD-002l` when the card has it, and `FR-SET-004` otherwise.
     - Its Armor equals its printed Armor N (`FR-KWD-002k`, counted only when N > 0). Skip the armor part when an event between the deploy event and that snapshot has the asset as `Target`.
  3. **`resources`** (`FR-CAP-001`, `FR-KWD-002i`, `FR-RES-002`, `FR-RES-003`, `KIM-S03`).
     - **Generate `action` event** ("before" and "after"):
       - The actor controls the asset, which is Readied before and Exhausted after.
       - The resource is allowed: a Unit in Reserves gives Requisition and in the Main Theater gives Tactics. A Versatile Unit (`FR-KWD-002i`, counted when it applies), a Structure or an HQ may give either.
       - The chosen pool rises by exactly the printed yield for that resource, and the other pool doesn't change. `KIM-S03` is counted when the yield is 0.
       - `FR-CAP-001` is counted on every Generate.
     - **Convert:** Requisition is at least 2 before. After, it is 2 lower and Tactics is 1 higher (`FR-RES-003`).
     - **Every `turn-start` snapshot:** both pools are 0 (`FR-RES-002`).
- **First violation wins.** A monitor records only the first violation of a game. Later events are still counted, so coverage stays complete.
- **`SimulateWith`** runs `NewMonitors` for every game unless `NoMonitors` is set. If `SimOptions.Observer` is also set, both receive every call, in the same order. After each game, a violation makes `SimulateWith` return that `*Violation` (an error, usable with `errors.As`) and play no more games. `Simulate` runs monitors too, because it calls `SimulateWith`.
- **Read-only:** monitors never change the game or touch any random source. So `SimStats` (apart from `Coverage`), `-events` and `-log` are byte-identical with and without them.
- **`cardtown sim`:**
  - `-monitors` (bool, default true). With `-monitors=false`, the output is exactly today's.
  - With monitors on, after the summary it prints a coverage table: a heading line, then one line per monitor and rule ID, sorted by monitor then ID, with the count, or `not reached` when the count is 0. The table is information, so `not reached` never changes the exit code.
  - On a violation, it prints `[ERROR] ` followed by `Violation.Error()`, then `ReplayCommand`'s two lines, and exits 1. The `simArgs` passed are the run's own flags, minus `-seed`, `-games`, `-log` and `-events`.
- **Where the contract lives:** add a "Rule monitors" paragraph under the design note's Boundaries section (`docs/architecture/m2-ability-model.md`). It covers what a monitor sees, the first-violation rule, coverage counting and the replay hint, and says that later shape tasks may add a monitor for their rule. Update the `pkg/engine/sim` row of the Boundaries table to list the monitors. ADR-0005 doesn't change: no event field is added in this task.
- **Not in this task:**
  - the damage-sequence monitor (`FR-REP-004`), the target-timing monitor (`FR-TGT-001`, `SVY-R3`) and the event facts they need (T-0026);
  - the APNAP monitor (`FR-TRG-003`) and the illegal-target monitor (`FR-TGT-003`), which starter cards don't reach yet;
  - Berserk secondary targets;
  - monitors in `cardtown play`.

## Acceptance Cases (harness tasks only; written by the lead)
Cases 1 to 5 go in a new `pkg/engine/sim/monitor_test.go`, plus the one authorized edit in case 2. All of it lands in the first `test(T-0025): …` commit, which must fail before the implementation. Cases 6 and 7 are CLI checks, run by the engineer and re-run by the reviewer. Record their output in the Log.
1. **Each monitor can fail.** Hand-built sequences of `Event` and `Settled` calls go to `NewMonitors(decks, 3)` with small test decks. Each of the following gives a `Violation` with the right `Monitor`, all of that monitor's `Rules`, `Game` 3, the seed from the hand-built match-start and the breaking event's `Seq`:
   - **attack-declaration:** an Exhausted attacker; an attacker in Reserves; a Structure attacker; a defender in Reserves; an Ethereal defender; a non-Guardian defender while the opponent has a Guardian in the Main Theater.
   - **entry-state:** a Readied Scratch Company; an Exhausted unit without Deploys Exhausted; a Readied unit with Deploys Exhausted; a unit with Armor 2 at Armor 0.
   - **resources:** Tactics from a non-Versatile Unit in Reserves; a pool rise different from the yield; a Convert that leaves Requisition unchanged; a turn-start with 1 Requisition.

   A valid sequence of each kind gives no violation, and its coverage counts are above 0 for the IDs it exercised.
2. **Clean on real games:**
   - `SimulateWith` on the starter decks, 500 games under `RandomPolicy` (no `Policy`) and 500 under `ProgressPolicy`, returns no error.
   - **Authorized edit to an existing test:** `TestRandomPlayoutsTerminateAndHoldInvariants` in `pkg/engine/sim/playout_test.go`. Its two decks move into a local variable, and its `engine.Config` gains `Observer: NewMonitors(decks, int(seed)+1)`. After `Playout`, it fails on `Violation() != nil`, printing the violation. Nothing else in that test changes: same decks, seeds, policies and game counts. That test's keyword deck reaches Guardian, Ethereal, Versatile, Armor and Deploys Exhausted, which starter games don't.
3. **Observing changes nothing:** on the starter decks, 20 games, one seed:
   - `SimStats` with `NoMonitors: true` equals the default run's `SimStats` with `Coverage` set to nil.
   - With collecting observers on both runs, the `WriteEventsJSONL` bytes are identical.
4. **Coverage is deterministic and complete:**
   - Two default runs (starter decks, 50 games, one seed) give equal `Coverage`.
   - It has an entry for every rule ID of every monitor, including zero counts.
   - `FR-CBT-001`, `FR-CAP-001` (both monitors), `FR-SET-004` and `FR-ZON-005` are above 0.
5. **Violation text:**
   - `Violation.Error()` contains the monitor, every rule ID, the game, the seed and the Seq.
   - `ReplayCommand([]string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-policy", "random"}, v)` returns exactly these two lines, for `Seed` 17 and `Seq` 240:
     ```
     cardtown sim -deck starter -p0 ANIMUS -p1 PATH -policy random -seed 17 -games 1 -events monitor-17.jsonl
     cardtown replay -events monitor-17.jsonl -game 1 -at 240
     ```
6. **CLI default:**
   - `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and bare `go run ./cmd/cardtown sim` both exit 0 and print the summary and then the coverage table.
   - Each is byte-identical across two runs.
   - With `-monitors=false`, each is byte-identical to `rework/executable-rules`' output for the same flags.
   - Removing the coverage table from the default output gives exactly the `-monitors=false` output.
7. **CLI `-events` and `-log`:**
   - `-games 20 -events <tmp>/a.jsonl`, run with monitors on and with `-monitors=false`, writes byte-identical files.
   - `-log -seed 7` prints the same event lines both ways.
   - Under `-policy random`, the starter run and the stress run exit 0 with monitors on.
8. `go tool task check` passes.

If a monitor finds a violation in case 2, 6 or 7, that may be a real engine bug. Don't fix the engine and don't loosen the monitor. Log the violation and its `ReplayCommand`, check that the replay reaches it, and stop at `blocked` (directive #23). If the cause is a wrong check in your monitor, fix the monitor and say so in the Log.

## Allowed Files
- `pkg/engine/sim/monitor.go` (new)
- `pkg/engine/sim/sim.go`
- `cmd/cardtown/main.go`
- `docs/architecture/m2-ability-model.md` (the Boundaries section only)
- `pkg/engine/sim/monitor_test.go` (new, the `test(T-0025)` commit)
- `pkg/engine/sim/playout_test.go` (only the case 2 edit, in the `test(T-0025)` commit)

## Done When
- [ ] Cases 1 to 5 have tests, committed before the implementation, and all pass
- [ ] Cases 6 and 7 hold, with their output and the starter coverage table in the Log
- [ ] `go tool task check` passes
- [ ] Existing tests pass, unchanged apart from the authorized edit

## Log
- 2026-10-21 lead-architect: Wrote the task from board directive #23 (part 1 of 2), with status `ready` (harness, sonnet). The lead chose the first monitors from what starter cards reach without new event facts: attack declaration, entry state and resources. The damage-sequence and target-timing monitors need additive event facts (the damage stages, and an event when an item is placed on the stack), so they go to T-0026, which keeps this task within four non-test files.
