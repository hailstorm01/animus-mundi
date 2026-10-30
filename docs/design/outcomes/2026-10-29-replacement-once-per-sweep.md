Status: applied: FRD 01 v3.18, 2026-10-29
Directive: #81
Rulings: SVY-R24
Audit to close: none

### 🎲 Design Outcome: A replacement saves an asset at most once per sweep (2026-10-29)

- **Question:** A destruction replacement's steps may leave the asset on the battlefield and still lethal (for example "exhaust it instead"). What happens then? Does the same sweep take it up again, may the same replacement apply to it again, and how does the loop end? Raised by structural review 5 as survey R24 (design note [Open Question 15](../../architecture/m2-ability-model.md#open-questions)). Today `sweepDeaths` takes each lethal asset up only once per sweep ([damage.go:143-156](../../../pkg/engine/damage.go)). Such an asset would stay in play while lethal, `CheckInvariants` would report "lethal but survived the death sweep", and the same replacement would apply again at every later sweep, repeating its side effects each time. No printed card reaches this yet.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R24` | Within one death sweep, a destruction replacement applies at most once to a given asset. An asset still lethal after its replacement is taken up again in the sweep's next pass. The replacement already used on it doesn't apply again, so a different replacement applies or the asset is destroyed. Each kill step is a separate window. A replacement that keeps its asset on the battlefield saves it from a kill, but saves it from lethal damage only if its own steps remove that damage. | None printed. *Shifting Impossibility* is unaffected. |

- **Decisions (user, 2026-10-29):**
  - How the physical table plays it: "'exhaust it instead' on its own wouldn't save it from lethal damage because it's not leaving the battlefield so the damage on it doesn't clear, it would save it from 'kill this creature'. The text would need some flavor of 'and remove all damage from it' I think."
  - Direction: chose "Only from kills (A)": one save per asset per sweep, then the next pass. This beat keeping today's behaviour (B), "surviving clears damage" (C), and deferring with a design constraint (D).
  - Scope: chose "Per kill step / per sweep": a replacement that saved an asset from a kill during a stack item may apply to it once more at the sweep right after that item.
- **Cards affected:**
  - *Shifting Impossibility*: in M2, with two copies in the Animus starter deck. Unchanged, because it always moves the unit to its owner's hand.
  - Future cards: a destruction replacement that keeps its asset on the battlefield must remove the damage in its own text (for example "and remove all damage from it") to save the asset from lethal damage. Without that, it only saves the asset from kill effects.
- **Alternatives rejected:**
  - **B, today's behaviour:** the asset stays in play while lethal and its replacement repeats at every sweep. That is board state players can't easily read, and the side effects repeat after every stack item.
  - **C, surviving clears damage:** it adds a healing step no card prints. The table says damage stays on an asset that never leaves the battlefield.
  - **D, defer with a design constraint:** an aura that drops Max Health to 0 or below still needs A's answer, and A costs little now.
  - **One use per stack item and its sweep (the `SVY-R21` moment):** the sweep would have to remember replacements used during resolution. The only difference is one extra side effect before the same death.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:**
  - FRD 01, line 462, `FR-REP-001` "Replacement Within a Death Sweep": "The sweep then repeats: assets that a byproduct newly left lethal die in the next pass, which is a new moment." It covers only assets newly left lethal.
  - FRD 01, line 456, `FR-REP-003`: it doesn't say what one discrete death event is when the same asset is lethal across several passes.
- **FRD 01 changes (v3.17 → v3.18).** Apply [2026-10-29-choosing-between-replacements.md](2026-10-29-choosing-between-replacements.md) (v3.16 → v3.17) first if it hasn't been applied yet.
  - **`FR-REP-003`, line 456:** replace the bullet with:
    > - **Single Application per Discrete Event (`FR-REP-003`, `SVY-R24`)**: A given replacement effect may apply at most once to each discrete event. For a destruction replacement, an asset's death is one event for each kill step and for each death sweep. Within one sweep, a replacement applies at most once to a given asset, however many of the sweep's passes take that asset up (`FR-LAY-002`).
  - **`FR-REP-001` "Replacement Within a Death Sweep", line 462:** replace the sentence "The sweep then repeats: assets that a byproduct newly left lethal die in the next pass, which is a new moment." with:
    > The sweep then repeats with every asset that is still lethal, which is a new moment. That covers an asset a byproduct newly left lethal, and an asset a replacement's steps left on the battlefield and still lethal. A replacement that has already applied to an asset in this sweep does not apply to it again (`FR-REP-003`, `SVY-R24`), so a different replacement applies or the asset is destroyed.

    At the end of the same bullet, add:
    > *Example:* a replacement reading *"If an allied unit would be killed, deal 1 damage to this instead"* applies to an allied unit at −2 Current Health. The unit stays on the battlefield and stays lethal, and the sweep's next pass destroys it. To save a unit from lethal damage, a replacement's steps must also leave it non-lethal, for example *"and remove all damage from it"*. The same replacement does save a unit from a kill effect, because a killed unit with Health left is not lethal afterwards. If one stack item both kills that unit and leaves it at 0 Current Health, the replacement applies to the kill and once more at the sweep after the item, and then the sweep's next pass destroys the unit.
  - **Revision summary:** `### Revision Summary (v3.17 -> v3.18): A Replacement Saves an Asset at Most Once per Sweep (2026-10-29)`
    1. **Once per Sweep (`SVY-R24`; `FR-REP-003`, `FR-REP-001`, `FR-LAY-002`)**: within one death sweep, a destruction replacement applies at most once to a given asset. An asset its replacement leaves on the battlefield and still lethal is taken up again in the sweep's next pass, where a different replacement applies or it is destroyed. A replacement that keeps its asset on the battlefield saves it from lethal damage only if its own steps remove that damage.
  - **Title line:** `Master FRD v3.17` becomes `Master FRD v3.18`.
- **Register rows:**
  - **Survey §4, row R24, status cell:** **Ruled (user, 2026-10-29): once per sweep.** Within one death sweep, a destruction replacement applies at most once to a given asset. An asset still lethal afterwards is taken up again in the sweep's next pass, where a different replacement applies or it is destroyed. Each kill step is a separate window. A replacement that keeps its asset on the battlefield saves it from lethal damage only if its own steps remove that damage. Recorded in FRD 01 v3.18 as `SVY-R24` (`FR-REP-003`, `FR-REP-001`, `FR-LAY-002`).
  - **Survey line 18:** add R24 (it becomes "R20–R24 on 2026-10-29"), and change the range to `SVY-R1`–`SVY-R24`.
- **Card text:** None.
- **Audit to close:** None.
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-10-29-replacement-once-per-sweep.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #81 in [`.agents/tasks/m2/BOARD.md`](../../../.agents/tasks/m2/BOARD.md), status `waiting: docs/design/outcomes/2026-10-29-replacement-once-per-sweep.md applied`. It replaces `newlyLethal`'s "only assets not yet seen" with "a replacement applies at most once to a given asset per sweep" (`replacementsFor` leaves out one already used on that asset in this sweep), and lists four settling tests that use test-only replacements: once per sweep, then destroyed; a second replacement in the next pass; a healing replacement saves; and a kill followed by the sweep applies twice. It lands after #79, whose `OPEN(SVY-R24)` comment bullet is dropped. Merged code that contradicts the ruling: `newlyLethal` and the `sweepDeaths` comment ([damage.go:120-156](../../../pkg/engine/damage.go)). No merged test pins the old behaviour. Checked by grepping every `WouldBeDestroyed` replacement in `pkg/engine/replacement_test.go` and `pkg/engine/cards/animus.go`: each one moves the asset off the battlefield.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. No M2 card reaches this, and it adds one directive to the existing wave G.1 order.
