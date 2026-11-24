package sim

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"animus-mundi/pkg/engine"
)

// The sim matrix (M3 design note, Deliverable 3) plays every ordered pairing
// of two different factions under each policy, with the same games and seed,
// so one command covers every deck against every other from both seats.

// MatrixPolicies are the policies every pairing runs under, in run order.
var MatrixPolicies = []string{"progress", "random"}

// MatrixRun is one run of the matrix: two factions in seat order and a policy.
type MatrixRun struct {
	Factions [2]string // seat 0, then seat 1
	Policy   string    // a PolicyByName name
}

func (r MatrixRun) String() string {
	return fmt.Sprintf("%s v %s, %s", r.Factions[0], r.Factions[1], r.Policy)
}

// MatrixRuns returns the runs of a matrix over factions, in a fixed order:
// seat 0's faction sorted, then seat 1's, then the policy. A faction never
// plays itself, so n factions give n*(n-1)*len(MatrixPolicies) runs.
// Repeated factions count once.
func MatrixRuns(factions []string) []MatrixRun {
	fs := slices.Clone(factions)
	slices.Sort(fs)
	fs = slices.Compact(fs)
	var runs []MatrixRun
	for _, a := range fs {
		for _, b := range fs {
			if a == b {
				continue
			}
			for _, p := range MatrixPolicies {
				runs = append(runs, MatrixRun{Factions: [2]string{a, b}, Policy: p})
			}
		}
	}
	return runs
}

// DeckFactions returns, sorted and upper-case, the factions that have a
// decklist <faction>-<deck>.txt under repoRoot (DecklistName).
func DeckFactions(repoRoot, deck string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(repoRoot, filepath.FromSlash(decklistDir)))
	if err != nil {
		return nil, err
	}
	suffix := "-" + deck + ".txt"
	var fs []string
	for _, e := range entries {
		if f, ok := strings.CutSuffix(e.Name(), suffix); ok && !e.IsDir() && f != "" {
			fs = append(fs, strings.ToUpper(f))
		}
	}
	slices.Sort(fs)
	return fs, nil
}

// StressFactions returns, sorted, the faction of every Headquarters in cards:
// the factions FactionDeck can build a stress deck for.
func StressFactions(cards map[string]*engine.Card) []string {
	var fs []string
	for _, c := range cards {
		if c.Type == engine.TypeHeadquarters && !slices.Contains(fs, strings.ToUpper(c.Faction)) {
			fs = append(fs, strings.ToUpper(c.Faction))
		}
	}
	slices.Sort(fs)
	return fs
}

// SimulateRun plays one matrix run exactly as cardtown sim plays the same
// flags: the decks of MatchDecks (deck "" means stress decks), the run's
// policy, and the monitors on.
func SimulateRun(repoRoot string, cards map[string]*engine.Card, deck string, run MatrixRun, games int, seed uint64, maxTurns int) (SimStats, error) {
	newPolicy, err := PolicyByName(run.Policy)
	if err != nil {
		return SimStats{}, err
	}
	decks, names, err := MatchDecks(repoRoot, cards, deck, run.Factions)
	if err != nil {
		return SimStats{}, err
	}
	return SimulateWith(decks, games, seed, maxTurns, SimOptions{DeckNames: names, Policy: newPolicy})
}

// MatrixCoverage sums the rule monitor coverage of a matrix's runs: Checks
// counts every check, and Runs counts the runs that checked a rule at least
// once. A rule no run reached is listed with 0 in both.
type MatrixCoverage struct {
	Checks Coverage
	Runs   Coverage
}

// NewMatrixCoverage returns an empty MatrixCoverage.
func NewMatrixCoverage() MatrixCoverage {
	return MatrixCoverage{Checks: Coverage{}, Runs: Coverage{}}
}

// Add sums one run's coverage into m.
func (m MatrixCoverage) Add(run Coverage) {
	m.Checks.add(run)
	reached := Coverage{}
	for mon, ids := range run {
		reached[mon] = map[string]int{}
		for id, n := range ids {
			reached[mon][id] = min(n, 1)
		}
	}
	m.Runs.add(reached)
}
