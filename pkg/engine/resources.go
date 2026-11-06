// Resources: Requisition and Tactics, generating and converting them, and paying with them (FR-RES-001, FR-RES-003, FR-CAP-001).

package engine

// Resource selects Requisition or Tactics.
type Resource int

const (
	Requisition Resource = iota
	Tactics
)

// String names the resource, as events and the play screen write it.
func (r Resource) String() string {
	if r == Tactics {
		return "Tactics"
	}
	return "Requisition"
}

// ConvertReq is the Requisition one conversion to Tactics takes (FR-RES-003).
const ConvertReq = 2

// generate carries out a checked Generate: the asset exhausts and its controller
// gains the asset's printed yield of the chosen resource (FR-RES-001).
func (g *Game) generate(a Action) {
	p := g.Players[a.Player]
	as := g.Asset(a.Asset)
	g.exhaust(as)
	yield := as.Printed().TacYield
	if a.Resource == Requisition {
		yield = as.Printed().ReqYield
		p.Req += yield
	} else {
		p.Tac += yield
	}
	g.logEvent(Event{Kind: EventGenerate, Player: int(a.Player), Source: as.ID, Amount: yield},
		"%s exhausts for resources (R%d T%d)", as, p.Req, p.Tac)
}

// convert carries out a checked conversion (FR-RES-003).
func (p *Player) convert() {
	p.pay(ConvertReq, 0)
	p.Tac++
}

// canYield implements the Generate Resources rows of FR-CAP-001. A zero yield
// is still a legal choice (Design Rationale KIM-S03).
func (g *Game) canYield(a *Asset, r Resource) bool {
	if a.Printed().Type != TypeUnit || g.Keywords(a).Versatile {
		return true
	}
	if a.Zone == ZoneReserves {
		return r == Requisition
	}
	return r == Tactics
}

// canPay reports whether p holds req Requisition and tac Tactics, the one read
// of "can pay".
func (p *Player) canPay(req, tac int) bool {
	return p.Req >= req && p.Tac >= tac
}

// pay takes req Requisition and tac Tactics from p, the one payment.
func (p *Player) pay(req, tac int) {
	p.Req -= req
	p.Tac -= tac
}
