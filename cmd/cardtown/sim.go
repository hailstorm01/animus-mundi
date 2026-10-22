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
	"animus-mundi/pkg/engine/eventlog"
	"animus-mundi/pkg/engine/sim"
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
	fs.Usage = func() {
		fmt.Println("Usage: cardtown sim [-deck <name>] [-games N] [-seed S] [-p0 FACTION] [-p1 FACTION] [-max-turns N] [-policy random|progress] [-log] [-events FILE] [-monitors=false]")
		fmt.Println("Random-policy playouts on the rules kernel (termination, invariants, win rates)")
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

	if *policy != "random" && *policy != "progress" {
		fmt.Printf("[ERROR] unknown -policy %q (want random or progress)\n", *policy)
		os.Exit(1)
	}

	repoRoot := findRepoRoot()
	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
	if err != nil {
		fmt.Printf("[ERROR] loading cards: %v\n", err)
		os.Exit(1)
	}
	var decks [2]engine.DeckList
	if *deck != "" {
		decksDir := filepath.Join(repoRoot, "data", "decks")
		decks, err = sim.SimDecks(cards, decksDir, *deck, [2]string{*p0, *p1})
		if err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	} else {
		for i, faction := range []string{*p0, *p1} {
			if decks[i], err = sim.FactionDeck(cards, faction); err != nil {
				fmt.Printf("[ERROR] %v\n", err)
				os.Exit(1)
			}
		}
	}

	var out *os.File
	if *eventsFile != "" {
		if out, err = os.Create(*eventsFile); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	}
	opt := sim.SimOptions{NoMonitors: !*monitors}
	if *policy == "progress" {
		opt.Policy = sim.ProgressPolicy
	}
	if *deck != "" {
		for i, f := range []string{*p0, *p1} {
			opt.DeckNames[i] = sim.DecklistName(f, *deck)
		}
	} else {
		opt.DeckNames = [2]string{sim.StressName(*p0), sim.StressName(*p1)}
	}
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
	fmt.Println("  CARD TOWN RULES KERNEL: RANDOM-POLICY PLAYOUTS           ")
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
		fmt.Println("  HQ abilities are ignored.")
	} else {
		fmt.Println("  Decks cycle each faction's kernel-executable cards; HQ abilities are ignored.")
	}
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
			if n := cov[mon][id]; n == 0 {
				fmt.Printf("    %-20s %-12s not reached\n", mon, id)
			} else {
				fmt.Printf("    %-20s %-12s %d\n", mon, id, n)
			}
		}
	}
}

// eventCollector is an Observer that only appends the events it is given.
type eventCollector struct{ events []engine.Event }

func (c *eventCollector) Event(e engine.Event)      { c.events = append(c.events, e) }
func (c *eventCollector) Settled(s engine.Snapshot) {}
