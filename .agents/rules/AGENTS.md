# Workspace Rules: Animus Mundi

## Progressive Skill Routers
Read a skill only when its trigger fires. Do not load persona skills for unrelated work.
- **Docs Front Door (`docs-concierge`):** On `@docs-concierge`, or when the user asks what tools/personas exist, what to do next, or how to launch something in the docs workflow. Read `.agents/skills/docs-concierge`.
- **Workflow Steward (`workflow-steward`):** On `@workflow-steward`, or when the user wants to check another persona's work, change how a persona, the routing rules or the build loop behave, add a board directive, or reconcile branches and worktrees. Owns `.agents/` skills, rules and task templates; asks before any change outside them and records it in `.agents/steward-log.md`. Opus, main session. Read `.agents/skills/workflow-steward`.
- **Product Vision & Portfolio Lead (`pm-lead`):** On `@pm-lead`, or when asked to review or update the product vision, milestones, or portfolio view in `docs/PRODUCT_SUMMARY.md`, or when the stakeholder wants to adapt the product vision (expand, contract, pivot). Dispatches analysis requests for active FRDs to `pm-frd-drafter`. Read `.agents/skills/pm-lead`.
- **FRD Authoring & Spec Analysis (`pm-frd-drafter`):** On `@pm-frd-drafter`, when asked to edit FRD 01 rules, to draft an FRD for a milestone that has started, or to analyze and report the state of an FRD (e.g., for `pm-lead`). Read `.agents/skills/pm-frd-drafter`.
- **Rules Red-Teaming (`pm-devils-advocate`):** After an agent makes major structural revisions to `docs/frd/01_Game_Rules_FRD.md` (edits made by `pm-audit-reviewer` do NOT count), or on `@pm-devils-advocate`. Frozen FRDs are never audited. Read `.agents/skills/pm-devils-advocate`.
- **Audit Triage & Resolution (`pm-audit-reviewer`):** Only on `@pm-audit-reviewer` or an explicit user request to review, triage, or resolve audits. Files sitting in `docs/audit/pending/` are NOT a trigger. Read `.agents/skills/pm-audit-reviewer`.
- **Game Design Discussions (`game-designer`):** On `@game-designer`, reviewer option `[E]`, or when the user wants to work through a design question: a ruling, card text, a mechanic, or an open survey or keyword-matrix item. Ends with a Design Outcome Report: a directive row on the build board for `lead-architect`, plus prompts the user runs for `pm-frd-drafter` (FRD 01, rulings, card text) and `pm-lead` (scope). Read `.agents/skills/game-designer`.
- **Technical Architecture (`lead-architect`):** On `@lead-architect`, when a milestone starts and needs a design note, when a technical decision needs a decision record, or when the user explicitly asks for a full TDD or for slicing a TDD into GitHub issues (`cardtown tdd slice`), or for a structural review of the code (Runbook E). Read `.agents/skills/lead-architect`.
- **Build Loop Lead (`lead-architect` Runbook D):** When asked to plan, start, continue or review build tasks in `.agents/tasks/`. Always Opus, in the main Claude Code session. Read `.agents/skills/lead-architect`.
- **Rules Test Author (`rules-test-author`):** On `@rules-test-author T-NNNN`, or when a task file is at `status: tests`. Sonnet. Read `.agents/skills/rules-test-author`.
- **Kernel Engineer (`kernel-engineer`):** On `@kernel-engineer T-NNNN`, or when a task file is at `status: ready` or `changes`. Gemini Flash for `tier: flash`, Sonnet for `tier: sonnet`. Read `.agents/skills/kernel-engineer`.
- **Card Town Verification (`cardtown-linter`):** When auditing specifications, engine scenario tests, TDDs, or CSV balance sheets. Read `.agents/skills/cardtown-linter`.

## Source of Truth
- **Rules:** `docs/frd/01_Game_Rules_FRD.md`, made executable by `pkg/engine`. A rule change lands as FRD 01 text plus a `TestFR_*` scenario test, in the same change. Precedence and workflow: `docs/frd/README.md`.
- **Frozen FRDs (02–05):** design intent only. Do not edit, audit, or design from them; they are rewritten when their milestone starts.

## Guard Rails
- **No unattended chains:** The reviewer -> DA -> reviewer clearing loop (`--clear-backlog`) starts only on an explicit human command. DA never invokes the reviewer. Reviewer edits never trigger the DA outside a human-started loop. Portfolio reconciliation via `pm-lead` starts only on an explicit human prompt; FRD edits never auto-trigger synchronization. Design notes and TDDs (`lead-architect` Runbooks A and B) never auto-trigger GitHub issue publication (Runbook C); always run `cardtown tdd slice -dry-run` first and require explicit human confirmation before `-publish`. The build loop (`lead-architect` Runbook D) runs **one task per human prompt**: once that task is `done` or `blocked`, or needs the user (a Flash dispatch or a ruling), the lead stops and reports. It never dispatches another task, or starts the next deliverable, without a new human prompt. It may write the current deliverable's other tasks, but leaves them `queued`. A merge never triggers the next task.
- **No agent `git push`:** Agents commit locally. Push only when the user asks.
- **Worktrees (all personas):** The main checkout belongs to the user. Never switch its branch, and never stage, stash or reset changes you did not make. Work there only when it is already on the branch you need and `git status --short` is empty. Otherwise work in a temporary worktree in your session's scratch folder (the one the harness provides, or a new folder under the system temp directory if it gives none), never beside the repository:
  ```sh
  git worktree add <scratch-dir>/<name> <branch>                       # existing branch
  git worktree add -b <new-branch> <scratch-dir>/<name> <base-branch>  # new branch
  # edit, run gates and commit inside <scratch-dir>/<name>, then before handing off:
  git worktree remove <scratch-dir>/<name>
  ```
  If `git worktree add` fails because the branch is checked out in another worktree (`git worktree list`), work there only when `git -C <that worktree> status --short` is empty. Otherwise another session is mid-edit: for the integration branch, commit to a side branch (`git worktree add -b <side-branch> <scratch-dir>/side <integration-branch>`) and ask the user to merge it later; for a task branch, stop and tell the user.
- **Pre-commit hook:** It is not versioned by git. After a fresh clone run `cardtown hook install`.
