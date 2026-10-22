// Command: play.

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/sim"
	"animus-mundi/pkg/view"
)

var errQuit = errors.New("quit")

func runPlay(args []string, in io.Reader, out io.Writer) (exitCode int) {
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(out)
	deck := fs.String("deck", "", "Decklist name under data/decks/ (<faction>-<name>.txt) to play instead of stress decks")
	seed := fs.Uint64("seed", 1, "Seed for the match")
	p0 := fs.String("p0", "SOULLESS", "Faction for player 0")
	p1 := fs.String("p1", "CONCORDIA", "Faction for player 1")
	seat := fs.Int("seat", 0, "Human seat: 0 or 1")
	policy := fs.String("policy", "progress", "Bot seat policy: progress or random")

	fs.Usage = func() {
		fmt.Fprintln(out, "Usage: cardtown play [-deck <name>] [-p0 FACTION] [-p1 FACTION] [-seat 0|1] [-seed S] [-policy progress|random]")
		fmt.Fprintln(out, "Play a text-mode match against a sim policy, drawn only from PlayerView")
		fmt.Fprintln(out, "\nFlags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(normalizeArgs(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if *seat != 0 && *seat != 1 {
		fmt.Fprintf(out, "[ERROR] invalid -seat %d: must be 0 or 1\n", *seat)
		return 1
	}
	if *policy != "progress" && *policy != "random" {
		fmt.Fprintf(out, "[ERROR] unknown -policy %q (want progress or random)\n", *policy)
		return 1
	}

	repoRoot := findRepoRoot()
	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
	if err != nil {
		fmt.Fprintf(out, "[ERROR] loading cards: %v\n", err)
		return 1
	}

	var decks [2]engine.DeckList
	if *deck != "" {
		decksDir := filepath.Join(repoRoot, "data", "decks")
		decks, err = sim.SimDecks(cards, decksDir, *deck, [2]string{*p0, *p1})
		if err != nil {
			fmt.Fprintf(out, "[ERROR] %v\n", err)
			return 1
		}
	} else {
		for i, faction := range []string{*p0, *p1} {
			if decks[i], err = sim.FactionDeck(cards, faction); err != nil {
				fmt.Fprintf(out, "[ERROR] %v\n", err)
				return 1
			}
		}
	}

	var deckNames [2]string
	if *deck != "" {
		for i, f := range []string{*p0, *p1} {
			deckNames[i] = sim.DecklistName(f, *deck)
		}
	} else {
		deckNames = [2]string{sim.StressName(*p0), sim.StressName(*p1)}
	}

	humanSeat := engine.PlayerID(*seat)
	botSeat := humanSeat.Opponent()

	botSeed := *seed*2 + 1
	if botSeat == 1 {
		botSeed = *seed*2 + 2
	}

	scanner := bufio.NewScanner(in)

	defer func() {
		if r := recover(); r != nil {
			if r == errQuit {
				fmt.Fprintln(out, "quit")
				exitCode = 0
				return
			}
			panic(r)
		}
	}()

	var choosers [2]engine.Chooser
	choosers[botSeat] = engine.RandomChooser(botSeed)
	choosers[humanSeat] = terminalChooser(scanner, out)

	cfg := engine.Config{
		Seed:      *seed,
		Decks:     decks,
		DeckNames: deckNames,
		Choosers:  choosers,
	}

	g, err := engine.NewGame(cfg)
	if err != nil {
		fmt.Fprintf(out, "[ERROR] NewGame: %v\n", err)
		return 1
	}

	var botPolicy sim.Policy
	if *policy == "progress" {
		botPolicy = sim.ProgressPolicy(botSeed)
	} else {
		botPolicy = sim.RandomPolicy(botSeed, 0.1)
	}

	seen := 0

	for {
		v := view.PlayerView(g, humanSeat)
		if v.Over {
			break
		}

		if v.Active == humanSeat {
			newLogs := view.LogLines(g.Log[seen:], humanSeat)
			seen = len(g.Log)

			printScreen(out, v, newLogs)

			legal := g.LegalActions()
			fmt.Fprintln(out, "Legal Actions:")
			for i, a := range legal {
				fmt.Fprintf(out, "  %d) %s\n", i+1, actionLabel(v, a))
			}

			chosen := promptAction(scanner, out, legal)
			if err := g.Apply(chosen); err != nil {
				fmt.Fprintf(out, "[ERROR] applying action: %v\n", err)
				return 1
			}
		} else {
			legal := g.LegalActions()
			a := botPolicy(g, legal)
			if err := g.Apply(a); err != nil {
				fmt.Fprintf(out, "[ERROR] applying bot action: %v\n", err)
				return 1
			}
			newLogs := view.LogLines(g.Log[seen:], humanSeat)
			seen = len(g.Log)
			for _, line := range newLogs {
				fmt.Fprintln(out, line)
			}
		}
	}

	v := view.PlayerView(g, humanSeat)
	newLogs := view.LogLines(g.Log[seen:], humanSeat)
	for _, line := range newLogs {
		fmt.Fprintln(out, line)
	}

	if v.Draw {
		fmt.Fprintf(out, "Result: Draw (%s)\n", v.Reason)
	} else if v.Winner >= 0 {
		fmt.Fprintf(out, "Result: Player %d won (%s)\n", v.Winner, v.Reason)
	} else {
		fmt.Fprintf(out, "Result: Match over (%s)\n", v.Reason)
	}

	return 0
}

func printScreen(out io.Writer, v view.View, newLogs []string) {
	if len(newLogs) > 0 {
		for _, line := range newLogs {
			fmt.Fprintln(out, line)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "=== Turn %d (%s) ===\n", v.Turn, v.Phase)
	fmt.Fprintf(out, "Pools: You (R%d T%d) | Opponent (R%d T%d)\n", v.You.Req, v.You.Tac, v.Opponent.Req, v.Opponent.Tac)

	// Opponent
	if v.Opponent.HQ != nil {
		fmt.Fprintf(out, "Opponent HQ: %s\n", formatAsset(*v.Opponent.HQ))
	} else {
		fmt.Fprintln(out, "Opponent HQ: (destroyed)")
	}
	if len(v.Opponent.Reserves) > 0 {
		fmt.Fprintln(out, "Opponent Reserves:")
		for _, a := range v.Opponent.Reserves {
			fmt.Fprintf(out, "  %s\n", formatAsset(a))
		}
	} else {
		fmt.Fprintln(out, "Opponent Reserves: (none)")
	}
	fmt.Fprintf(out, "Opponent Hand: %d cards | Deck: %d cards\n", v.Opponent.HandCount, v.Opponent.DeckCount)

	// Main Theater
	if len(v.Theater) > 0 {
		fmt.Fprintln(out, "Main Theater:")
		for _, a := range v.Theater {
			fmt.Fprintf(out, "  %s (Player %d)\n", formatAsset(a), a.Controller)
		}
	} else {
		fmt.Fprintln(out, "Main Theater: (empty)")
	}

	// You
	if v.You.HQ != nil {
		fmt.Fprintf(out, "Your HQ: %s\n", formatAsset(*v.You.HQ))
	} else {
		fmt.Fprintln(out, "Your HQ: (destroyed)")
	}
	if len(v.You.Reserves) > 0 {
		fmt.Fprintln(out, "Your Reserves:")
		for _, a := range v.You.Reserves {
			fmt.Fprintf(out, "  %s\n", formatAsset(a))
		}
	} else {
		fmt.Fprintln(out, "Your Reserves: (none)")
	}
	if len(v.You.Hand) > 0 {
		fmt.Fprintln(out, "Your Hand:")
		for i, c := range v.You.Hand {
			notExec := ""
			if !c.Executable {
				notExec = " [not executable yet]"
			}
			fmt.Fprintf(out, "  [%d] %s (R%d T%d, %d/%d)%s\n", i+1, c.Card, c.ReqCost, c.TacCost, c.Power, c.Health, notExec)
		}
	} else {
		fmt.Fprintln(out, "Your Hand: (empty)")
	}
	fmt.Fprintf(out, "Your Deck: %d cards\n", v.You.DeckCount)

	// Stack
	if len(v.Stack) > 0 {
		fmt.Fprintln(out, "Resolution Stack:")
		for _, item := range v.Stack {
			fmt.Fprintf(out, "  %s (Player %d, Source %d)\n", item.Kind, item.Controller, item.Source)
		}
	} else {
		fmt.Fprintln(out, "Resolution Stack: (empty)")
	}

	// Discards
	fmt.Fprintf(out, "Discards: You %d | Opponent %d\n\n", len(v.You.Discard), len(v.Opponent.Discard))
}

func formatAsset(a view.AssetView) string {
	status := "Ready"
	if a.Exhausted {
		status = "Exhausted"
	}
	var extra []string
	if a.Armor > 0 {
		extra = append(extra, fmt.Sprintf("Armor %d", a.Armor))
	}
	if a.Damage > 0 {
		extra = append(extra, fmt.Sprintf("Dmg %d", a.Damage))
	}
	statStr := ""
	if len(extra) > 0 {
		statStr = ", " + strings.Join(extra, ", ")
	}
	return fmt.Sprintf("%s#%d (%d/%d, %s%s)", a.Card, a.ID, a.Power, a.MaxHealth, status, statStr)
}

func actionLabel(v view.View, a engine.Action) string {
	findAsset := func(id int) string {
		if v.You.HQ != nil && v.You.HQ.ID == id {
			return fmt.Sprintf("%s#%d", v.You.HQ.Card, id)
		}
		for _, as := range v.You.Reserves {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		if v.Opponent.HQ != nil && v.Opponent.HQ.ID == id {
			return fmt.Sprintf("%s#%d", v.Opponent.HQ.Card, id)
		}
		for _, as := range v.Opponent.Reserves {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		for _, as := range v.Theater {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		return fmt.Sprintf("Asset#%d", id)
	}

	findHandCard := func(instID int) string {
		for _, c := range v.You.Hand {
			if c.Inst == instID {
				return c.Card
			}
		}
		return fmt.Sprintf("Card#%d", instID)
	}

	switch a.Kind {
	case engine.ActGenerate:
		res := "Requisition"
		if a.Resource == engine.Tactics {
			res = "Tactics"
		}
		return fmt.Sprintf("Generate: %s -> %s", findAsset(a.Asset), res)
	case engine.ActConvert:
		return "Convert: 2 Requisition -> 1 Tactics"
	case engine.ActDeploy:
		return fmt.Sprintf("Deploy: %s to %s", findHandCard(a.Card), a.Zone)
	case engine.ActMarch:
		return fmt.Sprintf("March: %s", findAsset(a.Asset))
	case engine.ActAttack:
		if a.Secondary > 0 {
			return fmt.Sprintf("Attack: %s -> %s, %s", findAsset(a.Asset), findAsset(a.Target), findAsset(a.Secondary))
		}
		return fmt.Sprintf("Attack: %s -> %s", findAsset(a.Asset), findAsset(a.Target))
	case engine.ActScratch:
		return fmt.Sprintf("Deploy Scratch Company: %s", findAsset(a.Asset))
	case engine.ActEndTurn:
		return "End Turn"
	default:
		return a.Kind.String()
	}
}

func promptAction(scanner *bufio.Scanner, out io.Writer, legal []engine.Action) engine.Action {
	for {
		fmt.Fprintf(out, "Action [1-%d, q to quit]: ", len(legal))
		if !scanner.Scan() {
			panic(errQuit)
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "q" {
			panic(errQuit)
		}
		val, err := strconv.Atoi(line)
		if err != nil {
			fmt.Fprintf(out, "Invalid input %q: please enter a number between 1 and %d\n", line, len(legal))
			continue
		}
		if val < 1 || val > len(legal) {
			fmt.Fprintf(out, "Invalid choice %d: must be between 1 and %d\n", val, len(legal))
			continue
		}
		return legal[val-1]
	}
}

func terminalChooser(scanner *bufio.Scanner, out io.Writer) engine.Chooser {
	return func(g *engine.Game, d engine.Decision) engine.Answer {
		fmt.Fprintln(out, d.Prompt)

		if d.Kind == engine.DecideOptional {
			for {
				fmt.Fprint(out, "Choose [y/n]: ")
				if !scanner.Scan() {
					panic(errQuit)
				}
				line := strings.TrimSpace(scanner.Text())
				if line == "q" {
					panic(errQuit)
				}
				if line == "" || strings.EqualFold(line, "n") || strings.EqualFold(line, "no") {
					return engine.Answer{Picks: []int{}}
				}
				if strings.EqualFold(line, "y") || strings.EqualFold(line, "yes") {
					return engine.Answer{Picks: []int{0}}
				}
				fmt.Fprintln(out, "Invalid input: please enter y or n")
			}
		}

		for i, opt := range d.Options {
			fmt.Fprintf(out, "  %d) %s\n", i+1, opt.Label)
		}

		for {
			if d.Min == 0 && d.Max > 0 {
				fmt.Fprintf(out, "Choose up to %d options (space-separated, empty for none): ", d.Max)
			} else if d.Min == 1 && d.Max == 1 {
				fmt.Fprint(out, "Choose option (1-N): ")
			} else if d.Kind == engine.DecideOrder {
				fmt.Fprintf(out, "Order all %d options (space-separated): ", len(d.Options))
			} else {
				fmt.Fprintf(out, "Choose %d-%d options (space-separated): ", d.Min, d.Max)
			}

			if !scanner.Scan() {
				panic(errQuit)
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "q" {
				panic(errQuit)
			}
			if line == "" {
				if d.Min == 0 {
					return engine.Answer{Picks: []int{}}
				}
				fmt.Fprintln(out, "Invalid input: selection required")
				continue
			}

			parts := strings.Fields(line)
			var picks []int
			valid := true
			for _, p := range parts {
				idx, err := strconv.Atoi(p)
				if err != nil || idx < 1 || idx > len(d.Options) {
					fmt.Fprintf(out, "Invalid choice %q: must be a number between 1 and %d\n", p, len(d.Options))
					valid = false
					break
				}
				picks = append(picks, idx-1)
			}
			if !valid {
				continue
			}

			if !d.Repeats {
				seen := map[int]bool{}
				hasDup := false
				for _, p := range picks {
					if seen[p] {
						hasDup = true
						break
					}
					seen[p] = true
				}
				if hasDup {
					fmt.Fprintln(out, "Invalid choice: duplicate selections not allowed")
					continue
				}
			}

			if len(picks) < d.Min || len(picks) > d.Max {
				fmt.Fprintf(out, "Invalid choice: selected %d options, want between %d and %d\n", len(picks), d.Min, d.Max)
				continue
			}

			return engine.Answer{Picks: picks}
		}
	}
}
