package sim

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"animus-mundi/pkg/engine"
)

// T-0297: the sim matrix (M3 design note, Deliverable 3).

func TestMatrixRuns_OrderAndCount(t *testing.T) {
	runs := MatrixRuns([]string{"PATH", "ANIMUS", "SOULLESS", "PATH"})
	var got []string
	for _, r := range runs {
		got = append(got, r.String())
	}
	want := []string{
		"ANIMUS v PATH, progress", "ANIMUS v PATH, random",
		"ANIMUS v SOULLESS, progress", "ANIMUS v SOULLESS, random",
		"PATH v ANIMUS, progress", "PATH v ANIMUS, random",
		"PATH v SOULLESS, progress", "PATH v SOULLESS, random",
		"SOULLESS v ANIMUS, progress", "SOULLESS v ANIMUS, random",
		"SOULLESS v PATH, progress", "SOULLESS v PATH, random",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MatrixRuns = %q, want %q", got, want)
	}
}

func TestMatrixRuns_FewerThanTwoFactions(t *testing.T) {
	for _, fs := range [][]string{nil, {"PATH"}, {"PATH", "PATH"}} {
		if runs := MatrixRuns(fs); len(runs) != 0 {
			t.Errorf("MatrixRuns(%q) = %v, want none", fs, runs)
		}
	}
}

func TestDeckFactions_FindsDecklistsByName(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "decks")
	if err := os.MkdirAll(filepath.Join(dir, "x-starter.txt"), 0o755); err != nil { // a folder, not a decklist
		t.Fatal(err)
	}
	for _, name := range []string{"path-starter.txt", "animus-starter.txt", "path-other.txt", "notes.md", "-starter.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := DeckFactions(root, "starter")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ANIMUS", "PATH"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("DeckFactions = %q, want %q", got, want)
	}
}

func TestDeckFactions_RealStarterDecks(t *testing.T) {
	got, err := DeckFactions(filepath.Join("..", ".."), "starter")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ANIMUS", "CONCORDIA", "PATH", "SOULLESS"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("DeckFactions = %q, want %q", got, want)
	}
}

func TestStressFactions_EveryHeadquartersFaction(t *testing.T) {
	cards, err := engine.LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := StressFactions(cards), []string{"ANIMUS", "CONCORDIA", "PATH", "SOULLESS"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("StressFactions = %q, want %q", got, want)
	}
}

func TestMatrixCoverage_SumsChecksAndCountsRuns(t *testing.T) {
	m := NewMatrixCoverage()
	m.Add(Coverage{"mon": {"A": 3, "B": 0, "C": 0}})
	m.Add(Coverage{"mon": {"A": 2, "B": 5, "C": 0}})
	if want := (Coverage{"mon": {"A": 5, "B": 5, "C": 0}}); !reflect.DeepEqual(m.Checks, want) {
		t.Errorf("Checks = %v, want %v", m.Checks, want)
	}
	if want := (Coverage{"mon": {"A": 2, "B": 1, "C": 0}}); !reflect.DeepEqual(m.Runs, want) {
		t.Errorf("Runs = %v, want %v", m.Runs, want)
	}
}

// A matrix run is the same playout as cardtown sim with the same flags, so its
// replay command reproduces it.
func TestSimulateRun_MatchesAPlainSim(t *testing.T) {
	root := filepath.Join("..", "..")
	cards, err := engine.LoadCards(filepath.Join(root, "data", "cards"))
	if err != nil {
		t.Fatal(err)
	}
	for _, deck := range []string{"starter", ""} {
		run := MatrixRun{Factions: [2]string{"PATH", "SOULLESS"}, Policy: "random"}
		got, err := SimulateRun(root, cards, deck, run, 3, 7, 400)
		if err != nil {
			t.Fatalf("deck %q: %v", deck, err)
		}
		decks, names, err := MatchDecks(root, cards, deck, run.Factions)
		if err != nil {
			t.Fatal(err)
		}
		pol, _ := PolicyByName("random")
		want, err := SimulateWith(decks, 3, 7, 400, SimOptions{DeckNames: names, Policy: pol})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("deck %q: SimulateRun = %+v, want %+v", deck, got, want)
		}
		if got.Games != 3 || got.Coverage == nil {
			t.Fatalf("deck %q: %d games, coverage %v; want 3 games with coverage", deck, got.Games, got.Coverage)
		}
	}
}

func TestSimulateRun_UnknownPolicy(t *testing.T) {
	if _, err := SimulateRun(filepath.Join("..", ".."), nil, "", MatrixRun{Factions: [2]string{"A", "B"}, Policy: "lazy"}, 1, 1, 10); err == nil {
		t.Fatal("SimulateRun with policy lazy: no error")
	}
}
