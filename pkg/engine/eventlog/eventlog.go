package eventlog

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"animus-mundi/pkg/engine"
)

// FormatEvent is one readable line for e, with no newline, starting with the
// Seq. The full snapshot of a turn-start is never printed.
func FormatEvent(e engine.Event) string {
	head := fmt.Sprintf("%d T%d %s %s", e.Seq, e.Turn, e.Phase, e.Kind)
	var body string
	switch {
	case e.Start != nil:
		st := e.Start
		body = fmt.Sprintf("seed %d; decks %q vs %q; HQs %s vs %s; player %d decides, play first=%t",
			st.Seed, st.Decks[0], st.Decks[1], st.HQs[0], st.HQs[1], st.Decider, st.PlayFirst)
	case e.Kind == "turn-start":
		body = fmt.Sprintf("turn %d, player %d is active", e.Turn, e.Player)
	case e.Decision != nil:
		d := e.Decision
		var picked []string
		for _, p := range d.Picks {
			if p >= 0 && p < len(d.Options) {
				picked = append(picked, d.Options[p])
			}
		}
		body = fmt.Sprintf("player %d, %s, %q: picked [%s] (%s)", d.Player, d.Kind, d.Prompt, strings.Join(picked, "; "), d.How)
	case e.End != nil:
		outcome := fmt.Sprintf("player %d wins", e.End.Winner)
		if e.End.Draw {
			outcome = "draw"
		}
		body = fmt.Sprintf("%s: %s after %d turns", outcome, e.End.Reason, e.End.Turns)
	default:
		body = e.Detail
	}
	return head + ": " + body
}

// WriteEventsJSONL writes one JSON object per event and per line, in order.
// Each object is the event's own JSON with a leading "game" key.
func WriteEventsJSONL(w io.Writer, game int, events []engine.Event) error {
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		line := fmt.Appendf(nil, `{"game":%d,`, game)
		line = append(line, b[1:]...)
		line = append(line, '\n')
		if _, err := w.Write(line); err != nil {
			return err
		}
	}
	return nil
}
