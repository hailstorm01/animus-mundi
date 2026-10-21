# 0005. Structured game events, one read-only observer, and turn-start snapshots

- **Status:** Accepted
- **Date:** 2026-10-20

## Context

The user wants match metrics later ("what did players do on their first turn", "which HQ did they bring"), a replay command (board directive #19), and the same data for M5 replays and M6 telemetry. Until T-0020 the kernel kept `Event{Turn, Kind, Detail}`: free text from scattered `logf` calls, read by tests and `Playout`. No player choice was recorded, and nothing described the match setup.

These consumers need a stable schema, and that schema is expensive to change once files exist. They also need a storage model. Events are small. A full-state snapshot is roughly 1–4 MB per game if taken per event. A run of the deterministic kernel (ADR-0002: one seeded stream) can rebuild any state from its inputs.

## Decision

**One emit path and one observer** (`pkg/engine/events.go`). Every event goes through `Game.emit`, which numbers it (`Seq` 1, 2, 3 … per game, no gaps), stamps `Turn` and `Phase`, appends it to `Game.Log` and passes a deep copy to the optional `Config.Observer`. `Game.Log` stays the default record, and `Kind`, `Detail` and `Turn` are unchanged for existing event kinds.

The observer has two methods:
- `Event(e Event)` gets every event, in order.
- `Settled(s Snapshot)` is called only when nothing is resolving and the death sweep and trigger flush are done: after setup, after each stack item, and at the end of each `Apply`.

It receives copies only and has no path to the game or its random source, so a game runs identically with or without one.

**Structured events.** `Event` adds the acting player (-1 if none), the card name (from card data), the instance ID, the source and target asset IDs, the zone and an amount. It may also carry one record:
- `decision`: every answer through `choose`, with the decision kind, the player, the options and refs offered, the picks, and `how` (answer, auto, timeout, invalid).
- `match-start`: the seed, deck names, HQs, the first-player decider and their choice, the cards mulliganed, and the build's VCS revision and dirty flag.
- `turn-start`: a full `Snapshot`, hidden information included.
- `match-end`: the outcome and the turn count.

Every accepted action emits one `action` event.

**Addendum (T-0026).** Damage events carry the mitigation facts as well as the committed `Amount`: `Raw` (before Stage 2), `Combat`, `Armor` (the target's pool before the damage) and `Absorbed` (Stage 4). Every item put on the Resolution Stack emits one `place` event right after it is pushed: `Player` is its controller, `Source`, `Target`, `Secondary`, `Card` and `Inst` describe the item, and `Amount` is the stack depth after the push. The `combat` event now also carries the attacker as `Source` and the defender as `Target`. All of these are additive and omitted when unused.

**Storage model.** Events are the permanent record. Full state is stored only in turn-start snapshots. Any other point is rebuilt by replaying the match-start record and the decision and action events on the same build (#19). The build revision makes "same build" checkable. Events have JSON tags, and T-0021 writes them as JSON Lines, one event per line, `match-start` first.

## Consequences

- Metrics, replay and telemetry read one schema. Adding a field is additive. Renaming or removing one supersedes this record.
- `Log` now holds turn-start snapshots, so a long simulated game holds one full state per turn in memory. Revisit this if `sim` memory becomes a problem: the snapshot can move to observer-only, with `Log` keeping a marker.
- Snapshots and event logs contain hidden information (hands, deck order, face-down identities). They are omniscient records for testing and analysis. Anything shown to a player (M5) must be redacted per seat first.
- Placing any item on the stack, a triggered ability included, now emits a `place` event. Resolving a printed triggered ability still emits no event of its own: only its effects are logged. Add an `ability` event (source, card, chosen targets) when the metrics work starts.
- A replay only reproduces a game exactly on the same build. A rules change makes old records non-replayable, but they stay readable.
