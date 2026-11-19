---
id: T-0283
title: "`ability.Keyword` becomes `KeywordTrigger`, and `Trigger.Keyword` and `StackItem.Keyword` become `FromKeyword` (directive #171; wave H.2 harness, flash, pure rename)"
kind: harness
tier: flash
wave: H.2 (board directive #171)
status: done
branch: m2/autoloop-wave-h.2/T-0283-keyword-trigger
depends_on: [T-0282]
claims: [type:ability.KeywordTrigger, field:ability.Trigger.FromKeyword, field:engine.StackItem.FromKeyword]
cards: []
requirements: []
---

## Goal
Since ADR 0008, `ability.Keyword`, the keyword a trigger comes from ([ability.go:211-227](../../../pkg/engine/ability/ability.go)), sits beside `ability.Keywords`, the keyword set ([keywords.go:13](../../../pkg/engine/ability/keywords.go)), and the fields `Trigger.Keyword` and `StackItem.Keyword` ([stack.go:43](../../../pkg/engine/stack.go)) sit beside `Step.Keywords` and `Static.Keywords`, so `go doc ./pkg/engine/ability` reads as if one were an element of the other. The type becomes `KeywordTrigger` (`KeywordTriggerNone`, `KeywordTriggerAggressive`), and both fields become `FromKeyword`. Every reference is updated in this task, with no alias ([ADR 0007](../../../docs/architecture/decisions/0007-moves-update-every-reference.md)). It is a pure rename: the same test names pass, and match output is byte for byte the same. When this merges, #171 is `applied`.

## Inputs
- Board directive #171 (structural review 14, finding 3), approved by the user 2026-11-19, `open`; wave H.2, after #170 (T-0282, `done`). Design note Deliverable 6o ([m2-ability-model.md:230](../../../docs/architecture/m2-ability-model.md)) names it.
- Where it is today (at `autoloop/m2-wave-h.2` 6cf38a1): `git grep -n -w -e Keyword -e KeywordNone -e KeywordAggressive -- "*.go"` and a search for `.Keyword` find the declarations in `ability.go` (the package doc's enum list at line 12, the type, its constants, `String`, `Aggressive` and `Trigger.Keyword`), the comment at `keywords.go:37` and `KeywordAggressive.String()` at `keywords.go:41`, `StackItem.Keyword` and `StackLabel` in `stack.go`, `fireAttacks` and `placeTrigger` in `trigger.go`, and the seven test files #171 names. The other `Keyword` hits are the words "Keyword interaction matrix" in `pkg/toolchain` (printed text and a test fixture), which stay.
- Design (lead decisions):
  - Names exactly as #171 gives them. The comments that name the type (the package doc, the type's doc, `String`'s doc, `keywordDefs`' doc, `StackItem.FromKeyword`'s comment) name the new one; two comments are re-wrapped to stay inside 80 columns. `gofmt` re-aligns the `Trigger` and `StackItem` field blocks, which is why `stack.go` shows eight changed lines.
  - The message `Keyword names = %q, %q` in `TestFR_KWD_002f_AggressiveConstructor` becomes `KeywordTrigger names = %q, %q`, so #176's search for the old type name finds nothing there.
  - Test names stay, including `TestView_AbilityItemShowsItsKeyword` and `TestFR_CBT_002_AggressiveItemIsAnAbilityItemWithTheKeyword`: they name the behavior, not the type.
- **Authorized test edits** (#171; D2.4): the seven test files under `Tests`, reference updates only, exactly as Edits 1-7, in the `test(T-0283)` commit. At that commit the tests reference names that don't exist yet, so the packages' tests don't build (case 1), as in T-0277.
- Docs: no doc names `ability.Keyword` except ADR 0008 (a decision's history, which stays) and the design note's #171 line (the record of this directive). The design note's Stack paragraph ([m2-ability-model.md:52](../../../docs/architecture/m2-ability-model.md)) names the field `Keyword`; it is the lead's to write, and the lead updates it in the commit that marks this task `done`. `docs/architecture/engine-structs.md` and `pkg/engine/doc.go` don't name it.
- Reuse search (D2.3): nothing is added.
- Size (D2.4): four non-test files, three of them reference-only (`keywords.go`, `stack.go`'s `StackLabel` lines, `trigger.go`); a rename's reference updates don't count toward the four.
- **Expected output difference:** none: `go tool task sim-compare` passes.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 6cf38a1: `gofmt -l pkg cmd` clean, `go vet ./pkg/engine/... ./pkg/view/...` clean, `go tool task check` exit 0, `go tool task sim-compare -- autoloop/m2-wave-h.2` `[PASS]`; with only Edits 1-7 applied, `go vet ./pkg/engine/` fails with `aggressive_item_test.go:53:72: (*items)[1].FromKeyword undefined`. Blobs: `ability.go` 7f1fb87, `keywords.go` cea96ce, `stack.go` 09edf45, `trigger.go` 18ccc1c, `aggressive_item_test.go` 5286426, `cards_path_allshan_test.go` c7322db, `cornered_keywords_test.go` 4402de8, `pkg/engine/stack_label_test.go` dc890fa, `watch_trigger_test.go` ece7dd6, `pkg/view/stack_label_test.go` 17ab7d2, `view_test.go` bd052bf.

## Commits
1. `test(T-0283): tests name KeywordTrigger and FromKeyword` with Edits 1-7 only. At this commit the `pkg/engine` and `pkg/view` tests do not build (case 1).
2. `refactor(engine): ability.Keyword becomes KeywordTrigger, its fields FromKeyword (T-0283)` with Edits 8-11.

No double quotes in either message: the allow-list refuses an escaped one.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
### 1. `pkg/engine/aggressive_item_test.go` (commit 1, reference only)
```diff
@@ -50,7 +50,7 @@ func TestFR_CBT_002_AggressiveIsListedFirstAndPlacedFirstByDefault(t *testing.T)
 	if got := placeDepths(s.g.Log); !reflect.DeepEqual(got, []int{1, 2, 3}) {
 		t.Errorf("place depths = %v, want [1 2 3]", got)
 	}
-	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].Keyword != ability.KeywordAggressive || (*items)[2].Keyword != ability.KeywordNone {
+	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].FromKeyword != ability.KeywordTriggerAggressive || (*items)[2].FromKeyword != ability.KeywordTriggerNone {
 		t.Errorf("placed items = %+v, want Attack, Aggressive, printed", *items)
 	}
 }
@@ -61,7 +61,7 @@ func TestFR_TRG_003_AggressiveCanBePlacedLast(t *testing.T) {
 	att, def := raiderWithPrinted(s)
 	items := recordPlaced(s)
 	s.do(s.attack(att, def))
-	if len(*items) != 3 || (*items)[1].Keyword != ability.KeywordNone || (*items)[2].Keyword != ability.KeywordAggressive {
+	if len(*items) != 3 || (*items)[1].FromKeyword != ability.KeywordTriggerNone || (*items)[2].FromKeyword != ability.KeywordTriggerAggressive {
 		t.Errorf("placed items = %+v, want Attack, printed, Aggressive on top", *items)
 	}
 }
@@ -73,12 +73,12 @@ func TestFR_CBT_002_AggressiveItemIsAnAbilityItemWithTheKeyword(t *testing.T) {
 	items := recordPlaced(s)
 	s.do(s.attack(att, def))
 	ag, printed := (*items)[1], (*items)[2]
-	if ag.Kind != StackAbility || ag.Keyword != ability.KeywordAggressive || ag.Source != att.ID || ag.Controller != 0 ||
+	if ag.Kind != StackAbility || ag.FromKeyword != ability.KeywordTriggerAggressive || ag.Source != att.ID || ag.Controller != 0 ||
 		!reflect.DeepEqual(ag.Effect, ability.Aggressive(2).Effect) {
 		t.Errorf("Aggressive item = %+v", ag)
 	}
-	if printed.Keyword != ability.KeywordNone {
-		t.Errorf("printed item has keyword %v", printed.Keyword)
+	if printed.FromKeyword != ability.KeywordTriggerNone {
+		t.Errorf("printed item has keyword %v", printed.FromKeyword)
 	}
 	pl := eventsOfKind(s.g.Log, "place")
 	if want := "player 0 places Aggressive trigger of Raider on the stack"; pl[1].Detail != want {
@@ -91,7 +91,7 @@ func TestFR_CBT_002_AggressiveItemIsAnAbilityItemWithTheKeyword(t *testing.T) {
 
 func TestFR_KWD_002f_AggressiveConstructor(t *testing.T) {
 	tr := ability.Aggressive(3)
-	want := ability.Trigger{Hook: ability.HookAttacks, Keyword: ability.KeywordAggressive,
+	want := ability.Trigger{Hook: ability.HookAttacks, FromKeyword: ability.KeywordTriggerAggressive,
 		Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpModifyStat, Subject: ability.SubjectSelf, Amount: ability.Value{N: 3}}}}}
 	if !reflect.DeepEqual(tr, want) {
 		t.Errorf("Aggressive(3) = %+v, want %+v", tr, want)
@@ -99,8 +99,8 @@ func TestFR_KWD_002f_AggressiveConstructor(t *testing.T) {
 	if tr.Optional || tr.Cornered || len(tr.Effect.Targets) != 0 {
 		t.Errorf("Aggressive(3) must not be Optional, Cornered or targeted: %+v", tr)
 	}
-	if ability.KeywordAggressive.String() != "Aggressive" || ability.KeywordNone.String() != "" {
-		t.Errorf("Keyword names = %q, %q", ability.KeywordAggressive, ability.KeywordNone)
+	if ability.KeywordTriggerAggressive.String() != "Aggressive" || ability.KeywordTriggerNone.String() != "" {
+		t.Errorf("KeywordTrigger names = %q, %q", ability.KeywordTriggerAggressive, ability.KeywordTriggerNone)
 	}
 }
 
@@ -126,7 +126,7 @@ func TestFR_KWD_002f_AggressiveIsReadWhenTheAttackHappens(t *testing.T) {
 	c.Keywords.Aggressive = 2
 	items = recordPlaced(s)
 	s.do(s.attack(att, def))
-	if len(*items) != 2 || (*items)[1].Keyword != ability.KeywordAggressive {
+	if len(*items) != 2 || (*items)[1].FromKeyword != ability.KeywordTriggerAggressive {
 		t.Errorf("one Aggressive item expected: placed %+v", *items)
 	}
 	s.wantDamage(def, 3)
```

### 2. `pkg/engine/cards_path_allshan_test.go` (commit 1, reference only)
```diff
@@ -119,7 +119,7 @@ func TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger(t *testing.T) {
 	count := func(items []StackItem) int {
 		n := 0
 		for _, it := range items {
-			if it.Kind == StackAbility && it.Keyword == ability.KeywordAggressive {
+			if it.Kind == StackAbility && it.FromKeyword == ability.KeywordTriggerAggressive {
 				n++
 			}
 		}
@@ -136,7 +136,7 @@ func TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger(t *testing.T) {
 			t.Fatalf("Cornered: Aggressive 3 places one Aggressive StackAbility item on attack: got %d in %+v", n, *items)
 		}
 		for _, it := range *items {
-			if it.Keyword == ability.KeywordAggressive && it.Source != al.ID {
+			if it.FromKeyword == ability.KeywordTriggerAggressive && it.Source != al.ID {
 				t.Errorf("the Aggressive item's source is Allshan: %+v", it)
 			}
 		}
```

### 3. `pkg/engine/cornered_keywords_test.go` (commit 1, reference only)
```diff
@@ -30,7 +30,7 @@ func corneredCard(name string, power, health int, c ability.WhileCornered, trigs
 func aggressiveItems(items []StackItem) int {
 	n := 0
 	for _, it := range items {
-		if it.Kind == StackAbility && it.Keyword == ability.KeywordAggressive {
+		if it.Kind == StackAbility && it.FromKeyword == ability.KeywordTriggerAggressive {
 			n++
 		}
 	}
@@ -145,7 +145,7 @@ func TestCard_Allshan_AggressiveComesFromTheKeyword(t *testing.T) {
 			t.Fatalf("Cornered attack places one Aggressive keyword item: got %d in %+v (FR-KWD-002f)", n, *items)
 		}
 		for _, it := range *items {
-			if it.Keyword == ability.KeywordAggressive && it.Source != al.ID {
+			if it.FromKeyword == ability.KeywordTriggerAggressive && it.Source != al.ID {
 				t.Errorf("the Aggressive item's source is Allshan: %+v", it)
 			}
 		}
```

### 4. `pkg/engine/stack_label_test.go` (commit 1, reference only)
```diff
@@ -29,7 +29,7 @@ func TestStackLabel_EachKind(t *testing.T) {
 		{&StackItem{Kind: StackScratch}, "Scratch"},
 		{&StackItem{Kind: StackManeuver, Inst: m}, "Maneuver Played"},
 		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst}, "ability of Open Source"},
-		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive}, "Aggressive trigger of Open Source"},
+		{&StackItem{Kind: StackAbility, Source: a.ID, Inst: a.Inst, FromKeyword: ability.KeywordTriggerAggressive}, "Aggressive trigger of Open Source"},
 		{&StackItem{Kind: StackAbility, Source: 999, Inst: gone, Tracked: true}, "ability of Gone"},
 	} {
 		if got := s.g.StackLabel(tc.it); got != tc.want {
@@ -69,7 +69,7 @@ func TestStackLabel_FaceDownSourceShowsOnlyTheScratchFace(t *testing.T) {
 func TestStackLabel_FaceUpSourcePlaceEventNamesItsInstance(t *testing.T) {
 	s := newScene(t)
 	a := s.put(0, ZoneMainTheater, unit("Open Source", 1, 1, ability.Keywords{}))
-	s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst, Keyword: ability.KeywordAggressive})
+	s.g.push(&StackItem{Kind: StackAbility, Controller: 0, Source: a.ID, Inst: a.Inst, FromKeyword: ability.KeywordTriggerAggressive})
 	pl := eventsOfKind(s.g.Log, "place")
 	if len(pl) == 0 {
 		t.Fatalf("push logged no place event")
```

### 5. `pkg/engine/watch_trigger_test.go` (commit 1, reference only)
```diff
@@ -100,7 +100,7 @@ func TestFR_TRG_003_OpposingWatcherResolvesBeforeAttackerTriggers(t *testing.T)
 	att := s.put(0, ZoneMainTheater, unit("Raider", 3, 9, ability.Keywords{Aggressive: 1}))
 	items := recordPlaced(s)
 	s.do(s.attack(att, s.g.Players[1].HQ))
-	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].Keyword != ability.KeywordAggressive || (*items)[2].Source != w.ID {
+	if len(*items) != 3 || (*items)[0].Kind != StackAttack || (*items)[1].FromKeyword != ability.KeywordTriggerAggressive || (*items)[2].Source != w.ID {
 		t.Fatalf("the non-active player's watcher is placed after, so above, the attacker's triggers and resolves first (FR-TRG-003): placed %+v", *items)
 	}
 }
```

### 6. `pkg/view/stack_label_test.go` (commit 1, reference only)
```diff
@@ -20,7 +20,7 @@ func TestView_StackItemLabel(t *testing.T) {
 	down := s.put(1, engine.ZoneMainTheater, testUnit("Secret Source", 1, 1, ability.Keywords{}))
 	down.FaceDown = true
 	s.g.Stack = []*engine.StackItem{
-		{Kind: engine.StackAbility, Controller: 0, Source: up.ID, Inst: up.Inst, Keyword: ability.KeywordAggressive},
+		{Kind: engine.StackAbility, Controller: 0, Source: up.ID, Inst: up.Inst, FromKeyword: ability.KeywordTriggerAggressive},
 		{Kind: engine.StackAbility, Controller: 1, Source: down.ID, Inst: down.Inst},
 		{Kind: engine.StackMarch, Controller: 0, Source: up.ID, Zone: engine.ZoneMainTheater},
 	}
```

### 7. `pkg/view/view_test.go` (commit 1, reference only)
```diff
@@ -704,7 +704,7 @@ func TestView_Case5_LogLines(t *testing.T) {
 func TestView_AbilityItemShowsItsKeyword(t *testing.T) {
 	s := newContentScene(t)
 	s.g.Stack = []*engine.StackItem{
-		{Kind: engine.StackAbility, Controller: 0, Source: s.lancer.ID, Inst: s.lancer.Inst, Keyword: ability.KeywordAggressive},
+		{Kind: engine.StackAbility, Controller: 0, Source: s.lancer.ID, Inst: s.lancer.Inst, FromKeyword: ability.KeywordTriggerAggressive},
 		{Kind: engine.StackAbility, Controller: 0, Source: s.sentry.ID, Inst: s.sentry.Inst},
 	}
 	v := view.PlayerView(s.g, 0)
```

### 8. `pkg/engine/ability/ability.go`: the type, its constants and `Trigger.FromKeyword` (commit 2)
```diff
@@ -9,8 +9,8 @@
 //   - An enum with no default starts at iota + 1, so an unset field (the zero
 //     value) is invalid and never passes for the first member (Hook, Op).
 //   - An enum whose zero value is a meaningful default starts at iota, and the
-//     first member's comment says "the zero value" (TurnOwner, Keyword, Zone,
-//     Allegiance, ScratchCompany, Subject, ValueSource).
+//     first member's comment says "the zero value" (TurnOwner, KeywordTrigger,
+//     Zone, Allegiance, ScratchCompany, Subject, ValueSource).
 package ability
 
 // Def is the executable definition of one card's printed abilities.
@@ -208,19 +208,20 @@ const (
 	TurnYours                  // only during the trigger controller's own turn
 )
 
-// Keyword names a keyword that is a triggered ability, the one a trigger comes
-// from (Aggressive, FR-KWD-002f). The zero value is a printed ability.
-type Keyword int
+// KeywordTrigger names a keyword that is a triggered ability, the one a
+// trigger comes from (Aggressive, FR-KWD-002f). The zero value is a printed
+// ability.
+type KeywordTrigger int
 
 const (
-	KeywordNone       Keyword = iota // the zero value: a printed ability
-	KeywordAggressive                // "Aggressive N" (FR-KWD-002f)
+	KeywordTriggerNone       KeywordTrigger = iota // the zero value: a printed ability
+	KeywordTriggerAggressive                       // "Aggressive N" (FR-KWD-002f)
 )
 
-// String returns the keyword's printed name, or "" for KeywordNone.
-func (k Keyword) String() string {
+// String returns the keyword's printed name, or "" for KeywordTriggerNone.
+func (k KeywordTrigger) String() string {
 	switch k {
-	case KeywordAggressive:
+	case KeywordTriggerAggressive:
 		return "Aggressive"
 	}
 	return ""
@@ -245,7 +246,7 @@ func (d Duration) String() string {
 // Aggressive is the trigger the keyword "Aggressive N" means: whenever this
 // unit attacks, it gets +N power until end of turn (FR-KWD-002f).
 func Aggressive(n int) Trigger {
-	return Trigger{Hook: HookAttacks, Keyword: KeywordAggressive,
+	return Trigger{Hook: HookAttacks, FromKeyword: KeywordTriggerAggressive,
 		Effect: Effect{Steps: []Step{{Op: OpModifyStat, Subject: SubjectSelf, Amount: Value{N: n}}}}}
 }
 
@@ -256,10 +257,10 @@ type Trigger struct {
 	ToZone Zone // HookDeploy and HookMoves only: fires only if the asset was deployed or moved into this zone
 	// ActiveZone is where the trigger works from (survey G1): ZoneAny, the zero value, is the
 	// battlefield; ZoneDiscard is its card's owner's discard pile. No other value is used.
-	ActiveZone Zone
-	Cornered   bool    // "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)
-	Optional   bool    // "you may": asked when the trigger is placed, before its targets; declined, it is not placed (SVY-R9)
-	Keyword    Keyword // the keyword this trigger comes from; zero for a printed ability
+	ActiveZone  Zone
+	Cornered    bool           // "Cornered: ...": fires only if its controller is Cornered when the event happens (FR-KWD-002e, FR-LAY-003)
+	Optional    bool           // "you may": asked when the trigger is placed, before its targets; declined, it is not placed (SVY-R9)
+	FromKeyword KeywordTrigger // the keyword this trigger comes from; zero for a printed ability
 	// Watches makes the trigger fire on the event of each battlefield asset matching Watch
 	// (judged against the trigger's controller and source when the event happens), never
 	// on its own source's event unless the source matches Watch. False, the zero value, is
```

### 9. `pkg/engine/ability/keywords.go` (commit 2, reference only)
```diff
@@ -34,11 +34,11 @@ type keywordDef struct {
 // keywordDefs lists the standalone keywords in printed order. ParseKeywordLine,
 // Keywords.Lines and Keywords.Add all walk it, so each keyword's name is
 // written only here, or, for a keyword that is a triggered ability, in
-// Keyword.String.
+// KeywordTrigger.String.
 var keywordDefs = []keywordDef{
 	{name: "Resistant", num: func(k *Keywords) *int { return &k.Resistant }},
 	{name: "Stalwart", flag: func(k *Keywords) *bool { return &k.Stalwart }},
-	{name: KeywordAggressive.String(), num: func(k *Keywords) *int { return &k.Aggressive }},
+	{name: KeywordTriggerAggressive.String(), num: func(k *Keywords) *int { return &k.Aggressive }},
 	{name: "Ethereal", flag: func(k *Keywords) *bool { return &k.Ethereal }},
 	{name: "Ward", flag: func(k *Keywords) *bool { return &k.Ward }},
 	{name: "Versatile", flag: func(k *Keywords) *bool { return &k.Versatile }},
```

### 10. `pkg/engine/stack.go`: `StackItem.FromKeyword` (commit 2)
```diff
@@ -37,12 +37,12 @@ type StackItem struct {
 	Source int // asset ID: marching unit, attacker, scratching HQ; 0 for a StackManeuver or a trigger from a card in a pile (neither is an asset)
 	Target int // StackAttack: the defender
 
-	Effect  ability.Effect  // StackAbility, StackManeuver: the effect resolving (for StackAbility, Source is its source asset)
-	Chosen  [][]int         // StackAbility, StackManeuver: per Effect.Targets slot, in pick order, the asset IDs chosen, or for a PileSlot the card instance IDs; empty if none
-	Tracked bool            // StackAbility: a death trigger or a trigger from its card's discard pile, so Inst is its card tracked into its owner's discard pile (FR-ZON-004); false for an activated ability
-	Keyword ability.Keyword // StackAbility: the keyword its trigger came from (ability.KeywordAggressive); zero for a printed trigger or an activated ability
-	Watched int             // StackAbility: the asset ID of the asset whose event fired its watching trigger (ability.SubjectWatched); 0 for none
-	Locked  *LockedZone     // StackAbility, StackManeuver: the zone locked for its source-relative filters (SVY-R44, SVY-R45): a move trigger's destination, or the source's zone for an effect that reads "in the same zone"; nil if it locked none
+	Effect      ability.Effect         // StackAbility, StackManeuver: the effect resolving (for StackAbility, Source is its source asset)
+	Chosen      [][]int                // StackAbility, StackManeuver: per Effect.Targets slot, in pick order, the asset IDs chosen, or for a PileSlot the card instance IDs; empty if none
+	Tracked     bool                   // StackAbility: a death trigger or a trigger from its card's discard pile, so Inst is its card tracked into its owner's discard pile (FR-ZON-004); false for an activated ability
+	FromKeyword ability.KeywordTrigger // StackAbility: the keyword its trigger came from (ability.KeywordTriggerAggressive); zero for a printed trigger or an activated ability
+	Watched     int                    // StackAbility: the asset ID of the asset whose event fired its watching trigger (ability.SubjectWatched); 0 for none
+	Locked      *LockedZone            // StackAbility, StackManeuver: the zone locked for its source-relative filters (SVY-R44, SVY-R45): a move trigger's destination, or the source's zone for an effect that reads "in the same zone"; nil if it locked none
 }
 
 // PileSlot reports whether Effect.Targets[i] chooses cards in discard piles, so
@@ -95,8 +95,8 @@ func (g *Game) StackLabel(it *StackItem) string {
 		return "Maneuver " + cardName(it.Inst)
 	case StackAbility:
 		name := g.sourceName(it)
-		if it.Keyword != ability.KeywordNone {
-			return it.Keyword.String() + " trigger of " + name
+		if it.FromKeyword != ability.KeywordTriggerNone {
+			return it.FromKeyword.String() + " trigger of " + name
 		}
 		return "ability of " + name
 	}
```

### 11. `pkg/engine/trigger.go` (commit 2, reference only)
```diff
@@ -204,7 +204,7 @@ func (g *Game) bufferTrigger(src triggerSource, label string, t ability.Trigger)
 func (g *Game) fireAttacks(a *Asset) {
 	src := assetSource(a)
 	if n := g.Keywords(a).Aggressive; n > 0 {
-		g.bufferTrigger(src, fmt.Sprintf("%s: %s %d", src.card.Name, ability.KeywordAggressive.String(), n), ability.Aggressive(n))
+		g.bufferTrigger(src, fmt.Sprintf("%s: %s %d", src.card.Name, ability.KeywordTriggerAggressive.String(), n), ability.Aggressive(n))
 	}
 	g.fireHook(src, ability.HookAttacks, nil)
 	for _, w := range g.assetsByID() {
@@ -276,7 +276,7 @@ func (g *Game) placeTrigger(p pendingTrigger) {
 	item := p.stackItem(StackAbility)
 	item.Effect, item.Chosen = g.chooseOnPlacement(p.placement, p.trig.Effect)
 	item.Watched = p.watched
-	item.Tracked, item.Keyword = p.trig.Hook == ability.HookKilled || p.trig.ActiveZone == ability.ZoneDiscard, p.trig.Keyword
+	item.Tracked, item.FromKeyword = p.trig.Hook == ability.HookKilled || p.trig.ActiveZone == ability.ZoneDiscard, p.trig.FromKeyword
 	g.push(item)
 }
 
```

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The test edits fail before the implementation:** at the `test(T-0283)` commit, `go -C W vet ./pkg/engine/` fails, reporting `FromKeyword undefined` in `aggressive_item_test.go`.
2. **The old names are gone:** `git -C W grep -n -w -e Keyword -e KeywordNone -e KeywordAggressive -- "*.go"` prints exactly three lines, each the words "Keyword interaction matrix" or "Keyword Interaction Matrix": `pkg/toolchain/docs/whatsup.go:159`, `pkg/toolchain/docs/whatsup_test.go:58` and `pkg/toolchain/linter/fences/fences_test.go:56`.
3. **The new names:** `git -C W grep -n "type KeywordTrigger int" -- pkg` prints exactly `pkg/engine/ability/ability.go:214:type KeywordTrigger int`, and `git -C W grep -n -w FromKeyword -- pkg ":!*_test.go"` prints exactly six lines: `ability.go:249` and `:263`, `stack.go:43`, `:98` and `:99`, and `trigger.go:279`.
4. **The same tests pass, with only the authorized edits:** `go -C W test ./pkg/... ./cmd/...` passes, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` prints exactly seven lines: `9	9	pkg/engine/aggressive_item_test.go`, `2	2	pkg/engine/cards_path_allshan_test.go`, `2	2	pkg/engine/cornered_keywords_test.go`, `2	2	pkg/engine/stack_label_test.go`, `1	1	pkg/engine/watch_trigger_test.go`, `1	1	pkg/view/stack_label_test.go` and `1	1	pkg/view/view_test.go`.
5. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` prints `[PASS]`.
6. **Only the listed edits:** `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` prints exactly eleven lines: the seven of case 4 and `16	15	pkg/engine/ability/ability.go`, `2	2	pkg/engine/ability/keywords.go`, `8	8	pkg/engine/stack.go` and `2	2	pkg/engine/trigger.go`.
7. **Nothing else changed:** `git -C W diff autoloop/m2-wave-h.2 --stat` lists only the eleven files of case 6 and this task file.
8. `gofmt -l W/pkg W/cmd` prints nothing; `go -C W tool task check` and `go -C W tool task sim` pass.

## Allowed Files
- `pkg/engine/ability/ability.go` (Edit 8)
- `pkg/engine/ability/keywords.go` (Edit 9, reference only)
- `pkg/engine/stack.go` (Edit 10)
- `pkg/engine/trigger.go` (Edit 11, reference only)

## Tests
Authorized reference-only edits (#171), in the `test(T-0283)` commit; no test is added, renamed or removed:
- `pkg/engine/aggressive_item_test.go` (Edit 1): `TestFR_CBT_002_AggressiveIsListedFirstAndPlacedFirstByDefault`, `TestFR_TRG_003_AggressiveCanBePlacedLast`, `TestFR_CBT_002_AggressiveItemIsAnAbilityItemWithTheKeyword`, `TestFR_KWD_002f_AggressiveConstructor` (also its message), `TestFR_KWD_002f_AggressiveIsReadWhenTheAttackHappens`
- `pkg/engine/cards_path_allshan_test.go` (Edit 2): `TestCard_Allshan_CorneredAggressiveIsTheKeywordTrigger`
- `pkg/engine/cornered_keywords_test.go` (Edit 3): the helper `aggressiveItems` and `TestCard_Allshan_AggressiveComesFromTheKeyword`
- `pkg/engine/stack_label_test.go` (Edit 4): `TestStackLabel_EachKind`, `TestStackLabel_FaceUpSourcePlaceEventNamesItsInstance`
- `pkg/engine/watch_trigger_test.go` (Edit 5): `TestFR_TRG_003_OpposingWatcherResolvesBeforeAttackerTriggers`
- `pkg/view/stack_label_test.go` (Edit 6): `TestView_StackItemLabel`
- `pkg/view/view_test.go` (Edit 7): `TestView_AbilityItemShowsItsKeyword`

## Done When
- [x] Cases 1-8 hold, with their commands and results in the Log
- [x] `go tool task sim-compare -- autoloop/m2-wave-h.2` passes (starter and stress sims byte-identical to the run branch)
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-19, lead-architect (wave H.2 run, session 3): written in full against 6cf38a1 for #171, next in wave H.2's order after #170 (T-0282, `done`). `main` (1859fe4) has nothing the run branch lacks, so no sync. Not `WAVE-COMPLETE` before the task: #172, #173 and #174 are `open: before the next Deliverables task`. #171 is `open`, not `open: before`, but the run file gives wave H.2's order as board order, and the blocking directives #172-#174 each say they come after it, so it goes first. This is a harness task, so it skips `tests` and goes straight to `ready`. Tier is `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list: one per line, no pipes, no `$`, and patterns with special characters are quoted. Waiting directives checked: #6, #7 (mode `serial`), #15 (2) (no draw or end-of-turn hook), #71, #91 (2), #106, #120, #142 (wave H.2 has no cards): none has fired. Next: `kernel-engineer` on Flash.
- 2026-11-19, kernel-engineer (wave H.2 run, session 3): implemented directive #171.
  - Commits:
    - e48477a6 test(T-0283): tests name KeywordTrigger and FromKeyword
    - bf8808a9 refactor(engine): ability.Keyword becomes KeywordTrigger, its fields FromKeyword (T-0283)
  - Changes:
    - Renamed enum type `Keyword` to `KeywordTrigger` and its constants `KeywordNone` / `KeywordAggressive` to `KeywordTriggerNone` / `KeywordTriggerAggressive` in `pkg/engine/ability/ability.go`.
    - Renamed fields `Trigger.Keyword` and `StackItem.Keyword` to `FromKeyword` with type `KeywordTrigger` in `pkg/engine/ability/ability.go` and `pkg/engine/stack.go`.
    - Updated references in `pkg/engine/ability/keywords.go`, `pkg/engine/trigger.go`, and `pkg/engine/stack.go`.
    - Updated reference tests in `pkg/engine/aggressive_item_test.go`, `pkg/engine/cards_path_allshan_test.go`, `pkg/engine/cornered_keywords_test.go`, `pkg/engine/stack_label_test.go`, `pkg/engine/watch_trigger_test.go`, `pkg/view/stack_label_test.go`, and `pkg/view/view_test.go`.
  - Reuse: pure rename and reference updates as prescribed by directive #171; no new shapes added.
  - Verification:
    - Case 1: `go -C W vet ./pkg/engine/` at test commit failed with `aggressive_item_test.go:53:72: (*items)[1].FromKeyword undefined`.
    - Case 2: `git -C W grep -n -w -e Keyword -e KeywordNone -e KeywordAggressive -- "*.go"` printed exactly 3 lines in `pkg/toolchain/` referencing "Keyword interaction matrix".
    - Case 3: `git -C W grep -n "type KeywordTrigger int" -- pkg` printed `pkg/engine/ability/ability.go:214:type KeywordTrigger int`; `git -C W grep -n -w FromKeyword -- pkg ':!*_test.go'` printed exactly the 6 lines (`ability.go:249`, `:263`, `stack.go:43`, `:98`, `:99`, `trigger.go:279`).
    - Case 4: `go -C W test ./pkg/... ./cmd/...` passed, and `git -C W diff autoloop/m2-wave-h.2 --numstat -- "*_test.go"` printed exactly the 7 test files with expected numstats.
    - Case 5: `go -C W tool task sim-compare -- autoloop/m2-wave-h.2` printed `[PASS] starter and stress sims match autoloop/m2-wave-h.2 byte for byte, exit codes included`.
    - Case 6: `git -C W diff autoloop/m2-wave-h.2 --numstat -- pkg cmd docs` printed exactly the 11 modified Go files.
    - Case 7: `git -C W diff autoloop/m2-wave-h.2 --stat` showed only the 11 code/test files plus this task file.
    - Case 8: `gofmt -l W/pkg W/cmd` printed nothing; `go -C W tool task check` and `go -C W tool task sim` passed.
  - Status moved to `review`. Ready for review.
- 2026-11-19, lead-architect (wave H.2 run, session 3, D5): **Approved, `done`.** Merged into `autoloop/m2-wave-h.2` as 2787369. Flash run `7b198333-52ef-4bbe-bdfb-04e8eed1c5c8`: exit 0, `status: SUCCESS`, no `denied_actions` key, and no step with `"status":"ERROR"` in its transcript. `.err` holds only agy's "root agent idle" notice. Review of `git diff autoloop/m2-wave-h.2...m2/autoloop-wave-h.2/T-0283-keyword-trigger` (1b3f48f):
  - Scope: the four Allowed Files, the seven authorized test files and this file. The test commit (e48477a) holds only the seven test files, and the implementation commit (bf8808a) only `ability.go`, `keywords.go`, `stack.go` and `trigger.go`; all eleven blobs match the lead's dry run, whose test-only state failed `go vet ./pkg/engine/` as case 1 says. No test is added, renamed or removed.
  - Fences and reuse: nothing is added. ADR 0007: `git grep -n -w -e Keyword -e KeywordNone -e KeywordAggressive -- "*.go"` finds only the three "Keyword interaction matrix" words in `pkg/toolchain`; no alias, re-export or forwarding wrapper keeps an old name.
  - Re-run by the reviewer on the merge: `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- 4d129e2d` `[PASS]`; `go tool task sim` exit 0, 200 games, P0 58.5% / P1 39.5% / draws 2.0%, P0 50/50 and P1 50/50 main cards executable, no monitor violation.
  - Docs: the design note's Stack paragraph now names the field `FromKeyword` (the lead's file, updated in the `done` commit, as Inputs said). #171 is `applied` (T-0283) and moves to the directive archive.
