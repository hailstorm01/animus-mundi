// Command: play.

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/sim"
	"animus-mundi/pkg/view"
)

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

	decks, deckNames, err := sim.MatchDecks(repoRoot, cards, *deck, [2]string{*p0, *p1})
	if err != nil {
		fmt.Fprintf(out, "[ERROR] %v\n", err)
		return 1
	}

	humanSeat := engine.PlayerID(*seat)
	botSeat := humanSeat.Opponent()

	botSeed := sim.SeatSeed(*seed, botSeat)

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
		botPolicy = sim.RandomPolicy(botSeed, sim.DefaultEndBias)
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
