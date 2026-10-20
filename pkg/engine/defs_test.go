package engine

import (
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
	"animus-mundi/pkg/engine/cards"
)

// checkStarterSet and checkDefs are the coverage checkers behind
// TestStarterSet_Executable and TestDefs_MatchRulesText. Later tasks change
// only cards.NotYetExecutable and cards.Defs, never this file.

func checkStarterSet(all map[string]*Card, allow []string) []string {
	return []string{"not implemented"}
}

func checkDefs(all map[string]*Card, defs map[string]*ability.Def) []string {
	return []string{"not implemented"}
}

func synthCard(name, faction string, typ CardType, unsupported ...string) *Card {
	return &Card{Name: name, Faction: faction, Type: typ, Unsupported: unsupported}
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

// 12. The real allowlist matches the real data.
func TestStarterSet_Executable(t *testing.T) {
	all, err := LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkStarterSet(all, cards.NotYetExecutable) {
		t.Error(p)
	}
}

// 13. In-scope unsupported card missing from the allowlist.
func TestCheckStarterSet_UnlistedUnsupported(t *testing.T) {
	all := map[string]*Card{"Alpha": synthCard("Alpha", "PATH", TypeUnit, "text")}
	wantOneProblem(t, checkStarterSet(all, nil), "Alpha", "NotYetExecutable")
}

// 14. Allowlisted card that is now supported.
func TestCheckStarterSet_ListedButSupported(t *testing.T) {
	all := map[string]*Card{"Alpha": synthCard("Alpha", "ANIMUS", TypeUnit)}
	wantOneProblem(t, checkStarterSet(all, []string{"Alpha"}), "Alpha", "remove")
}

// 15. Allowlisted names that are out of scope.
func TestCheckStarterSet_OutOfScopeEntry(t *testing.T) {
	all := map[string]*Card{
		"Rift":   synthCard("Rift", "ANIMUS", TypeHeadquarters, "text"),
		"Hollow": synthCard("Hollow", "SOULLESS", TypeUnit, "text"),
	}
	for _, name := range []string{"Rift", "Hollow", "No Such Card"} {
		wantOneProblem(t, checkStarterSet(all, []string{name}), name)
	}
}

// 16. Duplicate or unsorted allowlists.
func TestCheckStarterSet_DuplicateOrUnsorted(t *testing.T) {
	all := map[string]*Card{
		"Alpha": synthCard("Alpha", "PATH", TypeUnit, "text"),
		"Beta":  synthCard("Beta", "PATH", TypeUnit, "text"),
	}
	wantOneProblem(t, checkStarterSet(all, []string{"Alpha", "Alpha", "Beta"}))
	wantOneProblem(t, checkStarterSet(all, []string{"Beta", "Alpha"}))
	if p := checkStarterSet(all, []string{"Alpha", "Beta"}); len(p) != 0 {
		t.Fatalf("sorted list should pass, got %q", p)
	}
}

// 17. Unsupported cards of other factions need no entry.
func TestCheckStarterSet_OtherFactionIgnored(t *testing.T) {
	all := map[string]*Card{"Hollow": synthCard("Hollow", "SOULLESS", TypeUnit, "text")}
	if p := checkStarterSet(all, nil); len(p) != 0 {
		t.Fatalf("got %q, want none", p)
	}
}

// 18. The real defs match the real data.
func TestDefs_MatchRulesText(t *testing.T) {
	all, err := LoadCards("../../data/cards")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range checkDefs(all, cards.Defs) {
		t.Error(p)
	}
}

// 19. checkDefs on synthetic input.
func TestCheckDefs_Synthetic(t *testing.T) {
	all := map[string]*Card{"Alpha": synthCard("Alpha", "PATH", TypeUnit, "text")}
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Ghost": {Text: []string{"x"}}}), "Ghost")
	wantOneProblem(t, checkDefs(all, map[string]*ability.Def{"Alpha": {}}), "Alpha")
	if p := checkDefs(all, map[string]*ability.Def{"Alpha": {Text: []string{"text"}}}); len(p) != 0 {
		t.Fatalf("valid def should pass, got %q", p)
	}
}
