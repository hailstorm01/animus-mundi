package engine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	Berserk          int
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

	// Unsupported lists rules-text lines the M1 kernel cannot execute.
	// A card with unsupported lines cannot be deployed.
	Unsupported []string
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
}

// Numeric keywords always print N (KIM-B04 ruling); a bare "Berserk" is not
// a keyword and loads as unsupported.
var keywordLine = regexp.MustCompile(`^(Resistant|Aggressive|Armor|Berserk) (-?\d+)$`)

// ParseKeywordLine parses one standalone keyword line from rules text.
func ParseKeywordLine(line string, kw *Keywords) bool {
	if m := keywordLine.FindStringSubmatch(line); m != nil {
		n, _ := strconv.Atoi(m[2])
		switch m[1] {
		case "Resistant":
			kw.Resistant += n
		case "Aggressive":
			kw.Aggressive += n
		case "Armor":
			kw.Armor += n
		case "Berserk":
			kw.Berserk += n
		}
		return true
	}
	switch line {
	case "Stalwart":
		kw.Stalwart = true
	case "Ethereal":
		kw.Ethereal = true
	case "Ward":
		kw.Ward = true
	case "Versatile":
		kw.Versatile = true
	case "Guardian":
		kw.Guardian = true
	case "Deploys Exhausted":
		kw.DeploysExhausted = true
	default:
		return false
	}
	return true
}

// LoadCards reads every CSV in dir (data/cards) and returns cards by name.
func LoadCards(dir string) (map[string]*Card, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	cards := make(map[string]*Card)
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		loaded, err := readCards(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		for _, c := range loaded {
			cards[c.Name] = c
		}
	}
	return cards, nil
}

func readCards(r io.Reader) ([]*Card, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	col := make(map[string]int)
	for i, h := range rows[0] {
		col[strings.TrimPrefix(strings.TrimSpace(h), string(rune(0xFEFF)))] = i
	}
	get := func(row []string, name string) string {
		if i, ok := col[name]; ok && i < len(row) {
			v := strings.TrimSpace(row[i])
			if v == "--" {
				return ""
			}
			return v
		}
		return ""
	}
	num := func(row []string, name string) int {
		n, _ := strconv.Atoi(get(row, name))
		return n
	}

	var out []*Card
	for _, row := range rows[1:] {
		name := get(row, "name")
		if name == "" {
			continue
		}
		c := &Card{
			Name:      name,
			Faction:   get(row, "faction"),
			ReqCost:   num(row, "requisitionCost"),
			TacCost:   num(row, "tacticsCost"),
			Power:     num(row, "power"),
			Health:    num(row, "health"),
			ReqYield:  num(row, "requisitionYield"),
			TacYield:  num(row, "tacticsYield"),
			Rarity:    get(row, "rarity"),
			RulesText: get(row, "rulesText"),
		}
		switch get(row, "cardType") {
		case "UNIT":
			c.Type = TypeUnit
		case "STRUCTURE":
			c.Type = TypeStructure
		case "MANEUVER":
			c.Type = TypeManeuver
		case "HEADQUARTERS":
			c.Type = TypeHeadquarters
		default:
			return nil, fmt.Errorf("card %q: unknown cardType %q", name, get(row, "cardType"))
		}
		for _, s := range strings.Split(get(row, "subtypes"), ",") {
			if s = strings.TrimSpace(s); s != "" {
				c.Subtypes = append(c.Subtypes, s)
			}
		}
		for _, z := range strings.Split(get(row, "deploymentZones"), ",") {
			switch strings.TrimSpace(z) {
			case "RESERVES":
				c.DeployZones = append(c.DeployZones, ZoneReserves)
			case "MAIN_THEATER":
				c.DeployZones = append(c.DeployZones, ZoneMainTheater)
			case "HEADQUARTERS":
				c.DeployZones = append(c.DeployZones, ZoneHeadquarters)
			}
		}
		for _, line := range strings.Split(strings.ReplaceAll(c.RulesText, "\r", ""), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "--" {
				continue
			}
			if !ParseKeywordLine(line, &c.Keywords) {
				c.Unsupported = append(c.Unsupported, line)
			}
		}
		out = append(out, c)
	}
	return out, nil
}
