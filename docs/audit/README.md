# Product Management & Design Audits

Red-team audits of the Game Rules FRD and its keyword interaction matrix. Persona behavior is defined only in `.agents/skills/` (`pm-devils-advocate`, `pm-audit-reviewer`, `game-designer`); this page is the human-facing overview.

## Scope

Only `docs/frd/01_Game_Rules_FRD.md` and `docs/frd/keyword_interaction_matrix.md` are audited. FRDs 02–05 are frozen as design intent and are not audited (see `docs/frd/README.md`). Behavior already pinned by a passing `pkg/engine` scenario test, and items already open in the matrix, are not re-raised.

## Workflow

```text
FRD 01 majorly revised
        │  (agent revisions trigger the DA automatically; reviewer edits do not)
        ▼
pm-devils-advocate ──► docs/audit/pending/[date]-[doc].md (PENDING, committed locally)
        │                (zero findings is a normal, recorded result)
        ▼  (human asks for review)
pm-audit-reviewer ── [A] Accept [B] Modify [C] Dismiss [D] Defer [E] Design discussion [F] Prototype ruling
        │                                                        │
        │                                                        └──► game-designer
        ▼
FRD 01 text + TestFR_* scenario test (or a matrix entry if not executable yet)
        │
        ▼
cardtown audit resolve ──► lint, purge pending file, commit, mark RESOLVED in audit_registry.md
```

Nothing pushes automatically. Agents commit locally; push when you decide to.

## What Makes a Finding

Every finding names a concrete game state, the competing outcomes, whether today's cards can reach it (`R`/`U`), and the scenario test that would settle it. Findings without those are dropped. Keyword findings are written as proposals for a matrix cell.

## Clearing Loop (human-started only)

Say `@pm-audit-reviewer --clear-backlog` to run reviewer -> DA -> reviewer passes on FRD 01 that auto-fix findings below 🔴 Blocker / 🟠 High (and not rule-changing), up to 3 passes. It stops when only human-required items remain, on no-progress/oscillation, or when the DA returns `CONVERGED / PASS`, and then reports what needs you. It never starts on its own.

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
- 🔴 **Blocker / Critical:** reachable (`R`) undefined or contradictory outcome that changes who wins, or a deadlock or infinite loop. Human required.
- 🟠 **High:** reachable (`R`) undefined or ambiguous outcome that materially changes the game state. Human required.
- 🟡 **Medium:** unreachable today (`U`) but plausible with future cards, a minor reachable difference, or redundant requirements. Loop-eligible.
- 🔵 **Low / Clarification:** phrasing ambiguity that every reasonable reading resolves the same way (Socratic questions always need a human).
