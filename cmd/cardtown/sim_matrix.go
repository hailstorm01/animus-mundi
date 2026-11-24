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
