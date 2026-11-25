---
id: T-0298
title: "The M2 survey and the ability shapes write-up move to docs/design/survey/, and the fences linter reads every survey there (M3 Deliverable 3a; harness, sonnet)"
kind: harness
tier: sonnet
wave: none (M3 design note Deliverable 3a)
status: ready
branch: m3/T-0298-survey-folder
depends_on: [T-0297]
claims: [folder:docs/design/survey, func:fences.readMarkerIDs, func:fences.Reads, const:fences.surveyDir]
cards: []
requirements: []
---

## Goal
`docs/design/survey/` holds the M2 survey and the ability shapes write-up. `cardtown lint fences` checks `OPEN(SVY-…)` markers against the Ruling Questions tables of every survey in that folder, so the M3 survey (Deliverable 4) can add its own. It reports a ruling ID that is defined twice.

## Inputs
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/toolchain/linter/fences`), Deliverable 3a. ADR-0007 (a move updates every reference).
- Board: no open directive applies. #181 (4) changes two lines of the write-up after this merges. The lead does that, not this task.
- Where it is today (at `main` 91f55ed8):
  - `fences.SurveyFile` and `surveyDir` ([fences.go:34-39](../../../pkg/toolchain/linter/fences/fences.go)) name one file. Nothing outside `fences.go` uses `SurveyFile`.
  - `readMarkerIDs` ([fences.go:190-247](../../../pkg/toolchain/linter/fences/fences.go)) reads that file's `## 4. Ruling Questions` table, and errors when the file is missing. `Reads` ([fences.go:384-399](../../../pkg/toolchain/linter/fences/fences.go)) names it too.
  - The write-up has no `## 4. Ruling Questions` section (its §4 is "For M3"). The survey's table has 57 rows and no duplicate ID (checked 2026-11-24).
- Design (lead decisions):
  - `surveyDir` becomes `docs/design/survey`, and `SurveyFile` is deleted.
  - Surveys are the `*.md` files directly in `surveyDir`, read in name order. Move the per-file table parsing into its own function, and leave its row rules as they are.
  - A file with no Ruling Questions section contributes nothing and is not an error. If no file in the folder has the section, or the folder is missing or empty, `Lint` returns an error naming `docs/design/survey`. (Today's missing-file error would no longer guard anything, because the write-up stays in the folder.)
  - A ruling ID defined a second time, in another survey or in the same one, is a violation at the later row: `docs/design/survey/m3_x.md:7: SVY-R7 is already defined at docs/design/survey/m2_starter_set_ability_survey.md:7`. It goes through the same sorted violation list as the others. The first definition's status decides markers.
  - `Reads` is true for any `*.md` directly in `surveyDir`, and false for the old path.
- Tier `sonnet`: the move needs `git mv`, which is not on the headless allow-list.
- Trial (lead, throwaway worktree): after the `git mv` and the new `surveyDir`, `cardtown lint -all` passed. Only tests whose fixtures name the old path failed (`fences` and `TestSelectChanged_Case1_Paths`).
- **Expected output difference:** none in the sims.
- Size (D2.4): one file with a behavior change, `fences.go`. Every other file only moves or updates a reference.

## Acceptance Cases (harness tasks only; written by the lead)
1. The fixture's survey at `docs/design/survey/m2_starter_set_ability_survey.md` → every existing fences test passes, with only the edits listed under Tests.
2. A second survey `m3_x.md` with **R60** open and **R61** ruled → `OPEN(SVY-R60)` and the first survey's `OPEN(SVY-R7)` pass, `OPEN(SVY-R61)` fails as ruled, and `OPEN(SVY-R62)` fails as not in the survey.
3. **R7** in both surveys → fails, with one violation in the form above that names both files and lines.
4. The same ID twice in one survey → fails, naming both lines.
5. A write-up with a table but no `## 4. Ruling Questions` section, a `.txt` file, and an `.md` in a subfolder, all in the survey folder beside the survey → no IDs from them: `OPEN(SVY-…)` naming a row of theirs fails as not in the survey, and nothing else is reported.
6. The survey removed, the survey folder removed, and a folder holding only a write-up → each an error (not a violation) containing `docs/design/survey`, with `ok` false.
7. `Reads`: true for `docs/design/survey/m2_starter_set_ability_survey.md` and `docs/design/survey/m3_x.md`. False for `docs/design/m2_starter_set_ability_survey.md`, `docs/design/outcomes/x.md`, `docs/design/survey/sub/x.md` and `docs/design/survey/x.txt`.
8. `selectChanged` of the survey's new path → `lintSelection{fences: true}`.
9. Every relative link in the two moved files resolves from the new folder.
10. `git grep` for `design/m2_starter_set_ability_survey` and `design/m2_ability_shapes` finds only:
    - frozen history: outcome bodies, archived tasks and directives, run files, steward history;
    - this task file, the M3 note's Deliverable 3a, and directive #181's row (the lead's);
    - files other personas update after the merge: `docs/frd/`, `docs/PRODUCT_SUMMARY.md`, `.agents/skills/`, `.agents/tasks/README.md`, `.agents/steward-log.md`.
11. `git log --follow` on both moved files shows their history (they are moved with `git mv`).

## Tests
- New, in `pkg/toolchain/linter/fences/fences_test.go`, numbered from Case48 (the last one is Case47): one `TestFences_Case<NN>_*` for each of cases 2 to 5, and one for the parts of case 6 that Case24 doesn't cover.
- **Authorized edits to existing tests**, in the `test(T-0298)` commit:
  - `setupFixtureRepo`: the survey fixture is written under `docs/design/survey/`.
  - `TestFences_Case24_MissingDocsError`: the survey path is the new one, and for it the error must contain `docs/design/survey` (the matrix part is unchanged).
  - `TestReads`: the survey row gets the new path, and case 7's other paths are added.
  - `TestSelectChanged_Case1_Paths` in `cmd/cardtown/lint_test.go`: the "ability survey" path.
- No other existing test changes.

## Commits
1. `test(T-0298): …`: the test changes only. The edited and new tests fail on it.
2. `feat(fences): … (T-0298)`: the move, `fences.go` and the reference updates together, so no commit holds a moved file with stale references.

## Allowed Files
- `pkg/toolchain/linter/fences/fences.go`
- `docs/design/m2_starter_set_ability_survey.md` → `docs/design/survey/m2_starter_set_ability_survey.md` (move; only relative links change)
- `docs/design/m2_ability_shapes.md` → `docs/design/survey/m2_ability_shapes.md` (move; only relative links change)
- Reference updates only: `pkg/engine/doc.go` (line 170, whose comment now says the ruling questions are in the surveys under `docs/design/survey/`), `docs/AGENTS.md` (lines 12-13), `docs/architecture/m2-ability-model.md` (lines 3 and 244), `docs/architecture/m3-full-catalog.md` (line 5)

## Done When
- [ ] `go test ./pkg/toolchain/linter/fences ./cmd/cardtown -count=1` passes
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to `main`)
- [ ] Acceptance cases 9 to 11 checked, with the commands in the Log

## Log
- 2026-11-24 lead-architect: written from M3 design note Deliverable 3a. Status `ready` (a harness task skips `tests`), tier `sonnet` (needs `git mv`).
