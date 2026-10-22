# Product Management & Design Audits

Red-team audits of the Game Rules FRD and its keyword interaction matrix. Persona behavior is defined only in `.agents/skills/` (`pm-devils-advocate`, `game-designer`, `pm-frd-drafter`); this page is the human-facing overview.

## Scope

Only `docs/frd/01_Game_Rules_FRD.md` and `docs/frd/keyword_interaction_matrix.md` are audited. FRDs 02–05 are frozen as design intent and are not audited (see `docs/frd/README.md`). Behavior already pinned by a passing `pkg/engine` scenario test, and items already open in the matrix, are not re-raised.

## Workflow

```text
FRD 01 majorly revised
        │  (agent revisions trigger the DA automatically; applying a Design Outcome does not)
        ▼
pm-devils-advocate ──► docs/audit/pending/[date]-[doc].md (PENDING, committed locally)
        │                (zero findings is a normal, recorded result)
        ▼  (you ask to work through it)
game-designer ── per item: Rule | Prototype ruling | Dismiss (Design Rationale) | Defer (matrix) | Wording only
        │
        ▼  one Design Outcome Report for the whole audit
        ├──► board directive ──► lead-architect build loop writes the TestFR_* settling tests
        ▼
pm-frd-drafter ── FRD 01 text, callouts, matrix rows
        │
        ▼
cardtown audit resolve ──► lint, purge pending file, commit, mark RESOLVED in audit_registry.md
```

Nothing pushes automatically. Agents commit locally; push when you decide to.

## What Makes a Finding

Every finding names a concrete game state, the competing outcomes, whether today's cards can reach it (`R`/`U`), and the scenario test that would settle it. Findings without those are dropped. Keyword findings are written as proposals for a matrix cell.

## Working Through an Audit

Say `@game-designer` and name the pending audit. It takes every finding, redundancy row and Socratic question in severity order and asks you to decide each one; nothing is fixed without you. The result is one Design Outcome Report. You run the `@pm-frd-drafter` prompt it gives you, which applies the FRD 01 and matrix edits and closes the audit. Settling tests for new or changed rules reach the build loop as a board directive.

## Tooling
- `cardtown audit status`: read-only status of pending audits (commit-based `STALE`/`CURRENT`, severity counts, human-required flag).
- `cardtown audit resolve [-dry-run] [-push] <auditFile>`: atomic close-out. Lints first, rolls back on failure, pushes only with `-push`.
- `cardtown lint rules` and `go test ./pkg/engine`: what the kernel already pins.

## Layout
- [`audit_registry.md`](audit_registry.md): ledger of every audit (dates, commits, targets, resolution).
- `pending/`: active punch-lists as `[YYYY-MM-DD]-[document].md`.

## Staleness
An audit is stale when its target has commits after the audited commit recorded in the registry (Commit Hash column), or a newer audit exists for the same document. `cardtown audit status` computes this.

## Severity Taxonomy
Severity sets the order findings are taken in; every one is decided by you.
- 🔴 **Blocker / Critical:** reachable (`R`) undefined or contradictory outcome that changes who wins, or a deadlock or infinite loop.
- 🟠 **High:** reachable (`R`) undefined or ambiguous outcome that materially changes the game state.
- 🟡 **Medium:** unreachable today (`U`) but plausible with future cards, a minor reachable difference, or redundant requirements.
- 🔵 **Low / Clarification:** phrasing ambiguity that every reasonable reading resolves the same way.
