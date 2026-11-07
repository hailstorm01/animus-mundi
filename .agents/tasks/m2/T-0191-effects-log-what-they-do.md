---
id: T-0191
title: Effects log what they do (directive #129, harness, flash)
kind: harness
tier: flash
wave: D.2
status: in-progress
branch: m2/autoloop-wave-d.2/T-0191-effects-log-what-they-do
depends_on: [T-0190]
claims: [event:EventHeal, event:EventArmor, event:EventPower, event:EventCardMove, engine:stepHeal, engine:stepAddArmor, engine:stepModifyStat, engine:stepRelocate, test:TestEffectEvents_*, test:TestView_HealLineShowsToBothSeats, test:TestReplay_EffectEventsRoundTrip]
cards: []
requirements: [FR-TGT-003]
---

## Goal
Heal, Armor, power changes and cards moved out of a discard pile each log one event per asset or card they act on (#129). The play log then shows what *Macabre Dancers*' Heal 2 did, which the #14 (d) playtest needs. Play doesn't change, so the sims stay byte-identical.

## Inputs
- Board directive #129 (structural review 9, approved by the user, 2026-11-06). ADR-0005 says "only its effects are logged". The lead adds the ADR's addendum when this merges, so that file isn't yours to edit.
- **Event shapes (lead decisions).** All four use `Player: -1`, like `EventMove` and `EventLeave`: no player acts. `EventHeal`, `EventArmor` and `EventPower` carry `Target` and `Amount`, and no `Card`. Their text goes through `a.String()`, which hides a face-down asset. `EventCardMove` carries `Card`, `Inst` and `Zone` (where the card goes).
- **A heal that removes nothing still logs,** with `Amount` 0: #129 asks for one event per asset acted on.
- **A card-move event is logged before the card moves.** A log line may name a card only while it is public, which `TestView_Case4_NoLeak` enforces (T-0079). In the dry run, logging after the move made that test report 13 of its 20 seeds, for example `seat 1's log names "Unceasing Tide", hidden by player 0`. Every card `stepCards` returns comes from a discard pile, which is public. `stepRelocate`'s asset loop already logs first.
- **The guard changes, deliberately.** `takeCard` finds a card in any pile, so it can't guard a line that names the card: a card an earlier step put into a hand would be named and moved again. `DiscardCard` is the existing one lookup of a chosen pile card, and `pileTarget` re-verifies with it (`zones.go:158`). A card that has left is skipped, as `stepAssets` skips an asset that has left (`FR-TGT-003`). No card does this today: `sim-compare` passes. `TestEffectEvents_CardNoLongerInTheDiscardPileLogsNothing` pins it.
- **Out of scope:** deck searches stay as they are (#129), and draws log nothing because their card is secret. `TestEventKinds_ValuesAreTheSchema` has no row for `EventShuffle`, which this task doesn't add.
- **Existing tests that pin an exact event sequence (#129 asks for them by name): none.** With the edits below, every existing test passes unedited. The one existing test edited is `TestEventKinds_ValuesAreTheSchema`, which gains four rows in the `test(T-0191)` commit.
- **Why the monitors still hold:** `endTargetPrompts` needs a Target decision followed by its `place` event, and these events fire only as a step resolves. `sim-compare` compares summaries, and no summary counts events.
- Reuse search (D2.3): no existing kind covers these. `EventMove` and `EventLeave` are assets changing zone, and `EventDiscard` is a card entering a discard pile. `logEvent` and `DiscardCard` are reused, and there is no new non-test helper.
- The lead dry-ran these exact edits in a throwaway worktree of the run branch at 4de2a3f. With only Edit 1, the three `go vet` commands in Case 1 print the `undefined` errors shown there. With all the edits: `gofmt -l pkg` is clean, `go tool task check` and `go tool task sim` pass, and `go tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`. Every expected output below is from that dry run.

## Acceptance Cases (harness tasks only; written by the lead)
W is your worktree. Run each command on its own, pointed at W. Record each command and its result in the Log. The reviewer re-runs them.
1. **The tests fail first.** With only Edit 1 in place, each of these fails to compile: `go -C W vet ./pkg/engine` prints `undefined: EventHeal`, and `go -C W vet ./pkg/view` and `go -C W vet ./pkg/eventlog` each print `undefined: engine.EventHeal`. Then commit Edit 1.
2. **Each kind reads its fields:** `go -C W test ./pkg/engine -run TestEffectEvents_` passes. It covers: a heal's `Amount` is the damage actually removed, 0 included; Armor's is the Armor added; power's is the change, negative included; each face-down target stays hidden; a card move names the card and its destination, while the card is still in the discard pile; and a card no longer there moves nothing and logs nothing.
3. **The schema:** `go -C W test ./pkg/engine -run TestEventKinds_` passes, with the four new rows.
4. **The play log:** `go -C W test ./pkg/view -run TestView_HealLineShowsToBothSeats` passes: both seats see each heal line, and none names a face-down card.
5. **No leak:** `go -C W test ./pkg/view -run TestView_Case4_NoLeak` passes, with that test unedited.
6. **Replay:** `go -C W test ./pkg/eventlog -run TestReplay_EffectEventsRoundTrip -v` passes, and its log shows these counts for the 25 sample games: `heal: 98 events`, `armor: 18 events`, `power: 43 events`, `card-move: 72 events`.
7. **Sims byte-identical:** `go -C W tool task sim-compare -- autoloop/m2-wave-d.2` prints `[PASS]`.
8. **Only the listed edits:** `git -C W diff autoloop/m2-wave-d.2 --numstat -- pkg cmd` prints exactly these six lines:
   ```text
   193	0	pkg/engine/effect_events_test.go
   18	8	pkg/engine/effects.go
   4	0	pkg/engine/eventkinds_test.go
   5	1	pkg/engine/events.go
   38	0	pkg/eventlog/replay_test.go
   38	0	pkg/view/view_test.go
   ```
9. `go -C W tool task check` and `go -C W tool task sim` pass.

## Edits (lead decisions; copy exactly, tabs as in the surrounding code)
Line numbers are at 4de2a3f.

### 1. The tests, in the `test(T-0191)` commit
**`pkg/engine/effect_events_test.go`:** a new file:
```go
package engine

import (
	"fmt"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Tests for T-0191 (directive #129): an effect logs one event per asset or card
// it acts on. Cards are synthetic. Player 0 is the active player.

// effectEvents returns the events of kind logged from Log index from on.
func effectEvents(g *Game, from int, kind string) []Event {
	var out []Event
	for _, e := range g.Log[from:] {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// onAlliedKinds is a step acting on every allied asset of the given kinds.
func onAlliedKinds(op ability.Op, n int, kinds ability.Kinds) ability.Step {
	return ability.Step{Op: op, Amount: ability.Value{N: n}, Subject: ability.SubjectAll,
		Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: kinds}}
}

// effectSource is a unit in the Main Theater whose free ability runs step.
func (s *scene) effectSource(step ability.Step) *Asset {
	return s.put(0, ZoneMainTheater, actUnit("Engine", nil, actAbility(nil, step)))
}

func (s *scene) wall() *Asset {
	return s.put(0, ZoneMainTheater, structure("Wall", 0, 9, Keywords{}))
}

// wantAssetEvents checks one event per asset, in order: its target, its amount
// and its line, with no acting player and no card.
func wantAssetEvents(t *testing.T, got []Event, as []*Asset, amounts []int, format string) {
	t.Helper()
	if len(got) != len(as) {
		t.Fatalf("%d events, want one per asset (%d): %+v", len(got), len(as), got)
	}
	for i, a := range as {
		e := got[i]
		if e.Target != a.ID || e.Amount != amounts[i] || e.Player != -1 || e.Card != "" {
			t.Errorf("event %d: target %d, amount %d, player %d, card %q; want target %d, amount %d, player -1, no card",
				i, e.Target, e.Amount, e.Player, e.Card, a.ID, amounts[i])
		}
		if want := fmt.Sprintf(format, a, amounts[i]); e.Detail != want {
			t.Errorf("event %d: detail %q, want %q", i, e.Detail, want)
		}
	}
}

func TestEffectEvents_HealLogsTheDamageRemoved(t *testing.T) {
	s := newScene(t)
	src := s.effectSource(onAlliedKinds(ability.OpHeal, 2, ability.KindStructure))
	hurt, scratched, whole := s.wall(), s.wall(), s.wall()
	hurt.Damage, scratched.Damage = 3, 1
	from := len(s.g.Log)
	s.do(activate(src, 0))
	wantAssetEvents(t, effectEvents(s.g, from, EventHeal), []*Asset{hurt, scratched, whole}, []int{2, 1, 0}, "%s heals %d")
	if hurt.Damage != 1 || scratched.Damage != 0 || whole.Damage != 0 {
		t.Errorf("damage %d, %d, %d after Heal 2, want 1, 0, 0", hurt.Damage, scratched.Damage, whole.Damage)
	}
}

func TestEffectEvents_ArmorLogsTheArmorAdded(t *testing.T) {
	s := newScene(t)
	src := s.effectSource(onAlliedKinds(ability.OpAddArmor, 2, ability.KindStructure))
	bare, plated := s.wall(), s.wall()
	plated.Armor = 3
	from := len(s.g.Log)
	s.do(activate(src, 0))
	wantAssetEvents(t, effectEvents(s.g, from, EventArmor), []*Asset{bare, plated}, []int{2, 2}, "%s gains %d Armor")
	if bare.Armor != 2 || plated.Armor != 5 {
		t.Errorf("Armor %d, %d, want 2, 5", bare.Armor, plated.Armor)
	}
}

func TestEffectEvents_PowerLogsThePowerChange(t *testing.T) {
	for _, n := range []int{2, -1} {
		t.Run(fmt.Sprintf("%+d", n), func(t *testing.T) {
			s := newScene(t)
			src := s.effectSource(onAlliedKinds(ability.OpModifyStat, n, ability.KindStructure))
			a, b := s.wall(), s.wall()
			from := len(s.g.Log)
			s.do(activate(src, 0))
			wantAssetEvents(t, effectEvents(s.g, from, EventPower), []*Asset{a, b}, []int{n, n}, "%s gets %+d power until end of turn")
			if a.TempPower != n || b.TempPower != n {
				t.Errorf("TempPower %d, %d, want %d", a.TempPower, b.TempPower, n)
			}
		})
	}
}

func TestEffectEvents_FaceDownAssetStaysHidden(t *testing.T) {
	for _, c := range []struct {
		op   ability.Op
		kind string
	}{
		{ability.OpHeal, EventHeal},
		{ability.OpAddArmor, EventArmor},
		{ability.OpModifyStat, EventPower},
	} {
		t.Run(c.kind, func(t *testing.T) {
			s := newScene(t)
			src := s.effectSource(onAlliedKinds(c.op, 1, ability.KindUnit))
			fd := s.putScratch(0)
			from := len(s.g.Log)
			s.do(activate(src, 0))
			found := false
			for _, e := range effectEvents(s.g, from, c.kind) {
				if e.Target != fd.ID {
					continue
				}
				found = true
				if strings.Contains(e.Detail, "Filler") || e.Card != "" || !strings.HasPrefix(e.Detail, fd.String()) {
					t.Errorf("the event names the hidden card: card %q, detail %q", e.Card, e.Detail)
				}
			}
			if !found {
				t.Fatalf("no %s event for the face-down asset", c.kind)
			}
		})
	}
}

func TestEffectEvents_CardMoveNamesTheCardWhileItIsPublic(t *testing.T) {
	pile := ability.Target{Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}}
	for _, c := range []struct {
		op   ability.Op
		zone Zone
		text string
	}{
		{ability.OpReturnToHand, ZoneHand, "Relic returns to its owner's hand"},
		{ability.OpPutOnDeck, ZoneDeck, "Relic is put into its owner's deck"},
	} {
		t.Run(c.zone.String(), func(t *testing.T) {
			s := newScene(t)
			inst := s.inDiscard(0, unit("Relic", 1, 1, Keywords{}))[0]
			a := s.put(0, ZoneMainTheater, actUnit("Recaller", nil, ability.Activated{
				Effect: ability.Effect{Targets: []ability.Target{pile}, Steps: []ability.Step{ptStep(c.op, 0)}},
			}))
			s.script(0, Answer{Picks: []int{0}})
			inPileAtEvent := false
			s.g.obs = &recorder{onEvent: func(e Event) {
				if e.Kind == EventCardMove {
					inPileAtEvent = inDiscardPile(s.g.Players[0], inst)
				}
			}}
			from := len(s.g.Log)
			s.do(activate(a, 0))
			got := effectEvents(s.g, from, EventCardMove)
			if len(got) != 1 {
				t.Fatalf("%d card-move events, want 1: %+v", len(got), got)
			}
			e := got[0]
			if e.Card != "Relic" || e.Inst != inst.ID || e.Zone != c.zone.String() || e.Player != -1 || e.Detail != c.text {
				t.Errorf("event %+v, want card Relic, inst %d, zone %s, player -1, detail %q", e, inst.ID, c.zone, c.text)
			}
			if !inPileAtEvent {
				t.Errorf("the event was logged after the card left the discard pile")
			}
			if inDiscardPile(s.g.Players[0], inst) {
				t.Errorf("the card is still in the discard pile")
			}
		})
	}
}

// A card an earlier step moved out of the discard pile is gone (FR-TGT-003): the
// step doesn't move it again, doesn't count it and logs nothing, so a later move
// never names a card that went into a hand.
func TestEffectEvents_CardNoLongerInTheDiscardPileLogsNothing(t *testing.T) {
	s := newScene(t)
	inst := s.give(0, unit("Relic", 1, 1, Keywords{}))
	r := &resolution{it: &StackItem{Controller: 0}, cards: [][]*Instance{{inst}}}
	from := len(s.g.Log)
	if n := s.g.runStep(r, ptStep(ability.OpPutOnDeck, 0)); n != 0 {
		t.Errorf("the step did %d things, want 0", n)
	}
	if !inHand(s.g.Players[0], inst) {
		t.Errorf("the card left the hand")
	}
	if got := effectEvents(s.g, from, EventCardMove); len(got) != 0 {
		t.Errorf("card-move events %+v, want none", got)
	}
}
```

**`pkg/engine/eventkinds_test.go`:** after line 28, `{EventDiscard, "discard"},`, add:
```go
		{EventHeal, "heal"},
		{EventArmor, "armor"},
		{EventPower, "power"},
		{EventCardMove, "card-move"},
```

**`pkg/view/view_test.go`:** add this at the end of the file, with one blank line before it:
```go
// T-0191 (directive #129): an effect's event is the same log line for both seats,
// and it names no face-down card.
func TestView_HealLineShowsToBothSeats(t *testing.T) {
	s := newScene(t)
	src := testUnit("Healer", 1, 5, engine.Keywords{})
	src.Ability = &ability.Def{Activated: []ability.Activated{{
		Effect: ability.Effect{Steps: []ability.Step{{Op: ability.OpHeal, Amount: ability.Value{N: 2}, Subject: ability.SubjectAll,
			Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit}}}},
	}}}
	healer := s.put(0, engine.ZoneMainTheater, src)
	healer.Damage = 3
	hidden := s.put(0, engine.ZoneReserves, testUnit("Hidden Recruit", 1, 5, engine.Keywords{}))
	hidden.FaceDown, hidden.Damage = true, 1
	if err := s.g.Apply(engine.Action{Kind: engine.ActActivate, Player: 0, Asset: healer.ID, Ability: 0}); err != nil {
		t.Fatal(err)
	}
	want := map[int]string{
		healer.ID: fmt.Sprintf("%s heals 2", healer),
		hidden.ID: fmt.Sprintf("face-down Scratch Company#%d heals 1", hidden.ID),
	}
	lines := [2][]string{view.LogLines(s.g.Log, 0), view.LogLines(s.g.Log, 1)}
	seen := 0
	for i, e := range s.g.Log {
		if e.Kind != engine.EventHeal {
			continue
		}
		seen++
		for seat, l := range lines {
			if !strings.HasSuffix(l[i], want[e.Target]) || strings.Contains(l[i], "Hidden Recruit") {
				t.Errorf("seat %d's line %q, want it to end with %q and name no face-down card", seat, l[i], want[e.Target])
			}
		}
	}
	if seen != 2 {
		t.Errorf("%d heal events, want 2", seen)
	}
}
```

**`pkg/eventlog/replay_test.go`:** add this at the end of the file, with one blank line before it:
```go
// T-0191 (directive #129): a recorded game with an effect's events replays with
// no divergence after the JSON round trip, and the sample holds every kind.
func TestReplay_EffectEventsRoundTrip(t *testing.T) {
	seen := map[string]int{}
	for i, evs := range allGames(t) {
		found := false
		for _, e := range evs {
			switch e.Kind {
			case engine.EventHeal, engine.EventArmor, engine.EventPower:
				if e.Target == 0 {
					t.Errorf("game %d, seq %d: a %s event with no target", i+1, e.Seq, e.Kind)
				}
			case engine.EventCardMove:
				if e.Card == "" || e.Inst == 0 || e.Zone == "" {
					t.Errorf("game %d, seq %d: a card-move event with no card, instance or zone", i+1, e.Seq)
				}
			default:
				continue
			}
			seen[e.Kind]++
			found = true
		}
		if !found {
			continue
		}
		if _, err := eventlog.Replay(evs, len(evs), resolveDeck(t)); err != nil {
			t.Fatalf("game %d: %v", i+1, err)
		}
	}
	for _, kind := range []string{engine.EventHeal, engine.EventArmor, engine.EventPower, engine.EventCardMove} {
		if seen[kind] == 0 {
			t.Errorf("no %s event in the sample games", kind)
		} else {
			t.Logf("%s: %d events", kind, seen[kind])
		}
	}
}
```
Commit only these four files as `test(T-0191): ...`, after Acceptance Case 1. The three packages' tests don't compile until Edit 2, because the kinds don't exist yet. That is this commit's expected failure.

### 2. `pkg/engine/events.go`
1. Replace line 218, the `Amount` field, with:
```go
	Amount int    `json:"amount,omitempty"` // a place event: stack depth; an Activate action event: the ability's index; heal: the damage removed; armor: the Armor added; power: the power change
```
2. After line 260, `EventShuffle ...`, add:
```go
	EventHeal          = "heal"
	EventArmor         = "armor"
	EventPower         = "power"
	EventCardMove      = "card-move" // an effect moved a card out of a discard pile to Zone; deck searches and draws stay secret
```

### 3. `pkg/engine/effects.go`
1. Replace lines 186-194, the whole of `stepHeal` with its comment, with:
```go
// stepHeal removes damage from each asset the step acts on. Its event's Amount is
// the damage actually removed.
func (g *Game) stepHeal(r *resolution, st ability.Step) int {
	n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil)
	as := g.stepAssets(r, st)
	for _, a := range as {
		before := a.Damage
		a.Damage = max(0, a.Damage-n)
		g.logEvent(Event{Kind: EventHeal, Player: -1, Target: a.ID, Amount: before - a.Damage}, "%s heals %d", a, before-a.Damage)
	}
	return len(as)
}
```
2. After line 201, `a.Armor += n`, add:
```go
		g.logEvent(Event{Kind: EventArmor, Player: -1, Target: a.ID, Amount: n}, "%s gains %d Armor", a, n)
```
3. Replace line 234, `a.TempPower += g.value(...)`, with:
```go
		n := g.value(r.it.Controller, r.it.Source, r.it.Inst, st.Amount, nil)
		a.TempPower += n
		g.logEvent(Event{Kind: EventPower, Player: -1, Target: a.ID, Amount: n}, "%s gets %+d power until end of turn", a, n)
```
4. Replace lines 263-288, the whole of `stepRelocate` with its comment, with:
```go
// stepRelocate moves each asset of stepAssets, then each card of stepCards, to
// its owner's hand or deck, and returns how many moved. An asset leaves the
// battlefield with its last-known state (FR-ZON-004); the event carries no
// Card, as a face-down asset's card is private in a hand or deck. A card must
// still be in a discard pile, which is public: its event names it there, before
// the move can hide it.
func (g *Game) stepRelocate(r *resolution, st ability.Step) int {
	to, bottom := relocateTo(st)
	text := "%s returns to its owner's hand"
	if to == ZoneDeck {
		text = "%s is put into its owner's deck"
	}
	n := 0
	for _, a := range g.stepAssets(r, st) {
		g.logEvent(Event{Kind: EventLeave, Player: -1, Target: a.ID, Zone: to.String()}, text, a)
		g.leaveBattlefield(a, g.effectiveState(a))
		g.putCard(a.Inst, to, bottom)
		n++
	}
	for _, inst := range g.stepCards(r, st) {
		if g.DiscardCard(inst.ID) != inst {
			continue
		}
		g.logEvent(Event{Kind: EventCardMove, Player: -1, Card: inst.Card.Name, Inst: inst.ID, Zone: to.String()}, text, inst.Card.Name)
		g.takeCard(inst)
		g.putCard(inst, to, bottom)
		n++
	}
	return n
}
```

No other file changes.

## Allowed Files
- `pkg/engine/effects.go`
- `pkg/engine/events.go`
- In the `test(T-0191)` commit only: `pkg/engine/effect_events_test.go`, `pkg/engine/eventkinds_test.go`, `pkg/view/view_test.go` and `pkg/eventlog/replay_test.go`, for the additions in Edit 1. No other existing test changes.

## Done When
- [ ] Cases 1-9 hold, with their commands and results in the Log
- [ ] `go tool task sim-compare -- autoloop/m2-wave-d.2` passes (starter and stress sims byte-identical to the run branch)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim` runs clean on the starter decks

## Log
- 2026-11-07, lead-architect (wave D.2 run, session 5): wrote this task from board directive #129, next in board order, against the run branch at 4de2a3f. No `main` commits to sync. Not `WAVE-COMPLETE`: #129-#135 are `open: before the next Deliverables task`. Harness, so it skips `tests` and goes straight to `ready`; tier `flash`, because the edits are exact and were dry-run. Every command here is on the headless allow-list, one per line, with no pipes, `$`, double quotes or unquoted parentheses. Waiting directives checked: none has triggered (#6, #7: mode is `serial`; #15: no turn-boundary hook; #71, #91 (2), #106, #118 (b), #120, #128: no such card). Next: `kernel-engineer` on Flash.
