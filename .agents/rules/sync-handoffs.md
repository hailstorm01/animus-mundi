# Sync Handoffs (`pm-lead`, `pm-frd-drafter`, `lead-architect`)

Agreed by the user on 2026-11-23. When one of these three personas commits a change that another one's documents must follow, it passes that work on within the same prompt, instead of handing the user a prompt to run. (Before this, the catalog export's move to M5 took three prompts: `e3e0baca` pm-lead, `d372273e` pm-frd-drafter, `7001c9db` lead-architect.)

`game-designer` handoffs are not covered. The user reviews every Design Outcome before it goes anywhere (game-designer §5), because it decides what the other personas are told about the game.

## 1. Sync or Decision
A handoff is a **sync** when the receiver only brings its own documents in line with a decision already committed upstream, and nothing in it is the receiver's to choose. Examples: a milestone renamed or moved; a deliverable, check or requirement moving from one document to another; a version, status, frozen or active flag; counts, links and the doc map.

It is a **decision**, and goes to the user as a prompt, when any of these holds:
- It needs a ruling, card text, or a change to FRD 01 rule behavior. Those always go to `game-designer`.
- It needs a scope or sequencing trade-off (pm-lead Runbook A Step 3), or changes a milestone's "Done when" beyond what the upstream commit already says.
- It needs a technical choice that a decision record would hold, or supersedes one.
- The upstream commit conflicts with the receiver's own documents, or the receiver thinks it is wrong.

When in doubt, it is a decision.

## 2. The Handoff Prompt
```text
@<persona> handoff (chain: <persona> <sha>[, <persona> <sha> ...]): <one line: what changed upstream>. Update <files or sections> to match.
```
The chain lists every hop so far, oldest first, with the commit each one made. The receiver works from those commits (`git show <sha>`), not from the prompt's summary. If the commits don't support the request, it is a decision (§1, last case). Because the prompt names commits rather than chat, it works unchanged in a fresh session when a chain stops.

## 3. Coordinating
The persona the user prompted is the **coordinator**, in the main session. Receivers run as Claude Code subagents (`.claude/agents/pm-lead.md`, `pm-frd-drafter.md`, `lead-architect.md`), which cannot start subagents of their own, so every further hop comes back through the coordinator.
1. Commit your own change first. Then dispatch each sync handoff it causes, one subagent at a time, and wait for each to finish.
2. Read each subagent's `Ripples:` list. Do a sync ripple addressed to your own persona yourself (§4). Dispatch the others as the next hop, with the chain extended.
3. **Stop the chain** when a handoff is a decision, a persona would run a third time in this chain (the coordinator's own run counts), a hop made no commit, a gate failed, or a subagent stopped early. Hand the user the prompt for whatever is left.
4. **Report** one table, `Hop | Persona | Commit | What changed`, then each decision handoff and each stopped hop with its prompt to run, and where the commits landed (the integration branch, or a side branch to merge under the **Worktrees** guard rail). Never push.

## 4. Receiving
As a subagent, or as the coordinator doing its own ripple:
1. Read the chain's commits. Confirm the request is a sync (§1). If it isn't, change nothing and report it as a decision with the reason.
2. Make only the edits those commits imply, inside your persona's own write permissions. Run your persona's usual gate and freshness check.
3. Commit path-limited on the integration branch, or on a side branch under the **Worktrees** guard rail, with a message that names the source: `<type>(<scope>): <summary> (sync from <persona> <sha>)`.
4. End your report with these lines:
   - `Commit:` the hash, and the branch it is on.
   - `Files:` the paths changed.
   - `Gates:` each command run and its result.
   - `Ripples:` one handoff prompt (§2) per other persona your change now affects, each marked `sync` or `decision` with a one-line reason, or `none`.

## 5. Where Chains Never Run
- **Design Outcomes.** `pm-frd-drafter` applying one (its §2.1) starts no sync handoff. It lists anything the outcome's own handoff sections don't cover in its report, with a prompt for the user, because the user's review didn't cover it. `game-designer`'s `pm-lead` prompt for a scope change stays manual too.
- **Wave runs** (lead-architect D7). Never dispatch. Name the ripple in the session's run-file row and give its prompt in the report.
- **Build tasks.** A chain never writes, dispatches or moves a build task. A `lead-architect` build-loop prompt dispatches sync handoffs only after its one task is `done` or `blocked`.
