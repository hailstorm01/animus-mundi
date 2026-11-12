---
id: T-0222
title: OpLookTop, a look at the top cards of a deck with free deploys and a bottom order in one prompt (wave F shape, sonnet; #140 (3))
kind: shape
tier: sonnet
wave: F
status: queued
branch: m2/autoloop-wave-f/T-0222-look-top
depends_on: [T-0221]
claims: [shape:OpLookTop (ability.OpLookTop; stepLookTop; the one look decision DecideLook, its outcome options and fallback); the free-deploy zone list shared with stepDeployFree]
cards: []
requirements: [FR-MOD-001, FR-TRN-003a, FR-CST-001, FR-DEK-001]
---

## Goal
The `LOOK_TOP_CARDS` partition (`FR-MOD-001` §4, §6; `SVY-R41`–`SVY-R43`; directive #140 (3)): a resolving step has its controller look at the top N cards of their deck and, in one hidden prompt, choose which of them (up to a number, matching a card filter) are deployed for free, the zone of each, and the order the rest go to the bottom. The design note's Open Question 5 left "the answer encoding" to this task; it is fixed below. Second of three tasks for *Devilhorn Scouting Party* (T-0221 before it, T-0223 after).

Covers (Inputs, Tests and Allowed Files are written against the code of the day before this leaves `queued`):
- `ability.OpLookTop`, `SubjectController` only: `Amount` is N, `Filter` the card-level criteria a deployed card must meet (read with T-0221's `cardMeets`; no zone criterion), `UpTo` how many may be deployed (`UpTo` N > 0: up to N, possibly none, as `SubjectPick` reads it; 0: exactly one, the required pick of `SVY-R41`, when any card is eligible). The only destination besides the bottom is a free deploy; a look into a hand (*Hourglass Diviner*, *Light of the World Soul*, outside M2) is a later field.
- **The answer encoding:** one `Decision` of a new kind `DecideLook` (`String` "Look"), `Hidden` (a deck inspection: no fast path, `FR-MOD-001` §1), `Min` and `Max` 1, whose options are the whole outcomes, each `Label` naming the deploys with their zones and the bottom order. They are listed in this order: deploy sets smaller first, each set in look order; for each set, its zone assignments, Reserves before the Main Theater, card by card; for each, the bottom orders as permutations of the rest in look order, the order seen first. So the first listed option is "deploy nothing (or, for a required pick, the first eligible card into Reserves), the rest in the order seen", which is `SVY-R41`'s and `SVY-R42`'s timeout, and `fallbackPicks` adds `DecideLook` to the first-listed case. *Scouting Party* with three cards has at most 6 + 3 × 2 × 2 = 18 options.
- Each card's zone list comes from one helper shared with `stepDeployFree` (a pure move out of it: `Step.To` if it names a battlefield zone, else the printed Reserves and the Main Theater in that order, none for a card that isn't a supported Unit or Structure), so the eligibility and zone rule of `FR-TRN-003a` 1 and 2 is written once. `stepDeployFree` then asks its `DecideZone` only when that list has two zones.
- On the answer: the looked cards leave the deck, the rest go to the bottom in the chosen order (the first named ends highest), then each chosen card enters through `deployCard` in look order (deploy triggers fire, no move, `SVY-R1`). It returns how many cards it deployed. A deck shorter than N looks at what is there; an empty deck asks nothing.
- `doc.go`'s step list names `OpLookTop`, and its decisions line names `DecideLook` (`SVY-R41`–`SVY-R43`).
- Design note: Open Question 5's sentence on partitioned inspections names this encoding.
- Tests from #140 (3): a timeout deploys nothing, puts all 3 cards on the bottom in the order seen (top first, so the top card ends highest), and the next step (Scratch 1 on the card) still runs; a player's own bottom order is kept; the zone of a card printing both zones is chosen in the same prompt (`SVY-R43`), and no `DecideZone` is asked; a card that fails the filter, or isn't supported, is never offered for a deploy; a required pick (`UpTo` 0) times out to the first eligible card from the top, into Reserves. Synthetic cards only.

- Not this task: *Devilhorn Scouting Party* itself (T-0223, `queued`).

## Log
- 2026-11-11 lead-architect (wave F run, session 23): split off T-0221 under the D2 size rule, with T-0223; `queued` on T-0221.
