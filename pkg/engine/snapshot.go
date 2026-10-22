// Snapshots: deep copies of the full game state, hidden information included, for observers and replay.

package engine

import "slices"

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

// Snapshot deep-copies the full current state, hidden information included.
// Replay uses it to capture the state at an event.
func (g *Game) Snapshot() Snapshot { return g.snapshot() }

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
