// Command: replay.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/eventlog"
	"animus-mundi/pkg/engine/sim"
)

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
