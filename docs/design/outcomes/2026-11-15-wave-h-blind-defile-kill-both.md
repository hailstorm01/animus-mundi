Status: approved (user, 2026-11-15)
Directive: #155
Rulings: SVY-R49

### 🎲 Design Outcome: Blind Defile's attacker kill depends on its self-kill (2026-11-15)

- **Question:** Do *Blind Defile*'s two kills ("you may kill this structure and that attacking unit") depend on each other when one of the two has left the battlefield before the trigger resolves? Raised by lead-architect in the wave H run (design note Deliverable 6m) before card task T-0242 was written, as a `Status: open` stub at this path; design note Open Question 22.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R49` | *Blind Defile*'s self-kill is `SVY-R14`'s X and its kill of the attacking unit is Y: the attacker dies only if the Defile was killed, and a replaced kill still counts (`SVY-R15`). If the Defile has left the battlefield before the trigger resolves, nothing is killed. If the attacker has left, the Defile is still killed. Card text in which one optional action pays for another says "If you do" (provisional). Wording for actions meant to happen independently of each other, even after the source has left, is not ruled. | Blind Defile (reworded) |

- **Decisions (user, 2026-11-15):**
  - "go with A": *Blind Defile* is read as `SVY-R14` (you may X. If you do, Y), and its text is reworded to match.
  - On the wording convention, **provisional**: "that wording rule is fine for now. There's no non-empty stack interaction at the moment so the wordings have the same end result (both die), but it doesn't hurt to plan ahead." In M2 nothing can remove the Defile between its trigger's placement and resolution, and in the two-Defile case the second Defile dies under either reading, so the rewording changes no reachable M2 game.
  - On its limits: "I do want to leave room for the possibility of stuff like Mangara of Corandor where you can bounce him in response to his own trigger and still exile something, but that would have pretty specific wording I think." So the convention covers only an optional action that pays for another. It does not say that "X and Y" always makes Y depend on X, nor that it never does. **What reopens it:** the first card designed whose actions should still happen after its source has left the battlefield (a Mangara-style "remove this and chosen asset"). Its wording, and what "and" means there, are ruled then.
- **Cards affected:** *Blind Defile* (Path structure, `data/cards/units-and-structures-path.csv:7`). In M2 scope (wave H, on `cards.NotYetExecutable`), in neither starter deck.
- **Alternatives rejected:**
  - B, two independent kills (the literal "and"): when the Defile has already left, the attacker would die for free, which breaks the card as a trade. It also clashes with `FR-TRG-005`, which already treats the self-kill as `SVY-R14`'s X.
  - C, all or nothing (neither dies if either has left): spares the second Defile in the two-Defile case, but needs a new kind of check on resolution and wordier text for one common card. The player can already avoid that case by declining the second "you may" on placement (`SVY-R9`).

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - `data/cards/units-and-structures-path.csv:7`: *Blind Defile*'s "… you may kill this structure and that attacking unit." text.
  - `docs/frd/01_Game_Rules_FRD.md:460` (`FR-TRG-005`): quotes only *"you may kill this structure"* on *Blind Defile*. Still true, but the quote is updated to the new text.
  - `docs/design/m2_starter_set_ability_survey.md:182` (card row): "opposing unit attacks (may, kill self) → `KILL` attacking unit".
- **FRD 01 changes (v3.25 → v3.26):**
  1. `FR-TGT-001`, the **"You May X. If You Do, Y" (`SVY-R14`)** bullet (line 391): append: "*Blind Defile* (*"Whenever an opposing unit attacks, you may kill this structure. If you do, kill that attacking unit."*) is read the same way (`SVY-R49`): the kill of the Defile is X and the kill of the attacking unit is Y. If the Defile has left the battlefield before the trigger resolves, X selects nothing and the attacking unit is not killed. If the attacking unit has left, the Defile is still killed and Y does nothing, so a player with two Defiles who accepts both triggers loses the second Defile for nothing, unless they decline its *"you may"* on placement. Card text in which one optional action pays for another says *"If you do"* (`SVY-R49`). This does not settle a card whose actions are meant to happen independently of each other, even after its source has left the battlefield: that wording is ruled when such a card is designed."
  2. `FR-TRG-005` (line 460): replace `(*"you may kill this structure"* on *Blind Defile*; …)` with `(*"you may kill this structure. If you do, kill that attacking unit."* on *Blind Defile*, `SVY-R49`; …)`, keeping the *Infested Apostle* example.
  3. Revision summary, above v3.24 → v3.25, and bump the title to v3.26: `### Revision Summary (v3.25 -> v3.26): Blind Defile's Kills (2026-11-15)` with the item `1. **Blind Defile's Kills Depend on the Self-Kill (`SVY-R49`; `FR-TGT-001`, `FR-TRG-005`)**: the self-kill is `SVY-R14`'s X and the attacker's kill its Y. Card text in which one optional action pays for another says "If you do". Card text: *Blind Defile*.`
- **Register rows:**
  - Survey §4, add after R48: `| **R49** | Do *Blind Defile*'s two kills depend on each other when one has left the battlefield before the trigger resolves? (Open Question 22, before T-0242) | Blind Defile (reworded) | **Ruled (user, 2026-11-15).** Recorded in FRD 01 v3.26 as `SVY-R49` (`FR-TGT-001`, `FR-TRG-005`). The wording convention is provisional. |`
  - Survey card row (line 182): the effect column becomes `opposing unit attacks (may) → KILL self; if done, KILL attacking unit (`SVY-R14`, `SVY-R49`)`.
  - Keyword interaction matrix: none.
- **Card text:**
  - Old: `Whenever an opposing unit attacks, you may kill this structure and that attacking unit.`
  - New: `Whenever an opposing unit attacks, you may kill this structure. If you do, kill that attacking unit.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md`

#### → `lead-architect` (build board directive)
- **Directive row:**

  | # | Directive | Source | Status |
  | :---: | :--- | :--- | :--- |
  | 155 | **Blind Defile's attacker kill depends on its self-kill (`SVY-R49`; `FR-TGT-001`, `FR-TRG-005`; design note Open Question 22).** In T-0242, write *Blind Defile* as an optional watcher on opposing unit attacks (T-0241's `Trigger.Watches`) with two steps: `OpKill` on the source, then `OpKill` on `SubjectWatched` with `IfDone`. No kernel change. Tests: (a) accepted, the Defile and the attacker both die; (b) declined or timed out, nothing dies (`FR-MOD-001` §6); (c) two Defiles, one attack, both accepted: the first kills itself and the attacker, the second Defile still dies and its Y does nothing; (d) the Defile gone before its trigger resolves (set up the stack item directly, since no M2 card reaches it): the attacker lives; (e) a replaced kill of the Defile still kills the attacker (`SVY-R15`, with a test replacement). Mark design note Open Question 22 answered, citing this outcome. **Checked for contradictions:** none. On `autoloop/m2-wave-h`, `pkg/engine/watch_trigger_test.go` (T-0241) pins only the watcher hook and the watched subject, and `TestFR_TGT_001_WatchedSubjectGoneDoesNothing` agrees with (c). `git grep -i defile` in `pkg` and `cmd` finds only the `cards.NotYetExecutable` entry. *Blind Defile* is in neither starter deck, so the sim baseline doesn't move. | User via game-designer, 2026-11-15 ([outcome](../../../docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md)) | waiting: docs/design/outcomes/2026-11-15-wave-h-blind-defile-kill-both.md applied (then open: before T-0242) |

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None.
