package main

import (
	"bufio"
	"bytes"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0237 (directive #151): a single-pick prompt lists identical
// options once, and a pick answers the first option with that text; a prompt
// that may pick more than one lists every option.

func TestPlay_ChooserListsIdenticalOptionsOnce(t *testing.T) {
	keep := "Deploy Chainbreaker Irregular to Reserves; bottom: Boneyard Amalgam, Boneyard Amalgam"
	swap := "Deploy Boneyard Amalgam to Reserves; bottom: Chainbreaker Irregular, Boneyard Amalgam"
	look := []engine.Option{{Label: keep}, {Label: keep}, {Label: swap}, {Label: swap}}
	search := []engine.Option{{Label: "Boneyard Amalgam"}, {Label: "Chainbreaker Irregular"}, {Label: "Boneyard Amalgam"}}
	hand := []engine.Option{{Label: "Boneyard Amalgam"}, {Label: "Boneyard Amalgam"}, {Label: "Chainbreaker Irregular"}}
	for _, tc := range []struct {
		name  string
		d     engine.Decision
		input string
		picks []int
		want  string // the whole output
	}{
		{
			"(a) a look's identical outcomes are listed once",
			engine.Decision{Kind: engine.DecideLook, Prompt: "Look at the top 3 cards of your deck", Options: look, Min: 1, Max: 1},
			"2\n", []int{2},
			"Look at the top 3 cards of your deck\n  1) " + keep + "\n  2) " + swap + "\nChoose option (1-N): ",
		},
		{
			"(a) picking the first listed answers the first option",
			engine.Decision{Kind: engine.DecideLook, Prompt: "Look at the top 3 cards of your deck", Options: look, Min: 1, Max: 1},
			"1\n", []int{0},
			"Look at the top 3 cards of your deck\n  1) " + keep + "\n  2) " + swap + "\nChoose option (1-N): ",
		},
		{
			"(b) a single-pick deck search lists two copies once",
			engine.Decision{Kind: engine.DecideSearch, Prompt: "Search your deck for a card", Options: search, Min: 1, Max: 1},
			"2\n", []int{1},
			"Search your deck for a card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose option (1-N): ",
		},
		{
			"(b) a number past the listed options is refused",
			engine.Decision{Kind: engine.DecideSearch, Prompt: "Search your deck for a card", Options: search, Min: 1, Max: 1},
			"3\n1\n", []int{0},
			"Search your deck for a card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose option (1-N): " +
				"Invalid choice \"3\": must be a number between 1 and 2\nChoose option (1-N): ",
		},
		{
			"(b) an up-to-one discard lists two copies once",
			engine.Decision{Kind: engine.DecideHand, Prompt: "Discard up to 1 card", Options: hand, Min: 0, Max: 1},
			"2\n", []int{2},
			"Discard up to 1 card\n  1) Boneyard Amalgam\n  2) Chainbreaker Irregular\nChoose up to 1 options (space-separated, empty for none): ",
		},
		{
			"(c) discard 2 lists both copies",
			engine.Decision{Kind: engine.DecideHand, Prompt: "Discard 2 cards", Options: hand, Min: 2, Max: 2},
			"1 2\n", []int{0, 1},
			"Discard 2 cards\n  1) Boneyard Amalgam\n  2) Boneyard Amalgam\n  3) Chainbreaker Irregular\nChoose 2-2 options (space-separated): ",
		},
		{
			"(d) a decision with no duplicates is numbered as before",
			engine.Decision{Kind: engine.DecideTarget, Prompt: "Devilhorn Captain: trigger 1: choose target 1", Options: []engine.Option{{Label: "Foe#3"}, {Label: "Ally#4"}, {Label: "Foe#5"}}, Min: 1, Max: 1},
			"3\n", []int{2},
			"Devilhorn Captain: trigger 1: choose target 1\n  1) Foe#3\n  2) Ally#4\n  3) Foe#5\nChoose option (1-N): ",
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
