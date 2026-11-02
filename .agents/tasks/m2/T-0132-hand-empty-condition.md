---
id: T-0132
title: A step can run only if its controller has no cards in hand (wave C shape, sonnet)
kind: shape
tier: sonnet
wave: C
status: queued
branch: m2/autoloop-wave-c/T-0132-hand-empty-condition
depends_on: [T-0131]
claims: [condition:hand-empty]
cards: []
requirements: [FR-MNV-001]
---

## Goal
A step can print *"Then, if you have no cards in hand, ..."* (*Steel Rain*'s last clause: *"Then, if you have no cards in hand, draw 2 cards."*). The step runs only if the ability's controller has no cards in hand when it resolves, after the steps before it. Split from *Steel Rain* with T-0131 under the D2 size rule; the per-discard target choices and the card stay on directive #20 (b).

Covers:
- the condition, read as the step resolves (not on placement), on the ability's controller's hand;
- tests on a synthetic Maneuver: the step runs with an empty hand and is skipped with one card left; an earlier step that empties the hand (a discard) makes it run;
- no card uses it yet, so `go tool task sim-compare -- autoloop/m2-wave-c` passes.

Inputs, Shape, Tests, Allowed Files and Done When are written against the code of the day before it leaves `queued` (D2.4).

## Log
- 2026-11-02 lead-architect (wave C session 19): queued with T-0131's task commit (D2.4).
