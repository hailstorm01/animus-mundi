# Technical Design Documents (TDDs) & Deliverable Slicing

> [!IMPORTANT]
> **Status: FROZEN (2026-11-03).** The TDD path is retired. This page records how it worked; it is not an active workflow. No persona writes or slices TDDs, Runbooks B and C of `lead-architect` no longer exist, and the `cardtown lint tdd` and `cardtown tdd slice` commands below are being removed. Plan a milestone with a design note and decision records in [`docs/architecture/`](../architecture/) instead. [TDD 03](03_Card_Compiler_TDD.md) is kept as frozen design intent.

Authoritative, zero-guesswork Technical Design Documents (`docs/tdd/<XX>_<Name>_TDD.md`) that translate Card Town Functional Requirements Documents (`docs/frd/<XX>_<Name>_FRD.md` — **what** the system does) into concrete engineering blueprints (**how** to build it) and dependency-ordered GitHub Deliverable Issues (`SLICE-XX-YY`). Persona behavior is defined in `.agents/skills/lead-architect/SKILL.md`; this page is the human-facing overview.

## Workflow

```text
Authoritative FRD in docs/frd/<XX>_<Name>_FRD.md
        │
        ▼
lead-architect (Runbook B) ──► docs/tdd/<XX>_<Name>_TDD.md
        │
        ▼
cardtown lint tdd ──► verifies 0 dangling FR-XXX-YYY refs, 100% FRD coverage, valid SLICE DAG
        │
        ▼  (Human reviews TDD & confirms GitHub issue slicing)
lead-architect (Runbook C) ──► cardtown tdd slice -DryRun (preview) ──► -Publish (live)
        │
        ▼
GitHub Milestone + Topological Issues ([SLICE-XX-01], [SLICE-XX-02], ...) with 'Depends on #N'
```

Nothing pushes or creates GitHub Issues automatically. Agents commit TDDs locally and always run `-DryRun` before asking to publish live issues.

## Zero-Guesswork Engineering Standard

Every TDD in `docs/tdd/` must provide:
1. **Repository File Tree:** Exact paths of every source, schema, and test file to create or modify.
2. **Concrete Data Structures & Schemas:** Copy-pasteable Go structs/interfaces, PowerShell types, Protobuf messages, JSON schemas, or GDScript classes—zero hand-waving prose.
3. **Module & Algorithmic Specifications:** Exact function/method/cmdlet signatures, deterministic step-by-step execution logic, transaction/rollback boundaries, and `ERR_*` error dispatch.
4. **Verification Matrix & Test Fixtures:** Tabular test cases (`TST-XX-NNN`) mapped to `FR-XXX-YYY` and `ERR_*` codes.
5. **100% Requirement Traceability Matrix:** Every Active and Subsumed `FR-XXX-YYY` in the source FRD accounted for.
6. **Deliverable & Implementation Slice Manifest (`### [SLICE-XX-YY]`):** Self-contained, topologically ordered work packages with explicit `Depends On` links, target files, signatures, acceptance checkboxes, and CLI verification commands.

## Tooling
- `cardtown lint tdd [-Files <path>]`: validates header metadata, 0 dangling `FR-XXX-YYY` references, 100% source FRD coverage, and acyclic `SLICE-XX-YY` dependency graphs.
- `cardtown tdd slice -TddFile <path> [-DryRun] [-Publish] [-UpdateExisting]`: slices a lint-clean TDD into a GitHub Milestone, Labels, and topologically ordered Issues via `gh`, resolving upstream slice dependencies to real GitHub Issue numbers (`Depends on #N`).
