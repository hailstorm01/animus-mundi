// Card types: CardType, Keywords and Card, with the Scratch Company face.

package engine

import (
	"slices"
	"strconv"
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

// Keywords holds the standalone keyword values printed on (or granted to) an
// entity. Numeric keywords use 0 for "absent"; Resistant may be negative.
type Keywords struct {
	Resistant        int
	Stalwart         bool
	Aggressive       int
	Ethereal         bool
	Ward             bool
	Versatile        bool
	Guardian         bool
	Armor            int
	DeploysExhausted bool
}

// keywordDef is one standalone keyword: its printed name and its Keywords
// field. A numeric keyword (num) prints its N, as in "Resistant 2"; a flag
// keyword prints its name alone.
type keywordDef struct {
	name string
	num  func(*Keywords) *int
	flag func(*Keywords) *bool
}

// keywordDefs lists the standalone keywords in printed order. ParseKeywordLine
// and Keywords.Lines both walk it, so each keyword's name is written only here.
var keywordDefs = []keywordDef{
	{name: "Resistant", num: func(k *Keywords) *int { return &k.Resistant }},
	{name: "Stalwart", flag: func(k *Keywords) *bool { return &k.Stalwart }},
	{name: "Aggressive", num: func(k *Keywords) *int { return &k.Aggressive }},
	{name: "Ethereal", flag: func(k *Keywords) *bool { return &k.Ethereal }},
	{name: "Ward", flag: func(k *Keywords) *bool { return &k.Ward }},
	{name: "Versatile", flag: func(k *Keywords) *bool { return &k.Versatile }},
	{name: "Guardian", flag: func(k *Keywords) *bool { return &k.Guardian }},
	{name: "Armor", num: func(k *Keywords) *int { return &k.Armor }},
	{name: "Deploys Exhausted", flag: func(k *Keywords) *bool { return &k.DeploysExhausted }},
}

// Lines lists k's standalone keywords in printed form and order, never nil.
func (k Keywords) Lines() []string {
	out := []string{}
	for _, d := range keywordDefs {
		switch {
		case d.num != nil:
			if n := *d.num(&k); n != 0 {
				out = append(out, d.name+" "+strconv.Itoa(n))
			}
		case *d.flag(&k):
			out = append(out, d.name)
		}
	}
	return out
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
	Keywords    Keywords

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

// Supported reports whether the kernel can fully execute this card.
func (c *Card) Supported() bool {
	return len(c.Unsupported) == 0 && c.Type != TypeManeuver
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
	Keywords: Keywords{DeploysExhausted: true},
}

// ScratchFace returns a copy of the face-down Scratch Company characteristics (FR-ZON-005).
func ScratchFace() Card {
	c := *scratchFace
	c.Subtypes = slices.Clone(c.Subtypes)
	c.DeployZones = slices.Clone(c.DeployZones)
	return c
}
