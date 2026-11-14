// Command: play, input: action prompts and the terminal Chooser.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"animus-mundi/pkg/engine"
)

var errQuit = errors.New("quit")

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
		if d.Source != nil {
			fmt.Fprintln(out, sourceLine(d.Source))
		}

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

		shown := shownOptions(d)
		for k, i := range shown {
			fmt.Fprintf(out, "  %d) %s\n", k+1, d.Options[i].Label)
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
			picks := []int{}
			parsed := true
			for _, p := range strings.Fields(line) {
				idx, err := strconv.Atoi(p)
				if err != nil || idx < 1 || idx > len(shown) {
					fmt.Fprintf(out, "Invalid choice %q: must be a number between 1 and %d\n", p, len(shown))
					parsed = false
					break
				}
				picks = append(picks, shown[idx-1])
			}
			if !parsed {
				continue
			}

			if !d.Valid(picks) {
				suffix := ""
				if !d.Repeats {
					suffix = ", no repeats"
				}
				fmt.Fprintf(out, "Invalid choice: pick %d to %d option(s)%s\n", d.Min, d.Max, suffix)
				continue
			}

			return engine.Answer{Picks: picks}
		}
	}
}

// sourceLine is what terminalChooser prints under the prompt of a decision with
// a source card (Decision.Source, #149): the card's name and its printed rules
// text, each further line of the text indented under the first.
func sourceLine(c *engine.Card) string {
	text := strings.ReplaceAll(strings.TrimSpace(c.RulesText), "\n", "\n    ")
	if text == "" {
		return "  [" + c.Name + "]"
	}
	return "  [" + c.Name + "] " + text
}

// shownOptions returns the indices of d's options that terminalChooser lists,
// in order: the listed number k+1 answers option shownOptions(d)[k]. A decision
// that picks at most one lists each option text once, at its first option, so
// identical outcomes (two copies of a card) aren't offered twice (#151); any
// other decision lists every option, because picking two copies there is a
// real choice.
func shownOptions(d engine.Decision) []int {
	var shown []int
	seen := make(map[string]bool, len(d.Options))
	for i, opt := range d.Options {
		if d.Max == 1 {
			if seen[opt.Label] {
				continue
			}
			seen[opt.Label] = true
		}
		shown = append(shown, i)
	}
	return shown
}
