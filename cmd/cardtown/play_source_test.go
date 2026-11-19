package main

import (
	"bufio"
	"bytes"
	"slices"
	"strconv"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

// Tests for T-0235 (directive #149 (c)): the play screen prints a decision's
// source card and its rules text under the prompt, and the stack listing names
// each item's card.

func TestPlay_ChooserPrintsTheSourceCard(t *testing.T) {
	captain := &engine.Card{Name: "Devilhorn Captain", RulesText: "Unbowed: Deal 1 damage to chosen asset in the Main Theater."}
	volley := &engine.Card{Name: "Disorienting Pulse", RulesText: "Return chosen opposing unit in the Main Theater to its owner's hand."}
	ambush := &engine.Card{Name: "Ambush Tactics", RulesText: "Choose two:\n- Kill chosen unit.\n- Return chosen unit to its owner's hand."}
	apostle := &engine.Card{Name: "Infested Apostle", RulesText: "When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."}
	remnant := &engine.Card{Name: "Eerie Remnant", RulesText: "Deploys Exhausted\nWhen this card is put into your discard pile from your hand, you may deploy it from your discard pile for free."}
	vanilla := &engine.Card{Name: "Plain Unit"}
	two := []engine.Option{{Label: "Foe#3"}, {Label: "Ally#4"}}
	for _, tc := range []struct {
		name  string
		d     engine.Decision
		input string
		picks []int
		want  string // the whole output
	}{
		{
			"(a) a trigger's target",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Devilhorn Captain: trigger 1: choose target 1", Options: two, Min: 1, Max: 1, Source: captain},
			"2\n", []int{1},
			"Devilhorn Captain: trigger 1: choose target 1\n  [Devilhorn Captain] Unbowed: Deal 1 damage to chosen asset in the Main Theater.\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(b) a Maneuver's target",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Disorienting Pulse: choose target 1", Options: two, Min: 1, Max: 1, Source: volley},
			"1\n", []int{0},
			"Disorienting Pulse: choose target 1\n  [Disorienting Pulse] Return chosen opposing unit in the Main Theater to its owner's hand.\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(b) a mode choice, text of several lines",
			engine.Decision{Kind: engine.DecideMode, Prompt: "Ambush Tactics: choose 2 modes", Options: []engine.Option{{Label: "mode 1"}, {Label: "mode 2"}}, Min: 2, Max: 2, Source: ambush},
			"1 2\n", []int{0, 1},
			"Ambush Tactics: choose 2 modes\n  [Ambush Tactics] Choose two:\n    - Kill chosen unit.\n    - Return chosen unit to its owner's hand.\n  1) mode 1\n  2) mode 2\nChoose 2-2 options (space-separated): ",
		},
		{
			"(b) an optional trigger's you may",
			engine.Decision{Kind: engine.DecideOptional, Prompt: "Infested Apostle: trigger 1: you may", Min: 0, Max: 1, Source: apostle},
			"y\n", []int{0},
			"Infested Apostle: trigger 1: you may\n  [Infested Apostle] When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit.\nChoose [y/n]: ",
		},
		{
			"(c) a trigger from a discard pile",
			engine.Decision{Kind: engine.DecideOptional, Prompt: "Eerie Remnant: trigger 1: you may", Min: 0, Max: 1, Source: remnant},
			"n\n", []int{},
			"Eerie Remnant: trigger 1: you may\n  [Eerie Remnant] Deploys Exhausted\n    When this card is put into your discard pile from your hand, you may deploy it from your discard pile for free.\nChoose [y/n]: ",
		},
		{
			"a source without rules text",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Plain Unit: choose target 1", Options: two, Min: 1, Max: 1, Source: vanilla},
			"1\n", []int{0},
			"Plain Unit: choose target 1\n  [Plain Unit]\n  1) Foe#3\n  2) Ally#4\nChoose option (1-N): ",
		},
		{
			"(d) the trigger order prints as before",
			engine.Decision{Kind: engine.DecideOrder, Prompt: "order your triggers", Options: []engine.Option{{Label: "a"}, {Label: "b"}}, Min: 2, Max: 2},
			"2 1\n", []int{1, 0},
			"order your triggers\n  1) a\n  2) b\nOrder all 2 options (space-separated): ",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			a := terminalChooser(bufio.NewScanner(strings.NewReader(tc.input)), &out)(nil, tc.d)
			if !slices.Equal(a.Picks, tc.picks) {
				t.Errorf("picks %v, want %v", a.Picks, tc.picks)
			}
			if got := out.String(); got != tc.want {
				t.Errorf("output:\n%q\nwant:\n%q", got, tc.want)
			}
		})
	}
}

// A scripted session prints, under the prompt of a real decision with a source,
// that card's line (#149 (a)-(c)): the prompt names the card, and the line
// under it starts with the same name in brackets.
func TestPlay_ScriptedSessionPrintsSourceLines(t *testing.T) {
	// The first seed whose session reaches such a prompt is enough; which seed
	// does depends on the random answers, so try several.
	for seed := 1; seed <= 8; seed++ {
		args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", strconv.Itoa(seed)}
		var out bytes.Buffer
		if code := runPlay(args, &scriptedReader{out: &out}, &out); code != 0 {
			t.Fatalf("seed %d: runPlay exited with code %d", seed, code)
		}
		lines := strings.Split(out.String(), "\n")
		sources := 0
		for i := 1; i < len(lines); i++ {
			rest, ok := strings.CutPrefix(lines[i], "  [")
			name, _, closed := strings.Cut(rest, "]")
			if ok && closed && strings.Contains(lines[i-1], name+": ") {
				sources++
			}
		}
		if sources > 0 {
			return
		}
	}
	t.Errorf("no prompt is followed by its source card's line in any session of seeds 1 to 8")
}

func TestPlay_PrintScreenNamesEachStackItemsCard(t *testing.T) {
	v := view.View{Stack: []view.StackView{
		{Kind: "Ability", Controller: 1, Source: 12, Card: "Devilhorn Captain", Label: "ability of Devilhorn Captain"},
		{Kind: "Ability", Controller: 0, Source: 7, Card: "Ripper Pack", Label: "Aggressive trigger of Ripper Pack"},
		{Kind: "Ability", Controller: 0, Source: 5},
		{Kind: "Deploy", Controller: 0, Card: "Chainbreaker Irregular", Zone: "Reserves"},
		{Kind: "Maneuver", Controller: 0, Card: "Disorienting Pulse"},
		{Kind: "March", Controller: 1, Source: 9, Zone: "Reserves"},
		{Kind: "Attack", Controller: 1, Source: 7, Target: 3},
		{Kind: "Scratch", Controller: 0, Source: 1},
	}}
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	want := "Resolution Stack:\n" +
		"  Ability of Devilhorn Captain (Player 1, Source 12)\n" +
		"  Aggressive trigger of Ripper Pack (Player 0, Source 7)\n" +
		"  Ability (Player 0, Source 5)\n" +
		"  Deploy of Chainbreaker Irregular to Reserves (Player 0)\n" +
		"  Maneuver: Disorienting Pulse (Player 0)\n" +
		"  March (Player 1, Source 9)\n" +
		"  Attack (Player 1, Source 7)\n" +
		"  Scratch (Player 0, Source 1)\n"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("stack section:\n%s\nwant:\n%s", buf.String(), want)
	}
}
