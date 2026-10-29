// Package view builds what one seat may see of a game: PlayerView for the
// state and LogLines for the event log (board directive #24, FR-ZON-005).
//
// It is a client of the engine's exported API, like sim and eventlog: it
// imports engine and eventlog only, and engine never imports it. It is the
// first draft of the M4/M5 client boundary. What a seat sees:
//   - its own hand: card names, printed characteristics and instance IDs;
//   - the opponent's hand and both decks: counts only;
//   - public: both discard piles, HQs, Reserves and resource pools, the Main
//     Theater, the stack, the turn, the phase, the active player and the outcome.
//
// A face-down asset is a Scratch Company with the Scratch face's
// characteristics to both seats, its owner included (FR-ZON-005,
// double-blind). No asset and no discard card carries a card instance ID: an
// instance ID could identify a card seen earlier. Only the seat's own hand
// cards do, because an Action names them.
package view

import (
	"strconv"

	"animus-mundi/pkg/engine"
)

// View is what one seat may see. It is a deep copy: changing it changes no game.
type View struct {
	Seat, Active, First engine.PlayerID
	Turn                int
	Phase               string
	Over, Draw          bool
	Winner              int // -1 until someone has won
	Reason              string
	You, Opponent       Side
	Theater             []AssetView // the shared Main Theater, in battlefield order
	Stack               []StackView // bottom first
}

// Side is one seat's zones as the viewing seat sees them.
type Side struct {
	HQ        *AssetView // nil once destroyed
	Reserves  []AssetView
	Hand      []CardView // You only; empty for Opponent
	HandCount int        // both
	DeckCount int        // both; never the order or contents
	Discard   []CardView // both, top last; public
	Req, Tac  int
}

// CardView is a card off the battlefield, with its printed characteristics.
type CardView struct {
	Inst       int // only for the seat's own hand cards (an Action names them); 0 elsewhere
	Card       string
	Type       string
	ReqCost    int
	TacCost    int
	Power      int
	Health     int
	Keywords   []string // printed
	Text       string   // printed rules text
	Executable bool     // false while the kernel cannot run this card's printed text (Card.Supported)
}

// AssetView is a battlefield asset. Power, MaxHealth and Keywords are
// effective (the layer functions); a face-down asset shows the Scratch face.
type AssetView struct {
	ID                int // asset ID, as Actions and Decisions use
	Card              string
	Controller, Owner int
	Zone              string
	Exhausted         bool
	FaceDown          bool
	Damage            int
	Armor             int
	Power             int
	MaxHealth         int
	Keywords          []string // none for a face-down asset
}

// StackView is one Resolution Stack item.
type StackView struct {
	Kind       string // "Deploy", "March", "Attack", "Scratch" or "Ability"
	Controller int
	Source     int     // asset ID: marching unit, attacker, scratching HQ, or ability source
	Card       string  // Deploy: the card on the stack; Ability: its source card
	Zone       string  // Deploy: the declared zone; March: the origin zone
	Target     int     // Attack: the primary defender (asset ID)
	Keyword    string  // Ability: the keyword its trigger came from ("Aggressive"); empty for a printed ability
	Chosen     [][]int // Ability: per target slot, the chosen asset IDs; empty for a pile slot (see ChosenCards)
	// ChosenCards: Ability: per target slot, for a pile slot, the names of its chosen cards that are still in a discard pile, in pick order; empty for an asset slot
	ChosenCards [][]string
}

var stackKinds = map[engine.StackKind]string{
	engine.StackDeploy: "Deploy", engine.StackMarch: "March", engine.StackAttack: "Attack",
	engine.StackScratch: "Scratch", engine.StackAbility: "Ability",
}

// PlayerView is seat's view of g now.
func PlayerView(g *engine.Game, seat engine.PlayerID) View {
	v := View{
		Seat: seat, Active: g.Active, First: g.FirstPlayer, Turn: g.Turn, Phase: g.Phase.String(),
		Over: g.Outcome.Over, Draw: g.Outcome.Draw, Winner: -1, Reason: g.Outcome.Reason,
		You: side(g, g.Players[seat], true), Opponent: side(g, g.Players[seat.Opponent()], false),
		Theater: assetViews(g, g.Theater), Stack: []StackView{},
	}
	if g.Outcome.Over && !g.Outcome.Draw {
		v.Winner = int(g.Outcome.Winner)
	}
	for _, it := range g.Stack {
		v.Stack = append(v.Stack, stackView(g, it))
	}
	return v
}

func side(g *engine.Game, p *engine.Player, own bool) Side {
	s := Side{
		Reserves: assetViews(g, p.Reserves), Hand: []CardView{}, HandCount: len(p.Hand), DeckCount: len(p.Deck),
		Discard: []CardView{}, Req: p.Req, Tac: p.Tac,
	}
	if p.HQ != nil {
		hq := assetView(g, p.HQ)
		s.HQ = &hq
	}
	if own {
		for _, inst := range p.Hand {
			c := cardView(inst.Card)
			c.Inst = inst.ID
			s.Hand = append(s.Hand, c)
		}
	}
	for _, inst := range p.Discard {
		s.Discard = append(s.Discard, cardView(inst.Card))
	}
	return s
}

func cardView(c *engine.Card) CardView {
	return CardView{
		Card: c.Name, Type: c.Type.String(), ReqCost: c.ReqCost, TacCost: c.TacCost,
		Power: c.Power, Health: c.Health, Keywords: keywords(c.Keywords), Text: c.RulesText,
		Executable: c.Supported(),
	}
}

// assetView reads a through the layer functions. Printed is the Scratch face
// for a face-down asset, so its name and stats never reveal the card.
func assetView(g *engine.Game, a *engine.Asset) AssetView {
	return AssetView{
		ID: a.ID, Card: a.Printed().Name, Controller: int(a.Controller), Owner: int(a.Inst.Owner),
		Zone: a.Zone.String(), Exhausted: a.Exhausted, FaceDown: a.FaceDown, Damage: a.Damage, Armor: a.Armor,
		Power: g.Power(a), MaxHealth: g.MaxHealth(a), Keywords: keywords(g.Keywords(a)),
	}
}

func assetViews(g *engine.Game, as []*engine.Asset) []AssetView {
	out := make([]AssetView, len(as))
	for i, a := range as {
		out[i] = assetView(g, a)
	}
	return out
}

func stackView(g *engine.Game, it *engine.StackItem) StackView {
	s := StackView{Kind: stackKinds[it.Kind], Controller: int(it.Controller), Source: it.Source}
	switch it.Kind {
	case engine.StackDeploy:
		s.Card, s.Zone = it.Inst.Card.Name, it.Zone.String()
	case engine.StackMarch:
		s.Zone = it.Zone.String()
	case engine.StackAttack:
		s.Target = it.Target
	case engine.StackAbility:
		s.Keyword = it.Keyword.String()
		// A face-down asset has no abilities today; if one ever has, its
		// item still names only the Scratch face.
		if a := g.Asset(it.Source); a != nil && a.FaceDown {
			s.Card = a.Printed().Name
		} else if it.Inst != nil {
			s.Card = it.Inst.Card.Name
		}
		s.Chosen = make([][]int, len(it.Chosen))
		s.ChosenCards = make([][]string, len(it.Chosen))
		for i, ids := range it.Chosen {
			s.Chosen[i], s.ChosenCards[i] = []int{}, []string{}
			if !it.PileSlot(i) {
				s.Chosen[i] = append(s.Chosen[i], ids...)
				continue
			}
			for _, id := range ids {
				if inst := g.DiscardCard(id); inst != nil {
					s.ChosenCards[i] = append(s.ChosenCards[i], inst.Card.Name)
				}
			}
		}
	}
	return s
}

// keywords lists k's standalone keywords in printed form, never nil.
func keywords(k engine.Keywords) []string {
	out := []string{}
	num := func(name string, n int) {
		if n != 0 {
			out = append(out, name+" "+strconv.Itoa(n))
		}
	}
	flag := func(name string, on bool) {
		if on {
			out = append(out, name)
		}
	}
	num("Resistant", k.Resistant)
	flag("Stalwart", k.Stalwart)
	num("Aggressive", k.Aggressive)
	flag("Ethereal", k.Ethereal)
	flag("Ward", k.Ward)
	flag("Versatile", k.Versatile)
	flag("Guardian", k.Guardian)
	num("Armor", k.Armor)
	flag("Deploys Exhausted", k.DeploysExhausted)
	return out
}
