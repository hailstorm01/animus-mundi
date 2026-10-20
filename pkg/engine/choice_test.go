package engine

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// opts builds n options labelled A, B, C... with Ref 100+i.
func opts(n int) []Option {
	out := make([]Option, n)
	for i := range out {
		out[i] = Option{Label: string(rune('A' + i)), Ref: 100 + i}
	}
	return out
}

// validPicks reports whether picks is a valid answer to d (FR-MOD-001 §5: no
// repeated option unless the card allows it).
func validPicks(d Decision, picks []int) bool {
	if len(picks) < d.Min || len(picks) > d.Max {
		return false
	}
	seen := map[int]bool{}
	for _, p := range picks {
		if p < 0 || p >= len(d.Options) {
			return false
		}
		if seen[p] && !d.Repeats {
			return false
		}
		seen[p] = true
	}
	return true
}

func sorted(in []int) []int {
	out := append([]int{}, in...)
	sort.Ints(out)
	return out
}

func invalidLogs(g *Game) []Event {
	var out []Event
	for _, e := range g.Log {
		if e.Kind == "invalid-answer" {
			out = append(out, e)
		}
	}
	return out
}

// --- FR-MOD-001 §1: single-option fast path ---

func TestFR_MOD_001_FastPathSoleTargetNotAsked(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "target", Options: opts(1), Min: 1, Max: 1})
	if !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("a sole legal target is selected automatically: got %v, want [0]", got)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("a sole legal target must not pause for the player: chooser asked %d times", n)
	}
}

func TestFR_MOD_001_FastPathTakeEverythingNotAsked(t *testing.T) {
	s := newScene(t)
	s.script(0)
	got := s.g.choose(Decision{Player: 0, Kind: DecideMode, Prompt: "modes", Options: opts(3), Min: 3, Max: 3})
	if !reflect.DeepEqual(sorted(got), []int{0, 1, 2}) {
		t.Errorf("when the choice must take every option, all are selected automatically: got %v", got)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only one legal answer exists, so the player is not asked: asked %d times", n)
	}
}

func TestFR_MOD_001_FastPathRepeatsSingleOptionNotAsked(t *testing.T) {
	s := newScene(t)
	s.script(0)
	got := s.g.choose(Decision{Player: 0, Kind: DecideMode, Prompt: "repeat", Options: opts(1), Min: 2, Max: 2, Repeats: true})
	if !reflect.DeepEqual(got, []int{0, 0}) {
		t.Errorf("one option chosen twice where repeats are allowed is the only answer: got %v, want [0 0]", got)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("only one legal answer exists, so the player is not asked: asked %d times", n)
	}
}

func TestFR_MOD_001_MultipleOptionsAskedAndAnswerUsed(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{2}})
	d := Decision{Player: 0, Kind: DecideTarget, Prompt: "pick one", Options: opts(3), Min: 1, Max: 1}
	got := s.g.choose(d)
	if !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("a valid answer to a multi-option choice is used as given: got %v, want [2]", got)
	}
	if a := s.asked(0); len(a) != 1 || a[0].Prompt != "pick one" {
		t.Errorf("a multi-option choice asks the deciding player once with the decision: %+v", a)
	}
}

func TestFR_MOD_001_FastPathNotTakenWhenHidden(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "inspect hand", Options: opts(1), Min: 1, Max: 1, Hidden: true})
	if !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("hidden inspection answer: got %v, want [0]", got)
	}
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("a choice that inspects hidden cards never fast-paths: chooser asked %d times, want 1", n)
	}
}

func TestFR_MOD_001_OptionalAlwaysAsked(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideOptional, Prompt: "you may", Options: opts(1), Min: 0, Max: 1})
	if !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("accepting a \"you may\" effect: got %v, want [0]", got)
	}
	if n := len(s.asked(0)); n != 1 {
		t.Errorf("a \"you may\" choice has two answers (take or decline) and is always asked: asked %d times", n)
	}
}

func TestFR_MOD_001_OptionalDeclineAnswerHonoured(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideOptional, Prompt: "you may", Options: opts(1), Min: 0, Max: 1})
	if len(got) != 0 {
		t.Errorf("a player declining a \"you may\" effect gets no picks: got %v", got)
	}
	if n := len(invalidLogs(s.g)); n != 0 {
		t.Errorf("declining is a valid answer and not logged as invalid: %d events", n)
	}
}

func TestFR_MOD_001_NoLegalAnswerReturnsNothingAndAsksNoOne(t *testing.T) {
	s := newScene(t)
	s.script(0)
	got := s.g.choose(Decision{Player: 0, Kind: DecideTarget, Prompt: "impossible", Options: opts(1), Min: 2, Max: 2})
	if len(got) != 0 {
		t.Errorf("needing two distinct picks from one option is impossible, so no picks: got %v", got)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("with no legal answer nobody is asked: asked %d times", n)
	}
}

// --- FR-MOD-001 §5: distinct modes ---

func TestFR_MOD_001_DistinctModesRepeatedPickIsInvalid(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{1, 1}})
	d := Decision{Player: 0, Kind: DecideMode, Prompt: "choose two", Options: opts(3), Min: 2, Max: 2}
	got := s.g.choose(d)
	if !validPicks(d, got) {
		t.Errorf("\"Choose 2\" requires two different modes; got %v", got)
	}
	if !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("an answer repeating a mode is replaced by the first-listed-modes fallback: got %v, want [0 1]", got)
	}
	if n := len(invalidLogs(s.g)); n != 1 {
		t.Errorf("a repeated mode is an invalid answer and is logged as invalid-answer: %d events, want 1", n)
	}
}

func TestFR_MOD_001_RepeatedModeAllowedWhenCardSaysSo(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{1, 1}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideMode, Prompt: "choose two, may repeat", Options: opts(3), Min: 2, Max: 2, Repeats: true})
	if !reflect.DeepEqual(got, []int{1, 1}) {
		t.Errorf("the same mode may be chosen twice when the card allows it: got %v, want [1 1]", got)
	}
	if n := len(invalidLogs(s.g)); n != 0 {
		t.Errorf("an allowed repeat is not an invalid answer: %d events", n)
	}
}

// --- FR-MOD-001 §6: deterministic fallbacks ---

func TestFR_MOD_001_ModeFallbackOnTimeoutPicksFirstListed(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	got := s.g.choose(Decision{Player: 0, Kind: DecideMode, Prompt: "choose two", Options: opts(4), Min: 2, Max: 2})
	if !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("a timed-out modal choice selects the first listed legal modes: got %v, want [0 1]", got)
	}
	if n := len(invalidLogs(s.g)); n != 0 {
		t.Errorf("a timeout is not logged as an invalid answer: %d events", n)
	}
}

func TestFR_MOD_001_ModeFallbackWithNoChooserPicksFirstListed(t *testing.T) {
	s := newScene(t)
	got := s.g.choose(Decision{Player: 0, Kind: DecideMode, Prompt: "choose two", Options: opts(4), Min: 2, Max: 2})
	if !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("a seat with no chooser always times out; modal choice takes the first listed modes: got %v, want [0 1]", got)
	}
	if n := len(invalidLogs(s.g)); n != 0 {
		t.Errorf("a missing chooser is not logged as an invalid answer: %d events", n)
	}
}

func TestFR_MOD_001_OptionalFallbackOnTimeoutDeclines(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	got := s.g.choose(Decision{Player: 0, Kind: DecideOptional, Prompt: "you may", Options: opts(1), Min: 0, Max: 1})
	if len(got) != 0 {
		t.Errorf("a timed-out \"you may\" effect is declined: got %v", got)
	}
}

func TestFR_MOD_001_OptionalFallbackWithNoChooserDeclines(t *testing.T) {
	s := newScene(t)
	got := s.g.choose(Decision{Player: 0, Kind: DecideOptional, Prompt: "you may", Options: opts(2), Min: 0, Max: 1})
	if len(got) != 0 {
		t.Errorf("with no chooser a \"you may\" effect is declined: got %v", got)
	}
}

func TestFR_MOD_001_TargetFallbackOnTimeoutPicksLegalDistinctTargets(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Timeout: true})
	d := Decision{Player: 0, Kind: DecideTarget, Prompt: "two targets", Options: opts(5), Min: 2, Max: 2}
	got := s.g.choose(d)
	if !validPicks(d, got) || len(got) != 2 {
		t.Errorf("a timed-out mandatory target choice selects 2 distinct legal targets: got %v", got)
	}
}

func TestFR_MOD_001_TargetFallbackWithNoChooserPicksLegalTarget(t *testing.T) {
	s := newScene(t)
	d := Decision{Player: 1, Kind: DecideTarget, Prompt: "one target", Options: opts(4), Min: 1, Max: 1}
	got := s.g.choose(d)
	if !validPicks(d, got) || len(got) != 1 {
		t.Errorf("with no chooser a mandatory target is a random legal target: got %v", got)
	}
}

func TestFR_MOD_001_TargetFallbackDeterministicForSameSeed(t *testing.T) {
	run := func() [][]int {
		s := newScene(t) // same game seed every time
		d := Decision{Player: 0, Kind: DecideTarget, Prompt: "target", Options: opts(6), Min: 2, Max: 2}
		var out [][]int
		for i := 0; i < 10; i++ {
			out = append(out, s.g.choose(d))
		}
		return out
	}
	a, b := run(), run()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("the random target fallback is deterministic for a game seed:\n%v\n%v", a, b)
	}
}

func TestFR_MOD_001_TargetFallbackIsRandomNotFixed(t *testing.T) {
	s := newScene(t)
	d := Decision{Player: 0, Kind: DecideTarget, Prompt: "target", Options: opts(6), Min: 1, Max: 1}
	seen := map[string]bool{}
	for i := 0; i < 60; i++ {
		got := s.g.choose(d)
		if !validPicks(d, got) {
			t.Fatalf("fallback target must be a legal option: got %v", got)
		}
		seen[fmt.Sprint(got)] = true
	}
	if len(seen) < 2 {
		t.Errorf("a mandatory multi-option target fallback is random, but 60 draws gave only %v", seen)
	}
}

func TestFR_MOD_001_InvalidAnswerFallsBackAndIsLogged(t *testing.T) {
	cases := map[string][]int{
		"out of range":  {7},
		"negative":      {-1},
		"too few":       {},
		"too many":      {0, 1},
		"repeated pick": {0, 0},
	}
	for name, picks := range cases {
		t.Run(name, func(t *testing.T) {
			s := newScene(t)
			s.script(1, Answer{Picks: picks})
			d := Decision{Player: 1, Kind: DecideMode, Prompt: "choose wisely", Options: opts(3), Min: 1, Max: 1}
			got := s.g.choose(d)
			if !reflect.DeepEqual(got, []int{0}) {
				t.Errorf("an invalid answer %v takes the fallback (first listed mode): got %v, want [0]", picks, got)
			}
			logs := invalidLogs(s.g)
			if len(logs) != 1 {
				t.Fatalf("an invalid answer appends one invalid-answer event: got %d", len(logs))
			}
			if !strings.Contains(logs[0].Detail, "choose wisely") || !strings.Contains(logs[0].Detail, "1") {
				t.Errorf("invalid-answer detail must name the player and the prompt: %q", logs[0].Detail)
			}
		})
	}
}

func TestFR_MOD_001_InvalidOptionalAnswerDeclines(t *testing.T) {
	s := newScene(t)
	s.script(0, Answer{Picks: []int{0, 1}})
	got := s.g.choose(Decision{Player: 0, Kind: DecideOptional, Prompt: "you may", Options: opts(2), Min: 0, Max: 1})
	if len(got) != 0 {
		t.Errorf("an invalid answer to a \"you may\" effect falls back to declining: got %v", got)
	}
	if n := len(invalidLogs(s.g)); n != 1 {
		t.Errorf("invalid-answer events = %d, want 1", n)
	}
}

// --- FR-TRG-002: the non-active player is asked too ---

func TestFR_TRG_002_InactivePlayerIsAskedOwnChooser(t *testing.T) {
	s := newScene(t)
	if s.g.Active != 0 {
		t.Fatalf("scene should start with player 0 active, got %d", s.g.Active)
	}
	s.script(0)
	s.script(1, Answer{Picks: []int{1}})
	d := Decision{Player: 1, Kind: DecideTarget, Prompt: "opponent picks", Options: opts(3), Min: 1, Max: 1}
	got := s.g.choose(d)
	if !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("the non-active player's own answer is used: got %v, want [1]", got)
	}
	if n := len(s.asked(1)); n != 1 {
		t.Errorf("the non-active deciding player is asked: asked %d times, want 1", n)
	}
	if n := len(s.asked(0)); n != 0 {
		t.Errorf("the active player is not asked another player's choice: asked %d times", n)
	}
}

func TestFR_TRG_002_InactivePlayerFastPathsSoleOption(t *testing.T) {
	s := newScene(t)
	s.script(1)
	got := s.g.choose(Decision{Player: 1, Kind: DecideTarget, Prompt: "only one", Options: opts(1), Min: 1, Max: 1})
	if !reflect.DeepEqual(got, []int{0}) {
		t.Errorf("the non-active player's sole legal option is selected automatically: got %v", got)
	}
	if n := len(s.asked(1)); n != 0 {
		t.Errorf("the non-active player is not prompted for a sole non-inspection option: asked %d times", n)
	}
}

func TestFR_TRG_002_InactivePlayerHiddenInspectionStillAsked(t *testing.T) {
	s := newScene(t)
	s.script(1, Answer{Picks: []int{0}})
	s.g.choose(Decision{Player: 1, Kind: DecideTarget, Prompt: "inspect", Options: opts(1), Min: 1, Max: 1, Hidden: true})
	if n := len(s.asked(1)); n != 1 {
		t.Errorf("a non-active player's hidden inspection is not fast-pathed: asked %d times, want 1", n)
	}
}

// --- RandomChooser (lead's request; ADR-0003) ---

// decisionMix returns a deterministic spread of decisions of every kind,
// including Repeats, whose Options always admit a valid answer.
func decisionMix() []Decision {
	var ds []Decision
	for n := 1; n <= 6; n++ {
		for min := 0; min <= n; min++ {
			for max := min; max <= n; max++ {
				for _, k := range []DecisionKind{DecideTarget, DecideMode, DecideOptional} {
					ds = append(ds, Decision{Player: 0, Kind: k, Prompt: "mix", Options: opts(n), Min: min, Max: max})
				}
			}
		}
		// Repeats: Max may exceed the option count.
		ds = append(ds,
			Decision{Player: 1, Kind: DecideMode, Prompt: "rep", Options: opts(n), Min: 2, Max: 2, Repeats: true},
			Decision{Player: 1, Kind: DecideMode, Prompt: "rep", Options: opts(n), Min: 1, Max: 4, Repeats: true},
			Decision{Player: 1, Kind: DecideTarget, Prompt: "rep", Options: opts(n), Min: 3, Max: 3, Repeats: true},
		)
	}
	ds = append(ds, Decision{Player: 0, Kind: DecideOptional, Prompt: "may", Options: opts(1), Min: 0, Max: 1})
	return ds
}

func TestRandomChooser_AnswersAreAlwaysValid(t *testing.T) {
	s := newScene(t)
	ch := RandomChooser(42)
	for round := 0; round < 5; round++ {
		for _, d := range decisionMix() {
			a := ch(s.g, d)
			if a.Timeout {
				t.Fatalf("a random chooser gives a valid answer, never a timeout: %+v", d)
			}
			if !validPicks(d, a.Picks) {
				t.Fatalf("random answer %v is not valid for kind=%d options=%d min=%d max=%d repeats=%v",
					a.Picks, d.Kind, len(d.Options), d.Min, d.Max, d.Repeats)
			}
		}
	}
}

func TestRandomChooser_SameSeedSameAnswers(t *testing.T) {
	record := func(seed uint64) []Answer {
		s := newScene(t)
		ch := RandomChooser(seed)
		var out []Answer
		for _, d := range decisionMix() {
			out = append(out, ch(s.g, d))
		}
		return out
	}
	a, b := record(7), record(7)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("the same seed must give the same sequence of answers")
	}
}

func TestRandomChooser_NotConstant(t *testing.T) {
	s := newScene(t)
	ch := RandomChooser(3)
	d := Decision{Player: 0, Kind: DecideTarget, Prompt: "spread", Options: opts(6), Min: 1, Max: 1}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		seen[fmt.Sprint(ch(s.g, d).Picks)] = true
	}
	if len(seen) < 3 {
		t.Errorf("a random chooser should spread over the options: 100 draws gave %v", seen)
	}
	// Optional decisions: both taking and declining should occur.
	od := Decision{Player: 0, Kind: DecideOptional, Prompt: "may", Options: opts(1), Min: 0, Max: 1}
	lens := map[int]bool{}
	for i := 0; i < 100; i++ {
		lens[len(ch(s.g, od).Picks)] = true
	}
	if len(lens) < 2 {
		t.Errorf("a random chooser should both take and decline an optional effect over 100 draws: saw %v", lens)
	}
}

func TestRandomChooser_DifferentSeedsDiffer(t *testing.T) {
	draw := func(seed uint64) string {
		s := newScene(t)
		ch := RandomChooser(seed)
		d := Decision{Player: 0, Kind: DecideTarget, Prompt: "spread", Options: opts(8), Min: 1, Max: 1}
		var out []int
		for i := 0; i < 40; i++ {
			out = append(out, ch(s.g, d).Picks...)
		}
		return fmt.Sprint(out)
	}
	if draw(1) == draw(2) {
		t.Errorf("different seeds should give different answer sequences")
	}
}
