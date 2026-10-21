---
id: T-0020
title: Structured game events, one observer, settled notifications and match records
kind: harness
tier: sonnet
wave: none (board directive #18, part 1 of 2: engine)
status: review
branch: m2/T-0020-game-events-observer
depends_on: []
claims: [type:Event-fields, type:Observer, type:Snapshot, type:MatchStart, type:MatchEnd, type:DecisionRecord, func:Game.emit, config:Observer, config:DeckNames, event:action, event:decision, event:settled]
cards: []
requirements: []
---

## Goal
The kernel reports everything that happens through one read-only observer. Every event is structured and numbered, every `Chooser` answer is recorded, and each game opens with a match-start record, carries a full-state snapshot at every turn start, and closes with a match-end record. A game runs identically with or without an observer. This is the schema that metrics (later), replay (directive #19), M5 replays and M6 telemetry build on. T-0021 adds the CLI (`sim -log`, `sim -events`) on top of it.

## Inputs
- Board directive #18 (`.agents/tasks/m2/BOARD.md`), whose text is the source. The user chose a two-task split on 2026-10-20, so this task is (a), (b) and (c), and T-0021 is (d).
- Today: `Event{Turn, Kind, Detail}` and `Game.logf` in `state.go`, with call sites in `actions.go`, `choice.go`, `damage.go`, `stack.go` and `turn.go`. `Playout` (`playout.go`) scans `g.Log` for `invalid-answer`, and tests match on `Kind`.
- The storage model is from the directive: events stay small, and full state is stored only at turn starts.
- Rulings that apply: none. This task changes no rule.

### Shapes (`pkg/engine/events.go`, new, unless noted)
```go
// Event (state.go) keeps Turn, Kind and Detail unchanged and gains:
	Seq    int      `json:"seq"`    // 1, 2, 3 … per game, no gaps
	Phase  string   `json:"phase"`  // Phase.String(): "setup", "ready", "draw", "main", "end", "over"
	Player int      `json:"player"` // the acting player, -1 if none
	Card   string   `json:"card,omitempty"`   // name from card data (Asset.Printed().Name or Instance.Card.Name), never a literal
	Inst   int      `json:"inst,omitempty"`   // card instance ID
	Source int      `json:"source,omitempty"` // asset ID
	Target int      `json:"target,omitempty"` // asset ID
	Zone   string   `json:"zone,omitempty"`   // Zone.String()
	Amount int      `json:"amount,omitempty"`
	Decision *DecisionRecord `json:"decision,omitempty"`
	Start    *MatchStart     `json:"start,omitempty"`    // Kind "match-start"
	Snapshot *Snapshot       `json:"snapshot,omitempty"` // Kind "turn-start"
	End      *MatchEnd       `json:"end,omitempty"`      // Kind "match-end"
// Turn, Kind and Detail get the json tags "turn", "kind" and "detail".

// Observer receives a game's events. It must not change the game: it gets
// copies, never the *Game, and has no access to the random source.
type Observer interface {
	Event(e Event)        // every event, in Seq order, as it happens
	Settled(s Snapshot)   // after each settled step (see Lead decisions)
}

type DecisionRecord struct {
	Kind    string   `json:"kind"`    // DecisionKind name
	Player  int      `json:"player"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"` // the option labels offered, in order
	Refs    []int    `json:"refs"`    // the option refs, in order
	Picks   []int    `json:"picks"`   // indexes into Options, as applied
	How     string   `json:"how"`     // "answer", "auto" (fast path), "timeout" (fallback), "invalid" (fallback after an invalid answer)
}

type MatchStart struct {
	Seed      uint64      `json:"seed"`
	Decks     [2]string   `json:"decks"`     // Config.DeckNames; "" if unset
	HQs       [2]string   `json:"hqs"`       // HQ card names
	Decider   int         `json:"decider"`
	PlayFirst bool        `json:"playFirst"` // the decider's choice
	Mulligans [2][]string `json:"mulligans"` // names of the cards each player set aside
	Revision  string      `json:"revision"`  // runtime/debug.ReadBuildInfo vcs.revision, "" if absent
	Modified  bool        `json:"modified"`  // vcs.modified
}

type MatchEnd struct {
	Draw   bool   `json:"draw"`
	Winner int    `json:"winner"` // -1 on a draw
	Reason string `json:"reason"`
	Turns  int    `json:"turns"`
}

// Snapshot is a deep copy of the full state, hidden information included:
// both hands, both deck orders and face-down identities.
type Snapshot struct { /* turn, phase, active, first player, outcome, stack length,
	   per player: HQ, Reserves, Deck, Hand, Discard, Req, Tac;
	   Main Theater. Assets record ID, instance ID, card name, owner, controller,
	   zone, exhausted, face-down, damage, armor, temp power, and effective power and maximum health (Layer 6). Cards record
	   instance ID and card name. Fields have json tags. */ }

// Config (turn.go) gains:
	Observer  Observer  // optional; nil means only Game.Log
	DeckNames [2]string // optional; recorded in MatchStart
```
The engineer chooses the exact `Snapshot` field names and records them in the Log. Every other name above is fixed.

### Lead decisions (technical)
- **One emit path:** `g.emit(e Event)` assigns `Seq`, `Turn` and `Phase`, appends to `g.Log`, then calls the observer if one is set. `logf` stays as a thin wrapper around `emit` for the detail text. `Game.Log` is the default observer, so `Log` keeps every event, and existing `Kind`, `Detail` and `Turn` values don't change.
- **Structured fields at existing call sites:** each current `logf` call also fills the fields that apply:
  - `generate`: Player, Source, Amount (the yield).
  - `attack`: Player, Source, Target.
  - `deploy`: Player, Card, Inst, Source (the new asset), Zone.
  - `march`: Player, Source, Zone (the destination).
  - `scratch`: Player, Source.
  - `damage`: Source (0 if nil), Target, Amount (health damage committed).
  - `destroy`: Target, Card.
  - `trigger` and `fizzle`: Source.
  - `draw`: Player.
  - `result`: none.
- **Action events:** `Apply` emits one `action` event per accepted action, before its effects. Its Detail is the `ActionKind` name, and it carries Player, Card or Source, Target and Zone. Resource kinds go in Detail. Every `ActionKind` therefore appears.
- **Decision events:** `choose` emits one `decision` event per call that returns picks. It covers fast-path picks (`How: "auto"`), real answers, timeouts and invalid answers. A Decision with no satisfiable options emits nothing. The existing `invalid-answer` event stays as well, because `Playout` relies on it.
- **Settled:** `Settled(snapshot)` is called only when no item is resolving and the death sweep and trigger flush are done:
  1. after setup, at the end of `NewGame`;
  2. after each stack item's resolution, sweep and flush in `resolveStack`;
  3. at the end of each `Apply`, `ActEndTurn` included.
  Never call it mid-resolution, so an answer given inside a resolving item gets a decision event but no settled call. Build the snapshot only when an observer is set.
- **Records:**
  - `match-start` is the first event of every game, emitted at the end of setup, after the mulligan. Seq 1 is the match-start.
  - `turn-start` is emitted at the start of `beginTurn`, before the Ready phase, carrying a `Snapshot`. There is exactly one per turn.
  - `match-end` is emitted once, when `Outcome` is first set, after the `result` event.
- **Read-only and deterministic:** the observer gets value copies only. Nothing it does can reach `g.rng` or the game. The same seed gives the same `Log`, with or without an observer.
- **Not in this task:** any CLI flag or file output, sorted end reasons (#12), replay (#19), and metrics. All of those are T-0021 or later.

## Acceptance Cases (harness tasks only; written by the lead)
Each case becomes a test in a new `pkg/engine/events_test.go`, committed first as `test(T-0020): …` and failing before the implementation commit.
1. **Seq:** in a full playout of the starter decks (`LoadDeck` on `data/decks/*-starter.txt`, random policies and choosers, a fixed seed), `Log[i].Seq == i+1` for every event.
2. **Records:** in that playout, `Log[0]` is `match-start` with `Start` set: the seed, both HQ names, a decider of 0 or 1, and a `Mulligans` entry per player. The last event is `match-end`, and its `End` matches `g.Outcome` and `g.Turn`. There is exactly one `turn-start` per turn number from 1 to `g.Turn`, each with `Snapshot.Turn` equal to its `Turn`.
3. **Snapshot is complete:** in a turn-start snapshot, each player's deck lists the same instance IDs in the same order as the live deck at that moment, both hands are present, and face-down assets carry their real card names.
4. **Every action kind:** across a few seeds of starter-deck playouts, `action` events cover every `ActionKind` (Generate, Convert, Deploy, March, Attack, Scratch, EndTurn). Pick seeds that do, and assert all of them.
5. **Decisions:** a scenario with a scripted chooser answering a two-option `DecideTarget` (a T-0017 targeted trigger) emits a `decision` event with `How: "answer"`, the offered refs and the pick. A single-option target emits `How: "auto"`, and a timeout emits `How: "timeout"`.
6. **Structured fields:** a scenario deploy emits a `deploy` event with Player, Card (the card's name), Inst, Source and Zone set. A targeted damage step emits a `damage` event with Source, Target and Amount.
7. **Settled is settled:** an observer that collects snapshots during a full starter playout never sees a battlefield asset whose damage is at or above its effective maximum health, and is called at least once per action.
8. **Read-only and deterministic:** for several seeds, a starter playout with a recording observer and one without produce equal `Outcome`, `Turns` and `Log` (deep-equal). Mutating a received `Snapshot` (for example, emptying a hand slice) leaves the game unchanged.
9. **Back-compat:** every existing test passes unchanged, and the `Kind`, `Detail` and `Turn` of existing event kinds are as before.

## Allowed Files
- `pkg/engine/events.go` (new)
- `pkg/engine/state.go`, `choice.go`, `actions.go`, `stack.go`, `damage.go` and `turn.go`, for emit call sites and `Config`
- `pkg/engine/events_test.go` (new, the `test(T-0020)` commit)
- Size exception: 7 non-test files against the limit of 4. The user chose this two-task split on 2026-10-20, so the schema stays in one task.

## Done When
- [ ] Every acceptance case has a test in `events_test.go`, committed before the implementation, and all pass
- [ ] `go tool task check` passes
- [ ] `go run ./cmd/cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `go run ./cmd/cardtown sim` exit 0, and their summaries are byte-identical to `rework/executable-rules` (sort the end-reason lines; #12 lands in T-0021)
- [ ] At review, the lead writes ADR-0005 (event schema, observer contract, storage model) from the merged code

## Log
- 2026-10-20 lead-architect: Wrote the task from board directive #18 (part 1 of 2, after the user chose the two-task split), with status `ready` (harness, sonnet). T-0021 (CLI) follows when this is `done`.
- 2026-10-20 kernel-engineer: Implemented on `m2/T-0020-game-events-observer`. Tests first in commit 8157983 (they did not compile before the schema existed, which is the failing state). 9 acceptance cases are `TestEvents_Case1..9*` in `events_test.go`; cases 4 and 5 have several tests.
  - **Changed:** new `events.go` (Observer, DecisionRecord, MatchStart, MatchEnd, Snapshot types, `emit`, `logEvent`, `logf` wrapper, `settled`, snapshot/clone, `Phase.String`, `DecisionKind.String`, emit helpers). `state.go`: tagged Event fields, `Game.obs`, old `logf` removed. `turn.go`: `Config.Observer`/`DeckNames`, mulligan names, match-start, turn-start at the top of `beginTurn`, Settled after setup, draw fields. `actions.go`: one `action` event per accepted Apply, generate/attack fields, Settled at end of Apply (EndTurn included). `choice.go`: one `decision` event per choose that returns picks (auto/answer/timeout/invalid; `invalid-answer` kept). `stack.go`: settled after each item's sweep and flush, fields on trigger/deploy/march/scratch/fizzle. `damage.go`: damage/destroy fields, `match-end` after `result`.
  - **Snapshot fields:** `Snapshot{Turn, Phase, Active, First, Over, Draw, Winner (-1 until won), Reason, StackLen, Players [2]PlayerSnap, Theater []AssetSnap}`; `PlayerSnap{HQ *AssetSnap, Reserves []AssetSnap, Deck, Hand, Discard []CardSnap, Req, Tac}`; `CardSnap{Inst, Card}`; `AssetSnap{ID, Inst, Card (real name, also when face-down), Owner, Controller, Zone, Exhausted, FaceDown, Damage, Armor, TempPower, Power, MaxHealth}`. json tags are lowerCamel.
  - **Choices for the reviewer:** (1) action Detail is the ActionKind name; Generate is "Generate Requisition"/"Generate Tactics" (resource in Detail), so match on the first word. (2) `DecisionKind.String()` is "Target", "Mode", "Optional", "Order". (3) The turn-start snapshot is always built, because it is part of Log and Case 8 needs Log equal with or without an observer; only the Settled snapshot is built on demand. (4) The observer gets a deep clone of each event (Decision, Start, Snapshot, End), so editing it cannot change Log. (5) match-start has Turn 0 and phase "setup"; turn-start is emitted after Phase is set to ready, so its phase is "ready". (6) `trigger`, `fizzle`, `damage`, `destroy` have Player -1; `draw` has the drawing player.
  - **Reuse:** no existing event or snapshot code; `handCard` reused for the action event's card, `Power`/`MaxHealth` for effective values.
  - **Gates:** `go tool task check` green (fmt, vet, all tests, lint). `cardtown sim -deck starter -p0 ANIMUS -p1 PATH` and `cardtown sim` exit 0; summaries are identical to the pre-change output after sorting lines (win rates 51.0/17.0/32.0 and 66.0/30.0/4.0, average length 27.2 and 30.2 turns, end-reason counts 136/64 and 192/8). Existing tests untouched.
