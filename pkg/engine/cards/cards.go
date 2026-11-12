// Package cards holds the card definitions built from ability shapes
// (ADR-0004) and the allowlist of starter-set cards that cannot run yet.
// It imports only package ability.
package cards

import "animus-mundi/pkg/engine/ability"

// Defs maps a card name to its ability definition.
var Defs = map[string]*ability.Def{}

// NotYetExecutable lists, sorted, every starter-set card the kernel cannot
// execute yet. Card tasks remove names from it; M2 is done when it is empty.
var NotYetExecutable = []string{
	"Ancestral Monument",
	"Animoceps Node",
	"Animoceps Spore Cloud",
	"Animoceps Stalk",
	"Blind Defile",
	"Blindside",
	"Brigand Highlander",
	"Devilhorn Scouting Party",
	"Drifting Concept",
	"Eerie Remnant",
	"Hammerhand Bivouac",
	"Hulking Incongruity",
	"Melodic Pummeler",
	"Mutation Swell",
	"Poisoned Rations",
	"Rhapsody's Hollow",
	"Solarheart Monk",
}
