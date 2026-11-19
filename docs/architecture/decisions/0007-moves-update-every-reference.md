# 0007. A move or rename updates every reference; no aliases or forwarding shims

- **Status:** Accepted
- **Date:** 2026-11-18

## Context

Structural review 13 found keywords in four shapes (board directive #159). The first fix proposed moving `engine.Keywords` into `ability` and keeping `type Keywords = ability.Keywords` in `engine`, so that no caller had to change. The user rejected the alias: a name that points to a name that points to the real thing means a reader has to chase two hops, `go doc` and grep show two homes, and the old location never goes away.

The same habit already shows up at function level. When a task generalizes a function, it keeps the old signature as a wrapper over the new one, so the task's diff stays inside its `Allowed Files`:

- `fireHook` → `fireHookFor(src, hook, nil)` (T-0241)
- `chooseDiscards` → `chooseHandCards(pid, pid, …)` (T-0243)
- `matchesFilter` → `matchesFilterAt(…, nil, …)` (T-0206)

The D2 size rule (four non-test files per task) pushes in the same direction. A move that touches every caller looks too big, so it gets split, and the split needs a bridge.

The repository has no outside consumers of its Go packages (the client protocol is M5), so nothing needs the old names to stay.

## Decision

When a change moves a declaration to another package or file, renames it, or changes its signature, the same change updates every reference and deletes the old name. "Every reference" means code, tests, comments and docs, in every package. The old name survives in no form:

- **No type alias** (`type X = pkg.Y`) and **no re-exported** `var`, `const` or `func` standing in for a moved one.
- **No forwarding wrapper.** A forwarding wrapper is an old declaration whose body, in the change that adds the new one, becomes a call to the new one so that old call sites keep compiling. Callers call the new declaration and pass the argument themselves (`nil`, the zero value).
- **Not covered:** a helper written as its own concept from the start, such as `fireDeploy(a)` over `fireHook`. What decides it is where the wrapper came from, not what its body looks like.

A move is one task. Updating references doesn't count toward the D2 size rule's four-file limit. If the task still has to split, split it by behavior, never in a way that leaves an alias or a wrapper bridging two tasks.

## Consequences

- Moves produce large but mechanical diffs, including renames in test files that the task names as authorized edits. Review checks that the old name is gone (`git grep`) rather than reading each line.
- `cardtown lint fences` rejects type alias declarations in non-test Go code (board directive #165). Forwarding wrappers can't be detected mechanically, so plans (lead-architect D2) and reviews (D5, Runbook E) reject them.
- The three wrappers listed above predate this record. Removing them is a separate decision for the user; until then they are known exceptions, not precedents.
- Revisit this if a package gains outside consumers, for example a published client SDK in M5. A compatibility alias may then be the lesser cost, and it needs its own record.
