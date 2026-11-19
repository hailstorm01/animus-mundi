// Card types: CardType and Card, with the Scratch Company face.

package engine

import (
	"slices"
	"strings"

	"animus-mundi/pkg/engine/ability"
)

// CardType is the mutually exclusive primary classification (FR-CRD-101).
type CardType int

const (
	TypeUnit CardType = iota
	TypeStructure
	TypeManeuver
	TypeHeadquarters
)

func (t CardType) String() string {
	return [...]string{"Unit", "Structure", "Maneuver", "Headquarters"}[t]
}

// Card holds a card's printed characteristics (FR-CRD-101).
type Card struct {
	Name        string
	Type        CardType
	Faction     string
	Subtypes    []string
	ReqCost     int
	TacCost     int
	Power       int
	Health      int
	ReqYield    int
	TacYield    int
	DeployZones []Zone
	Rarity      string
	RulesText   string
	Keywords    ability.Keywords

	// Unsupported lists rules-text lines the kernel cannot execute.
	// A card with unsupported lines cannot be deployed.
	Unsupported []string

	// Ability is the card's executable definition from package cards, if any.
	Ability *ability.Def
}

// HasSubtype reports whether the card carries the subtype (case-insensitive).
func (c *Card) HasSubtype(s string) bool {
	for _, st := range c.Subtypes {
		if strings.EqualFold(st, s) {
			return true
		}
	}
	return false
}

// IsScratchCompany reports whether the card is a Scratch Company (FR-ZON-005).
func (c *Card) IsScratchCompany() bool {
	return c.HasSubtype("Scratch Company")
}

// Deployable reports whether the card is deployed rather than played: only
// Unit and Structure cards are (FR-TRN-003; a Maneuver is played, FR-MNV-001).
func (c *Card) Deployable() bool {
	return c.Type == TypeUnit || c.Type == TypeStructure
}

// Supported reports whether the kernel can fully execute this card.
func (c *Card) Supported() bool {
	if len(c.Unsupported) != 0 {
		return false
	}
	if c.Type == TypeManeuver {
		// A Maneuver is executable only through its Effect (FR-MNV-001).
		return c.Ability != nil && (len(c.Ability.Effect.Steps) > 0 || len(c.Ability.Effect.Modes) > 0)
	}
	return true
}

// scratchFace is the face-down Scratch Company characteristic set (FR-ZON-005).
var scratchFace = &Card{
	Name:     "Scratch Company",
	Type:     TypeUnit,
	Faction:  "NEUTRAL",
	Subtypes: []string{"Infantry", "Scratch Company"},
	Power:    1, Health: 1, ReqYield: 1, TacYield: 1,
	DeployZones: []Zone{ZoneReserves},
	// The face prints Deploys Exhausted (FR-ZON-005, FR-KWD-002l, KIM-S01).
	Keywords: ability.Keywords{DeploysExhausted: true},
}

// ScratchFace returns a copy of the face-down Scratch Company characteristics (FR-ZON-005).
func ScratchFace() Card {
	c := *scratchFace
	c.Subtypes = slices.Clone(c.Subtypes)
	c.DeployZones = slices.Clone(c.DeployZones)
	return c
}
