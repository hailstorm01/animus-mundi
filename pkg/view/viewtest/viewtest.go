// Package viewtest is the hidden-information oracle for leak tests: which card names are public, and which a seat hides, in a game's current state. It is test support, like net/http/httptest, and imports only engine.
package viewtest

import (
	"slices"

	"animus-mundi/pkg/engine"
)

// PublicNames are the card names publicly visible now: both discard piles,
// face-up battlefield assets (HQs included) and the cards on the stack.
func PublicNames(g *engine.Game) map[string]bool {
	pub := map[string]bool{}
	for _, p := range g.Players {
		for _, inst := range p.Discard {
			pub[inst.Card.Name] = true
		}
	}
	for _, a := range g.Assets() {
		if !a.FaceDown {
			pub[a.Inst.Card.Name] = true
		}
	}
	for _, it := range g.Stack {
		if (it.Kind == engine.StackDeploy || it.Kind == engine.StackAbility) && it.Inst != nil {
			pub[it.Inst.Card.Name] = true
		}
	}
	return pub
}

// HiddenNames are the names seat hides now: its hand, its deck and the real
// names of its face-down assets.
func HiddenNames(g *engine.Game, seat engine.PlayerID) map[string]bool {
	hid := map[string]bool{}
	for _, inst := range append(slices.Clone(g.Players[seat].Hand), g.Players[seat].Deck...) {
		hid[inst.Card.Name] = true
	}
	for _, a := range g.Assets() {
		if a.FaceDown && (a.Controller == seat || a.Inst.Owner == seat) {
			hid[a.Inst.Card.Name] = true
		}
	}
	return hid
}
