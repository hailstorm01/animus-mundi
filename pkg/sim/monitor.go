package sim

import (
	"fmt"
	"slices"
	"strings"

	"animus-mundi/pkg/engine"
)

// Rule monitors check FRD 01 rules on every game. A monitor is not a game
// reader: it sees only the Event and Settled calls plus the decklists' card
// data, so it cannot change a game (ADR-0005).

// Monitor names.
const (
	monAttack    = "attack-declaration"
	monEntry     = "entry-state"
	monResources = "resources"
	monDamage    = "damage-sequence"
	monTargeting = "targeting"
)

// monitorRules lists every rule ID each monitor names, sorted.
var monitorRules = map[string][]string{
	monAttack:    {"FR-CAP-001", "FR-CBT-001", "FR-KWD-002g", "FR-KWD-002j"},
	monEntry:     {"FR-KWD-002k", "FR-KWD-002l", "FR-SET-004", "FR-ZON-005"},
	monResources: {"FR-CAP-001", "FR-KWD-002i", "FR-RES-002", "FR-RES-003", "KIM-S03"},
	monDamage:    {"FR-CBT-002", "FR-KWD-002b", "FR-KWD-002c", "FR-KWD-002k", "FR-REP-004"},
	monTargeting: {"FR-KWD-002h", "FR-TGT-001", "SVY-R3"},
}

// Violation is the first broken rule a game's monitors found.
type Violation struct {
	Monitor string   // e.g. "attack-declaration"
	Rules   []string // every ID the monitor names
	Game    int      // counts from 1, as in -events
	Seed    uint64   // the game's seed, from its match-start
	Seq     int      // the event at which the rule broke
	Detail  string   // what was wrong, in words
}

func (v *Violation) Error() string {
	return fmt.Sprintf("rule monitor %s (%s) broke in game %d, seed %d, at event %d: %s",
		v.Monitor, strings.Join(v.Rules, ", "), v.Game, v.Seed, v.Seq, v.Detail)
}

// Coverage counts checks: monitor name -> rule ID -> number of checks.
type Coverage map[string]map[string]int

// add sums o into c.
func (c Coverage) add(o Coverage) {
	for mon, ids := range o {
		if c[mon] == nil {
			c[mon] = map[string]int{}
		}
		for id, n := range ids {
			c[mon][id] += n
		}
	}
}

// ReplayCommand is the two commands, one per line, that reach v. simArgs are
// the sim flags the run used, minus -seed, -games, -log and -events.
func ReplayCommand(simArgs []string, v *Violation) string {
	file := fmt.Sprintf("monitor-%d.jsonl", v.Seed)
	sim := append(append([]string{"cardtown", "sim"}, simArgs...),
		"-seed", fmt.Sprint(v.Seed), "-games", "1", "-events", file)
	return strings.Join(sim, " ") + "\n" +
		fmt.Sprintf("cardtown replay -events %s -game 1 -at %d", file, v.Seq)
}

// pendingCheck waits for the next settled snapshot after the event at seq.
type pendingCheck struct {
	seq      int
	asset    int  // the asset whose Target events void the check's armor part
	targeted bool // an event after seq had asset as its Target
	run      func(after engine.Snapshot, targeted bool)
}

// Monitors is one game's set of rule monitors, as one engine.Observer.
type Monitors struct {
	decks   [2]engine.DeckList
	game    int
	seed    uint64
	last    *engine.Snapshot // the most recent settled snapshot
	pending []pendingCheck
	v       *Violation
	cov     Coverage

	absorbed map[int]int    // asset ID -> Armor absorbed by damage events since the last settled snapshot
	combat   combatPair     // the combat event whose two damage events are still due
	targets  []targetPrompt // Target decisions waiting for the event that follows them
}

// NewMonitors makes the monitors for one game. decks gives each seat's card
// characteristics, looked up by the asset's owner and card name.
func NewMonitors(decks [2]engine.DeckList, game int) *Monitors {
	return &Monitors{decks: decks, game: game, cov: zeroCoverage(), absorbed: map[int]int{}}
}

func zeroCoverage() Coverage {
	c := Coverage{}
	for mon, ids := range monitorRules {
		c[mon] = map[string]int{}
		for _, id := range ids {
			c[mon][id] = 0
		}
	}
	return c
}

// Violation returns the first broken rule, nil while every rule held.
func (m *Monitors) Violation() *Violation { return m.v }

// Coverage returns every monitor and every rule ID it names, zero counts included.
func (m *Monitors) Coverage() Coverage {
	c := Coverage{}
	c.add(m.cov)
	return c
}

func (m *Monitors) count(mon string, ids ...string) {
	for _, id := range ids {
		m.cov[mon][id]++
	}
}

// fail records the game's first violation; later ones are ignored.
func (m *Monitors) fail(mon string, seq int, format string, args ...any) {
	if m.v != nil {
		return
	}
	m.v = &Violation{Monitor: mon, Rules: slices.Clone(monitorRules[mon]), Game: m.game, Seed: m.seed, Seq: seq,
		Detail: fmt.Sprintf(format, args...)}
}

// Event handles one event.
func (m *Monitors) Event(e engine.Event) {
	for i := range m.pending {
		if p := &m.pending[i]; e.Seq > p.seq && e.Target == p.asset {
			p.targeted = true
		}
	}
	if e.Kind != "decision" {
		m.endTargetPrompts(e)
	}
	switch e.Kind {
	case "decision":
		m.checkTargetDecision(e)
	case "combat":
		m.openCombat(e)
	case "damage":
		m.checkDamage(e)
	case "match-start":
		if e.Start != nil {
			m.seed = e.Start.Seed
		}
	case "turn-start":
		m.checkTurnStart(e)
	case "deploy", "scratch":
		m.watchEntry(e)
	case "action":
		switch kind, _, _ := strings.Cut(e.Detail, " "); kind {
		case "Attack":
			m.checkAttack(e)
		case "Generate":
			m.checkGenerate(e)
		case "Convert":
			m.checkConvert(e)
		}
	}
}

// Settled runs the checks that were waiting for a snapshot, then keeps it as
// the "before" of the next action.
func (m *Monitors) Settled(s engine.Snapshot) {
	pend := m.pending
	m.pending = nil
	for _, p := range pend {
		p.run(s, p.targeted)
	}
	m.last = &s
	clear(m.absorbed)
}

// find returns the battlefield asset with id in s, nil if there is none.
func find(s engine.Snapshot, id int) *engine.AssetSnap {
	var found *engine.AssetSnap
	eachAsset(s, func(a *engine.AssetSnap) bool {
		if a.ID == id {
			found = a
		}
		return found == nil
	})
	return found
}

// eachAsset calls f on every battlefield asset in s, the Main Theater first,
// then each player's Reserves and Headquarters, until f returns false.
func eachAsset(s engine.Snapshot, f func(*engine.AssetSnap) bool) {
	for i := range s.Theater {
		if !f(&s.Theater[i]) {
			return
		}
	}
	for p := range s.Players {
		for i := range s.Players[p].Reserves {
			if !f(&s.Players[p].Reserves[i]) {
				return
			}
		}
		if hq := s.Players[p].HQ; hq != nil && !f(hq) {
			return
		}
	}
}

// card is a's printed characteristics: the Scratch Company face if face-down,
// else the owner's decklist card of that name. nil if the decklist lacks it.
func (m *Monitors) card(a engine.AssetSnap) *engine.Card {
	if a.FaceDown {
		c := engine.ScratchFace()
		return &c
	}
	d := m.decks[a.Owner]
	if d.HQ != nil && d.HQ.Name == a.Card {
		return d.HQ
	}
	for _, c := range d.Main {
		if c.Name == a.Card {
			return c
		}
	}
	return nil
}
