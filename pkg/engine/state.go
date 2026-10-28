// Core state: the game, player and asset types and their lookups, with last-known information (FR-ZON-004).

package engine

import (
	"fmt"
	"math/rand/v2"
	"slices"
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

func (p Phase) String() string {
	return [...]string{"setup", "ready", "draw", "main", "end", "over"}[p]
}

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

	arrived int // arrival number: when it entered its current battlefield zone, counted across the match (SVY-R7)
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
	arrivals int
	choosers [2]Chooser // per-seat answerers of Decisions; nil always times out
	obs      Observer   // optional read-only observer of events and settled states

	// lastKnown holds each asset's characteristics as it left the battlefield,
	// for effects that still need them afterwards (FR-ZON-004).
	lastKnown map[int]*lastKnownRecord

	pending []pendingTrigger // triggers fired while an item resolves, awaiting the flush (FR-TRG-004)

	// firings counts the firings of limited triggers (Trigger.MaxPerTurn), per asset, per
	// trigger and per turn. A new turn number is a fresh count, which is how the limits
	// reset every turn (FR-TRN-001). Not part of snapshots or events.
	firings map[firingKey]int
}

// firingKey names one trigger of one asset in one turn.
type firingKey struct {
	asset, trigger, turn int
}

// EffectiveState is an asset's effective characteristics at one moment, read
// through the layer pipeline (FR-LAY-001).
type EffectiveState struct {
	Power     int
	MaxHealth int
	Keywords  Keywords
}

// effectiveState reads a's effective characteristics now.
func (g *Game) effectiveState(a *Asset) EffectiveState {
	return EffectiveState{Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: g.Keywords(a)}
}

// lastKnownRecord is an asset as it left the battlefield: a raw copy and its
// effective characteristics immediately prior to departure (FR-ZON-004, board
// directive #13).
type lastKnownRecord struct {
	asset Asset
	eff   EffectiveState
}

// LastKnown returns the asset as it was when it last left the battlefield,
// or nil if it never has.
func (g *Game) LastKnown(id int) *Asset {
	if r := g.lastKnown[id]; r != nil {
		return &r.asset
	}
	return nil
}

// LastKnownEffective returns the effective characteristics the asset had
// immediately prior to leaving the battlefield, or nil if it never has. Death
// triggers and other leaves-battlefield effects read these (FR-ZON-004).
func (g *Game) LastKnownEffective(id int) *EffectiveState {
	if r := g.lastKnown[id]; r != nil {
		return &r.eff
	}
	return nil
}

// newID returns the next ID from the game's one counter. Asset IDs and card
// instance IDs both come from it, so they never collide, and three things rely
// on that: fixedOrderPositions keys asset IDs and instance IDs in one map, and
// Option.Ref and StackItem.Chosen each hold either kind.
func (g *Game) newID() int {
	g.nextID++
	return g.nextID
}

// Battlefield orders. The battlefield is listed in exactly three orders, and
// each rule that walks it uses one of them. New code picks one of these three
// rather than adding a fourth.
//
// Assets: the Headquarters, then every player's Reserves, then the Main
// Theater, each zone in arrival order. The death sweep uses it. Simultaneous
// triggers are placed in the same order, by where each source stood when it
// fired (sourcePos in trigger.go), so a step that visits assets in another
// order doesn't change it (SVY-R7, FR-TRG-003). Uses where order doesn't
// matter (lookups, readying, invariants, legal actions) also use it.
//
// assetsByID: oldest (lowest ID) first. Decision option lists and "all" steps
// use it (SVY-R11).
//
// fixedOrderPositions: the choosing player's assets before the opponent's;
// within a player the Headquarters, then Reserves, then the Main Theater, each
// in arrival order. Then the cards in discard piles, after every battlefield
// asset: the choosing player's pile from the top down, then the opponent's, from
// the top down (the top of a pile is its last card). Timeout fallbacks use it
// (SVY-R16, FR-MOD-001 §6, directive #59 (1)).

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

// assetsByID returns every battlefield asset, oldest (lowest ID) first.
func (g *Game) assetsByID() []*Asset {
	assets := g.Assets()
	slices.SortFunc(assets, func(x, y *Asset) int { return x.ID - y.ID })
	return assets
}

// fixedOrderPositions maps each battlefield asset ID to its position in the
// FR-MOD-001 §6 fixed order for choosing player p (SVY-R16): p's assets before the
// opponent's; within a player the Headquarters, then Reserves, then the Main Theater
// assets they control, each zone in arrival order. Cards in discard piles
// follow, p's pile from the top down, then the opponent's, from the top down.
func (g *Game) fixedOrderPositions(p PlayerID) map[int]int {
	pos := map[int]int{}
	add := func(a *Asset) { pos[a.ID] = len(pos) }
	seat := func(id PlayerID) {
		pl := g.Players[id]
		if pl.HQ != nil {
			add(pl.HQ)
		}
		for _, a := range pl.Reserves {
			add(a)
		}
		for _, a := range g.TheaterOf(id) {
			add(a)
		}
	}
	if int(p) >= 0 && int(p) < len(g.Players) {
		seat(p)
	}
	for i := range g.Players {
		if PlayerID(i) != p {
			seat(PlayerID(i))
		}
	}
	for _, id := range []PlayerID{p, p.Opponent()} {
		if int(id) < 0 || int(id) >= len(g.Players) {
			continue
		}
		pile := g.Players[id].Discard
		for i := len(pile) - 1; i >= 0; i-- {
			pos[pile[i].ID] = len(pos)
		}
	}
	return pos
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
