---
id: T-0168
title: The play log hides an opponent's deck search pick, and a search replays (FR-DEK-001, harness, flash)
kind: harness
tier: flash
wave: D
status: queued
branch: m2/autoloop-wave-d/T-0168-search-stays-private
depends_on: [T-0167]
claims: [view:LogLines search redaction, comment:doc.go OpSearchDeck]
cards: []
requirements: [FR-DEK-001]
---

## Goal
`FR-DEK-001` gives the searching player, and only that player, sight of the deck. `view.LogLines` prints a decision's picked card names to either seat, so `cardtown play` (`play.go:111`) would tell a human which card the opponent found. Close that before any card can search, and pin that a game with a search replays.

Covers (write the Acceptance Cases and Allowed Files against the code of the day, once T-0167 has merged):
- `pkg/view/log.go`: for a `DecideSearch` decision event (`e.Decision.Kind == engine.DecideSearch.String()`) whose player isn't `seat`, the line names no card and no index: a redacted line, through the existing `redact`, that says the player searched and found a card, or found none. The searching seat's own line is unchanged. The doc comment's "except two records" becomes three.
- `pkg/view/view_test.go`: both seats, with the opponent's line naming no card from the options.
- `pkg/eventlog/replay_test.go`: a synthetic deck with an `OpSearchDeck` ability, modeled on `activationDeck`, round-trips through `Replay`. The order of the options and the shuffle's draws from the game PRNG must both reproduce.
- `pkg/engine/doc.go`: the built-steps list names `OpSearchDeck` and `DecideSearch` (T-0167 was at its file limit).

## Log
- 2026-11-04 lead-architect (wave D run, session 12): split off T-0167 (D2.4). Queued.
