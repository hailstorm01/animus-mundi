---
id: T-0036
title: PlayerView, a per-seat view of the game and its event log (directive #24)
kind: harness
tier: opus (effort high, board directive #27)
wave: none (design note Deliverable 5, part 1 of 2; board directive #24)
status: in-progress
branch: m2/T-0036-player-view
depends_on: [T-0035]
claims: [pkg:pkg/engine/view, type:View, func:PlayerView, func:LogLines]
cards: []
requirements: [FR-ZON-005]
---

## Goal
`cardtown play` (T-0037) can draw a human's screen without seeing anything that player may not see. A new package, `pkg/engine/view`, builds one seat's view of a game: its own hand, both boards, public zones and counts. It also turns the event log into the lines that seat may read.

This is the cheap first draft of the M4/M5 client boundary that board directive #24 asks for. Package `engine` doesn't change.

## Inputs
- Board directive #24 (`.agents/tasks/m2/BOARD.md`): `PlayerView(g, p)`, a redacted event log, and the leak test (b). The lead picks the package.
- FRD 01 `FR-ZON-005`: a Scratch Company's face is unknown to *both* players while it is on the battlefield. It is revealed when it enters the Discard pile.
- **What the play code can read today:**
  - `Game`'s exported fields: `Players`, `Theater`, `Stack`, `Active`, `FirstPlayer`, `Turn`, `Phase`, `Outcome` and `Log`.
  - `Player`'s fields: `HQ`, `Deck`, `Hand`, `Discard`, `Reserves`, `Req` and `Tac`.
  - `Asset`'s fields, including `Inst.Card` even when face-down.
  - `StackItem`'s fields.
  - The layer functions `Power`, `MaxHealth` and `Keywords`.
- **Events and leaks:**
  - `Asset.String()` prints a face-down asset as "face-down Scratch Company#ID", so event `Detail` text already hides face-down identities.
  - A `destroy` event names the card, which is correct: the card is revealed on entering the Discard pile.
  - Ordinary draws emit no event.
  - The leaks are in the records:
    - `match-start` carries the seed, which reveals every shuffle, and the names of the cards each player mulliganed;
    - `turn-start` carries the full omniscient `Snapshot`.
  - `eventlog.FormatEvent` prints the seed in its `match-start` line.
- The design note's Open Question 4, "Hidden information", says `cardtown play` shows the human only their own zones by convention. Directive #24 says to replace that line with a pointer to the view when the view lands.
- Rulings that apply: `FR-ZON-005`, double-blind Scratch Companies.

### Shapes (fields are the engineer's choice where unnamed; the lead fixes what is visible)
```go
// pkg/engine/view: imports engine (and eventlog) only; engine never imports it.

// View is what one seat may see. It is a deep copy: changing it changes no game.
type View struct {
	Seat, Active, First   engine.PlayerID
	Turn                  int
	Phase                 string
	Over, Draw            bool
	Winner                int    // -1 until someone has won
	Reason                string
	You, Opponent         Side
	Theater               []AssetView // the shared Main Theater, in battlefield order
	Stack                 []StackView // bottom first
}

type Side struct {
	HQ        *AssetView
	Reserves  []AssetView
	Hand      []CardView // You only; empty for Opponent
	HandCount int        // both
	DeckCount int        // both; never the order or contents
	Discard   []CardView // both, top last; public
	Req, Tac  int
}

type CardView struct {
	Inst int    // only for the seat's own hand cards (an Action names them); 0 elsewhere
	Card string
	// plus printed characteristics the screen needs (cost, type, power, health)
}

type AssetView struct {
	ID         int    // asset ID, as Actions and Decisions use
	Card       string // "Scratch Company" for every face-down asset, for both seats
	Controller, Owner int
	Zone       string
	Exhausted, FaceDown bool
	Damage, Armor, Power, MaxHealth int // effective (layers); a face-down asset shows the Scratch face's values
	// keywords: effective, none for face-down
}

type StackView struct { /* kind, controller, source asset ID, the card for a Deploy or ability item, an attack's targets */ }

func PlayerView(g *engine.Game, seat engine.PlayerID) View

// LogLines formats events for seat, one line per event, in order, hiding what
// seat may not see. It is the play screen's event log.
func LogLines(events []engine.Event, seat engine.PlayerID) []string
```

### Lead decisions (technical)
- **Package.** The view lives in `pkg/engine/view`, which imports only `engine` and `eventlog`. Package `engine` and every existing file stay unchanged. Like `sim` and `eventlog`, it is a client of the engine's exported API. That is the M4/M5 boundary in miniature.
- **Visibility:**
  - **Own hand:** card names, printed characteristics and instance IDs.
  - **Opponent's hand:** a count only.
  - **Decks:** counts only, for both seats.
  - **Public:** both discard piles, both HQs, both Reserves, the Main Theater, both resource pools, the stack, the turn, the phase, the active player and the outcome.
- **Face-down assets** show as `"Scratch Company"`, with the Scratch face's power, health and no keywords, to **both** seats, the owner included (`FR-ZON-005`, double-blind).
  - **No card instance ID appears for any asset**, nor for any discard card. An instance ID can identify a card someone saw earlier, for example a card put on the bottom of the deck and later scratched face-down.
  - Instance IDs appear only on the seat's own hand cards, which an Action needs.
- **`LogLines`:**
  - Every event uses `eventlog.FormatEvent`, except two records:
    - `match-start` shows the decks, the HQs, the decider and their choice, and the seat's own mulliganed card names. It does **not** show the seed, and shows only a count for the opponent's mulligan.
    - `turn-start` shows only the turn and the active player, never snapshot contents.
  - No `Detail` text needs changing: it already hides face-down identities.
  - Ordinary draws emit no event, so there is nothing to redact. Record that as a gap for M4/M5, together with the lack of "opponent draws a card" lines.
- **Shortcuts allowed** (directive #24: don't refactor engine internals for the view). Building the view from `Game`'s exported fields and the layer functions is fine. Record anything awkward as an M4/M5 gap in the Log.
- **Docs:**
  - The design note's Boundaries table gains a `pkg/engine/view` row.
  - Open Question 4's "Hidden information" line is replaced with a pointer to `view.PlayerView` and `view.LogLines`, and a note that the boundary is a first draft for M4/M5.
- **Not in this task:**
  - the play loop and terminal chooser (T-0037);
  - per-view handles in place of engine IDs;
  - wire formats;
  - draw events;
  - leak-checking rule monitors (all M4/M5 per directive #24).

## Acceptance Cases (harness tasks only; written by the lead)
All go in a new `pkg/engine/view/view_test.go`, committed first as `test(T-0036): …` and failing before the implementation.
1. **Contents, on scenes:**
   - Own hand: names and instance IDs.
   - Opponent's hand: a count, and no names anywhere in the `View`.
   - Decks: both shown as counts.
   - Discard piles: both shown with names.
   - HQs, Reserves and Main Theater: shown with damage, armor, effective power and exhaustion.
   - Resource pools and the stack: shown.
   - Turn, phase, active player and outcome: shown.
2. **Face-down:** a face-down asset shows `"Scratch Company"` and the Scratch face's stats in **both** seats' views, its owner's included. No instance ID appears for it. Once destroyed, it appears in its owner's discard under its real name.
3. **Deep copy:** changing every slice and pointer in a `View` leaves `g` unchanged.
4. **No leak (directive #24 b):**
   - Play 20 seeded starter matches (ANIMUS vs PATH), each to completion, using `sim.Playout` with `ProgressPolicy` and an observer collecting events.
   - Take `PlayerView` for each seat at every settled point (the observer's `Settled` calls; build the view from the game at that moment).
   - The view, marshaled to JSON, and the seat's `LogLines` so far never contain, for the *other* seat:
     - the name of a card in that seat's hand or deck, or
     - the real name of a face-down asset,
     unless that name is publicly visible in the view at that moment (in a discard, face-up on the battlefield, on the stack, or an HQ).
   - Check both seats.
5. **`LogLines`:**
   - A `match-start` line for each seat lacks the seed, and contains that seat's own mulliganed names and the opponent's mulligan count.
   - A `turn-start` line has only the turn and the active player.
   - Every other event's line equals `eventlog.FormatEvent`'s line.
6. **Engine untouched:** no file outside `pkg/engine/view/` and the design note changes. `cardtown sim` output is byte-identical to `rework/executable-rules`.
7. `go tool task check` passes, and `cardtown lint fences` accepts the new package (no card names).

## Allowed Files
- `pkg/engine/view/view.go` (new)
- `pkg/engine/view/log.go` (new; or keep it in `view.go`)
- `docs/architecture/m2-ability-model.md` (the Boundaries row and Open Question 4)
- `pkg/engine/view/view_test.go` (new, the `test(T-0036)` commit)

## Done When
- [ ] Cases 1 to 5 have tests, committed before the implementation, and pass
- [ ] Cases 6 and 7 hold, with output in the Log
- [ ] M4/M5 gaps recorded in the Log

## Log
- 2026-10-21 lead-architect: Wrote the task (Deliverable 5 part 1, board directive #24), with status `ready`.
  - Harness task, tier opus at high effort (board directive #27, the user's choice).
  - Package `pkg/engine/view` is the lead's pick under directive #24.
