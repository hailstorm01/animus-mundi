---
id: T-0297
title: "cardtown sim -matrix plays every pairing of the starter decks or the stress decks, both seat orders, both policies, with one coverage table (M3 Deliverable 3; harness, flash)"
kind: harness
tier: flash
wave: none (M3 design note Deliverable 3)
status: ready
branch: m3/T-0297-sim-matrix
depends_on: [T-0296]
claims: [type:sim.MatrixRun, func:sim.MatrixRuns, func:sim.DeckFactions, func:sim.StressFactions, func:sim.SimulateRun, type:sim.MatrixCoverage, flag:cardtown-sim-matrix, func:main.eachRule, task:sim-matrix]
cards: []
requirements: []
---

## Goal
`cardtown sim -deck starter -matrix` plays every ordered pairing of the factions that have a `<faction>-starter.txt`, under the `progress` and `random` policies. Each run plays `-games` games from the same `-seed`. It prints one summary line per run, then one coverage table summed over every run, with how many runs reached each rule. Without `-deck` it does the same with the stress decks. It exits non-zero on the first run that fails, naming the run and, for a monitor violation, its replay command. It never fails on an unreached rule. `go tool task sim-matrix` runs both deck sets.

## Inputs
- Design note: `docs/architecture/m3-full-catalog.md`, Boundaries (`pkg/sim`: the starter sim matrix; `cmd/cardtown`: `cardtown sim -matrix`), Test Plan (Sim, Monitor coverage), Deliverable 3.
- Board: no open directive applies.
- Where it is today (at `main` eb30d21c):
  - `cmd/cardtown/sim.go` plays one pairing: `-p0`, `-p1`, one `-policy`. It prints the coverage table with `printCoverage` ([sim.go:177-198](../../../cmd/cardtown/sim.go)).
  - `sim.MatchDecks` ([decks.go:58](../../../pkg/sim/decks.go)) builds both seats' decks (decklists with `-deck`, `FactionDeck` stress decks without). `sim.SimulateWith` plays the games. `sim.Coverage.add` (unexported, [monitor.go:52](../../../pkg/sim/monitor.go)) sums coverage, and `sim.ReplayCommand` builds a replay hint.
  - `data/decks/` holds all four starter decks, and every Headquarters faction has a stress deck.
- Design (lead decisions):
  - **Run order:** seat 0's faction sorted, then seat 1's, then the policy (`MatrixPolicies`: `progress`, `random`). A faction never plays itself, so four factions make 24 runs per deck set.
  - **Factions:** with `-deck`, every `<faction>-<deck>.txt` under `data/decks/`, found through `decklistDir` (`DeckFactions`, so the folder's path stays in one place). Without it, the faction of every Headquarters (`StressFactions`).
  - **`SimulateRun`** plays one run exactly as `cardtown sim -p0 … -p1 … -policy …` would: the same `MatchDecks`, `SimulateWith` and policy, monitors on. A run's replay command therefore reproduces it. `TestSimulateRun_MatchesAPlainSim` pins this, and the dry run's `ANIMUS v PATH, progress` line matched `go tool task sim` exactly.
  - **`MatrixCoverage`** sums `Checks` and counts `Runs` (runs with at least one check). A rule no run reached stays listed with 0 and prints `not reached`.
  - **Flags:** `-matrix` refuses `-p0`, `-p1`, `-policy`, `-log`, `-events` and `-monitors`, and exits 1. It also exits 1 when fewer than two factions have the decklist. It accepts `-deck`, `-games`, `-seed` and `-max-turns`. A failed run exits 1 with `[ERROR] run i of N (<run>): <error>`. For a monitor violation, the replay command follows, with that run's `-deck`, `-p0`, `-p1` and `-policy`, plus `-max-turns` if it was set.
  - **No second copy of the coverage loop:** `printCoverage`'s sorted walk over monitors and rule IDs moves into `eachRule`, which both tables use. The single-pairing output is byte-identical (`go tool task sim-compare`).
  - The matrix is information and a gate for invariants and monitors only. It doesn't replace `go tool task sim`, which every review still runs (Runbook D5).
  - **Expected output difference:** none for `cardtown sim` without `-matrix`.
- Size (D2.4): four non-test files: `pkg/sim/matrix.go` (new), `cmd/cardtown/sim_matrix.go` (new), `cmd/cardtown/sim.go` and `Taskfile.yml`.
- The lead dry-ran these exact files in a throwaway worktree of `main` at eb30d21c:
  - The new tests failed to build without `matrix.go`, and passed with it. `gofmt -l pkg cmd` was clean, `go vet` was clean, `go tool task check` exited 0, and `go tool task sim-compare` printed `[PASS]`.
  - `cardtown sim -deck starter -matrix` exited 0 in about 35 s (24 runs, `[PASS]`).
  - `cardtown sim -matrix` exited 0 in about 57 s (24 runs, `[PASS]`).
  - `go tool task sim-matrix -- -games 5` exited 0.
  - `-matrix -p0 ANIMUS -policy random`, `-deck nosuch -matrix` and `-deck starter -matrix -games 2 -max-turns 5` each exited 1 with the messages above.

  Blobs: `matrix.go` 234c6b2, `matrix_test.go` 208e279, `sim_matrix.go` 62f4ea3, `cmd/cardtown/sim.go` 345dfc9, `Taskfile.yml` d5687a5.

## Acceptance Cases (harness tasks only; written by the lead)
1. `MatrixRuns` of `PATH, ANIMUS, SOULLESS, PATH` → the 12 runs in the fixed order, with repeats counted once (`TestMatrixRuns_OrderAndCount`).
2. Fewer than two distinct factions → no runs (`TestMatrixRuns_FewerThanTwoFactions`).
3. `DeckFactions` finds `<faction>-<deck>.txt` files only: not another deck name, not a folder, not a file with an empty faction (`TestDeckFactions_FindsDecklistsByName`). On the real repository it gives the four factions (`TestDeckFactions_RealStarterDecks`).
4. `StressFactions` on the real cards → the four Headquarters factions (`TestStressFactions_EveryHeadquartersFaction`).
5. `MatrixCoverage.Add` sums checks and counts runs that reached each rule, and keeps unreached rules at 0 (`TestMatrixCoverage_SumsChecksAndCountsRuns`).
6. `SimulateRun` equals a plain `MatchDecks` + `SimulateWith` with the same flags, with decklists and with stress decks (`TestSimulateRun_MatchesAPlainSim`). An unknown policy is an error (`TestSimulateRun_UnknownPolicy`).
7. `go run ./cmd/cardtown sim -deck starter -matrix -games 2` → exit 0: 24 run lines, then `[PASS] All 24 runs terminated with all invariants holding.`, then the coverage table.
8. `go run ./cmd/cardtown sim -matrix -games 2` → exit 0, the same shape with the stress decks.
9. `go run ./cmd/cardtown sim -deck starter -matrix -p0 ANIMUS` → exit 1, naming `-p0`.
10. `go run ./cmd/cardtown sim -deck starter -matrix -games 2 -max-turns 5` → exit 1, `[ERROR] run 1 of 24 (ANIMUS v CONCORDIA, progress): seed 1: match did not end within 5 turns`.
11. `go tool task sim-compare` passes: `cardtown sim` without `-matrix` prints the same bytes.

## Tests (copy exactly; a new file, tabs as shown)
`pkg/sim/matrix_test.go`:
```go
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
```

## Code (copy exactly)
New file:

`pkg/sim/matrix.go`:
```go
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
```

New file:

`cmd/cardtown/sim_matrix.go`:
```go
// Command: sim -matrix.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/sim"
)

// runSimMatrix plays the sim matrix (M3 design note, Deliverable 3): with
// deck, every faction that has a <faction>-<deck>.txt; without it, every
// faction's stress deck. It prints one line per run, then the coverage summed
// over every run, and exits 1 on the first run that fails, naming it and how
// to replay it. An unreached rule never fails the matrix.
func runSimMatrix(fs *flag.FlagSet, repoRoot string, cards map[string]*engine.Card, deck string, games int, seed uint64, maxTurns int) {
	var conflicts []string
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "p0", "p1", "policy", "log", "events", "monitors":
			conflicts = append(conflicts, "-"+f.Name)
		}
	})
	if len(conflicts) > 0 {
		fmt.Printf("[ERROR] -matrix plays every pairing under every policy, with the monitors on; drop %v\n", conflicts)
		os.Exit(1)
	}

	var factions []string
	decks := "stress decks"
	if deck != "" {
		var err error
		if factions, err = sim.DeckFactions(repoRoot, deck); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
		decks = sim.DecklistName("<faction>", deck)
	} else {
		factions = sim.StressFactions(cards)
	}
	runs := sim.MatrixRuns(factions)
	if len(runs) == 0 {
		fmt.Printf("[ERROR] -matrix needs at least two factions with %s; found %v\n", decks, factions)
		os.Exit(1)
	}

	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN RULES KERNEL: SIM MATRIX                       ")
	fmt.Println("==========================================================")
	fmt.Printf("  %d runs (%v, %s), %d games each, seeds %d..%d\n", len(runs), factions, decks, games, seed, seed+uint64(games)-1)
	cov := sim.NewMatrixCoverage()
	for i, run := range runs {
		st, err := sim.SimulateRun(repoRoot, cards, deck, run, games, seed, maxTurns)
		if err != nil {
			fmt.Printf("[ERROR] run %d of %d (%s): %v\n", i+1, len(runs), run, err)
			var v *sim.Violation
			if errors.As(err, &v) {
				var args []string
				if deck != "" {
					args = append(args, "-deck", deck)
				}
				args = append(args, "-p0", run.Factions[0], "-p1", run.Factions[1], "-policy", run.Policy)
				fs.Visit(func(f *flag.Flag) {
					if f.Name == "max-turns" {
						args = append(args, "-max-turns", f.Value.String())
					}
				})
				fmt.Println(sim.ReplayCommand(args, v))
			}
			os.Exit(1)
		}
		cov.Add(st.Coverage)
		pct := func(n int) float64 { return 100 * float64(n) / float64(st.Games) }
		fmt.Printf("  %2d  %-31s P0 %5.1f%%  P1 %5.1f%%  draws %5.1f%%  first %5.1f%%  %5.1f turns\n",
			i+1, run, pct(st.Wins[0]), pct(st.Wins[1]), pct(st.Draws), pct(st.FirstPlayerWin), float64(st.TotalTurns)/float64(st.Games))
	}
	fmt.Printf("  [PASS] All %d runs terminated with all invariants holding.\n", len(runs))
	fmt.Println("==========================================================")
	printMatrixCoverage(cov, len(runs))
}

// printMatrixCoverage prints one line per monitor and rule ID, sorted, with
// the checks summed over the matrix and how many runs reached the rule, or
// "not reached". It is information and never fails a run.
func printMatrixCoverage(cov sim.MatrixCoverage, runs int) {
	fmt.Printf("  RULE MONITOR COVERAGE OVER %d RUNS (checks, then the runs that reached the rule)\n", runs)
	eachRule(cov.Checks, func(mon, id string, n int) {
		if n == 0 {
			fmt.Printf("    %-20s %-12s not reached\n", mon, id)
		} else {
			fmt.Printf("    %-20s %-12s %d in %d of %d runs\n", mon, id, n, cov.Runs[mon][id], runs)
		}
	})
}
```

Replace the whole of `cmd/cardtown/sim.go`. The changes are a `-matrix` flag, a second usage line, the branch to `runSimMatrix` before `MatchDecks`, and `printCoverage` using the new `eachRule`:

`cmd/cardtown/sim.go`:
```go
// Command: sim.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/eventlog"
	"animus-mundi/pkg/sim"
	"animus-mundi/pkg/toolchain/linter/cardsheet"
)

func runSim(args []string) {
	fs := flag.NewFlagSet("sim", flag.ExitOnError)
	deck := fs.String("deck", "", "Decklist name under data/decks/ (<faction>-<name>.txt) to play instead of stress decks")
	games := fs.Int("games", 200, "Number of matches to play")
	seed := fs.Uint64("seed", 1, "First seed; match i uses seed+i")
	p0 := fs.String("p0", "SOULLESS", "Faction for player 0")
	p1 := fs.String("p1", "CONCORDIA", "Faction for player 1")
	maxTurns := fs.Int("max-turns", 400, "Fail if a match runs longer than this")
	logGame := fs.Bool("log", false, "Play only the -seed game and print its events, then the summary (-games must be 1)")
	policy := fs.String("policy", "progress", "Seat policy: progress (deploys, generates toward cards in hand) or random")
	eventsFile := fs.String("events", "", "Write every game's events to this file as JSON Lines (game, then the event)")
	monitors := fs.Bool("monitors", true, "Check FRD 01 rules in every game and print a coverage table")
	matrix := fs.Bool("matrix", false, "Play every pairing of two factions, both seat orders, both policies: with -deck, the factions that have that decklist; without, the stress decks")
	fs.Usage = func() {
		fmt.Println("Usage: cardtown sim [-deck <name>] [-games N] [-seed S] [-p0 FACTION] [-p1 FACTION] [-max-turns N] [-policy random|progress] [-log] [-events FILE] [-monitors=false]")
		fmt.Println("       cardtown sim -matrix [-deck <name>] [-games N] [-seed S] [-max-turns N]")
		fmt.Println("Playouts on the rules kernel (termination, invariants, win rates)")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
	}
	fs.Parse(normalizeArgs(args))

	if *logGame {
		gamesSet := false
		fs.Visit(func(f *flag.Flag) { gamesSet = gamesSet || f.Name == "games" })
		if gamesSet && *games != 1 {
			fmt.Println("[ERROR] -log plays exactly one game (the one at -seed); drop -games or use -games 1")
			os.Exit(1)
		}
		*games = 1
	}

	newPolicy, err := sim.PolicyByName(*policy)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	repoRoot := findRepoRoot()
	cards, err := engine.LoadCards(filepath.Join(repoRoot, cardsheet.Dir))
	if err != nil {
		fmt.Printf("[ERROR] loading cards: %v\n", err)
		os.Exit(1)
	}
	if *matrix {
		runSimMatrix(fs, repoRoot, cards, *deck, *games, *seed, *maxTurns)
		return
	}
	decks, deckNames, err := sim.MatchDecks(repoRoot, cards, *deck, [2]string{*p0, *p1})
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	var out *os.File
	if *eventsFile != "" {
		if out, err = os.Create(*eventsFile); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	}
	opt := sim.SimOptions{DeckNames: deckNames, Policy: newPolicy, NoMonitors: !*monitors}
	var played []*eventCollector // one per game, only when -log or -events asks for events
	if *logGame || out != nil {
		opt.Observer = func(game int, seed uint64) engine.Observer {
			c := &eventCollector{}
			played = append(played, c)
			return c
		}
	}
	writeEvents := func() {
		var werr error
		for i, c := range played {
			if werr = eventlog.WriteEventsJSONL(out, i+1, c.events); werr != nil {
				break
			}
		}
		if cerr := out.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			fmt.Printf("[ERROR] writing %s: %v\n", *eventsFile, werr)
			os.Exit(1)
		}
	}
	st, err := sim.SimulateWith(decks, *games, *seed, *maxTurns, opt)
	if err != nil {
		var v *sim.Violation
		if errors.As(err, &v) {
			// Keep the events up to the violation, so the replay hint has a file to read.
			if out != nil {
				writeEvents()
			}
			fmt.Printf("[ERROR] %v\n", v)
			var simArgs []string
			fs.Visit(func(f *flag.Flag) {
				switch f.Name {
				case "seed", "games", "log", "events":
				default:
					if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
						simArgs = append(simArgs, "-"+f.Name+"="+f.Value.String())
					} else {
						simArgs = append(simArgs, "-"+f.Name, f.Value.String())
					}
				}
			})
			fmt.Println(sim.ReplayCommand(simArgs, v))
			os.Exit(1)
		}
		if out != nil {
			out.Close()
		}
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}
	if out != nil {
		writeEvents()
	}
	if *logGame {
		for _, e := range played[0].events {
			fmt.Println(eventlog.FormatEvent(e))
		}
	}
	pct := func(n int) float64 { return 100 * float64(n) / float64(st.Games) }
	fmt.Println("==========================================================")
	fmt.Println("  CARD TOWN RULES KERNEL: PLAYOUTS                         ")
	fmt.Println("==========================================================")
	fmt.Printf("  %s (%s) vs %s (%s), %d games, seeds %d..%d\n", *p0, decks[0].HQ.Name, *p1, decks[1].HQ.Name, st.Games, *seed, *seed+uint64(st.Games)-1)
	if *deck != "" {
		countSupported := func(main []*engine.Card) int {
			n := 0
			for _, c := range main {
				if c.Supported() {
					n++
				}
			}
			return n
		}
		fmt.Printf("  P0 (%s): %d/%d main cards executable now (Supported())\n", sim.DecklistName(*p0, *deck), countSupported(decks[0].Main), len(decks[0].Main))
		fmt.Printf("  P1 (%s): %d/%d main cards executable now (Supported())\n", sim.DecklistName(*p1, *deck), countSupported(decks[1].Main), len(decks[1].Main))
	} else {
		fmt.Println("  Decks cycle each faction's kernel-executable cards.")
	}
	fmt.Printf("  Headquarters executable now (Supported()): P0 (%s) %t, P1 (%s) %t\n", decks[0].HQ.Name, decks[0].HQ.Supported(), decks[1].HQ.Name, decks[1].HQ.Supported())
	fmt.Printf("  P0 wins %.1f%%  P1 wins %.1f%%  draws %.1f%%  first player wins %.1f%%\n", pct(st.Wins[0]), pct(st.Wins[1]), pct(st.Draws), pct(st.FirstPlayerWin))
	fmt.Printf("  Average match length: %.1f turns\n", float64(st.TotalTurns)/float64(st.Games))
	fmt.Println("  End reasons:")
	reasons := make([]string, 0, len(st.Reasons))
	for reason := range st.Reasons {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		fmt.Printf("    %-45s %d\n", reason, st.Reasons[reason])
	}
	fmt.Println("  [PASS] Every match terminated with all invariants holding.")
	fmt.Println("==========================================================")
	if st.Coverage != nil {
		printCoverage(st.Coverage)
	}
}

// printCoverage prints one line per monitor and rule ID, sorted, with the
// number of checks or "not reached". It is information and never fails a run.
func printCoverage(cov sim.Coverage) {
	fmt.Println("  RULE MONITOR COVERAGE (checks per rule; the monitors found no violation)")
	eachRule(cov, func(mon, id string, n int) {
		if n == 0 {
			fmt.Printf("    %-20s %-12s not reached\n", mon, id)
		} else {
			fmt.Printf("    %-20s %-12s %d\n", mon, id, n)
		}
	})
}

// eachRule calls f for every monitor and rule ID in cov, sorted by monitor,
// then by rule ID, with its count.
func eachRule(cov sim.Coverage, f func(mon, id string, n int)) {
	monitors := make([]string, 0, len(cov))
	for mon := range cov {
		monitors = append(monitors, mon)
	}
	sort.Strings(monitors)
	for _, mon := range monitors {
		ids := make([]string, 0, len(cov[mon]))
		for id := range cov[mon] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			f(mon, id, cov[mon][id])
		}
	}
}

// eventCollector is an Observer that only appends the events it is given.
type eventCollector struct{ events []engine.Event }

func (c *eventCollector) Event(e engine.Event)      { c.events = append(c.events, e) }
func (c *eventCollector) Settled(s engine.Snapshot) {}
```

In `Taskfile.yml`, insert this block right before the line `  sim-log:` (two spaces of indent, then a blank line before `sim-log:`):
```yaml
  sim-matrix:
    desc: 'Play every pairing of the starter decks, then of the stress decks: both seat orders, both policies, one coverage table each (cardtown sim -matrix)'
    deps: [build]
    cmds:
      - '{{.BIN}} sim -deck starter -matrix {{.CLI_ARGS}}'
      - '{{.BIN}} sim -matrix {{.CLI_ARGS}}'

```

## Commits
1. `test(T-0297): the sim matrix runs, factions and coverage` holds only `pkg/sim/matrix_test.go`. On this commit `go test ./pkg/sim -run TestMatrix` fails to build.
2. `feat(sim): cardtown sim -matrix plays every pairing under both policies with one coverage table (T-0297)` holds `matrix.go`, `sim_matrix.go`, `sim.go` and `Taskfile.yml`.

No double quotes in either message: the allow-list refuses an escaped one.

## Allowed Files
- `pkg/sim/matrix.go` (new)
- `cmd/cardtown/sim_matrix.go` (new)
- `cmd/cardtown/sim.go`
- `Taskfile.yml`

## Done When
- [ ] `go test ./pkg/sim -run TestMatrix -count=1` passes
- [ ] `go test ./pkg/sim -run TestDeckFactions -count=1` passes
- [ ] `go test ./pkg/sim -run TestStressFactions -count=1` passes
- [ ] `go test ./pkg/sim -run TestSimulateRun -count=1` passes
- [ ] `gofmt -l pkg cmd` prints nothing
- [ ] `go run ./cmd/cardtown sim -deck starter -matrix -games 2` exits 0 (acceptance case 7)
- [ ] `go run ./cmd/cardtown sim -matrix -games 2` exits 0 (acceptance case 8)
- [ ] `go tool task check` passes (gofmt, `go vet`, every test, `cardtown lint --all`)
- [ ] `go tool task sim-compare` passes (starter and stress sims byte-identical to `main`)

## Log
- 2026-11-24 lead-architect: written from M3 design note Deliverable 3. Status `ready` (a harness task skips `tests`), tier `flash`.
