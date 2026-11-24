// Card loading: the CSV sheets and keyword lines that become Card values (FR-CRD-101).

package engine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/engine/cards"
)

// LoadCards reads every CSV in dir (data/cards) and returns cards by name.
func LoadCards(dir string) (map[string]*Card, error) {
	return loadCards(dir, cards.Defs)
}

// loadCards reads the sheets in dir, then attaches each def to its card and
// removes the lines it covers from Unsupported. Defs are processed in name
// order so the first error is stable; a def whose card is not in dir is
// ignored (fixtures hold subsets; TestDefs_MatchRulesText checks real data).
func loadCards(dir string, defs map[string]*ability.Def) (map[string]*Card, error) {
	loaded, err := loadSheets(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(defs))
	for name := range defs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c, ok := loaded[name]
		if !ok {
			continue
		}
		rest := append([]string(nil), c.Unsupported...)
		for _, line := range defs[name].Text {
			i := indexOf(rest, line)
			if i < 0 {
				return nil, fmt.Errorf("card %q: ability text %q is not an unclaimed non-keyword rules-text line", name, line)
			}
			rest = append(rest[:i], rest[i+1:]...)
		}
		if err := checkShapeReads(defs[name]); err != nil {
			return nil, fmt.Errorf("card %q: %w", name, err)
		}
		c.Ability = defs[name]
		c.Unsupported = rest
	}
	return loaded, nil
}

// checkShapeReads rejects a def that sets a field in a context no engine code
// reads it in, where it would otherwise load and silently do nothing or the
// wrong thing. The rules mirror the readers: Trigger.Watches only for
// HookAttacks (fireHook), Step.Split only for an OpAddArmor SubjectPick over
// assets with UpTo 0 (stepAddArmor), Step.ControllerChooses only for OpDiscard
// (stepDiscard), and an aura's Filter.Stat stays StatNone (Game.Power would
// recurse through auras). A shape task that adds a new reader of one of these
// fields widens the matching rule in the same task.
func checkShapeReads(d *ability.Def) error {
	if err := checkEffectReads(d.Effect); err != nil {
		return err
	}
	for _, t := range d.Triggers {
		if t.Watches && t.Hook != ability.HookAttacks {
			return fmt.Errorf("Trigger.Watches is read only for HookAttacks")
		}
		if err := checkEffectReads(t.Effect); err != nil {
			return err
		}
	}
	for _, a := range d.Activated {
		if err := checkEffectReads(a.Effect); err != nil {
			return err
		}
	}
	for _, r := range d.Replacements {
		if err := checkStepReads(r.Steps); err != nil {
			return err
		}
	}
	for _, s := range d.Statics {
		if s.Kind == ability.StaticAura && s.Filter.Stat != ability.StatNone {
			return fmt.Errorf("Static.Filter.Stat must stay StatNone on an aura (Game.Power would recurse)")
		}
	}
	return nil
}

// checkEffectReads checks the steps of e and of each of its Modes (Parts).
func checkEffectReads(e ability.Effect) error {
	for _, p := range e.Parts() {
		if err := checkStepReads(p.Steps); err != nil {
			return err
		}
	}
	return nil
}

func checkStepReads(steps []ability.Step) error {
	for _, s := range steps {
		if s.Split && (s.Op != ability.OpAddArmor || s.Subject != ability.SubjectPick || s.UpTo != 0 || s.Filter.InPile()) {
			return fmt.Errorf("Step.Split is read only for OpAddArmor over SubjectPick assets with UpTo 0")
		}
		if s.ControllerChooses && s.Op != ability.OpDiscard {
			return fmt.Errorf("Step.ControllerChooses is read only for OpDiscard")
		}
	}
	return nil
}

func indexOf(lines []string, s string) int {
	for i, l := range lines {
		if l == s {
			return i
		}
	}
	return -1
}

// sheetErrors lists every data error one load of the card sheets found
// (FR-DAT-002, FR-DAT-006), in file then row order. Each names the file, the
// row (the header is row 1), the card and the column.
type sheetErrors []string

func (e sheetErrors) Error() string {
	return fmt.Sprintf("card sheets: %d data errors:\n  %s", len(e), strings.Join(e, "\n  "))
}

// loadSheets reads every CSV in dir, in file name order. A duplicate name, a
// numeric cell that is not a whole number, a Unit or Structure with no zone,
// or an unknown cardType is a data error: loading goes on through every sheet
// and returns them all together, and no card is returned.
func loadSheets(dir string) (map[string]*Card, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	cards := make(map[string]*Card)
	seen := make(map[string]string) // card name → "file row N" of its first row
	var problems sheetErrors
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		loaded, found, err := readCards(f, filepath.Base(p), seen)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		problems = append(problems, found...)
		for _, c := range loaded {
			cards[c.Name] = c
		}
	}
	if len(problems) > 0 {
		return nil, problems
	}
	return cards, nil
}

// readCards reads one sheet. file names it in data errors, and seen maps each
// card name already read, in this sheet or an earlier one, to its row.
func readCards(r io.Reader, file string, seen map[string]string) ([]*Card, sheetErrors, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, nil
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

	var out []*Card
	var problems sheetErrors
	for i, row := range rows[1:] {
		name := get(row, "name")
		if name == "" {
			continue
		}
		at := fmt.Sprintf("%s row %d", file, i+2)
		bad := func(column, format string, args ...any) {
			problems = append(problems, fmt.Sprintf("%s (%s): %s: %s", at, name, column, fmt.Sprintf(format, args...)))
		}
		if first, dup := seen[name]; dup {
			bad("name", "the same name as %s (FR-DAT-002)", first)
		} else {
			seen[name] = at
		}
		c := &Card{
			Name:      name,
			Faction:   get(row, "faction"),
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
			bad("cardType", "unknown cardType %q", get(row, "cardType"))
			continue
		}
		num := func(column string) int {
			v := get(row, column)
			n, err := strconv.Atoi(v)
			switch {
			case err == nil:
				return n
			case v == "":
				bad(column, "empty, want a whole number (FR-DAT-006)")
			default:
				bad(column, "%q is not a whole number (FR-DAT-006)", v)
			}
			return 0
		}
		// Costs are numeric on every card; power, health and yields on every
		// card but a Maneuver, which has none (FR-DAT-005).
		c.ReqCost, c.TacCost = num("requisitionCost"), num("tacticsCost")
		if c.Type != TypeManeuver {
			c.Power, c.Health = num("power"), num("health")
			c.ReqYield, c.TacYield = num("requisitionYield"), num("tacticsYield")
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
		if c.Deployable() && len(c.DeployZones) == 0 {
			bad("deploymentZones", "a Unit or Structure needs a deployment zone (FR-DAT-006)")
		}
		for _, line := range strings.Split(strings.ReplaceAll(c.RulesText, "\r", ""), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "--" {
				continue
			}
			if !ability.ParseKeywordLine(line, &c.Keywords) {
				c.Unsupported = append(c.Unsupported, line)
			}
		}
		out = append(out, c)
	}
	return out, problems, nil
}
