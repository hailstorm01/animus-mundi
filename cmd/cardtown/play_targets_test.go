package main

import (
	"bytes"
	"strings"
	"testing"

	"animus-mundi/pkg/view"
)

// Tests for T-0289: the stack listing shows each item's chosen targets
// (directive #175, acceptance cases (a) to (g)).

// targetsView is a view with Ripper Pack#7 and Devilhorn Captain#12 in the Main
// Theater, a face-down asset #9 in the opponent's Reserves, and stack.
func targetsView(stack ...view.StackView) view.View {
	return view.View{
		Theater: []view.AssetView{
			{ID: 7, Card: "Ripper Pack", Controller: 0, Zone: "Main Theater"},
			{ID: 12, Card: "Devilhorn Captain", Controller: 1, Zone: "Main Theater"},
		},
		Opponent: view.Side{Reserves: []view.AssetView{
			{ID: 9, Card: "Scratch Company", Controller: 1, Zone: "Reserves", FaceDown: true},
		}},
		Stack: stack,
	}
}

// stackSection checks that printScreen's output for v holds want.
func stackSection(t *testing.T, v view.View, want string) {
	t.Helper()
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	if !strings.Contains(buf.String(), want) {
		t.Errorf("stack section:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestPlay_StackLineShowsChosenAsset(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
		Label: "ability of Devilhorn Captain", Chosen: [][]int{{7}}, ChosenCards: [][]string{{}}})
	stackSection(t, v, "Resolution Stack:\n  Ability of Devilhorn Captain (Player 0, Source 5) -> Ripper Pack#7\n")
}

func TestPlay_StackLineSeparatesSlotsAndPicks(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{7}, {12}}, ChosenCards: [][]string{{}, {}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{12, 7}}, ChosenCards: [][]string{{}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Ripper Pack#7; Devilhorn Captain#12\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Devilhorn Captain#12, Ripper Pack#7\n")
}

func TestPlay_StackLineShowsChosenPileCards(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}}, ChosenCards: [][]string{{"Eerie Remnant"}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}, {7}},
			ChosenCards: [][]string{{"Eerie Remnant", "Ripper Pack"}, {}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Eerie Remnant (discard pile)\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Eerie Remnant, Ripper Pack (discard pile); Ripper Pack#7\n")
}

func TestPlay_StackLineLeavesOutGoneTargets(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{40}, {41, 12}}, ChosenCards: [][]string{{}, {}}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{40}, {}}, ChosenCards: [][]string{{}, {}}},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5) -> Devilhorn Captain#12\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n")
}

func TestPlay_StackLineShowsFaceDownTargetAsViewed(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
		Label: "ability of Devilhorn Captain", Chosen: [][]int{{9}}, ChosenCards: [][]string{{}}})
	stackSection(t, v, "Resolution Stack:\n  Ability of Devilhorn Captain (Player 0, Source 5) -> Scratch Company#9\n")
}

func TestPlay_StackLineWithoutTargetsIsUnchanged(t *testing.T) {
	v := targetsView(
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{}, ChosenCards: [][]string{}},
		view.StackView{Kind: "Ability", Controller: 0, Source: 5, Card: "Devilhorn Captain",
			Label: "ability of Devilhorn Captain", Chosen: [][]int{{}}, ChosenCards: [][]string{{}}},
		view.StackView{Kind: "Deploy", Controller: 0, Card: "Ripper Pack", Zone: "Reserves", Label: "Deploy of Ripper Pack"},
		view.StackView{Kind: "March", Controller: 0, Source: 7, Zone: "Main Theater", Label: "March"},
		view.StackView{Kind: "Attack", Controller: 0, Source: 7, Target: 12, Label: "Attack"},
		view.StackView{Kind: "Scratch", Controller: 1, Source: 12, Label: "Scratch"},
	)
	stackSection(t, v, "Resolution Stack:\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n"+
		"  Ability of Devilhorn Captain (Player 0, Source 5)\n"+
		"  Deploy of Ripper Pack to Reserves (Player 0)\n"+
		"  March (Player 0, Source 7)\n"+
		"  Attack (Player 0, Source 7)\n"+
		"  Scratch (Player 1, Source 12)\n")
}

func TestPlay_StackLineShowsManeuverTargets(t *testing.T) {
	v := targetsView(view.StackView{Kind: "Maneuver", Controller: 1, Card: "Alpha Volley",
		Label: "Maneuver Alpha Volley", Chosen: [][]int{{7}, {}}, ChosenCards: [][]string{{}, {"Eerie Remnant"}}})
	stackSection(t, v, "Resolution Stack:\n  Maneuver Alpha Volley (Player 1) -> Ripper Pack#7; Eerie Remnant (discard pile)\n")
}
