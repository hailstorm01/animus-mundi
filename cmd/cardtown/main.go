package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/eventlog"
	"animus-mundi/pkg/engine/sim"
	"animus-mundi/pkg/toolchain/audit"
	"animus-mundi/pkg/toolchain/docs"
	"animus-mundi/pkg/toolchain/git"
	"animus-mundi/pkg/toolchain/hook"
	"animus-mundi/pkg/toolchain/linter/cards"
	"animus-mundi/pkg/toolchain/linter/fences"
	"animus-mundi/pkg/toolchain/linter/frd"
	"animus-mundi/pkg/toolchain/linter/rules"
	"animus-mundi/pkg/toolchain/linter/syntax"
	"animus-mundi/pkg/toolchain/linter/tdd"
	"animus-mundi/pkg/toolchain/tddslice"
)

func findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "lint":
		runLint(os.Args[2:])
	case "hook":
		runHook(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "tdd":
		runTdd(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	case "sim":
		runSim(os.Args[2:])
	case "whats-up", "whatsup", "docs":
		runWhatsUp(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Card Town Unified Developer CLI (cardtown)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cardtown <command> [subcommand] [flags] [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  lint      Unified developer linter suite (FRD, TDD, executable rules, CSV sheets, rules syntax, fences)")
	fmt.Println("  hook      Manage Git lifecycle hooks (install pre-commit)")
	fmt.Println("  audit     Manage PM Devil's Advocate audit lifecycle (status, resolve)")
	fmt.Println("  tdd       Technical Design Document tools (slicing & GitHub issue publishing)")
	fmt.Println("  sim       Random-policy playouts on the rules kernel (termination, invariants, win rates)")
	fmt.Println("  replay    Rebuild a recorded game from a sim -events file and print its state at any event")
	fmt.Println("  whats-up  Live capability and state snapshot for the docs workflow")
	fmt.Println()
	fmt.Println("Run 'cardtown <command> -h' for command-specific flags and options.")
}

func runHook(args []string) {
	if len(args) == 0 || args[0] != "install" {
		fmt.Println("Usage: cardtown hook install")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	if err := hook.Install(repoRoot); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}
}

func runAudit(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cardtown audit [status | resolve]")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	sub := args[0]
	switch sub {
	case "status":
		if err := audit.Status(repoRoot); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	case "resolve":
		fs := flag.NewFlagSet("audit resolve", flag.ExitOnError)
		source := fs.String("source", "Agentic (pm-audit-reviewer)", "Resolution source label")
		msg := fs.String("msg", "", "Custom commit message for remediation")
		push := fs.Bool("push", false, "Run git push after resolution")
		dryRun := fs.Bool("dry-run", false, "Validate without committing or modifying files")
		fs.Parse(normalizeArgs(args[1:]))

		if fs.NArg() == 0 {
			fmt.Println("Usage: cardtown audit resolve [-dry-run] [-push] [-source <label>] [-msg <msg>] <auditFile>")
			os.Exit(1)
		}
		auditFile := fs.Arg(0)
		opts := audit.ResolveOptions{
			AuditFile:     auditFile,
			Source:        *source,
			CommitMessage: *msg,
			Push:          *push,
			DryRun:        *dryRun,
		}
		if err := audit.Resolve(repoRoot, opts); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown audit subcommand %q. Use 'status' or 'resolve'.\n", sub)
		os.Exit(1)
	}
}

func runTdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cardtown tdd [slice]")
		os.Exit(1)
	}
	repoRoot := findRepoRoot()
	sub := args[0]
	switch sub {
	case "slice":
		fs := flag.NewFlagSet("tdd slice", flag.ExitOnError)
		repo := fs.String("repo", "", "Optional GitHub owner/repo override")
		tddFileFlag := fs.String("tdd-file", "", "Path to target TDD markdown file")
		dryRun := fs.Bool("dry-run", true, "Preview mode (default)")
		publish := fs.Bool("publish", false, "Execute live creation/synchronization to GitHub")
		updateExisting := fs.Bool("update-existing", false, "Update existing issues instead of skipping")
		showBodies := fs.Bool("show-bodies", false, "Print rendered Markdown body for all slices")
		fs.Parse(normalizeArgs(args[1:]))

		tddFile := *tddFileFlag
		if tddFile == "" && fs.NArg() > 0 {
			tddFile = fs.Arg(0)
		}
		if tddFile == "" {
			fmt.Println("Usage: cardtown tdd slice [-dry-run] [-publish] [-update-existing] [-show-bodies] [-repo <owner/repo>] <TddFile>")
			os.Exit(1)
		}
		opts := tddslice.SliceOptions{
			TddFile:        tddFile,
			Repo:           *repo,
			DryRun:         *dryRun && !*publish,
			Publish:        *publish,
			UpdateExisting: *updateExisting,
			ShowBodies:     *showBodies,
		}
		if err := tddslice.Slice(repoRoot, opts); err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown tdd subcommand %q. Use 'slice'.\n", sub)
		os.Exit(1)
	}
}

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

func runWhatsUp(args []string) {
	fs := flag.NewFlagSet("whats-up", flag.ExitOnError)
	section := fs.String("section", "all", "Section to show: all, skills, tools, or state")
	fs.Parse(normalizeArgs(args))

	repoRoot := findRepoRoot()
	docs.WhatsUp(repoRoot, *section)
}

func runLint(args []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		sub := args[0]
		remaining := args[1:]
		repoRoot := findRepoRoot()

		switch sub {
		case "frd":
			ok, err := frd.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "tdd":
			ok, err := tdd.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "rules":
			ok, err := rules.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "sheets":
			ok, err := cards.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "fences":
			ok, err := fences.Lint(repoRoot, remaining)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "syntax":
			ok, err := syntax.Lint(repoRoot, remaining, true)
			if err != nil || !ok {
				os.Exit(1)
			}
			os.Exit(0)
		case "cards":
			ok1, err1 := cards.Lint(repoRoot, remaining)
			fmt.Println()
			ok2, err2 := syntax.Lint(repoRoot, remaining, true)
			if err1 != nil || err2 != nil || !ok1 || !ok2 {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	allFlag := fs.Bool("all", false, "Force full audit across all files")
	frdFlag := fs.Bool("frd", false, "Run FRD linter")
	tddFlag := fs.Bool("tdd", false, "Run TDD linter")
	rulesFlag := fs.Bool("rules", false, "Run executable rules traceability linter")
	cardsFlag := fs.Bool("cards", false, "Run sheet integrity & syntax linters")
	sheetsFlag := fs.Bool("sheets", false, "Run sheet integrity linter only")
	syntaxFlag := fs.Bool("syntax", false, "Run rules syntax linter only")
	fencesFlag := fs.Bool("fences", false, "Run engine architecture fences linter")
	stagedFlag := fs.Bool("staged", false, "Smart-run only on Git staged files")
	filesFlag := fs.String("files", "", "Comma-separated list of target files")
	repoRootFlag := fs.String("repo-root", "", "Path to repository root")
	normalizedArgs := normalizeArgs(args)
	fs.Parse(normalizedArgs)

	repoRoot := *repoRootFlag
	if repoRoot == "" {
		repoRoot = findRepoRoot()
	}

	var parsedFiles []string
	if *filesFlag != "" {
		for _, part := range strings.Split(*filesFlag, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				parsedFiles = append(parsedFiles, trimmed)
			}
		}
	}
	parsedFiles = append(parsedFiles, fs.Args()...)

	fmt.Println("==========================================================")
	fmt.Println("               CARD TOWN LINTER DISPATCHER                ")
	fmt.Println("==========================================================")
	fmt.Println("Repo Root:", repoRoot)
	fmt.Println()

	runFrd := false
	runTdd := false
	runRules := false
	runSheets := false
	runSyntax := false
	runFences := false

	var changedFrds []string
	var changedTdds []string
	var changedCsvs []string

	if *allFlag {
		runFrd = true
		runTdd = true
		runRules = true
		runSheets = true
		runSyntax = true
		runFences = true
	} else if *frdFlag || *tddFlag || *rulesFlag || *cardsFlag || *sheetsFlag || *syntaxFlag || *fencesFlag {
		if *frdFlag {
			runFrd = true
		}
		if *tddFlag {
			runTdd = true
		}
		if *rulesFlag {
			runRules = true
		}
		if *cardsFlag || *sheetsFlag {
			runSheets = true
		}
		if *cardsFlag || *syntaxFlag {
			runSyntax = true
		}
		if *fencesFlag {
			runFences = true
		}
	} else {
		// Smart change detection
		if *stagedFlag {
			fmt.Println("[Detecting Staged Files via Git...]")
			staged, err := git.GetStagedFiles(repoRoot)
			if err != nil {
				fmt.Printf("Git staged diff unavailable: %v. Running all linters...\n", err)
				runFrd = true
				runTdd = true
				runRules = true
				runSheets = true
				runSyntax = true
			} else {
				for _, path := range staged {
					pLower := strings.ToLower(filepath.ToSlash(path))
					if isRulesPath(pLower) {
						runRules = true
					}
					if isFencesPath(pLower) {
						runFences = true
					}
					if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
						changedFrds = append(changedFrds, path)
					} else if strings.HasPrefix(pLower, "docs/tdd/") && strings.HasSuffix(pLower, ".md") && filepath.Base(pLower) != "readme.md" {
						changedTdds = append(changedTdds, path)
					} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
						changedCsvs = append(changedCsvs, path)
					}
				}
			}
		} else {
			fmt.Println("[Detecting Changed Files via Git...]")
			statusFiles, err := git.GetStatus(repoRoot)
			if err != nil {
				fmt.Println("  Git change detection unavailable. Running all linters...")
				runFrd = true
				runTdd = true
				runRules = true
				runSheets = true
				runSyntax = true
			} else {
				for _, sf := range statusFiles {
					pLower := strings.ToLower(filepath.ToSlash(sf.Path))
					if isRulesPath(pLower) {
						runRules = true
					}
					if isFencesPath(pLower) {
						runFences = true
					}
					if strings.HasPrefix(pLower, "docs/frd/") && strings.HasSuffix(pLower, ".md") {
						changedFrds = append(changedFrds, sf.Path)
					} else if strings.HasPrefix(pLower, "docs/tdd/") && strings.HasSuffix(pLower, ".md") && filepath.Base(pLower) != "readme.md" {
						changedTdds = append(changedTdds, sf.Path)
					} else if strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv") {
						changedCsvs = append(changedCsvs, sf.Path)
					}
				}
			}
		}

		if len(changedFrds) > 0 {
			fmt.Printf("  Detected %d modified/added FRD document(s).\n", len(changedFrds))
			runFrd = true
		}
		if len(changedTdds) > 0 {
			fmt.Printf("  Detected %d modified/added TDD document(s).\n", len(changedTdds))
			runTdd = true
		}
		if len(changedCsvs) > 0 {
			fmt.Printf("  Detected %d modified/added Card CSV sheet(s).\n", len(changedCsvs))
			runSheets = true
			runSyntax = true
		}

		if !runFrd && !runTdd && !runRules && !runSheets && !runSyntax && !runFences {
			fmt.Println("  No changes detected in docs/frd/, docs/tdd/, pkg/engine/, or data/cards/.")
			fmt.Println("  Nothing to lint! (Use --all to force full audit across all files)")
			fmt.Println()
			fmt.Println("==========================================================")
			os.Exit(0)
		}
	}

	anyFailed := false

	// 1. FRD Linter
	if runFrd {
		fmt.Println("\n>>> Running FRD Requirements Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedFrds) > 0 && !*allFlag {
			files = changedFrds
		}
		ok, err := frd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 2. TDD Linter
	if runTdd {
		fmt.Println("\n>>> Running TDD Coverage & Slice Manifest Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedTdds) > 0 && !*allFlag {
			files = changedTdds
		}
		ok, err := tdd.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 2b. Executable Rules Traceability Linter
	if runRules {
		fmt.Println("\n>>> Running Executable Rules Traceability Linter...")
		ok, err := rules.Lint(repoRoot, nil)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 3. Card Sheet CSV Integrity Linter
	if runSheets {
		fmt.Println("\n>>> Running Card Sheet CSV Integrity Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedCsvs) > 0 && !*allFlag {
			files = changedCsvs
		}
		ok, err := cards.Lint(repoRoot, files)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 4. Card Rules Syntax & Terminology Linter
	if runSyntax {
		fmt.Println("\n>>> Running Card Rules Syntax & Terminology Linter...")
		files := parsedFiles
		if len(files) == 0 && len(changedCsvs) > 0 && !*allFlag {
			files = changedCsvs
		}
		ok, err := syntax.Lint(repoRoot, files, true)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	// 5. Engine Architecture Fences Linter
	if runFences {
		fmt.Println("\n>>> Running Engine Architecture Fences Linter...")
		ok, err := fences.Lint(repoRoot, nil)
		if err != nil || !ok {
			anyFailed = true
		}
	}

	fmt.Println("\n==========================================================")
	if anyFailed {
		fmt.Println("  LINTER SUMMARY: One or more checks FAILED!             ")
		fmt.Println("==========================================================")
		os.Exit(1)
	} else {
		fmt.Println("  LINTER SUMMARY: All executed checks PASSED!             ")
		fmt.Println("==========================================================")
		os.Exit(0)
	}
}

func normalizeArgs(args []string) []string {
	var normalized []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			parts := strings.SplitN(a, "=", 2)
			flagName := strings.TrimLeft(parts[0], "-")
			lower := strings.ToLower(flagName)
			switch lower {
			case "all":
				lower = "all"
			case "frd":
				lower = "frd"
			case "tdd":
				lower = "tdd"
			case "cards":
				lower = "cards"
			case "sheets":
				lower = "sheets"
			case "syntax":
				lower = "syntax"
			case "fences":
				lower = "fences"
			case "staged":
				lower = "staged"
			case "files":
				lower = "files"
			case "reporoot", "repo-root":
				lower = "repo-root"
			case "dryrun", "dry-run":
				lower = "dry-run"
			case "publish":
				lower = "publish"
			case "updateexisting", "update-existing":
				lower = "update-existing"
			case "showbodies", "show-bodies":
				lower = "show-bodies"
			case "tddfile", "tdd-file":
				lower = "tdd-file"
			case "source":
				lower = "source"
			case "msg", "commitmessage", "commit-message":
				lower = "msg"
			case "push":
				lower = "push"
			case "section":
				lower = "section"
			case "rules":
				lower = "rules"
			case "deck":
				lower = "deck"
			case "games":
				lower = "games"
			case "seed":
				lower = "seed"
			case "p0":
				lower = "p0"
			case "p1":
				lower = "p1"
			case "maxturns", "max-turns":
				lower = "max-turns"
			}
			newFlag := "--" + lower
			if len(parts) == 2 {
				normalized = append(normalized, newFlag+"="+parts[1])
			} else {
				normalized = append(normalized, newFlag)
			}
		} else {
			normalized = append(normalized, a)
		}
	}
	return normalized
}

// isRulesPath reports whether a change can affect executable rules traceability.
func isRulesPath(pLower string) bool {
	return pLower == "docs/frd/01_game_rules_frd.md" || (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go"))
}

// isFencesPath reports whether a change can affect engine architecture fences.
func isFencesPath(pLower string) bool {
	return (strings.HasPrefix(pLower, "pkg/engine/") && strings.HasSuffix(pLower, ".go")) ||
		(strings.HasPrefix(pLower, "data/cards/") && strings.HasSuffix(pLower, ".csv")) ||
		pLower == "docs/frd/keyword_interaction_matrix.md" ||
		pLower == "docs/design/m2_starter_set_ability_survey.md"
}

// runReplay rebuilds one recorded game from a `cardtown sim -events` file and
// prints the full state after one event as indented JSON.
func runReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	eventsFile := fs.String("events", "", "Events file written by cardtown sim -events")
	game := fs.Int("game", 0, "Game number in the file (counts from 1)")
	at := fs.Int("at", 0, "Event Seq to stop at (default: the game's last event)")
	fs.Usage = func() {
		fmt.Println("Usage: cardtown replay -events <file> -game N [-at SEQ]")
		fmt.Println("Rebuild a recorded game deterministically and print the full state after event SEQ as JSON")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
	}
	fs.Parse(normalizeArgs(args))
	fail := func(format string, a ...any) {
		fmt.Printf("[ERROR] "+format+"\n", a...)
		os.Exit(1)
	}
	if *eventsFile == "" || *game < 1 {
		fail("replay needs -events <file> and -game N (see cardtown replay -h)")
	}
	f, err := os.Open(*eventsFile)
	if err != nil {
		fail("%v", err)
	}
	games, err := eventlog.ReadEventsJSONL(f)
	f.Close()
	if err != nil {
		fail("%s: %v", *eventsFile, err)
	}
	events, ok := games[*game]
	if !ok {
		fail("%s has no game %d (it has %d games)", *eventsFile, *game, len(games))
	}
	if *at == 0 {
		*at = events[len(events)-1].Seq
	}

	repoRoot := findRepoRoot()
	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
	if err != nil {
		fail("loading cards: %v", err)
	}
	resolve := func(name string) (engine.DeckList, error) { return sim.ResolveDeck(repoRoot, cards, name) }
	snap, err := eventlog.Replay(events, *at, resolve)
	if err != nil {
		fail("%v", err)
	}
	if w := eventlog.ReplayWarning(events); w != "" {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", w)
	}
	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		fail("%v", err)
	}
	fmt.Println(string(out))
}
