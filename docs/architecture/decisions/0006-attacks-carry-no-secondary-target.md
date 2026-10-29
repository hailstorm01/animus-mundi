# 0006. Attack events carry no secondary target

- **Status:** Accepted
- **Date:** 2026-10-28

## Context

ADR-0005 fixed the event schema. Its T-0026 addendum lists `Secondary` among a `place` event's fields, and `Event` carries it as the JSON key `secondary` (`omitempty`) on an Attack's `action` and `place` events: the Berserk secondary target. Its Consequences say that removing a field supersedes the record.

Berserk left the kernel in wave B.2 (board directive #68, the user's code review, 2026-10-28). Since T-0089 an Attack with a secondary target is illegal, so every event since has carried 0 there, which `omitempty` drops. Survey `SVY-R12` replaced Berserk with Overkill (`FR-KWD-002p`, [outcome](../../design/outcomes/2026-10-28-berserk-replaced-by-overkill.md)). Overkill picks no extra target: it is a trigger on `HookCombatKill` with one more gate (board directive #72). No planned keyword needs the field.

## Decision

`Event` loses `Secondary` and its `secondary` key (T-0092). An Attack's `action` and `place` events carry the attacker as `Source` and the defender as `Target`, and nothing else. `Action.Secondary` and `StackItem.Secondary` go too (T-0093). The rest of ADR-0005 stands.

## Consequences

- Event files written by builds since T-0089 don't change: the key was always 0 and omitted. An older file that holds a `secondary` key still decodes, because `pkg/eventlog` doesn't reject unknown keys. It came from a build with other rules, so it doesn't replay (ADR-0005: a replay is exact only on the same build).
- If a keyword picks a second combat target again, the field comes back under a new decision record.
