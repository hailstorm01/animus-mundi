package main

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/sim"
)

// Tests for T-0037: cardtown play (Acceptance Cases 1 to 4).

type scriptedReader struct {
	out         *bytes.Buffer
	lastSeenLen int
}

func (r *scriptedReader) Read(p []byte) (int, error) {
	s := r.out.String()
	recent := s[r.lastSeenLen:]
	r.lastSeenLen = len(s)

	response := "1\n"
	if strings.Contains(recent, "[y/n]") || strings.Contains(recent, "(y/n)") {
		response = "y\n"
	}
	n := copy(p, response)
	return n, nil
}

// Case 1: A scripted session finishes.
//   - runPlay with -deck starter -p0 ANIMUS -p1 PATH -seat 0 -seed 3, and an input
//     stream that answers 1 to every prompt and y to every yes/no question,
//     finishes the match and returns 0.
//   - Its output ends with the result line.
//   - It contains [not executable yet] at least once (the starter decks hold such cards).
//   - It is byte-identical across two runs.
func TestPlay_Case1_ScriptedSessionFinishes(t *testing.T) {
	args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", "3"}

	runOnce := func() (int, string) {
		var out bytes.Buffer
		in := &scriptedReader{out: &out}
		code := runPlay(args, in, &out)
		return code, out.String()
	}

	code1, out1 := runOnce()
	if code1 != 0 {
		t.Fatalf("first run exited with code %d, want 0; output:\n%s", code1, out1)
	}

	trimmed1 := strings.TrimSpace(out1)
	if !strings.HasPrefix(trimmed1[strings.LastIndex(trimmed1, "\n")+1:], "Result:") {
		t.Errorf("output does not end with result line; last lines:\n%s", trimmed1[max(0, len(trimmed1)-200):])
	}

	if !strings.Contains(out1, "[not executable yet]") {
		t.Errorf("output does not contain %q", "[not executable yet]")
	}

	code2, out2 := runOnce()
	if code2 != 0 {
		t.Fatalf("second run exited with code %d, want 0", code2)
	}

	if out1 != out2 {
		t.Errorf("output is not byte-identical across two runs")
	}
}

// Case 2: Re-prompting.
// An input of x, then 999, then 1 takes the first legal action. The output shows two re-prompts.
func TestPlay_Case2_Reprompting(t *testing.T) {
	args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", "3"}
	var out bytes.Buffer
	// Input: invalid non-number "x", invalid out-of-range "999", valid "1", then "q" to quit.
	in := strings.NewReader("x\n999\n1\nq\n")

	code := runPlay(args, in, &out)
	if code != 0 {
		t.Fatalf("runPlay exited with code %d, want 0", code)
	}

	s := out.String()
	// The CLI should re-prompt twice before accepting action 1.
	repromptCount := strings.Count(s, "Invalid")
	if repromptCount < 2 {
		t.Errorf("expected at least 2 re-prompts in output, got %d; output:\n%s", repromptCount, s)
	}
}

// Case 3: Quit.
// q at the first prompt prints "quit" and returns 0. An empty input stream (EOF) does the same.
func TestPlay_Case3_Quit(t *testing.T) {
	args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", "3"}

	t.Run("q at first prompt", func(t *testing.T) {
		var out bytes.Buffer
		code := runPlay(args, strings.NewReader("q\n"), &out)
		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if !strings.HasSuffix(strings.TrimSpace(out.String()), "quit") {
			t.Errorf("output does not end with 'quit'; got:\n%s", out.String())
		}
	})

	t.Run("EOF at first prompt", func(t *testing.T) {
		var out bytes.Buffer
		code := runPlay(args, strings.NewReader(""), &out)
		if code != 0 {
			t.Errorf("exit code = %d, want 0", code)
		}
		if !strings.HasSuffix(strings.TrimSpace(out.String()), "quit") {
			t.Errorf("output does not end with 'quit'; got:\n%s", out.String())
		}
	})
}

// Case 4: No hidden information on screen.
//   - In case 1's whole output, no name of a card in the bot seat's hand or deck
//     appears unless it is publicly visible at that time.
//   - No face-down asset's real name appears unless it is publicly visible.
//   - Checked with the same rule as T-0036's case 4, by replaying the seed.
func TestPlay_Case4_NoHiddenInformation(t *testing.T) {
	// Replay seed 3 match with the same policy and starter decks.
	repoRoot := findRepoRoot()
	cards, err := engine.LoadCards(filepath.Join(repoRoot, "data", "cards"))
	if err != nil {
		t.Fatalf("load cards: %v", err)
	}
	decks, err := sim.SimDecks(cards, filepath.Join(repoRoot, "data", "decks"), "starter", [2]string{"ANIMUS", "PATH"})
	if err != nil {
		t.Fatalf("SimDecks: %v", err)
	}

	botSeat := engine.PlayerID(1)
	seed := uint64(3)
	botSeed := seed*2 + 2

	cfg := engine.Config{
		Seed:  seed,
		Decks: decks,
		Choosers: [2]engine.Chooser{
			func(g *engine.Game, d engine.Decision) engine.Answer {
				// Human always picks first option / accepts optional
				if d.Kind == engine.DecideOptional {
					return engine.Answer{Picks: []int{0}}
				}
				if d.Min > 0 {
					picks := make([]int, d.Min)
					for i := range picks {
						picks[i] = i
					}
					return engine.Answer{Picks: picks}
				}
				return engine.Answer{}
			},
			engine.RandomChooser(botSeed),
		},
	}
	g, err := engine.NewGame(cfg)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}

	// Capture Case 1 output by running runPlay
	args := []string{"-deck", "starter", "-p0", "ANIMUS", "-p1", "PATH", "-seat", "0", "-seed", "3"}
	var out bytes.Buffer
	in := &scriptedReader{out: &out}
	code := runPlay(args, in, &out)
	if code != 0 {
		t.Fatalf("runPlay exited with code %d", code)
	}

	// Split output into screens (each screen starts with turn header)
	screens := strings.Split(out.String(), "=== Turn ")

	botPolicy := sim.ProgressPolicy(botSeed)

	publicNames := func(g *engine.Game) map[string]bool {
		pub := map[string]bool{}
		for _, p := range g.Players {
			for _, inst := range p.Discard {
				pub[inst.Card.Name] = true
			}
		}
		for _, a := range g.Assets() {
			if !a.FaceDown {
				pub[a.Inst.Card.Name] = true
			}
		}
		for _, it := range g.Stack {
			if (it.Kind == engine.StackDeploy || it.Kind == engine.StackAbility) && it.Inst != nil {
				pub[it.Inst.Card.Name] = true
			}
		}
		return pub
	}

	hiddenNames := func(g *engine.Game, q engine.PlayerID) map[string]bool {
		hid := map[string]bool{}
		for _, inst := range append(slices.Clone(g.Players[q].Hand), g.Players[q].Deck...) {
			hid[inst.Card.Name] = true
		}
		for _, a := range g.Assets() {
			if a.FaceDown && (a.Controller == q || a.Inst.Owner == q) {
				hid[a.Inst.Card.Name] = true
			}
		}
		return hid
	}

	screenIdx := 1 // screens[0] is text before first screen, if any
	checkScreen := func(screenText string) {
		screenOnly := screenText
		if idx := strings.Index(screenText, "Legal Actions:"); idx != -1 {
			screenOnly = screenText[:idx]
		}
		pub := publicNames(g)
		for name := range hiddenNames(g, botSeat) {
			if pub[name] {
				continue
			}
			if strings.Contains(screenOnly, name) {
				t.Errorf("turn %d: screen leaks hidden card %q of bot seat", g.Turn, name)
			}
		}
	}

	// Replay step-by-step
	for !g.Outcome.Over {
		if g.Active == 0 {
			if screenIdx < len(screens) {
				checkScreen(screens[screenIdx])
				screenIdx++
			}
			legal := g.LegalActions()
			if err := g.Apply(legal[0]); err != nil {
				t.Fatalf("apply action: %v", err)
			}
		} else {
			legal := g.LegalActions()
			a := botPolicy(g, legal)
			if err := g.Apply(a); err != nil {
				t.Fatalf("apply bot action: %v", err)
			}
		}
	}
}
