# 0003. Synchronous Chooser for player choices

- **Status:** Accepted
- **Date:** 2026-10-19

## Context

The M1 kernel never asks a player anything after an action is declared: `Apply` resolves the stack to empty in one call. The M2 starter set needs choices made partway through declaration and resolution: APNAP trigger order (`FR-TRG-003`), targets and resolution-time targets (`FR-TGT-001`, `SVY-R3`, `SVY-R4`), "you may" triggers, modes, distributed pools (`FR-TGT-002`), sacrifice and discard costs (`FR-TRN-003a`), and replacement ordering (`SVY-R5`). The non-active player makes some of them (`FR-TRG-002`). Choices have a single-option fast path and timeout fallbacks (`FR-MOD-001`).

FRD 01 never lets a player act in response: manual actions need an empty stack (`FR-QUE-001`), and resolution only "pauses for interactive choices". So a choice is a question with an answer, never a window for new actions.

There are two ways to ask. A **suspendable state machine** stores a pending decision on `Game`, returns from `Apply`, and resumes when an answer action arrives. That makes every effect a resumable step program. A **synchronous chooser** is a callback the kernel calls and waits on, so effects stay ordinary Go code.

## Decision

Use a synchronous `Chooser`, one per seat in `Config`. The kernel builds an exported `Decision` (player, kind, options, min/max, whether it inspects hidden cards) and calls the seat's `Chooser`.

- The kernel applies the `FR-MOD-001` fast path itself, so a `Chooser` is never asked a question with one legal answer that inspects no hidden cards.
- A nil `Chooser`, a "timeout" answer or an invalid answer gets the `FR-MOD-001` §6 fallback. An invalid answer is also logged, and playouts fail on it.
- `LegalActions` offers a declaration only if at least one legal completion of its choices exists, so a `Chooser` is never trapped in a declaration it cannot finish.
- Existing Attack fields (`Target`, `Secondary`) stay on `Action`. New declarations take their choices through the `Chooser`.

Implementations: a scripted one for scenario tests, a random one for playouts and `cardtown sim`, and a terminal prompt for `cardtown play`.

## Consequences

- Effects stay straight-line Go code, which is what the M2 card loop needs from Sonnet- and Flash-tier engineers.
- A network server (M5) has to run each match in its own goroutine and block on the remote player's reply. Timeouts become the server's job, answered as "timeout".
- A match cannot be serialized mid-choice. Revisit this before M5's reconnect and persisted replays: if mid-choice snapshots are needed, move to the state-machine form, with the `Decision` type unchanged.
