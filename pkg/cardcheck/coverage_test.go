package cardcheck

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/engine/cards"
)

func synthCard(name, faction string, typ engine.CardType, unsupported ...string) *engine.Card {
	return &engine.Card{Name: name, Faction: faction, Type: typ, Unsupported: unsupported}
}

func wantOneProblem(t *testing.T, problems []string, parts ...string) {
	t.Helper()
	if len(problems) != 1 {
		t.Fatalf("got %d problems %q, want 1", len(problems), problems)
	}
	for _, p := range parts {
		if !strings.Contains(problems[0], p) {
			t.Fatalf("problem %q should contain %q", problems[0], p)
		}
	}
}

// The real allowlist matches the real data: every card in data/cards that
// cannot run is listed, and nothing else is.
func TestCards_Executable(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkExecutable(all, cards.NotYetExecutable) {
		t.Error(p)
	}
}

// A card that cannot run and is missing from the allowlist.
func TestCheckExecutable_UnlistedUnsupported(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text")}
	wantOneProblem(t, checkExecutable(all, nil), "Alpha", "NotYetExecutable")
}

// An allowlisted card that is now supported.
func TestCheckExecutable_ListedButSupported(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "ANIMUS", engine.TypeUnit)}
	wantOneProblem(t, checkExecutable(all, []string{"Alpha"}), "Alpha", "remove")
}

// An allowlisted name that matches no card.
func TestCheckExecutable_UnknownEntry(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit)}
	wantOneProblem(t, checkExecutable(all, []string{"No Such Card"}), "No Such Card", "matches no card")
}

// Duplicate or unsorted allowlists.
func TestCheckExecutable_DuplicateOrUnsorted(t *testing.T) {
	all := map[string]*engine.Card{
		"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text"),
		"Beta":  synthCard("Beta", "PATH", engine.TypeUnit, "text"),
	}
	wantOneProblem(t, checkExecutable(all, []string{"Alpha", "Alpha", "Beta"}))
	wantOneProblem(t, checkExecutable(all, []string{"Beta", "Alpha"}))
	if p := checkExecutable(all, []string{"Alpha", "Beta"}); len(p) != 0 {
		t.Fatalf("sorted list should pass, got %q", p)
	}
}

// Every faction counts, Headquarters included: there is no starter-set scope.
func TestCheckExecutable_EveryFactionCounts(t *testing.T) {
	for _, c := range []*engine.Card{
		synthCard("Hollow", "SOULLESS", engine.TypeUnit, "text"),
		synthCard("Bastion", "CONCORDIA", engine.TypeManeuver),
		synthCard("Outpost", "PATH", engine.TypeHeadquarters, "text"),
	} {
		all := map[string]*engine.Card{c.Name: c}
		wantOneProblem(t, checkExecutable(all, nil), c.Name)
		if p := checkExecutable(all, []string{c.Name}); len(p) != 0 {
			t.Fatalf("listed %q should pass, got %q", c.Name, p)
		}
	}
}

// The real defs match the real data.
func TestDefs_MatchRulesText(t *testing.T) {
	all, err := engine.LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkDefs(all, cards.Defs) {
		t.Error(p)
	}
}

// checkDefs on synthetic input.
func TestCheckDefs_Synthetic(t *testing.T) {
	all := map[string]*engine.Card{"Alpha": synthCard("Alpha", "PATH", engine.TypeUnit, "text")}
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Ghost": {Text: []string{"x"}}}), "Ghost")
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Alpha": {}}), "Alpha")
	if p := checkDefs(all, map[string]*ability.Def{"Alpha": {Text: []string{"text"}}}); len(p) != 0 {
		t.Fatalf("valid def should pass, got %q", p)
	}
}
