package main

import (
	"bytes"
	"strings"
	"testing"

	"animus-mundi/pkg/view"
)

// Tests for T-0236 (directive #150): each battlefield line on the play screen
// shows the asset's resource yields, on both seats' sides.

func TestPlay_FormatAssetShowsYields(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    view.AssetView
		want string
	}{
		{"(a) a face-up unit", view.AssetView{ID: 14, Card: "Chainbreaker Irregular", Power: 2, MaxHealth: 3, ReqYield: 1},
			"Chainbreaker Irregular#14 (2/3, Ready, Yield R1 T0)"},
		{"(b) a face-down asset", view.AssetView{ID: 9, Card: "Scratch Company", FaceDown: true, Exhausted: true, Power: 1, MaxHealth: 1, ReqYield: 1, TacYield: 1},
			"Scratch Company#9 (1/1, Exhausted, Yield R1 T1)"},
		{"(c) a Headquarters", view.AssetView{ID: 1, Card: "Animoceps Rager", Power: 3, MaxHealth: 20, ReqYield: 1, TacYield: 2},
			"Animoceps Rager#1 (3/20, Ready, Yield R1 T2)"},
		{"armor and damage follow the yields", view.AssetView{ID: 5, Card: "Wall", Power: 0, MaxHealth: 6, TacYield: 3, Armor: 1, Damage: 2},
			"Wall#5 (0/6, Ready, Yield R0 T3, Armor 1, Dmg 2)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatAsset(tc.a); got != tc.want {
				t.Errorf("formatAsset = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPlay_PrintScreenShowsYieldsOnBothSides(t *testing.T) {
	v := view.View{
		You: view.Side{
			HQ:       &view.AssetView{ID: 1, Card: "Your HQ", Power: 3, MaxHealth: 20, ReqYield: 1, TacYield: 2},
			Reserves: []view.AssetView{{ID: 2, Card: "Your Unit", Power: 2, MaxHealth: 2, ReqYield: 2}},
		},
		Opponent: view.Side{
			HQ:       &view.AssetView{ID: 3, Card: "Their HQ", Power: 2, MaxHealth: 18, ReqYield: 2, TacYield: 1},
			Reserves: []view.AssetView{{ID: 4, Card: "Their Unit", Power: 1, MaxHealth: 4, TacYield: 1}},
		},
		Theater: []view.AssetView{{ID: 5, Card: "Scratch Company", Controller: 1, FaceDown: true, Power: 1, MaxHealth: 1, ReqYield: 1, TacYield: 1}},
	}
	var buf bytes.Buffer
	printScreen(&buf, v, nil)
	out := buf.String()
	for _, want := range []string{
		"Opponent HQ: Their HQ#3 (2/18, Ready, Yield R2 T1)\n",
		"Opponent Reserves:\n  Their Unit#4 (1/4, Ready, Yield R0 T1)\n",
		"Main Theater:\n  Scratch Company#5 (1/1, Ready, Yield R1 T1) (Player 1)\n",
		"Your HQ: Your HQ#1 (3/20, Ready, Yield R1 T2)\n",
		"Your Reserves:\n  Your Unit#2 (2/2, Ready, Yield R2 T0)\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("screen lacks %q:\n%s", want, out)
		}
	}
}
