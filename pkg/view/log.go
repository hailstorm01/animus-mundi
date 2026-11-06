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
// Every line is eventlog.FormatEvent's, except three records:
//   - match-start drops the seed (it reveals every shuffle) and names only
//     seat's own mulliganed cards, with a count for the opponent's;
//   - turn-start shows the turn and the active player, never its snapshot;
//   - another seat's hidden decision names no option and no pick: a deck search
//     says whether a card was found, and any other kind says only that a choice
//     was made.
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
		// The search clause is only for T-0168's hand-built records, which don't set
		// the flag: every search in a game does. Directive #123 drops it.
		case e.Decision != nil && e.Decision.Player != int(seat) && (e.Decision.Hidden || e.Decision.Kind == engine.DecideSearch.String()):
			out = append(out, eventlog.FormatEvent(redact(e, hiddenDecision(e.Decision))))
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

// hiddenDecision is what the other seat sees of d. It names no option and no pick.
func hiddenDecision(d *engine.DecisionRecord) string {
	switch {
	case d.Kind != engine.DecideSearch.String():
		return fmt.Sprintf("player %d makes a hidden choice", d.Player)
	case len(d.Picks) > 0:
		return fmt.Sprintf("player %d searches their deck and finds a card", d.Player)
	}
	return fmt.Sprintf("player %d searches their deck and finds no card", d.Player)
}

func matchStart(st *engine.MatchStart, seat engine.PlayerID) string {
	opp := seat.Opponent()
	return fmt.Sprintf("decks %q vs %q; HQs %s vs %s; player %d decides, play first=%t; you (player %d) set aside [%s]; player %d set aside %d",
		st.Decks[0], st.Decks[1], st.HQs[0], st.HQs[1], st.Decider, st.PlayFirst,
		seat, strings.Join(st.Mulligans[seat], "; "), opp, len(st.Mulligans[opp]))
}
