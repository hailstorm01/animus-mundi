# 0009. Cards are identified by their printed name through M4

- **Status:** Accepted
- **Date:** 2026-11-23

## Context
`FR-DAT-002` makes a card's printed name its identity: it must be unique across every sheet, and deck lists and ability definitions (`FR-ABL-001`) already name cards by it. FRD 03 leaves the catalog's identifier form open: the printed name, or a derived slug (`FR-CMP-023`). The pm-lead review of 2026-11-23 settled the product side. The printed name is the identity through M3 and M4. A card rename policy is a required item at the start of M5, because replays, the session's catalog checksum (`FR-SES-001`, `FR-CMP-025`) and M6 deck codes all store card identifiers. The form chosen now must leave room for a stable identifier column later.

Today the name is the key everywhere: `cards.Defs` and `LoadCards`' map, `data/decks/*.txt`, `NotYetExecutable`, event records (`eventlog`) and the view.

## Decision
- The catalog identifies a card by its printed name, exactly as the sheet spells it. No slug is derived.
- Entries are written as a **list of objects**, each with a `name` field, sorted by name in byte order of its UTF-8 text (Go's `sort.Strings`). The name is a field of the entry, never a map key, so the catalog's shape doesn't depend on which field identifies a card.
- The checksum (`FR-CMP-025`) covers the name like any other printed value, so renaming a card changes the checksum.

## Consequences
- Nothing is added now: the name already is the key in every file and package, and a slug would be a second identity that nothing reads.
- A rename is a change to every reference in one task ([ADR 0007](0007-moves-update-every-reference.md), `FR-DAT-002`): the sheet row, the `Defs` key, deck lists, and the allowlists in `pkg/engine/cards`. Old replays and event logs then name a card that no longer exists. Until M5 that is acceptable, because nothing stores them across catalog versions.
- **What adding a stable identifier column would change** (decided at M5 start with the rename policy):
  - **Sheets:** a new `id` column in all ten CSVs, a new header for `FR-DAT-006` and `cardtown lint sheets`, and a uniqueness check for it beside the name check (`FR-DAT-002`, `FR-CMP-027`).
  - **Catalog:** each entry gains an `id` field, and the sort key moves to it. Every entry's position and the checksum change at once. That is a structural change, so it is a MAJOR dataset version bump (`FR-CMP-026`). Because entries are already a list of objects, no reader has to change how it walks the file.
  - **Kernel and tools:** `cards.Defs`, `LoadCards` and the decklist loader keep or drop name keys by the policy's choice. If deck lists and events switch to the id, `eventlog`'s records and `sim.ResolveDeck` change with them, and the old `.jsonl` files need a name-to-id mapping or become unreadable.
  - **Unchanged:** ability definitions, the rules and every scenario test. They refer to cards only by the key.
- Revisit sooner if any M3 or M4 work stores card identifiers outside the repository, or the stakeholder renames a card that a committed replay or log names.
