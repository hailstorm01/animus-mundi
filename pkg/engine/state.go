package engine

import (
	"fmt"
	"math/rand/v2"
)

// PlayerID identifies a seat. Players are 0 and 1.
type PlayerID int

// Opponent returns the other seat.
func (p PlayerID) Opponent() PlayerID { return 1 - p }

// Zone enumerates game zones (FR-ZON-001).
type Zone int

const (
	ZoneDeck Zone = iota
	ZoneHand
	ZoneDiscard
	ZoneReserves
	ZoneMainTheater
	ZoneHeadquarters
	ZoneStack
)

func (z Zone) String() string {
	return [...]string{"Deck", "Hand", "Discard", "Reserves", "Main Theater", "Headquarters Zone", "Resolution Stack"}[z]
}

// IsBattlefield reports whether entities in z are Assets (FR-ZON-003).
func (z Zone) IsBattlefield() bool {
	return z == ZoneReserves || z == ZoneMainTheater || z == ZoneHeadquarters
}

// Phase is the current turn phase (FR-TRN-001).
type Phase int

const (
	PhaseSetup Phase = iota
	PhaseReady
	PhaseDraw
	PhaseMain
	PhaseEnd
	PhaseOver
)

// Instance is one physical card for the whole match. Its ID never changes,
// which is what delayed triggers bind to after the card leaves the battlefield.
type Instance struct {
	ID    int
	Card  *Card
	Owner PlayerID
}

// Asset is an Instance while it occupies a battlefield zone. A new Asset is
// created every time a card enters the battlefield, so state from a previous
// visit can never leak back (FR-ZON-004).
type Asset struct {
	ID         int
	Inst       *Instance
	Controller PlayerID
	Zone       Zone
	Exhausted  bool
	FaceDown   bool
	Damage     int
	Armor      int

	// TempPower is the Layer 5 "until end of turn" power modifier.
	TempPower int
}

// Printed returns the Layer 1 characteristics: the card face, or the 1/1/1/1
// Infantry Scratch Company if face-down.
func (a *Asset) Printed() *Card {
	if a.FaceDown {
		return scratchFace
	}
	return a.Inst.Card
}

func (a *Asset) String() string {
	name := a.Printed().Name
	if a.FaceDown {
		name = "face-down Scratch Company"
	}
	return fmt.Sprintf("%s#%d", name, a.ID)
}

// Player holds one seat's private and per-player zones.
type Player struct {
	ID       PlayerID
	HQ       *Asset
	Deck     []*Instance // index 0 is the top of the deck
	Hand     []*Instance
	Discard  []*Instance
	Reserves []*Asset
	Req      int
	Tac      int

	// DrewFromEmpty records a required draw from an empty deck since the last
	// death sweep; the sweep turns it into a loss (FR-WIN-001).
	DrewFromEmpty bool
}

// Outcome is the match result once the game is over.
type Outcome struct {
	Over   bool
	Draw   bool
	Winner PlayerID
	Reason string
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

	Secondary int    `json:"secondary,omitempty"` // action events: an Attack's Berserk secondary target (asset ID)
	Resource  string `json:"resource,omitempty"`  // action events: a Generate's "Requisition" or "Tactics"

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

// Game is the complete match state.
type Game struct {
	Players     [2]*Player
	Theater     []*Asset // the shared Main Theater
	Stack       []*StackItem
	Active      PlayerID
	FirstPlayer PlayerID
	Turn        int
	Phase       Phase
	Outcome     Outcome
	Log         []Event

	rng      *rand.Rand
	nextID   int
	choosers [2]Chooser // per-seat answerers of Decisions; nil always times out
	obs      Observer   // optional read-only observer of events and settled states

	// lastKnown holds each asset's characteristics as it left the battlefield,
	// for effects that still need them afterwards (FR-ZON-004).
	lastKnown map[int]*Asset

	pending []pendingTrigger // triggers fired while an item resolves, awaiting the flush (FR-TRG-004)
}

// LastKnown returns the asset as it was when it last left the battlefield,
// or nil if it never has.
func (g *Game) LastKnown(id int) *Asset { return g.lastKnown[id] }

// leaveBattlefield removes a from the battlefield to a non-battlefield zone,
// keeping a last-known snapshot.
func (g *Game) leaveBattlefield(a *Asset) {
	if g.lastKnown == nil {
		g.lastKnown = make(map[int]*Asset)
	}
	snap := *a
	g.lastKnown[a.ID] = &snap
	g.removeFromBattlefield(a)
}

func (g *Game) newID() int {
	g.nextID++
	return g.nextID
}

// Assets returns every asset on the battlefield, HQs first.
func (g *Game) Assets() []*Asset {
	var out []*Asset
	for _, p := range g.Players {
		if p.HQ != nil {
			out = append(out, p.HQ)
		}
	}
	for _, p := range g.Players {
		out = append(out, p.Reserves...)
	}
	return append(out, g.Theater...)
}

// Asset finds a battlefield asset by ID, or nil if it is no longer there.
func (g *Game) Asset(id int) *Asset {
	for _, a := range g.Assets() {
		if a.ID == id {
			return a
		}
	}
	return nil
}

// TheaterOf returns the Main Theater assets controlled by p.
func (g *Game) TheaterOf(p PlayerID) []*Asset {
	var out []*Asset
	for _, a := range g.Theater {
		if a.Controller == p {
			out = append(out, a)
		}
	}
	return out
}

func remove[T comparable](s []T, v T) []T {
	for i, x := range s {
		if x == v {
			return append(s[:i:i], s[i+1:]...)
		}
	}
	return s
}

// removeFromBattlefield detaches a from whichever battlefield zone holds it.
func (g *Game) removeFromBattlefield(a *Asset) {
	switch a.Zone {
	case ZoneMainTheater:
		g.Theater = remove(g.Theater, a)
	case ZoneReserves:
		p := g.Players[a.Controller]
		p.Reserves = remove(p.Reserves, a)
	case ZoneHeadquarters:
		g.Players[a.Controller].HQ = nil
	}
}

// placeOnBattlefield puts a into zone under its controller.
func (g *Game) placeOnBattlefield(a *Asset, zone Zone) {
	a.Zone = zone
	switch zone {
	case ZoneMainTheater:
		g.Theater = append(g.Theater, a)
	case ZoneReserves:
		p := g.Players[a.Controller]
		p.Reserves = append(p.Reserves, a)
	case ZoneHeadquarters:
		g.Players[a.Controller].HQ = a
	}
}
