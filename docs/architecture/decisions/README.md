# Architecture Decision Records

One short file per decision that is expensive to reverse or that someone will later ask "why did we do it this way?" about. Write the record when the decision is made in code, not ahead of it.

- File name: `NNNN-short-slug.md`, numbered in order.
- Length: one page. If it needs more, the decision is probably several decisions.
- Never edit a decided record's substance. Supersede it with a new record and set the old one's status to `Superseded by NNNN`.

## Template

```markdown
# NNNN. Title

- **Status:** Proposed | Accepted | Superseded by NNNN
- **Date:** YYYY-MM-DD

## Context
What forces are at play, including what FRD 01 requires (cite FR IDs; do not restate rules).

## Decision
What we are doing.

## Consequences
What becomes easier, what becomes harder, and what would make us revisit this.
```

## Index

- [0001. Executable rules kernel as the rules oracle; freeze FRDs 02–05](0001-executable-rules-and-frozen-frds.md)
- [0002. Single seeded PCG stream in the M1 kernel](0002-single-prng-stream-in-m1.md)
- [0003. Synchronous Chooser for player choices](0003-synchronous-chooser-for-player-choices.md)
- [0004. Abilities as data values in Go, interpreted by the kernel](0004-abilities-as-data-in-go.md)
