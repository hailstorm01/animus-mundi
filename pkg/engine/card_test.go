package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// Fixture lines. A carries an apostrophe so case 6 can swap it.
const (
	lineA = "Return chosen unit to its owner's hand."
	lineB = "Whenever you generate Tactics, draw a card."
	lineM = "Draw two cards."
)

const (
	unitHeader     = `"name","cardType","faction","subtypes","requisitionCost","tacticsCost","deploymentZones","rulesText","power","health","requisitionYield","tacticsYield","rarity","needsRework"`
	maneuverHeader = `"name","cardType","faction","requisitionCost","tacticsCost","rulesText","rarity","needsRework"`
)

// fixtureDir writes a data/cards-style directory: a unit "Test Unit" printing
// Ward then unitLines, a unit "Twin Unit" printing A twice, and a Maneuver
// "Test Maneuver" printing M.
func fixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	units := unitHeader + "\n" +
		`"Test Unit","UNIT","PATH","Infantry","1","0","MAIN_THEATER","Ward` + "\n" + lineA + "\n" + lineB + `","1","1","0","0","COMMON","false"` + "\n" +
		`"Twin Unit","UNIT","PATH","Infantry","1","0","MAIN_THEATER","` + lineA + "\n" + lineA + `","1","1","0","0","COMMON","false"` + "\n"
	maneuvers := maneuverHeader + "\n" +
		`"Test Maneuver","MANEUVER","PATH","0","1","` + lineM + `","COMMON","false"` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "units.csv"), []byte(units), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "maneuvers.csv"), []byte(maneuvers), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func defsFor(name string, text ...string) map[string]*ability.Def {
	return map[string]*ability.Def{name: {Text: text}}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mustLoad(t *testing.T, defs map[string]*ability.Def) map[string]*Card {
	t.Helper()
	cs, err := loadCards(fixtureDir(t), defs)
	if err != nil {
		t.Fatalf("loadCards: %v", err)
	}
	return cs
}

// 1. No defs: Ability nil, Unsupported in printed order, Ward parsed as keyword.
func TestLoadCards_NoDefs(t *testing.T) {
	u := mustLoad(t, nil)["Test Unit"]
	if u.Ability != nil {
		t.Fatalf("Ability = %v, want nil", u.Ability)
	}
	if !equalStrings(u.Unsupported, []string{lineA, lineB}) {
		t.Fatalf("Unsupported = %q", u.Unsupported)
	}
	if !u.Keywords.Ward {
		t.Fatal("Ward should parse as a keyword")
	}
}

// 2. Def covering A: Ability is the same pointer, Unsupported is [B].
func TestLoadCards_PartialCover(t *testing.T) {
	defs := defsFor("Test Unit", lineA)
	u := mustLoad(t, defs)["Test Unit"]
	if u.Ability != defs["Test Unit"] {
		t.Fatal("Ability should be the def pointer")
	}
	if !equalStrings(u.Unsupported, []string{lineB}) {
		t.Fatalf("Unsupported = %q", u.Unsupported)
	}
}

// 3. Def covering A and B: nothing unsupported, card Supported.
func TestLoadCards_FullCover(t *testing.T) {
	u := mustLoad(t, defsFor("Test Unit", lineA, lineB))["Test Unit"]
	if len(u.Unsupported) != 0 || !u.Supported() {
		t.Fatalf("Unsupported = %q, Supported = %v", u.Unsupported, u.Supported())
	}
}

// 4. Fully covered Maneuver stays unsupported.
func TestLoadCards_ManeuverStaysUnsupported(t *testing.T) {
	m := mustLoad(t, defsFor("Test Maneuver", lineM))["Test Maneuver"]
	if len(m.Unsupported) != 0 {
		t.Fatalf("Unsupported = %q, want empty", m.Unsupported)
	}
	if m.Supported() {
		t.Fatal("Maneuver must stay unsupported")
	}
}

func wantLoadError(t *testing.T, defs map[string]*ability.Def, mustName ...string) {
	t.Helper()
	cs, err := loadCards(fixtureDir(t), defs)
	if err == nil {
		t.Fatal("expected an error")
	}
	if cs != nil {
		t.Fatalf("expected nil map, got %d cards", len(cs))
	}
	for _, s := range mustName {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("error %q should mention %q", err, s)
		}
	}
}

// 5. A line that is not on the card.
func TestLoadCards_LineNotOnCard(t *testing.T) {
	bad := "Return chosen unit to its owner's deck."
	wantLoadError(t, defsFor("Test Unit", bad), "Test Unit", bad)
}

// 6. Case or apostrophe differences are errors.
func TestLoadCards_CaseAndApostropheExact(t *testing.T) {
	lower := strings.ToLower(lineA)
	wantLoadError(t, defsFor("Test Unit", lower), "Test Unit", lower)
	curly := strings.ReplaceAll(lineA, "'", "’")
	wantLoadError(t, defsFor("Test Unit", curly), "Test Unit", curly)
}

// 7. A keyword line is never ability text.
func TestLoadCards_KeywordLineRejected(t *testing.T) {
	wantLoadError(t, defsFor("Test Unit", "Ward"), "Test Unit", "Ward")
}

// 8. A line claimed more often than it is printed.
func TestLoadCards_LineClaimedTwice(t *testing.T) {
	wantLoadError(t, defsFor("Test Unit", lineA, lineA), "Test Unit", lineA)
}

// 9. A card printing A twice, def covering it once: one A remains.
func TestLoadCards_DuplicatePrintedLine(t *testing.T) {
	u := mustLoad(t, defsFor("Twin Unit", lineA))["Twin Unit"]
	if !equalStrings(u.Unsupported, []string{lineA}) {
		t.Fatalf("Unsupported = %q", u.Unsupported)
	}
}

// 10. A def keyed by an unknown name is ignored.
func TestLoadCards_UnknownDefIgnored(t *testing.T) {
	cs := mustLoad(t, defsFor("No Such Card", "anything"))
	if len(cs) != 3 {
		t.Fatalf("loaded %d cards, want 3", len(cs))
	}
	if !equalStrings(cs["Test Unit"].Unsupported, []string{lineA, lineB}) {
		t.Fatalf("Unsupported = %q", cs["Test Unit"].Unsupported)
	}
}

// 11. With two failing defs, the alphabetically first card is named.
func TestLoadCards_FirstErrorStable(t *testing.T) {
	defs := map[string]*ability.Def{
		"Twin Unit": {Text: []string{"bogus two"}},
		"Test Unit": {Text: []string{"bogus one"}},
	}
	for i := 0; i < 20; i++ {
		_, err := loadCards(fixtureDir(t), defs)
		if err == nil || !strings.Contains(err.Error(), "Test Unit") || strings.Contains(err.Error(), "Twin Unit") {
			t.Fatalf("error = %v, want one naming Test Unit only", err)
		}
	}
}

// 20. realCard returns the CSV card.
func TestRealCard_ReturnsCSVCard(t *testing.T) {
	c, err := realCard("Chainbreaker Sneak")
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != TypeUnit || c.Faction != "PATH" || !c.Supported() {
		t.Fatalf("unexpected card: %+v", c)
	}
}

// 21. Each call returns an independent shallow copy.
func TestRealCard_ShallowCopy(t *testing.T) {
	a, err := realCard("Chainbreaker Sneak")
	if err != nil {
		t.Fatal(err)
	}
	b, err := realCard("Chainbreaker Sneak")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("realCard returned the same pointer twice")
	}
	orig := b.Power
	a.Power += 10
	if b.Power != orig {
		t.Fatal("changing one copy changed the other")
	}
}

// 22. Unknown names error, naming the card.
func TestRealCard_UnknownName(t *testing.T) {
	c, err := realCard("No Such Card")
	if err == nil || c != nil {
		t.Fatalf("got card=%v err=%v, want error", c, err)
	}
	if !strings.Contains(err.Error(), "No Such Card") {
		t.Fatalf("error %q should name the card", err)
	}
}

// 22 (cont.). s.card returns the real card.
func TestScene_CardHelper(t *testing.T) {
	s := newScene(t)
	if c := s.card("Chainbreaker Sneak"); c.Name != "Chainbreaker Sneak" {
		t.Fatalf("got %q", c.Name)
	}
}

// shapeDef is a def on "Test Unit" that claims lineA and carries the given parts.
func shapeDef(d ability.Def) map[string]*ability.Def {
	d.Text = []string{lineA}
	return map[string]*ability.Def{"Test Unit": &d}
}

func badSplit() ability.Step {
	return ability.Step{Op: ability.OpDealDamage, Subject: ability.SubjectPick, Split: true}
}

// T-0278: a field set where no engine code reads it is a load error.
func TestLoadCards_WatchesOffAttacksRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Triggers: []ability.Trigger{
		{Hook: ability.HookDeploy, Watches: true},
	}}), "Test Unit", "Trigger.Watches")
}

func TestLoadCards_SplitOffAddArmorRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Effect: ability.Effect{Steps: []ability.Step{badSplit()}}}),
		"Test Unit", "Step.Split")
}

func TestLoadCards_SplitWithUpToRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Effect: ability.Effect{Steps: []ability.Step{
		{Op: ability.OpAddArmor, Subject: ability.SubjectPick, UpTo: 2, Split: true},
	}}}), "Test Unit", "Step.Split")
}

func TestLoadCards_SplitOffPickRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Effect: ability.Effect{Steps: []ability.Step{
		{Op: ability.OpAddArmor, Subject: ability.SubjectAll, Split: true},
	}}}), "Test Unit", "Step.Split")
}

func TestLoadCards_SplitOverPileRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Effect: ability.Effect{Steps: []ability.Step{
		{Op: ability.OpAddArmor, Subject: ability.SubjectPick, Split: true, Filter: ability.Filter{Zone: ability.ZoneDiscard}},
	}}}), "Test Unit", "Step.Split")
}

func TestLoadCards_ControllerChoosesOffDiscardRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Effect: ability.Effect{Steps: []ability.Step{
		{Op: ability.OpDraw, Subject: ability.SubjectController, Amount: ability.Value{N: 1}, ControllerChooses: true},
	}}}), "Test Unit", "Step.ControllerChooses")
}

func TestLoadCards_AuraFilterStatRejected(t *testing.T) {
	wantLoadError(t, shapeDef(ability.Def{Statics: []ability.Static{
		{Kind: ability.StaticAura, Filter: ability.Filter{Stat: ability.StatPower, AtMost: 3}, Power: 1},
	}}), "Test Unit", "Static.Filter.Stat")
}

func TestLoadCards_ShapeCheckReachesEveryStep(t *testing.T) {
	bad := ability.Effect{Steps: []ability.Step{badSplit()}}
	rows := []struct {
		name string
		def  ability.Def
	}{
		{"effect steps", ability.Def{Effect: bad}},
		{"effect mode", ability.Def{Effect: ability.Effect{Modes: []ability.Effect{bad}, ChooseN: 1}}},
		{"nested mode", ability.Def{Effect: ability.Effect{Modes: []ability.Effect{{Modes: []ability.Effect{bad}, ChooseN: 1}}, ChooseN: 1}}},
		{"trigger steps", ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookDeploy, Effect: bad}}}},
		{"trigger mode", ability.Def{Triggers: []ability.Trigger{{Hook: ability.HookDeploy, Effect: ability.Effect{Modes: []ability.Effect{bad}, ChooseN: 1}}}}},
		{"activated steps", ability.Def{Activated: []ability.Activated{{Effect: bad}}}},
		{"activated mode", ability.Def{Activated: []ability.Activated{{Effect: ability.Effect{Modes: []ability.Effect{bad}, ChooseN: 1}}}}},
		{"replacement steps", ability.Def{Replacements: []ability.Replacement{{Steps: []ability.Step{badSplit()}}}}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			wantLoadError(t, shapeDef(r.def), "Test Unit", "Step.Split")
		})
	}
}

func TestLoadCards_ShapeReadsWhereReadLoad(t *testing.T) {
	defs := shapeDef(ability.Def{
		Triggers: []ability.Trigger{{Hook: ability.HookAttacks, Watches: true}},
		Effect: ability.Effect{Steps: []ability.Step{
			{Op: ability.OpAddArmor, Subject: ability.SubjectPick, Split: true},
			{Op: ability.OpDiscard, Subject: ability.SubjectController, ControllerChooses: true},
		}},
		Statics: []ability.Static{{Kind: ability.StaticAura, Power: 1}},
	})
	if got := mustLoad(t, defs)["Test Unit"].Ability; got != defs["Test Unit"] {
		t.Fatalf("Ability = %p, want the def pointer %p", got, defs["Test Unit"])
	}
}
