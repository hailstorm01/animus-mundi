# 0002. Single seeded PCG stream in the M1 kernel

- **Status:** Accepted
- **Date:** 2026-10-19

## Context

Shuffles (`FR-DEK-002`), first-player determination (`FR-SET-003`) and random timeout fallbacks (`FR-MOD-001`) need randomness that is reproducible from a seed, so a match can be replayed and a failing fuzz seed can be re-run. The frozen [FRD 05](../../frd/05_Headless_Game_Engine_FRD.md) proposes three isolated PCG32/SplitMix64 streams so that, for example, a targeting roll cannot shift later shuffles.

## Decision

The M1 kernel uses one `math/rand/v2` PCG generator per match, seeded from `Config.Seed`. Policies used in playouts have their own generators, so the policy never perturbs game randomness.

## Consequences

- Every match and every fuzz failure reproduces from its seed (`TestPlayoutsAreDeterministicPerSeed`).
- Adding a new random draw anywhere in the rules changes later shuffles for the same seed. That is fine while there are no stored replays.
- Revisit before the first persisted replay format (M5): that is when stream isolation starts to pay for itself, and the FRD 05 proposal is the starting point.
