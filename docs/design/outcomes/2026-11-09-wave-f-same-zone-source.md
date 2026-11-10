Status: applied: FRD 01 v3.23, 2026-11-09
Directive: #141
Rulings: SVY-R44, SVY-R45

### 🎲 Design Outcome: "in the same zone" is locked on placement, and "that zone" names a move's destination (2026-11-09)

- **Question:** *Broken Horn Armorer*'s "chosen asset in the same zone" is checked again just before its trigger resolves. If its source has moved to the other battlefield zone by then, or has left the battlefield, which zone does "the same zone" mean? Asked by T-0205 (wave F run, `autoloop/m2-wave-f`, session 4); design note Open Question 20. A knock-on case came up in the discussion: the untargeted same-zone picks.
- **Rulings:**

  | Ruling | Settles | Cards |
  | :--- | :--- | :--- |
  | `SVY-R44` | For a card or ability on the Resolution Stack, "in the same zone" means the zone its source is in when the item is placed, or the source's last-known zone if the source has already left the battlefield. That zone is locked with the targets. What happens to the source afterwards doesn't matter, for targeted clauses and for untargeted picks made on resolution. A card that needs its source to still be there says so in its text. A continuous ability isn't on the stack and uses its source's current zone. | Hammerhand Bivouac; Tempest Runner, Mobile Armory, Land Fortress, Popup Bunker, Flickering Ember (outside M2); the same-zone auras (Animoceps Stalk, Animoceps Node, Melodic Pummeler, Hulking Incongruity, Ancestral Monument, Animoceps Spore Cloud) keep the current-zone reading |
  | `SVY-R45` | In a trigger on a move, "that zone" is the zone the triggering move went to, fixed by the event. Card text says "that zone" when an effect is restricted to a move's destination, and "the same zone" only for the source's zone on placement. | Broken Horn Armorer (reworded) |

- **Decisions (user, 2026-11-09):**
  - Intent: "the intention is for armorer to buff someone in the zone it moved to, like it arrives and bolsters a comrade." On *Tempest Runner*, *Mobile Armory* and *Land Fortress*: "they use the "same zone" to establish who it can target before putting on the stack in the first place, and once the ability is committed to the stack it doesn't matter what happened to the activator." On the gap: "we don't currently have a standard for capturing "restricted to the destination zone" and "in the same zone" is ambiguous".
  - `SVY-R44`: "1 - yes. I think if I want "it has to still be there", it'll be extra text and not an implied restriction you're supposed to remember." Untargeted picks: "3 - yes, bunker, bivouac and ember would all track last known zone on placement as well."
  - `SVY-R45`: "2 - agreed." (rewording the Armorer to "chosen asset in that zone").
- **Cards affected:**
  - In M2 and in the Path starter deck: *Broken Horn Armorer* (reworded; no engine definition yet, still in `cards.NotYetExecutable`).
  - In M2, not in a starter deck: *Hammerhand Bivouac* (untargeted, zone locked on placement; deferred by #14).
  - In M2, behavior unchanged: the same-zone auras, which keep using their source's current zone.
  - Outside M2, no text change: *Tempest Runner*, *Mobile Armory*, *Land Fortress* (targeted), *Popup Bunker*, *Flickering Ember* (untargeted).
- **Alternatives rejected:**
  - Reading the source's zone at each check, with its last-known zone once it has left (the lead's recommendation, an extension of `SVY-R10`): the zone is part of picking the target, not a value the effect reads from its source, and once an ability is committed to the stack the activator's fate shouldn't matter.
  - A source that has left the battlefield has no zone, so the targets are illegal: a hidden rule players would have to remember. A card that wants this says so.
  - Keeping "in the same zone" on the Armorer and relying on `SVY-R44`: if the Armorer moved twice before its first trigger was placed, both triggers would lock the final zone. "That zone" states the intent outright and gives future cards a standard wording.

#### → `pm-frd-drafter` (FRD 01 and the ruling register)
- **Superseded wording:** no text in FRD 01 states a source-relative zone reading. Places to update:
  - `docs/frd/01_Game_Rules_FRD.md:370` (`FR-TGT-001` Zone): its examples list only "in the same zone".
  - `docs/design/m2_starter_set_ability_survey.md:198` (*Broken Horn Armorer* row): describes "chosen asset in same zone".
  - Outside pm-frd-drafter's files: design note `docs/architecture/m2-ability-model.md:39` ("same zone as source") and Open Question 20 (`:207`), and T-0205's Covers on `autoloop/m2-wave-f`. The lead updates these (directive #141 (4)).
- **FRD 01 changes** (v3.22 → v3.23):
  - `FR-TGT-001` Zone (line 370): replace the examples with *(e.g., *"in the Main Theater"*, *"in the same zone"* (`SVY-R44`), *"in that zone"* (`SVY-R45`), *"from your discard pile"*)*.
  - `FR-TGT-001`, a new bullet after *Untargeted Selections* (line 377): **"In the Same Zone" Is Locked on Placement (`SVY-R44`)**: *"For a card or ability on the Resolution Stack, "in the same zone" means the battlefield zone its source is in when it is placed on the stack (for Reserves and the Headquarters Zone, that player's own, `FR-ZON-001`), or the source's last-known zone (`FR-ZON-004`) if the source has already left the battlefield. That zone is locked with the targets (`SVY-R3`). What happens to the source afterwards doesn't change it: a target still in that zone at the second verification stage stays legal even if the source has moved or left the battlefield (`FR-TGT-003`), and an untargeted selection (Hammerhand Bivouac, Popup Bunker, Flickering Ember) picks, as it resolves, among the entities in that zone. A card that needs its source to still be there says so in its text. A continuous ability (Animoceps Stalk: "Other allied assets in the same zone have +1 power.") is not on the stack, and uses its source's current zone."*
  - `FR-TGT-001`, a new bullet after it: **"That Zone" After a Move (`SVY-R45`)**: *"In a trigger on a move ("Whenever this unit moves to a new zone", `FR-ZON-002`), "that zone" is the zone the triggering move went to. It is fixed by the event, whatever happens to the source later. Card text says "that zone" when an effect is restricted to a move's destination, and "the same zone" only for the source's zone on placement (`SVY-R44`). Broken Horn Armorer: "Whenever this unit moves to a new zone, chosen asset in that zone gains 1 armor.""*
  - `FR-TGT-003` *Two-Stage Verification* (line 389): after *"…no longer satisfies the ability's zone restriction)"*, add *"A zone restriction relative to the source ("in the same zone", "in that zone") is checked against the zone locked when the item was placed (`SVY-R44`, `SVY-R45`), not the source's zone at resolution."*
  - `FR-ZON-004` *Values Read From the Source* (line 242): add at the end *"The zone of "in the same zone" is not such a value: it is locked when the item is placed (`SVY-R44`)."*
  - Revision summary `(v3.22 -> v3.23): "In the Same Zone" and "That Zone" (2026-11-09)`:
    1. **Same Zone Locked on Placement (`SVY-R44`; `FR-TGT-001`, `FR-TGT-003`, `FR-ZON-004`)**: "in the same zone" on a stack item is the source's zone when placed (last-known if already gone), locked with the targets, for targeted and untargeted clauses. Continuous abilities use the source's current zone.
    2. **"That Zone" After a Move (`SVY-R45`; `FR-TGT-001`)**: "that zone" is a move trigger's destination zone. Card text: *Broken Horn Armorer*.
- **Register rows:**
  - Survey §4: add `R44` and `R45`, each with status **"Ruled (user, 2026-11-09).** Recorded in FRD 01 v3.23 as `SVY-R<n>` (`FR-TGT-001`…)".
  - Survey card rows:
    - *Broken Horn Armorer* (`:198`): effect `self moves zone → ADD_ARMOR 1 chosen asset in that zone (the move's destination)`; rulings `R1, R45`.
    - *Hammerhand Bivouac* (`:174`): note "zone locked on placement (`SVY-R44`)".
  - No keyword-matrix row changes.
- **Card text** (`rulesText` cell only; the Armorer has no definition yet, so `TestDefs_MatchRulesText` doesn't pin it):
  - *Broken Horn Armorer* (`data/cards/units-and-structures-path.csv:8`): `Whenever this unit moves to a new zone, chosen asset in the same zone gains 1 armor.` → `Whenever this unit moves to a new zone, chosen asset in that zone gains 1 armor.`
- **Prompt to run:** `@pm-frd-drafter apply Design Outcome docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md`

#### → `lead-architect` (build board directive)
- **Directive row:** #141 in [`.agents/tasks/m2/BOARD.md`](../../../.agents/tasks/m2/BOARD.md), status `waiting: docs/design/outcomes/2026-11-09-wave-f-same-zone-source.md applied (then open: before T-0205)`.
  - **(1) One locked reference zone on the stack item.** A zone criterion relative to the source compares the entity's zone with a zone stored on the stack item, set once and read at both target checks and by untargeted picks on resolution. "That zone" (`SVY-R45`) sets it from the triggering move's destination (T-0203's move hook). "The same zone" (`SVY-R44`) sets it from the source's zone on placement, or from its last-known record (`itemSource`) if it has already left. `matchesFilter` doesn't look up the source's current zone for stack items. Build it once: a second copy for the other wording would be duplication.
  - **(2) T-0205 (*Broken Horn Armorer*).** Rewrite its Covers for "chosen asset in that zone" (text as the CSV now reads; the Armorer itself may be chosen). Tests to pin: after a March into the Main Theater, a chosen asset there gains 1 armor. If the Armorer moves again or leaves the battlefield before resolution, the target still gains the armor (in principle only in M2; set up the stack item directly). A target that has left that zone is illegal (`FR-TGT-003`).
  - **(3) Continuous auras** (deferred by #14 (b)) use their source's current zone, since they aren't stack items. *Hammerhand Bivouac*'s task uses (1)'s placement zone for its untargeted pick.
  - **(4)** Mark the design note's Open Question 20 answered, and update its `Filter` vocabulary (`:39`, "same zone as source") to "same zone (locked on placement), that zone (move destination)".
  - **Checked for contradictions:** none found. `git grep` for "same zone" and `SameZone` in `pkg` and `cmd` on `autoloop/m2-wave-f` finds only the deploy-zone test helper `sameZones`, and `ability.Filter` has no same-zone criterion. *Broken Horn Armorer* and *Hammerhand Bivouac* are in `cards.NotYetExecutable` with no definitions. T-0205 is still `queued`.

#### → `pm-lead` (portfolio)
- **Milestone or scope impact:** None. No cards move, and no done-when criteria change.
