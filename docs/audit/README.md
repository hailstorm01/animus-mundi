# Product Management & Design Audits

Formal audit reports and red-team critiques for Card Town FRDs. Persona behavior is defined only in `.agents/skills/` (`pm-devils-advocate`, `pm-audit-reviewer`, `game-designer`); this page is the human-facing overview.

## Workflow

```text
FRD drafted/revised in docs/frd/
        │  (agent-drafted FRDs trigger the DA automatically; reviewer edits do not)
        ▼
pm-devils-advocate ──► docs/audit/pending/[date]-[doc].md (PENDING, committed locally)
        │
        ▼  (human asks for review)
pm-audit-reviewer ── interactive: [A] Accept [B] Modify [C] Dismiss [D] Defer [E] Design discussion ──► game-designer
        │
        ▼
cardtown audit resolve ──► lint, purge pending file, commit FRD, mark RESOLVED in audit_registry.md
```

Nothing pushes automatically. Agents commit locally; push when you decide to.

## Clearing Loop (human-started only)

Say `@pm-audit-reviewer --clear-backlog` to run reviewer -> DA -> reviewer passes that auto-fix findings below 🔴 Blocker / 🟠 High (and not rule-changing), up to 5 passes per FRD. It stops when only human-required items remain, on no-progress/oscillation, or when the DA returns `CONVERGED / PASS`, and then reports what needs you. It never starts on its own.

## Tooling
- `cardtown audit status`: read-only status of pending audits (commit-based `STALE`/`CURRENT`, severity counts, human-required flag).
- `cardtown audit resolve -AuditFile <file> [-DryRun] [-Push]`: atomic close-out. Lints first, rolls back on failure, pushes only with `-Push`.

## Layout
- [`audit_registry.md`](audit_registry.md): ledger of every audit (dates, commits, targets, resolution).
- `pending/`: active punch-lists as `[YYYY-MM-DD]-[document].md`.

## Staleness
An audit is stale when its target FRD has commits after the audited commit recorded in the registry (Commit Hash column), or a newer audit exists for the same document. `cardtown audit status` computes this.

## Severity Taxonomy
- 🔴 **Blocker / Critical:** unresolvable loops, deadlocks, catastrophic exploits. Human required.
- 🟠 **High:** major onboarding friction, significant ambiguity, clear abuse vectors. Human required.
- 🟡 **Medium:** over-engineering, subtle edge-case friction, collapsible overlap. Loop-eligible.
- 🔵 **Low / Clarification:** phrasing ambiguity, cosmetic overlap, Socratic questions (questions always need a human).
