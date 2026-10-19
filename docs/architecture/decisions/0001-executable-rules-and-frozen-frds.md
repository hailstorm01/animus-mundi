# 0001. Executable rules kernel as the rules oracle; freeze FRDs 02–05

- **Status:** Accepted
- **Date:** 2026-10-19

## Context

By 2026-10-19 the project had five FRDs, each audited six or seven times by `pm-devils-advocate` and resolved by `pm-audit-reviewer` over eight days, plus a 1,693-line TDD for the card compiler, but no game code. The physical prototype already shows the core loop is fun, so the open risk is not whether the game is good. It is whether the digital rules are complete and consistent.

The audit loop never converged: most audits came back with one or two Blockers, two Highs, two Mediums and a Low, whatever the document's maturity. Some real defects came from documents restating each other's rules. The 2026-10-19 Blocker was Berserk behaving differently in FRD 01 and FRD 05. Meanwhile FRDs 03–05 had fixed technical choices (wire format, PRNG layout, hosting, performance budgets) before any code existed to test them.

The devil's-advocate audits were genuinely useful at one thing: finding "this situation is undefined" and "what happens if…" cases in the rules.

## Decision

1. **FRD 01 stays the rules source of truth**, and `pkg/engine` becomes its executable form. Scenario tests named after FR IDs (`TestFR_<DOM>_<NNN>_*`) pin each ruling. Random-policy playouts fuzz for non-termination and broken invariants. `cardtown lint rules` reports which active requirements have tests.
2. **Keyword edge cases go into a bounded register**, the [keyword interaction matrix](../../frd/keyword_interaction_matrix.md). It is finished when every cell has a ruling, unlike an open-ended audit.
3. **FRDs 02–05 are frozen as design intent.** They are no longer audited or synced. Each one is rewritten when its milestone starts, from what the code needed. TDD 03 is parked with FRD 03.
4. **Technical decisions are recorded here**, as one-page records written when the code that needs them is built.
5. **The roadmap is ordered by playable slices** (M1–M6 in `PRODUCT_SUMMARY.md`): rules kernel → full FRD 01 coverage → card data pipeline shaped by the kernel → local client → networked play → metagame.

## Consequences

- "What happens if…" questions get answered by a test that passes or fails, instead of another round of prose.
- The DA's useful role survives, scoped to FRD 01 and the matrix, with explicit permission to report that it found nothing.
- The card ability format (FRD 02/03) will be designed after the kernel has executed real abilities, so about 140 hand-authored AST entries won't be written against a format that may change.
- Risk: frozen FRDs drift from reality. That is accepted, because they are labelled as intent and will be rewritten. Revisit if a frozen document starts being treated as authoritative again.
