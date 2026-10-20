# 0004. Abilities as data values in Go, interpreted by the kernel

- **Status:** Accepted
- **Date:** 2026-10-19

## Context

M2 makes 75 Animus and Path cards executable with no compiler (`PRODUCT_SUMMARY.md` §3.2), and its write-up is the input for rewriting FRD 02/03 in M3. Card work is split into `shape` tasks, which add machinery, and `card` tasks, which only combine existing shapes and run on a cheaper model. Engine code may not contain card names (`cardtown lint fences`). `rules-test-author` writes card tests as package-`engine` test files, so they must be able to reach real card definitions without an import cycle.

Abilities could be Go closures per card (`func(*Resolution)`) or data values (a hook, targets and a list of primitive steps) that the kernel interprets.

## Decision

- Abilities are **data**. `pkg/engine/ability` defines the vocabulary (hooks, conditions, filters, costs, targets, steps, values, statics, replacements) as plain types with no game logic. `pkg/engine` interprets them.
- Card definitions are Go values, `map[string]ability.Def` keyed by card name, in `pkg/engine/cards`. Each `Def` lists the exact rules-text lines it implements. Loading fails if a listed line is missing from the CSV.
- Imports run `engine` → `cards` → `ability`. The fence covers `pkg/engine` and `pkg/engine/ability`, and exempts `pkg/engine/cards`.

## Consequences

- A card task cannot hide logic in a closure: anything new has to become a named shape, claimed by a shape task and reviewed once.
- The `ability` package is the M2 write-up: its types map directly onto a future JSON/AST format, so the M3 compiler targets the same structure.
- Some cards will need a shape used by only one card. That is accepted, and the cost is a slightly larger vocabulary.
- Card-name logic stays out of the interpreter by construction, and the fence lint enforces it.
- Revisit if a starter card cannot be expressed without a one-off primitive that has no plausible second user. Record that card as an exception in the write-up rather than weakening the rule.
