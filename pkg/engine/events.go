package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"slices"
	"strings"
)

// Observer receives a game's events. It must not change the game: it gets
// copies, never the *Game, and has no access to the random source.
type Observer interface {
	Event(e Event)      // every event, in Seq order, as it happens
	Settled(s Snapshot) // after each settled step: setup, each resolved stack item, each Apply
}

// DecisionRecord is one answered Decision, as the log stores it.
type DecisionRecord struct {
	Kind    string   `json:"kind"` // DecisionKind name
	Player  int      `json:"player"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"` // the option labels offered, in order
	Refs    []int    `json:"refs"`    // the option refs, in order
	Picks   []int    `json:"picks"`   // indexes into Options, as applied
	How     string   `json:"how"`     // "answer", "auto" (fast path), "timeout" (fallback), "invalid" (fallback after an invalid answer)
}

// MatchStart is the first event of a game.
type MatchStart struct {
	Seed      uint64      `json:"seed"`
	Decks     [2]string   `json:"decks"` // Config.DeckNames; "" if unset
	HQs       [2]string   `json:"hqs"`   // HQ card names
	Decider   int         `json:"decider"`
	PlayFirst bool        `json:"playFirst"` // the decider's choice
	Mulligans [2][]string `json:"mulligans"` // names of the cards each player set aside
	Revision  string      `json:"revision"`  // vcs.revision from the build info, "" if absent
	Modified  bool        `json:"modified"`  // vcs.modified
}

// MatchEnd is the last event of a game.
type MatchEnd struct {
	Draw   bool   `json:"draw"`
	Winner int    `json:"winner"` // -1 on a draw
	Reason string `json:"reason"`
	Turns  int    `json:"turns"`
}

// CardSnap is a card instance in a hidden or off-battlefield zone.
type CardSnap struct {
	Inst int    `json:"inst"`
	Card string `json:"card"`
}

// AssetSnap is a battlefield asset. Card is always the real card name, even
// when the asset is face-down; Power and MaxHealth are effective (Layer 6).
type AssetSnap struct {
	ID         int    `json:"id"`
	Inst       int    `json:"inst"`
	Card       string `json:"card"`
	Owner      int    `json:"owner"`
	Controller int    `json:"controller"`
	Zone       string `json:"zone"`
	Exhausted  bool   `json:"exhausted"`
	FaceDown   bool   `json:"faceDown"`
	Damage     int    `json:"damage"`
	Armor      int    `json:"armor"`
	TempPower  int    `json:"tempPower"`
	Power      int    `json:"power"`
	MaxHealth  int    `json:"maxHealth"`
}

// PlayerSnap is one seat in a Snapshot.
type PlayerSnap struct {
	HQ       *AssetSnap  `json:"hq"`
	Reserves []AssetSnap `json:"reserves"`
	Deck     []CardSnap  `json:"deck"` // index 0 is the top
	Hand     []CardSnap  `json:"hand"`
	Discard  []CardSnap  `json:"discard"`
	Req      int         `json:"req"`
	Tac      int         `json:"tac"`
}

// Snapshot is a deep copy of the full state, hidden information included:
// both hands, both deck orders and face-down identities.
type Snapshot struct {
	Turn     int           `json:"turn"`
	Phase    string        `json:"phase"`
	Active   int           `json:"active"`
	First    int           `json:"first"`
	Over     bool          `json:"over"`
	Draw     bool          `json:"draw"`
	Winner   int           `json:"winner"` // -1 until a player has won
	Reason   string        `json:"reason"`
	StackLen int           `json:"stackLen"`
	Players  [2]PlayerSnap `json:"players"`
	Theater  []AssetSnap   `json:"theater"`
}

func (p Phase) String() string {
	return [...]string{"setup", "ready", "draw", "main", "end", "over"}[p]
}

func (k DecisionKind) String() string {
	return [...]string{"Target", "Mode", "Optional", "Order"}[k]
}

// emit numbers e, stamps the turn and phase, appends it to Log (the default
// record) and passes a copy to the observer if one is set.
func (g *Game) emit(e Event) {
	e.Seq = len(g.Log) + 1
	e.Turn = g.Turn
	e.Phase = g.Phase.String()
	g.Log = append(g.Log, e)
	if g.obs != nil {
		g.obs.Event(e.clone())
	}
}

// logEvent emits e with its Detail formatted from format and args.
func (g *Game) logEvent(e Event, format string, args ...any) {
	e.Detail = fmt.Sprintf(format, args...)
	g.emit(e)
}

// logf emits an event with no structured fields beyond its kind and detail.
func (g *Game) logf(kind, format string, args ...any) {
	g.logEvent(Event{Kind: kind, Player: -1}, format, args...)
}

// settled tells the observer the state is settled: nothing is resolving and the
// death sweep and trigger flush are done. The snapshot is built only if needed.
func (g *Game) settled() {
	if g.obs != nil {
		g.obs.Settled(g.snapshot())
	}
}

// cardName is the name of inst's card, "" for nil.
func cardName(inst *Instance) string {
	if inst == nil {
		return ""
	}
	return inst.Card.Name
}

func assetID(a *Asset) int {
	if a == nil {
		return 0
	}
	return a.ID
}

func (g *Game) assetSnap(a *Asset) AssetSnap {
	return AssetSnap{
		ID: a.ID, Inst: a.Inst.ID, Card: a.Inst.Card.Name,
		Owner: int(a.Inst.Owner), Controller: int(a.Controller), Zone: a.Zone.String(),
		Exhausted: a.Exhausted, FaceDown: a.FaceDown, Damage: a.Damage, Armor: a.Armor,
		TempPower: a.TempPower, Power: g.Power(a), MaxHealth: g.MaxHealth(a),
	}
}

func (g *Game) assetSnaps(as []*Asset) []AssetSnap {
	out := make([]AssetSnap, len(as))
	for i, a := range as {
		out[i] = g.assetSnap(a)
	}
	return out
}

func cardSnaps(cards []*Instance) []CardSnap {
	out := make([]CardSnap, len(cards))
	for i, c := range cards {
		out[i] = CardSnap{Inst: c.ID, Card: c.Card.Name}
	}
	return out
}

// snapshot deep-copies the full game state.
func (g *Game) snapshot() Snapshot {
	s := Snapshot{
		Turn: g.Turn, Phase: g.Phase.String(), Active: int(g.Active), First: int(g.FirstPlayer),
		Over: g.Outcome.Over, Draw: g.Outcome.Draw, Winner: -1, Reason: g.Outcome.Reason,
		StackLen: len(g.Stack), Theater: g.assetSnaps(g.Theater),
	}
	if g.Outcome.Over && !g.Outcome.Draw {
		s.Winner = int(g.Outcome.Winner)
	}
	for i, p := range g.Players {
		ps := PlayerSnap{
			Reserves: g.assetSnaps(p.Reserves), Deck: cardSnaps(p.Deck), Hand: cardSnaps(p.Hand),
			Discard: cardSnaps(p.Discard), Req: p.Req, Tac: p.Tac,
		}
		if p.HQ != nil {
			hq := g.assetSnap(p.HQ)
			ps.HQ = &hq
		}
		s.Players[i] = ps
	}
	return s
}

func (s Snapshot) clone() Snapshot {
	for i, p := range s.Players {
		if p.HQ != nil {
			hq := *p.HQ
			p.HQ = &hq
		}
		p.Reserves = slices.Clone(p.Reserves)
		p.Deck = slices.Clone(p.Deck)
		p.Hand = slices.Clone(p.Hand)
		p.Discard = slices.Clone(p.Discard)
		s.Players[i] = p
	}
	s.Theater = slices.Clone(s.Theater)
	return s
}

// clone copies e and everything it points to, so an observer that edits what it
// receives cannot change the Log.
func (e Event) clone() Event {
	if e.Decision != nil {
		d := *e.Decision
		d.Options, d.Refs, d.Picks = slices.Clone(d.Options), slices.Clone(d.Refs), slices.Clone(d.Picks)
		e.Decision = &d
	}
	if e.Start != nil {
		st := *e.Start
		for i := range st.Mulligans {
			st.Mulligans[i] = slices.Clone(st.Mulligans[i])
		}
		e.Start = &st
	}
	if e.Snapshot != nil {
		s := e.Snapshot.clone()
		e.Snapshot = &s
	}
	if e.End != nil {
		end := *e.End
		e.End = &end
	}
	return e
}

// emitDecision records the picks choose is about to return for d.
func (g *Game) emitDecision(d Decision, picks []int, how string) {
	rec := &DecisionRecord{
		Kind: d.Kind.String(), Player: int(d.Player), Prompt: d.Prompt, How: how,
		Options: make([]string, len(d.Options)), Refs: make([]int, len(d.Options)),
		Picks: slices.Clone(picks),
	}
	if rec.Picks == nil {
		rec.Picks = []int{}
	}
	for i, o := range d.Options {
		rec.Options[i], rec.Refs[i] = o.Label, o.Ref
	}
	g.emit(Event{Kind: "decision", Detail: fmt.Sprintf("player %d decides %q: %v (%s)", d.Player, d.Prompt, rec.Picks, how),
		Player: int(d.Player), Decision: rec})
}

// emitAction records an accepted action, before its effects.
func (g *Game) emitAction(a Action) {
	e := Event{Kind: "action", Detail: a.Kind.String(), Player: int(a.Player)}
	switch a.Kind {
	case ActGenerate:
		e.Source = a.Asset
		e.Detail += " Requisition"
		if a.Resource == Tactics {
			e.Detail = a.Kind.String() + " Tactics"
		}
	case ActMarch, ActScratch:
		e.Source = a.Asset
	case ActAttack:
		e.Source, e.Target = a.Asset, a.Target
	case ActDeploy:
		if inst := handCard(g.Players[a.Player], a.Card); inst != nil {
			e.Card, e.Inst = inst.Card.Name, inst.ID
		}
		e.Zone = a.Zone.String()
	}
	g.emit(e)
}

// emitMatchStart is the first event of the game, after setup and the mulligan.
func (g *Game) emitMatchStart(cfg Config, decider PlayerID, playFirst bool, aside [2][]string) {
	st := &MatchStart{
		Seed: cfg.Seed, Decks: cfg.DeckNames, Decider: int(decider), PlayFirst: playFirst, Mulligans: aside,
	}
	for i, p := range g.Players {
		st.HQs[i] = p.HQ.Inst.Card.Name
		if st.Mulligans[i] == nil {
			st.Mulligans[i] = []string{}
		}
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				st.Revision = s.Value
			case "vcs.modified":
				st.Modified = s.Value == "true"
			}
		}
	}
	g.emit(Event{Kind: "match-start", Detail: fmt.Sprintf("seed %d, player %d decides, play first %t", cfg.Seed, decider, playFirst),
		Player: -1, Start: st})
}

// emitTurnStart carries the full-state snapshot for the turn that begins now.
func (g *Game) emitTurnStart() {
	s := g.snapshot()
	g.emit(Event{Kind: "turn-start", Detail: fmt.Sprintf("turn %d begins for player %d", g.Turn, g.Active),
		Player: int(g.Active), Snapshot: &s})
}

// emitMatchEnd is emitted once, right after the result event.
func (g *Game) emitMatchEnd(o Outcome) {
	end := &MatchEnd{Draw: o.Draw, Winner: -1, Reason: o.Reason, Turns: g.Turn}
	if !o.Draw {
		end.Winner = int(o.Winner)
	}
	g.emit(Event{Kind: "match-end", Detail: o.Reason, Player: -1, End: end})
}

// FormatEvent is one readable line for e, with no newline, starting with the
// Seq. The full snapshot of a turn-start is never printed.
func FormatEvent(e Event) string {
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
func WriteEventsJSONL(w io.Writer, game int, events []Event) error {
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
