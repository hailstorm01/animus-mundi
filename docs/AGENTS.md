# Card Town Engineering & Design Guidelines

This directory houses the authoritative master specifications and card balance data for Card Town.

## Master Documentation Architecture
- **Master Game Rules Spec:** `frd/01_Game_Rules_FRD.md` (pure game rules, card & asset taxonomy, zone topology, turn phases, 6-layer rules, 5-stage damage mitigation, combat, keyword dictionary).
- **Master Card Data Pipeline Spec:** `frd/02_Card_Data_Pipeline_FRD.md` (ingestion lifecycle, blueprint JSON schema, declarative ability AST, event hooks, 25 atomic primitives, MTG-aligned text standards, review register).
- **Master Card Compiler Spec:** `frd/03_Card_Compiler_FRD.md` (CLI compiler, parsing grammar, AST synthesis, parity verification, fingerprinting).
- **Master Networking & Protocol Spec:** `frd/04_Client_Engine_Protocol_Networking_FRD.md` (wire contract, Protobuf schemas, WebSockets, Fog of War, choice prompts).
- **Master Headless Game Engine Spec:** `frd/05_Headless_Game_Engine_FRD.md` (Go headless kernel, 3-tier runtime memory structs, seeded PRNG, transactional state, stack runner, AST evaluator, stateless 6-layer evaluation, test harness).
- **Card Data Sheets:** CSVs in `../data/cards/` across Concordia, Soulless, Path, Animus, HQ, and Common.

> [!TIP]
> The FRDs are large (40-70 KB each). Read the section you need by line range rather than whole files.

## Linting
Run `tools/linter/lint_cardtown.ps1` (auto-detects changed files; `-All` forces everything). Details live in the `cardtown-linter` skill. The pre-commit hook lints staged FRDs and CSVs automatically.

## Product Management, FRD Authoring & Audits
Product vision, portfolio roadmap, and `PRODUCT_SUMMARY.md` maintenance are owned by `pm-lead`. FRD authoring from product vision and roadmap documentation is driven by `pm-frd-drafter`. Audit workflow, severity taxonomy, and registry: see `audit/README.md`. Persona behavior lives only in `.agents/skills/` (`pm-lead`, `pm-frd-drafter`, `pm-devils-advocate`, `pm-audit-reviewer`, `game-designer`).
