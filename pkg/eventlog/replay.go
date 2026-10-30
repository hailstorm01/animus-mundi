package eventlog

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"animus-mundi/pkg/engine"
)

// ReplayWarning is a one-line warning when a replay of events is only
// best-effort because the record or this build has uncommitted changes. It is
// "" otherwise. The divergence check still guards correctness.
func ReplayWarning(events []engine.Event) string {
	_, buildMod := engine.BuildInfo()
	recMod := len(events) > 0 && events[0].Start != nil && events[0].Start.Modified
	if buildMod || recMod {
		return "the record or this build has uncommitted changes; the replay may diverge"
	}
	return ""
}

// Replay rebuilds the game recorded in events (match-start first, Seq order)
// and returns the state immediately after event `at` was emitted. That point
// may be mid-resolution. resolveDeck turns a recorded deck name into a DeckList.
// Any difference between the replayed game and the record is an error naming
// the first Seq where they differ.
func Replay(events []engine.Event, at int, resolveDeck func(name string) (engine.DeckList, error)) (engine.Snapshot, error) {
	if len(events) == 0 || events[0].Kind != engine.EventMatchStart || events[0].Start == nil {
		return engine.Snapshot{}, fmt.Errorf("replay: the record must start with a match-start event")
	}
	for i, e := range events {
		if e.Seq != i+1 {
			return engine.Snapshot{}, fmt.Errorf("replay: event %d has seq %d, want %d", i, e.Seq, i+1)
		}
	}
	last := len(events)
	if at < 1 || at > last {
		return engine.Snapshot{}, fmt.Errorf("replay: seq %d is out of range 1..%d", at, last)
	}
	st := events[0].Start
	if rev, _ := engine.BuildInfo(); st.Revision != rev { // modified never bypasses a mismatch; it only adds the CLI warning
		return engine.Snapshot{}, fmt.Errorf("replay: the record is from revision %q, this build is revision %q", st.Revision, rev)
	}
	for _, e := range events[:at] {
		if e.Decision != nil && e.Decision.How == "invalid" {
			return engine.Snapshot{}, fmt.Errorf("replay: seq %d: cannot replay an invalid answer", e.Seq)
		}
	}
	if resolveDeck == nil {
		return engine.Snapshot{}, fmt.Errorf("replay: no deck resolver")
	}

	r := &replayer{events: events, at: at}
	r.captureSeq = at
	if at == 1 {
		// Event 1 is emitted inside NewGame, before there is a Game to
		// snapshot. Nothing changes until turn 1's turn-start (event 2), so
		// that event's snapshot gives the state, with the pre-turn scalars.
		r.captureSeq = 2
		if last < 2 {
			return engine.Snapshot{}, fmt.Errorf("replay: seq 1: the record has no turn-start to rebuild the setup state from")
		}
	}
	var decks [2]engine.DeckList
	for i, name := range st.Decks {
		d, err := resolveDeck(name)
		if err != nil {
			return engine.Snapshot{}, fmt.Errorf("replay: deck %q: %w", name, err)
		}
		decks[i] = d
	}
	for _, e := range events {
		if e.Decision != nil && e.Decision.How != "auto" && (e.Decision.Player == 0 || e.Decision.Player == 1) {
			r.queue[e.Decision.Player] = append(r.queue[e.Decision.Player], e)
		}
	}

	cfg := engine.Config{
		Seed: st.Seed, Decks: decks, DeckNames: st.Decks, Observer: r,
		PlayFirst: func(engine.PlayerID) bool { return st.PlayFirst },
		Choosers:  [2]engine.Chooser{r.chooser(0), r.chooser(1)},
	}
	if st.MulliganCallback {
		cfg.Mulligan = func(p engine.PlayerID, _ []*engine.Instance) []int { return slices.Clone(st.MulliganInsts[p]) }
	}
	g, err := engine.NewGame(cfg)
	if err != nil {
		return engine.Snapshot{}, fmt.Errorf("replay: %w", err)
	}
	r.g = g

	for _, e := range events {
		if r.done() || e.Seq > at {
			break
		}
		if e.Kind != engine.EventAction {
			continue
		}
		if r.seen+1 != e.Seq {
			r.fail(r.seen+1, "the replay emitted %d events before the next recorded action (seq %d)", r.seen, e.Seq)
			break
		}
		a, err := actionOf(e)
		if err != nil {
			r.fail(e.Seq, "%v", err)
			break
		}
		if err := g.Apply(a); err != nil {
			r.fail(e.Seq, "the recorded action was rejected: %v", err)
			break
		}
	}
	if r.err != nil {
		return engine.Snapshot{}, r.err
	}
	if !r.captured {
		return engine.Snapshot{}, fmt.Errorf("replay: seq %d: the replay ended after %d events without reaching seq %d", r.seen+1, r.seen, at)
	}
	return r.snap, nil
}

// replayer is Replay's Observer: it checks each replayed event against the
// record and captures the state at the wanted event.
type replayer struct {
	events     []engine.Event
	at         int
	captureSeq int
	g          *engine.Game // nil until NewGame returns
	seen       int          // events received
	err        error
	captured   bool
	snap       engine.Snapshot
	queue      [2][]engine.Event // recorded decisions still to answer, per seat
}

func (r *replayer) done() bool { return r.err != nil || r.captured }

func (r *replayer) fail(seq int, format string, args ...any) {
	if r.err == nil {
		r.err = fmt.Errorf("replay diverged at seq %d: %s", seq, fmt.Sprintf(format, args...))
	}
}

func (r *replayer) Settled(engine.Snapshot) {}

func (r *replayer) Event(e engine.Event) {
	if r.done() {
		return
	}
	r.seen = e.Seq
	if e.Seq < 1 || e.Seq > len(r.events) {
		r.fail(e.Seq, "the replay emitted an event the record does not have: %s", e.Detail)
		return
	}
	if msg := diff(r.events[e.Seq-1], e); msg != "" {
		r.fail(e.Seq, "%s", msg)
		return
	}
	if e.Seq != r.captureSeq {
		return
	}
	switch {
	case r.at == 1 && e.Snapshot != nil:
		s := *e.Snapshot
		s.Turn, s.Phase, s.Active = 0, "setup", 0
		r.snap = s
	case e.Snapshot != nil:
		r.snap = *e.Snapshot
	case r.g != nil:
		r.snap = r.g.Snapshot()
	default:
		r.fail(e.Seq, "the state during setup is not available")
		return
	}
	r.captured = true
}

// chooser replays seat's recorded decisions in order.
func (r *replayer) chooser(seat int) engine.Chooser {
	return func(_ *engine.Game, d engine.Decision) engine.Answer {
		if r.done() {
			return engine.Answer{Timeout: true}
		}
		if len(r.queue[seat]) == 0 {
			r.fail(r.seen+1, "player %d was asked %q, but the record has no more decisions for that seat", seat, d.Prompt)
			return engine.Answer{Timeout: true}
		}
		rec := r.queue[seat][0]
		r.queue[seat] = r.queue[seat][1:]
		dr := rec.Decision
		refs := make([]int, len(d.Options))
		for i, o := range d.Options {
			refs[i] = o.Ref
		}
		if dr.Kind != d.Kind.String() || dr.Player != int(d.Player) || !slices.Equal(dr.Refs, refs) {
			r.fail(rec.Seq, "the replay asked player %d a %s decision %q with refs %v; the record has a %s decision with refs %v",
				d.Player, d.Kind, d.Prompt, refs, dr.Kind, dr.Refs)
			return engine.Answer{Timeout: true}
		}
		if dr.How == "answer" {
			return engine.Answer{Picks: slices.Clone(dr.Picks)}
		}
		return engine.Answer{Timeout: true}
	}
}

// diff describes the first difference between the recorded event and the
// replayed one, or "" if they are the same.
func diff(rec, got engine.Event) string {
	switch {
	case rec.Kind != got.Kind:
		return fmt.Sprintf("kind: recorded %q, replayed %q", rec.Kind, got.Kind)
	case rec.Detail != got.Detail:
		return fmt.Sprintf("detail: recorded %q, replayed %q", rec.Detail, got.Detail)
	case rec.Turn != got.Turn || rec.Phase != got.Phase:
		return fmt.Sprintf("turn and phase: recorded T%d %s, replayed T%d %s", rec.Turn, rec.Phase, got.Turn, got.Phase)
	case rec.Player != got.Player || rec.Card != got.Card || rec.Inst != got.Inst || rec.Source != got.Source ||
		rec.Target != got.Target || rec.Zone != got.Zone || rec.Amount != got.Amount || rec.Resource != got.Resource:
		return fmt.Sprintf("%s event fields: recorded %+v, replayed %+v", rec.Kind, structured(rec), structured(got))
	}
	if (rec.Decision == nil) != (got.Decision == nil) ||
		(rec.Decision != nil && !decisionEqual(*rec.Decision, *got.Decision)) {
		return "decision record differs"
	}
	if (rec.Start == nil) != (got.Start == nil) {
		return "match-start record differs"
	}
	if rec.Start != nil {
		a, b := *rec.Start, *got.Start
		a.Revision, b.Revision, a.Modified, b.Modified = "", "", false, false // checked up front
		if !reflect.DeepEqual(a, b) {
			return fmt.Sprintf("match-start record differs: recorded %+v, replayed %+v", a, b)
		}
	}
	if !reflect.DeepEqual(rec.End, got.End) {
		return "match-end record differs"
	}
	if !reflect.DeepEqual(rec.Snapshot, got.Snapshot) {
		return "turn-start snapshot differs"
	}
	return ""
}

func structured(e engine.Event) [8]any {
	return [8]any{e.Player, e.Card, e.Inst, e.Source, e.Target, e.Zone, e.Amount, e.Resource}
}

func decisionEqual(a, b engine.DecisionRecord) bool {
	return a.Kind == b.Kind && a.Player == b.Player && a.Prompt == b.Prompt && a.How == b.How &&
		slices.Equal(a.Options, b.Options) && slices.Equal(a.Refs, b.Refs) && slices.Equal(a.Picks, b.Picks)
}

// actionOf rebuilds the Action an action event recorded. The kind is the first
// word of Detail.
func actionOf(e engine.Event) (engine.Action, error) {
	word, _, _ := strings.Cut(e.Detail, " ")
	a := engine.Action{Player: engine.PlayerID(e.Player), Asset: e.Source, Target: e.Target, Card: e.Inst}
	found := false
	for k := engine.ActGenerate; k <= engine.ActEndTurn; k++ {
		if k.String() == word {
			a.Kind, found = k, true
		}
	}
	if !found {
		return a, fmt.Errorf("unknown action %q", e.Detail)
	}
	if a.Kind == engine.ActGenerate && e.Resource == "Tactics" {
		a.Resource = engine.Tactics
	}
	if e.Zone != "" {
		zoneFound := false
		for z := engine.ZoneDeck; z <= engine.ZoneStack; z++ {
			if z.String() == e.Zone {
				a.Zone, zoneFound = z, true
			}
		}
		if !zoneFound {
			return a, fmt.Errorf("unknown zone %q", e.Zone)
		}
	}
	return a, nil
}
