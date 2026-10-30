package view

import (
	"fmt"
	"strings"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/eventlog"
)

// LogLines formats events for seat, one line per event, in order, hiding what
// seat may not see. It is the play screen's event log.
//
// Every line is eventlog.FormatEvent's, except two records:
//   - match-start drops the seed (it reveals every shuffle) and names only
//     seat's own mulliganed cards, with a count for the opponent's;
//   - turn-start shows the turn and the active player, never its snapshot.
//
// Event Detail text already hides face-down identities, so no other line
// changes.
func LogLines(events []engine.Event, seat engine.PlayerID) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		switch {
		case e.Kind == engine.EventMatchStart && e.Start != nil:
			out = append(out, eventlog.FormatEvent(redact(e, matchStart(e.Start, seat))))
		case e.Kind == engine.EventTurnStart:
			out = append(out, eventlog.FormatEvent(redact(e, fmt.Sprintf("turn %d, player %d is active", e.Turn, e.Player))))
		default:
			out = append(out, eventlog.FormatEvent(e))
		}
	}
	return out
}

// redact keeps e's header fields and replaces everything else with detail,
// which FormatEvent then prints as the line's body.
func redact(e engine.Event, detail string) engine.Event {
	return engine.Event{Seq: e.Seq, Turn: e.Turn, Phase: e.Phase, Kind: e.Kind, Player: e.Player, Detail: detail}
}

func matchStart(st *engine.MatchStart, seat engine.PlayerID) string {
	opp := seat.Opponent()
	return fmt.Sprintf("decks %q vs %q; HQs %s vs %s; player %d decides, play first=%t; you (player %d) set aside [%s]; player %d set aside %d",
		st.Decks[0], st.Decks[1], st.HQs[0], st.HQs[1], st.Decider, st.PlayFirst,
		seat, strings.Join(st.Mulligans[seat], "; "), opp, len(st.Mulligans[opp]))
}
