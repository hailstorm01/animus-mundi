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
)

// monitorRules lists every rule ID each monitor names, sorted.
var monitorRules = map[string][]string{
	monAttack:    {"FR-CAP-001", "FR-CBT-001", "FR-KWD-002g", "FR-KWD-002j"},
	monEntry:     {"FR-KWD-002k", "FR-KWD-002l", "FR-SET-004", "FR-ZON-005"},
	monResources: {"FR-CAP-001", "FR-KWD-002i", "FR-RES-002", "FR-RES-003", "KIM-S03"},
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

// scratchFace is the face-down Scratch Company face (FR-ZON-005): a Unit, 1/1,
// yields 1/1, no keywords.
var scratchFace = &engine.Card{Name: "Scratch Company", Type: engine.TypeUnit, Power: 1, Health: 1, ReqYield: 1, TacYield: 1}

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
}

// NewMonitors makes the monitors for one game. decks gives each seat's card
// characteristics, looked up by the asset's owner and card name.
func NewMonitors(decks [2]engine.DeckList, game int) *Monitors {
	return &Monitors{decks: decks, game: game, cov: zeroCoverage()}
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
	switch e.Kind {
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
}

// find returns the battlefield asset with id in s, nil if there is none.
func find(s engine.Snapshot, id int) *engine.AssetSnap {
	for i := range s.Theater {
		if s.Theater[i].ID == id {
			return &s.Theater[i]
		}
	}
	for p := range s.Players {
		for i := range s.Players[p].Reserves {
			if s.Players[p].Reserves[i].ID == id {
				return &s.Players[p].Reserves[i]
			}
		}
		if hq := s.Players[p].HQ; hq != nil && hq.ID == id {
			return hq
		}
	}
	return nil
}

// card is a's printed characteristics: the Scratch Company face if face-down,
// else the owner's decklist card of that name. nil if the decklist lacks it.
func (m *Monitors) card(a engine.AssetSnap) *engine.Card {
	if a.FaceDown {
		return scratchFace
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

func (m *Monitors) checkTurnStart(e engine.Event) {
	m.count(monResources, "FR-RES-002")
	if e.Snapshot == nil {
		return
	}
	for p, ps := range e.Snapshot.Players {
		if ps.Req != 0 || ps.Tac != 0 {
			m.fail(monResources, e.Seq, "player %d starts turn %d with Requisition %d and Tactics %d, want both 0", p, e.Snapshot.Turn, ps.Req, ps.Tac)
		}
	}
}

// opponentAssets are the assets p's opponent controls in the Main Theater and
// the Headquarters Zone.
func opponentAssets(s engine.Snapshot, p int) []engine.AssetSnap {
	var out []engine.AssetSnap
	for _, a := range s.Theater {
		if a.Controller != p {
			out = append(out, a)
		}
	}
	if hq := s.Players[1-p].HQ; hq != nil {
		out = append(out, *hq)
	}
	return out
}

func (m *Monitors) checkAttack(e engine.Event) {
	const mon = monAttack
	m.count(mon, "FR-CBT-001", "FR-CAP-001")
	b := m.last
	if b == nil {
		m.fail(mon, e.Seq, "attack declared before any settled snapshot")
		return
	}
	main, theater, hqZone := "main", engine.ZoneMainTheater.String(), engine.ZoneHeadquarters.String()
	if b.Phase != main || b.StackLen != 0 {
		m.fail(mon, e.Seq, "attack declared in phase %s with %d items on the stack, want phase %s and an empty stack", b.Phase, b.StackLen, main)
	}
	if e.Player != b.Active {
		m.fail(mon, e.Seq, "player %d attacked on player %d's turn", e.Player, b.Active)
	}
	att, def := find(*b, e.Source), find(*b, e.Target)
	if att == nil || def == nil {
		m.fail(mon, e.Seq, "attacker %d or defender %d is not on the battlefield", e.Source, e.Target)
		return
	}
	ac, dc := m.card(*att), m.card(*def)
	if ac == nil || dc == nil {
		m.fail(mon, e.Seq, "attacker %s or defender %s is not in its owner's decklist", att.Card, def.Card)
		return
	}
	if att.Controller != e.Player {
		m.fail(mon, e.Seq, "%s#%d is not controlled by the attacking player %d", att.Card, att.ID, e.Player)
	}
	if att.Exhausted {
		m.fail(mon, e.Seq, "attacker %s#%d is Exhausted", att.Card, att.ID)
	}
	ownHQ := b.Players[e.Player].HQ
	isHQ := ownHQ != nil && ownHQ.ID == att.ID
	if !(ac.Type == engine.TypeUnit && att.Zone == theater) && !(isHQ && att.Zone == hqZone) {
		m.fail(mon, e.Seq, "attacker %s#%d (%s in %s) must be a Unit in the Main Theater or the Headquarters in its zone", att.Card, att.ID, ac.Type, att.Zone)
	}

	if def.Controller != 1-e.Player {
		m.fail(mon, e.Seq, "defender %s#%d is not controlled by the opponent", def.Card, def.ID)
	}
	oppHQ := b.Players[1-e.Player].HQ
	defIsHQ := oppHQ != nil && oppHQ.ID == def.ID
	if !((dc.Type == engine.TypeUnit || dc.Type == engine.TypeStructure) && def.Zone == theater) && !(defIsHQ && def.Zone == hqZone) {
		m.fail(mon, e.Seq, "defender %s#%d (%s in %s) must be a Unit or Structure in the Main Theater or the opposing Headquarters in its zone", def.Card, def.ID, dc.Type, def.Zone)
	}

	var guardians []int
	for _, o := range opponentAssets(*b, e.Player) {
		oc := m.card(o)
		if oc == nil {
			continue
		}
		if oc.Keywords.Ethereal {
			m.count(mon, "FR-KWD-002g")
		} else if oc.Keywords.Guardian && o.Zone == theater {
			guardians = append(guardians, o.ID)
		}
	}
	if dc.Keywords.Ethereal {
		m.fail(mon, e.Seq, "defender %s#%d is Ethereal", def.Card, def.ID)
	}
	if len(guardians) > 0 {
		m.count(mon, "FR-KWD-002j")
		if !slices.Contains(guardians, def.ID) {
			m.fail(mon, e.Seq, "defender %s#%d is not a Guardian while the opponent has Guardians %v in the Main Theater", def.Card, def.ID, guardians)
		}
	}
}

// pool is player p's Requisition or Tactics in s.
func pool(s engine.Snapshot, p int, resource string) int {
	if resource == "Requisition" {
		return s.Players[p].Req
	}
	return s.Players[p].Tac
}

func (m *Monitors) checkGenerate(e engine.Event) {
	const mon = monResources
	m.count(mon, "FR-CAP-001")
	b := m.last
	if b == nil {
		m.fail(mon, e.Seq, "Generate before any settled snapshot")
		return
	}
	a := find(*b, e.Source)
	if a == nil {
		m.fail(mon, e.Seq, "Generate from asset %d, which is not on the battlefield", e.Source)
		return
	}
	c := m.card(*a)
	if c == nil {
		m.fail(mon, e.Seq, "%s is not in its owner's decklist", a.Card)
		return
	}
	if a.Controller != e.Player {
		m.fail(mon, e.Seq, "%s#%d is not controlled by the generating player %d", a.Card, a.ID, e.Player)
	}
	if a.Exhausted {
		m.fail(mon, e.Seq, "%s#%d was Exhausted before it generated", a.Card, a.ID)
	}
	res, other := e.Resource, "Tactics"
	yield := c.TacYield
	if res == "Requisition" {
		other, yield = "Tactics", c.ReqYield
	} else if res == "Tactics" {
		other = "Requisition"
	} else {
		m.fail(mon, e.Seq, "Generate names resource %q", res)
		return
	}
	if c.Type == engine.TypeUnit {
		reserves := engine.ZoneReserves.String()
		switch {
		case c.Keywords.Versatile:
			m.count(mon, "FR-KWD-002i")
		case a.Zone == reserves && res != "Requisition":
			m.fail(mon, e.Seq, "%s#%d in Reserves generated %s, not Requisition", a.Card, a.ID, res)
		case a.Zone != reserves && res != "Tactics":
			m.fail(mon, e.Seq, "%s#%d in %s generated %s, not Tactics", a.Card, a.ID, a.Zone, res)
		}
	}
	if yield == 0 {
		m.count(mon, "KIM-S03")
	}
	p, id, name := e.Player, a.ID, a.Card
	beforeRes, beforeOther := pool(*b, p, res), pool(*b, p, other)
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, run: func(after engine.Snapshot, _ bool) {
		if n := find(after, id); n == nil || !n.Exhausted {
			m.fail(mon, e.Seq, "%s#%d is not Exhausted after it generated", name, id)
		}
		if got := pool(after, p, res) - beforeRes; got != yield {
			m.fail(mon, e.Seq, "%s rose by %d after %s#%d generated, want its yield %d", res, got, name, id, yield)
		}
		if got := pool(after, p, other) - beforeOther; got != 0 {
			m.fail(mon, e.Seq, "%s changed by %d when %s#%d generated %s", other, got, name, id, res)
		}
	}})
}

func (m *Monitors) checkConvert(e engine.Event) {
	const mon = monResources
	m.count(mon, "FR-RES-003")
	b := m.last
	if b == nil {
		m.fail(mon, e.Seq, "Convert before any settled snapshot")
		return
	}
	p := e.Player
	req, tac := b.Players[p].Req, b.Players[p].Tac
	if req < 2 {
		m.fail(mon, e.Seq, "player %d converted with %d Requisition, want at least 2", p, req)
	}
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, run: func(after engine.Snapshot, _ bool) {
		if after.Players[p].Req != req-2 || after.Players[p].Tac != tac+1 {
			m.fail(mon, e.Seq, "Convert took player %d from R%d T%d to R%d T%d, want R%d T%d",
				p, req, tac, after.Players[p].Req, after.Players[p].Tac, req-2, tac+1)
		}
	}})
}

// watchEntry queues the entry-state check for a deploy or scratch event.
func (m *Monitors) watchEntry(e engine.Event) {
	const mon = monEntry
	m.pending = append(m.pending, pendingCheck{seq: e.Seq, asset: e.Source, run: func(after engine.Snapshot, targeted bool) {
		a := find(after, e.Source)
		if a == nil {
			return // left the battlefield before the state settled: skip
		}
		if e.Kind == "scratch" {
			m.count(mon, "FR-ZON-005")
			if a.Zone != engine.ZoneReserves.String() || a.Controller != e.Player || !a.FaceDown || !a.Exhausted {
				m.fail(mon, e.Seq, "Scratch Company %d entered as zone %s, controller %d, face-down %t, Exhausted %t; want Reserves, controller %d, face-down, Exhausted",
					a.ID, a.Zone, a.Controller, a.FaceDown, a.Exhausted, e.Player)
			}
			return
		}
		c := m.card(*a)
		if c == nil {
			m.fail(mon, e.Seq, "%s is not in its owner's decklist", a.Card)
			return
		}
		if a.FaceDown || a.Zone != e.Zone {
			m.fail(mon, e.Seq, "%s#%d entered face-down %t in %s, want face-up in %s", a.Card, a.ID, a.FaceDown, a.Zone, e.Zone)
		}
		if c.Keywords.DeploysExhausted {
			m.count(mon, "FR-KWD-002l")
		} else {
			m.count(mon, "FR-SET-004")
		}
		if a.Exhausted != c.Keywords.DeploysExhausted {
			m.fail(mon, e.Seq, "%s#%d entered with Exhausted %t, want %t", a.Card, a.ID, a.Exhausted, c.Keywords.DeploysExhausted)
		}
		if n := c.Keywords.Armor; n > 0 && !targeted {
			m.count(mon, "FR-KWD-002k")
			if a.Armor != n {
				m.fail(mon, e.Seq, "%s#%d entered with Armor %d, want its printed Armor %d", a.Card, a.ID, a.Armor, n)
			}
		}
	}})
}
