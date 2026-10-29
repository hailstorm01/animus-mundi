// Events: the Observer, the event and match records, and the emit helpers.

package engine

import (
	"fmt"
	"slices"
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

	MulliganInsts    [2][]int `json:"mulliganInsts"`    // instance IDs of the cards set aside, never nil
	MulliganCallback bool     `json:"mulliganCallback"` // Config.Mulligan was set (a set callback shuffles, even if it sets nothing aside)
}

// MatchEnd is the last event of a game.
type MatchEnd struct {
	Draw   bool   `json:"draw"`
	Winner int    `json:"winner"` // -1 on a draw
	Reason string `json:"reason"`
	Turns  int    `json:"turns"`
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
			st.MulliganInsts[i] = slices.Clone(st.MulliganInsts[i])
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
		e.Resource = "Requisition"
		if a.Resource == Tactics {
			e.Resource = "Tactics"
		}
		e.Detail += " " + e.Resource
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
func (g *Game) emitMatchStart(cfg Config, decider PlayerID, playFirst bool, aside [2][]string, asideInsts [2][]int) {
	st := &MatchStart{
		Seed: cfg.Seed, Decks: cfg.DeckNames, Decider: int(decider), PlayFirst: playFirst, Mulligans: aside,
		MulliganInsts: asideInsts, MulliganCallback: cfg.Mulligan != nil,
	}
	for i, p := range g.Players {
		st.HQs[i] = p.HQ.Inst.Card.Name
		if st.Mulligans[i] == nil {
			st.Mulligans[i] = []string{}
		}
		if st.MulliganInsts[i] == nil {
			st.MulliganInsts[i] = []int{}
		}
	}
	st.Revision, st.Modified = BuildInfo()
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

// Event is an entry in the game log. Tests and the simulator read it; the
// kernel never reads it back.
type Event struct {
	Seq    int    `json:"seq"` // 1, 2, 3 ... per game, no gaps
	Turn   int    `json:"turn"`
	Phase  string `json:"phase"` // Phase.String()
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
	Player int    `json:"player"` // the acting player, -1 if none

	Card   string `json:"card,omitempty"`   // card name from card data
	Inst   int    `json:"inst,omitempty"`   // card instance ID
	Source int    `json:"source,omitempty"` // asset ID
	Target int    `json:"target,omitempty"` // asset ID
	Zone   string `json:"zone,omitempty"`   // Zone.String()
	Amount int    `json:"amount,omitempty"`

	Resource string `json:"resource,omitempty"` // action events: a Generate's "Requisition" or "Tactics"

	// Damage events only. Amount is the damage committed to Health (Stage 5).
	Raw      int  `json:"raw,omitempty"`      // the incoming damage before Stage 2
	Combat   bool `json:"combat,omitempty"`   // combat damage (Stalwart applies)
	Armor    int  `json:"armor,omitempty"`    // the target's Armor pool before this damage
	Absorbed int  `json:"absorbed,omitempty"` // Stage 4: damage absorbed by Armor

	Decision *DecisionRecord `json:"decision,omitempty"`
	Start    *MatchStart     `json:"start,omitempty"`    // Kind "match-start"
	Snapshot *Snapshot       `json:"snapshot,omitempty"` // Kind "turn-start"
	End      *MatchEnd       `json:"end,omitempty"`      // Kind "match-end"
}
