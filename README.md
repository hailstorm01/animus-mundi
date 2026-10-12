# Animus Mundi

A strategy card game engine and client built with Godot.

## Repository Structure

- `docs/` - System documentation, specifications, and architecture.
  - `docs/frd/` - Functional Requirements Documents (FRDs) tracking game rules, networking, and card mechanics.
  - `docs/architecture/` - High-level engine and state architecture.
  - `docs/design/` - Gameplay design, lore, and aesthetic guides.
- `data/` - Authoring datasets and balance sheets.
  - `data/cards/` - Faction balance sheets and raw card data CSVs.
  - `data/schemas/` - Data validation schemas.
- `tools/` - Development utilities, linters, and asset pipelines.
  - `tools/linter/` - FRD requirement cross-reference and CSV validation tools.
  - `tools/importer/` - Pipelines converting raw balance data into engine resources.
- `src/` - Godot project root (`res://`).
