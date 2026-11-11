---
id: T-0217
title: The play screen names a deploy from the discard pile (wave F harness, flash)
kind: harness
tier: flash
wave: F
status: review
branch: m2/autoloop-wave-f/T-0217-discard-deploy-label
depends_on: [T-0214]
claims: [file:pkg/view/view.go (Inst on the seat's own discard cards), file:cmd/cardtown/play_screen.go (a discard deploy's label; findHandCard becomes findCard over a pile)]
cards: []
requirements: [FR-TRN-003]
---

## Goal
`cardtown play` labels a deploy from the discard pile with its card's name and its origin, so a human can play *Veil Render* (T-0216) from the discard pile. The rest of directive #128's last part, split off T-0215 under the D2 size rule.

## Inputs
- Board directive #128 (structural review 9) cites `play_screen.go:149` among the places that assume a deployed card is in hand. T-0214 (merge d7d9858) added `Action.From` (`pkg/engine/actions.go:47`; `ZoneDiscard` for a deploy permission, the zero value or `ZoneHand` for the hand), and `LegalActions` offers `Action{Kind: ActDeploy, Card: inst.ID, Zone: z, From: ZoneDiscard}` (`actions.go:306`). Today:
  - `actionLabel` (`cmd/cardtown/play_screen.go:117`) looks every Deploy's `Card` up in `v.You.Hand` (`findHandCard`, `:149`), so a discard deploy reads `Deploy: Card#N to Reserves`.
  - `view.side` (`pkg/view/view.go:116`) sets `CardView.Inst` only on the seat's own hand cards; discard cards carry 0 (`:133-135`), so the play screen has no way to name the card.
- **The view's rule:** its package comment says only the seat's own hand cards carry an instance ID, "because an Action names them". An Action now names the seat's own discard cards too, so they carry it as well. The opponent's discard cards keep 0, and no asset or stack item carries one. Nothing hidden is exposed: the seat's own discard pile is public, its cards are ones the seat may act on, and an ID shown there never appears on a face-down asset (the asset view carries asset IDs only). The package comment and the `Inst` field comment are updated to say so.
- **Tests that pin the old rule** (searched `Inst` and `instance ID` in `pkg/view`, `cmd/cardtown` and `pkg/view/viewtest`; no golden file holds a view): four existing tests in `pkg/view/view_test.go` assert that no discard card's instance ID appears in seat 0's own view. Each is narrowed to "nowhere outside the seat's own discard pile", and also checks that the own discard card does carry it: `TestView_Case1_Contents`, `TestView_Case2_FaceDown` (the destroyed face-down card, now in its owner's pile), `TestView_PileTargetShowsCardNamesNotInstanceIDs` and `TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs`. The stack item still shows a pile target by name only (T-0081), and seat 1 still sees no instance ID of seat 0's cards. Each edit is under Tests, in the `test(T-0217)` commit.
- **One lookup:** `findHandCard` becomes `findCard(pile, instID)`, called with `v.You.Hand` for a hand Deploy and an Enact, and with `v.You.Discard` for a discard deploy, so there is one copy of the name lookup and its `Card#N` fallback.
- `FormatEvent` and the view's log print an action event's `Detail` only, and the screen's `Discards:` line prints counts, so neither changes. `cardtown sim` doesn't read the view, so the sims don't change.
- Size: two non-test files.
- **Dry run:** the lead made exactly these edits and tests in a throwaway worktree of the run branch at c7e989a. `gofmt -l pkg cmd` clean; `go tool task check` exit 0; `go tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS]`. With the tests added and the two non-test files as at c7e989a, every new and edited test fails: `TestView_Case1_Contents`, `TestView_Case2_FaceDown`, `TestView_PileTargetShowsCardNamesNotInstanceIDs`, `TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs`, `TestView_OwnDiscardCardsCarryInstanceIDs` and `TestPlay_ActionLabelNamesADiscardDeploy`. Blobs after the edits: `pkg/view/view.go` cd848d8, `cmd/cardtown/play_screen.go` 98f2250, `pkg/view/view_test.go` 9d8ef49, `cmd/cardtown/play_test.go` edfc3d4.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log, for the test commit (each test fails) and for the implementation commit (each passes). The reviewer re-runs them.
1. **The seat's own discard cards carry their instance IDs, the opponent's don't:** `go -C W test ./pkg/view -run TestView_OwnDiscardCardsCarryInstanceIDs -count=1` passes.
2. **No instance ID outside the seat's own discard pile and hand** (the four narrowed tests): `go -C W test ./pkg/view -run TestView_Case -count=1` passes (it runs `TestView_Case1_Contents` and `TestView_Case2_FaceDown` among the `Case` tests), and `go -C W test ./pkg/view -run PileTargetShowsCardNamesNotInstanceIDs -count=1` passes (both pile-target tests).
3. **The label:** `go -C W test ./cmd/cardtown -run TestPlay_ActionLabelNamesADiscardDeploy -count=1` passes.
4. **No other change to the sims:** `go -C W tool task sim-compare -- autoloop/m2-wave-f` prints `[PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included`.
5. **Only the listed files:** `git -C W diff autoloop/m2-wave-f --stat` lists only the four Allowed Files and this task file.
6. `go -C W tool task check` and `go -C W tool task sim` pass.

## Tests (copy exactly; the `test(T-0217)` commit)
Line numbers are at c7e989a. The hunks below are the whole change to the two test files: four authorized edits to existing tests in `pkg/view/view_test.go` (named under Inputs), one new test appended to it, and one new test appended to `cmd/cardtown/play_test.go`. Tabs as in the surrounding code. Make them with your file-editing tool, not with a shell command. Both files' imports already cover everything the hunks use.
```diff
diff --git a/cmd/cardtown/play_test.go b/cmd/cardtown/play_test.go
index 5f688ff..edfc3d4 100644
--- a/cmd/cardtown/play_test.go
+++ b/cmd/cardtown/play_test.go
@@ -417,3 +417,28 @@ func TestPlay_ActionLabelFindsAssetsInEveryZone(t *testing.T) {
 		t.Errorf("actionLabel for Scratch = %q, want %q", gotScratch, wantScratch)
 	}
 }
+
+// T-0217 (directive #128, FR-TRN-003): the play screen labels a deploy from the
+// discard pile with its card's name, from the seat's own discard pile, and its
+// origin. A hand deploy's label is unchanged, and each looks only in its own pile.
+func TestPlay_ActionLabelNamesADiscardDeploy(t *testing.T) {
+	v := view.View{You: view.Side{
+		Hand:    []view.CardView{{Inst: 41, Card: "Alpha Courier"}},
+		Discard: []view.CardView{{Inst: 42, Card: "Veil Render"}},
+	}}
+	for _, tc := range []struct {
+		a    engine.Action
+		want string
+	}{
+		{engine.Action{Kind: engine.ActDeploy, Card: 42, Zone: engine.ZoneReserves, From: engine.ZoneDiscard}, "Deploy: Veil Render from discard pile to Reserves"},
+		{engine.Action{Kind: engine.ActDeploy, Card: 99, Zone: engine.ZoneMainTheater, From: engine.ZoneDiscard}, "Deploy: Card#99 from discard pile to Main Theater"},
+		{engine.Action{Kind: engine.ActDeploy, Card: 41, Zone: engine.ZoneReserves, From: engine.ZoneDiscard}, "Deploy: Card#41 from discard pile to Reserves"},
+		{engine.Action{Kind: engine.ActDeploy, Card: 41, Zone: engine.ZoneReserves}, "Deploy: Alpha Courier to Reserves"},
+		{engine.Action{Kind: engine.ActDeploy, Card: 41, Zone: engine.ZoneReserves, From: engine.ZoneHand}, "Deploy: Alpha Courier to Reserves"},
+		{engine.Action{Kind: engine.ActDeploy, Card: 42, Zone: engine.ZoneReserves}, "Deploy: Card#42 to Reserves"},
+	} {
+		if got := actionLabel(v, tc.a); got != tc.want {
+			t.Errorf("actionLabel(%+v) = %q, want %q", tc.a, got, tc.want)
+		}
+	}
+}
diff --git a/pkg/view/view_test.go b/pkg/view/view_test.go
index 564420c..9d8ef49 100644
--- a/pkg/view/view_test.go
+++ b/pkg/view/view_test.go
@@ -202,16 +202,20 @@ func TestView_Case1_Contents(t *testing.T) {
 		t.Errorf("deck counts = %d/%d, want 4/3", v.You.DeckCount, v.Opponent.DeckCount)
 	}
 
-	// Discard piles: both, with names, no instance IDs.
+	// Discard piles: both, with names; instance IDs on the seat's own only,
+	// because an Action names them (T-0217, FR-TRN-003).
 	if got := cardNames(v.You.Discard); !slices.Equal(got, []string{"Alpha Fallen"}) {
 		t.Errorf("own discard = %v, want [Alpha Fallen]", got)
 	}
 	if got := cardNames(v.Opponent.Discard); !slices.Equal(got, []string{"Beta Fallen"}) {
 		t.Errorf("opponent discard = %v, want [Beta Fallen]", got)
 	}
-	for _, c := range append(slices.Clone(v.You.Discard), v.Opponent.Discard...) {
+	if want := g.Players[0].Discard[0].ID; len(v.You.Discard) != 1 || v.You.Discard[0].Inst != want {
+		t.Errorf("own discard = %+v, want Alpha Fallen with instance ID %d", v.You.Discard, want)
+	}
+	for _, c := range v.Opponent.Discard {
 		if c.Inst != 0 {
-			t.Errorf("discard card %q carries instance ID %d, want 0", c.Card, c.Inst)
+			t.Errorf("opponent's discard card %q carries instance ID %d, want 0", c.Card, c.Inst)
 		}
 	}
 
@@ -368,8 +372,19 @@ func TestView_Case2_FaceDown(t *testing.T) {
 		if got := cardNames(owner.Discard); !slices.Equal(got, []string{"Alpha Hidden Gem"}) {
 			t.Errorf("seat %d sees player 0's discard as %v, want [Alpha Hidden Gem]", seat, got)
 		}
-		if strings.Contains(mustJSON(t, v), "424242") {
-			t.Errorf("seat %d's view contains the destroyed card's instance ID", seat)
+		js := mustJSON(t, v)
+		if seat == 0 {
+			// The owner's own discard card carries its instance ID, because an
+			// Action names it (T-0217, FR-TRN-003); nothing else does.
+			if len(owner.Discard) != 1 || owner.Discard[0].Inst != 424242 {
+				t.Errorf("seat 0's own discard = %+v, want the destroyed card with instance ID 424242", owner.Discard)
+			}
+			rest := v
+			rest.You.Discard = nil
+			js = mustJSON(t, rest)
+		}
+		if strings.Contains(js, "424242") {
+			t.Errorf("seat %d's view contains the destroyed card's instance ID outside its owner's own discard pile", seat)
 		}
 	}
 }
@@ -734,9 +749,19 @@ func TestView_PileTargetShowsCardNamesNotInstanceIDs(t *testing.T) {
 			t.Errorf("seat %d: ChosenCards is empty for the asset slot and only the name of the card still in a pile for the pile slot: %v", seat, it.ChosenCards)
 		}
 		js := mustJSON(t, v)
+		if seat == 0 {
+			// The owner's own discard card carries its instance ID, because an
+			// Action names it (T-0217, FR-TRN-003); nothing else does.
+			if len(v.You.Discard) != 1 || v.You.Discard[0].Inst != inPile.ID {
+				t.Errorf("seat 0: own discard = %+v, want Card In Pile with instance ID %d", v.You.Discard, inPile.ID)
+			}
+			rest := v
+			rest.You.Discard = nil
+			js = mustJSON(t, rest)
+		}
 		for _, id := range []int{inPile.ID, gone.ID} {
 			if regexp.MustCompile(fmt.Sprintf(`\b%d\b`, id)).MatchString(js) {
-				t.Errorf("seat %d: the view carries a card instance ID %d, which is never shown", seat, id)
+				t.Errorf("seat %d: the view carries a card instance ID %d outside its owner's own discard pile", seat, id)
 			}
 		}
 		if strings.Contains(js, "Card Gone From Piles") {
@@ -835,9 +860,19 @@ func TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs(t *testing.T) {
 			t.Errorf("seat %d: ChosenCards is empty for the asset slot and only the name of the card still in a pile for the pile slot: %v", seat, it.ChosenCards)
 		}
 		js := mustJSON(t, v)
+		if seat == 0 {
+			// The owner's own discard card carries its instance ID, because an
+			// Action names it (T-0217, FR-TRN-003); nothing else does.
+			if len(v.You.Discard) != 1 || v.You.Discard[0].Inst != inPile.ID {
+				t.Errorf("seat 0: own discard = %+v, want Card In Pile with instance ID %d", v.You.Discard, inPile.ID)
+			}
+			rest := v
+			rest.You.Discard = nil
+			js = mustJSON(t, rest)
+		}
 		for _, id := range []int{inPile.ID, gone.ID} {
 			if regexp.MustCompile(fmt.Sprintf(`\b%d\b`, id)).MatchString(js) {
-				t.Errorf("seat %d: the view carries a card instance ID %d, which is never shown", seat, id)
+				t.Errorf("seat %d: the view carries a card instance ID %d outside its owner's own discard pile", seat, id)
 			}
 		}
 		if strings.Contains(js, "Card Gone From Piles") {
@@ -1529,3 +1564,28 @@ func TestView_HealLineShowsToBothSeats(t *testing.T) {
 		t.Errorf("%d heal events, want 2", seen)
 	}
 }
+
+// T-0217 (directive #128, FR-TRN-003): each seat's own discard cards carry
+// their instance IDs, in pile order, because a deploy from the discard pile
+// names its card; they carry no cost now. The opponent's discard cards carry
+// none.
+func TestView_OwnDiscardCardsCarryInstanceIDs(t *testing.T) {
+	s := newContentScene(t)
+	for seat := range engine.PlayerID(2) {
+		v := view.PlayerView(s.g, seat)
+		own, opp := s.g.Players[seat].Discard, s.g.Players[1-seat].Discard
+		if len(own) == 0 || len(opp) == 0 || len(v.You.Discard) != len(own) || len(v.Opponent.Discard) != len(opp) {
+			t.Fatalf("seat %d: discards %+v and %+v, want one entry per card in %d and %d", seat, v.You.Discard, v.Opponent.Discard, len(own), len(opp))
+		}
+		for i, c := range v.You.Discard {
+			if c.Inst != own[i].ID || c.ReqNow != 0 || c.TacNow != 0 {
+				t.Errorf("seat %d: own discard card %d = %+v, want instance ID %d and no cost now", seat, i, c, own[i].ID)
+			}
+		}
+		for _, c := range v.Opponent.Discard {
+			if c.Inst != 0 {
+				t.Errorf("seat %d: opponent's discard card %q carries instance ID %d, want 0", seat, c.Card, c.Inst)
+			}
+		}
+	}
+}
```

## Edits (copy exactly; the implementation commit)
Line numbers are at c7e989a. The hunks below are the whole change to the two non-test files, tabs as in the surrounding code. Make them with your file-editing tool, not with a shell command.
```diff
diff --git a/cmd/cardtown/play_screen.go b/cmd/cardtown/play_screen.go
index 761fe15..98f2250 100644
--- a/cmd/cardtown/play_screen.go
+++ b/cmd/cardtown/play_screen.go
@@ -146,8 +146,8 @@ func actionLabel(v view.View, a engine.Action) string {
 		return fmt.Sprintf("Asset#%d", id)
 	}
 
-	findHandCard := func(instID int) string {
-		for _, c := range v.You.Hand {
+	findCard := func(pile []view.CardView, instID int) string {
+		for _, c := range pile {
 			if c.Inst == instID {
 				return c.Card
 			}
@@ -161,9 +161,12 @@ func actionLabel(v view.View, a engine.Action) string {
 	case engine.ActConvert:
 		return fmt.Sprintf("Convert: %d Requisition -> %d Tactics", engine.ConvertReq, engine.ConvertTac)
 	case engine.ActDeploy:
-		return fmt.Sprintf("Deploy: %s to %s", findHandCard(a.Card), a.Zone)
+		if a.From == engine.ZoneDiscard {
+			return fmt.Sprintf("Deploy: %s from discard pile to %s", findCard(v.You.Discard, a.Card), a.Zone)
+		}
+		return fmt.Sprintf("Deploy: %s to %s", findCard(v.You.Hand, a.Card), a.Zone)
 	case engine.ActEnact:
-		return fmt.Sprintf("Enact: %s", findHandCard(a.Card))
+		return fmt.Sprintf("Enact: %s", findCard(v.You.Hand, a.Card))
 	case engine.ActActivate:
 		if as := lookupAsset(a.Asset); as != nil {
 			return fmt.Sprintf("Activate: %s: %s", findAsset(a.Asset), as.Text)
diff --git a/pkg/view/view.go b/pkg/view/view.go
index d1f64a4..cd848d8 100644
--- a/pkg/view/view.go
+++ b/pkg/view/view.go
@@ -7,14 +7,16 @@
 //   - its own hand: card names, printed characteristics, what each costs now
 //     and instance IDs;
 //   - the opponent's hand and both decks: counts only;
-//   - public: both discard piles, HQs, Reserves and resource pools, the Main
-//     Theater, the stack, the turn, the phase, the active player and the outcome.
+//   - public: both discard piles (the seat's own with instance IDs), HQs,
+//     Reserves and resource pools, the Main Theater, the stack, the turn, the
+//     phase, the active player and the outcome.
 //
 // A face-down asset is a Scratch Company with the Scratch face's
 // characteristics to both seats, its owner included (FR-ZON-005,
-// double-blind). No asset and no discard card carries a card instance ID: an
-// instance ID could identify a card seen earlier. Only the seat's own hand
-// cards do, because an Action names them.
+// double-blind). No asset and no card in the opponent's discard pile carries
+// a card instance ID: an instance ID could identify a card seen earlier. Only
+// the seat's own hand and discard cards do, because an Action names them (a
+// deploy from the discard pile names its card, FR-TRN-003).
 package view
 
 import "animus-mundi/pkg/engine"
@@ -45,7 +47,7 @@ type Side struct {
 
 // CardView is a card off the battlefield, with its printed characteristics.
 type CardView struct {
-	Inst       int // only for the seat's own hand cards (an Action names them); 0 elsewhere
+	Inst       int // only for the seat's own hand and discard cards (an Action names them); 0 elsewhere
 	Card       string
 	Type       string
 	ReqCost    int
@@ -131,7 +133,11 @@ func side(g *engine.Game, p *engine.Player, own bool) Side {
 		}
 	}
 	for _, inst := range p.Discard {
-		s.Discard = append(s.Discard, cardView(inst.Card))
+		c := cardView(inst.Card)
+		if own {
+			c.Inst = inst.ID
+		}
+		s.Discard = append(s.Discard, c)
 	}
 	return s
 }
```

## Allowed Files
- `pkg/view/view.go`
- `cmd/cardtown/play_screen.go`
- `pkg/view/view_test.go` (the four named test edits and one new test, exactly as under Tests)
- `cmd/cardtown/play_test.go` (one new test appended only; no existing test changes)

## Done When
- [x] Cases 1-6 hold, with their commands and results in the Log, each test failing at the test commit
- [x] `go tool task sim-compare -- autoloop/m2-wave-f` passes, byte for byte
- [x] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [x] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-11 lead-architect (wave F run, session 16): queued with T-0215, whose original Covers needed five non-test files (D2.4). T-0216 depends on it. Directive #128 is `open: before T-0216`, and this task applies its last piece.
- 2026-11-11 lead-architect (wave F run, session 17): wrote the Inputs, Acceptance Cases, Tests and Edits against the run branch at c7e989a, and dry-ran them (Inputs). The Covers' check "no golden output or test pins a discard `CardView`'s `Inst` at 0" found four tests that do; each is narrowed, not dropped, and named under Inputs and Tests. The view's rule stays "an instance ID only where an Action names it", now the seat's own hand and discard cards. Harness, so it skips `tests` and goes to `ready`; tier `flash`, because the edits and tests are exact and dry-run. Every command here is on the headless allow-list, one per line. Next: `kernel-engineer` on Flash.
- 2026-11-11 kernel-engineer (wave F run, Flash): implemented T-0217.
  - Test commit 27c424c (`test(T-0217): own discard cards carry instance IDs and play screen discard deploy label`): narrowed four tests in `pkg/view/view_test.go` (`TestView_Case1_Contents`, `TestView_Case2_FaceDown`, `TestView_PileTargetShowsCardNamesNotInstanceIDs`, `TestView_ManeuverPileTargetShowsCardNamesNotInstanceIDs`), added `TestView_OwnDiscardCardsCarryInstanceIDs`, and added `TestPlay_ActionLabelNamesADiscardDeploy` in `cmd/cardtown/play_test.go`.
    - `go test ./pkg/view -run TestView_OwnDiscardCardsCarryInstanceIDs -count=1` failed (Inst 0 != ID).
    - `go test ./pkg/view -run TestView_Case -count=1` failed (`TestView_Case1_Contents` and `TestView_Case2_FaceDown` failed on Inst 0).
    - `go test ./pkg/view -run PileTargetShowsCardNamesNotInstanceIDs -count=1` failed (both pile-target tests failed on Inst 0).
    - `go test ./cmd/cardtown -run TestPlay_ActionLabelNamesADiscardDeploy -count=1` failed (actionLabel printed `Card#42 to Reserves` instead of discard origin label).
  - Implementation commit 7622bc5 (`feat(view): the play screen names a deploy from the discard pile (T-0217)`):
    - `pkg/view/view.go`: updated package doc and `CardView.Inst` comment; in `side()`, populated `c.Inst = inst.ID` for cards in `p.Discard` when `own` is true.
    - `cmd/cardtown/play_screen.go`: generalized `findHandCard` to `findCard(pile []view.CardView, instID int)`, used `findCard(v.You.Discard, a.Card)` for discard deploys with `"Deploy: %s from discard pile to %s"`, and reused `findCard(v.You.Hand, ...)` for hand deploy and enact.
  - Acceptance cases re-run and passed:
    - Case 1: `go test ./pkg/view -run TestView_OwnDiscardCardsCarryInstanceIDs -count=1` -> PASS (ok 0.644s)
    - Case 2: `go test ./pkg/view -run TestView_Case -count=1` -> PASS (ok 2.552s); `go test ./pkg/view -run PileTargetShowsCardNamesNotInstanceIDs -count=1` -> PASS (ok 0.651s)
    - Case 3: `go test ./cmd/cardtown -run TestPlay_ActionLabelNamesADiscardDeploy -count=1` -> PASS (ok 0.574s)
    - Case 4: `go tool task sim-compare -- autoloop/m2-wave-f` -> `[PASS] starter and stress sims match autoloop/m2-wave-f byte for byte, exit codes included`
    - Case 5: `git diff autoloop/m2-wave-f --stat` -> only 4 allowed files and task file modified
    - Case 6: `go tool task check` -> PASS (gofmt, go vet, all tests, cardtown lint --all); `go tool task sim` -> PASS (starter sim 200 games clean)
