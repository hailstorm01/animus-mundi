package engine

import (
	"errors"
	"fmt"
	"slices"
)

// ErrIllegal wraps every rejected action.
var ErrIllegal = errors.New("illegal action")

func illegal(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrIllegal, fmt.Sprintf(format, args...))
}

// ActionKind enumerates manual actions a player can declare.
type ActionKind int

const (
	ActGenerate ActionKind = iota // off-stack (FR-RES-001)
	ActConvert                    // off-stack 2R -> 1T (FR-RES-003)
	ActDeploy                     // Unit or Structure from Hand (FR-TRN-003)
	ActMarch                      // stack-pushed (FR-CAP-001)
	ActAttack                     // stack-pushed (FR-CBT-002)
	ActScratch                    // stack-pushed, HQ only (FR-CAP-001)
	ActEndTurn
)

func (k ActionKind) String() string {
	return [...]string{"Generate", "Convert", "Deploy", "March", "Attack", "Scratch", "EndTurn"}[k]
}

// Resource selects Requisition or Tactics.
type Resource int

const (
	Requisition Resource = iota
	Tactics
)

// Action is one manual declaration. Fields not used by Kind are ignored.
type Action struct {
	Kind      ActionKind
	Player    PlayerID
	Asset     int      // acting asset: Generate, March, Attack, Scratch
	Resource  Resource // Generate
	Card      int      // Deploy: Instance.ID in hand
	Zone      Zone     // Deploy: declared battlefield zone
	Target    int      // Attack: primary defender
	Secondary int      // Attack: Berserk secondary target (0 = none)
}

// Apply validates and executes a, then resolves the stack until it is empty.
func (g *Game) Apply(a Action) error {
	if err := g.check(a); err != nil {
		return err
	}
	p := g.Players[a.Player]
	switch a.Kind {
	case ActGenerate:
		as := g.Asset(a.Asset)
		as.Exhausted = true
		if a.Resource == Requisition {
			p.Req += as.Printed().ReqYield
		} else {
			p.Tac += as.Printed().TacYield
		}
		g.logf("generate", "%s exhausts for resources (R%d T%d)", as, p.Req, p.Tac)
	case ActConvert:
		p.Req -= 2
		p.Tac++
	case ActDeploy:
		inst := handCard(p, a.Card)
		p.Req -= inst.Card.ReqCost
		p.Tac -= inst.Card.TacCost
		p.Hand = remove(p.Hand, inst)
		g.push(&StackItem{Kind: StackDeploy, Controller: a.Player, Inst: inst, Zone: a.Zone})
	case ActMarch:
		as := g.Asset(a.Asset)
		as.Exhausted = true
		g.push(&StackItem{Kind: StackMarch, Controller: a.Player, Source: as.ID, Zone: as.Zone})
	case ActAttack:
		g.declareAttack(a)
	case ActScratch:
		p.HQ.Exhausted = true
		g.push(&StackItem{Kind: StackScratch, Controller: a.Player, Source: p.HQ.ID})
	case ActEndTurn:
		g.endTurn()
		return nil
	}
	g.resolveStack()
	return nil
}

// declareAttack is Steps 1 and 2 of combat (FR-CBT-002).
func (g *Game) declareAttack(a Action) {
	att := g.Asset(a.Asset)
	att.Exhausted = true
	g.push(&StackItem{Kind: StackAttack, Controller: a.Player, Source: att.ID, Target: a.Target, Secondary: a.Secondary})
	g.logf("attack", "%s attacks %s", att, g.Asset(a.Target))
	// Step 2: attack triggers enqueue on top of the Attack action.
	if n := g.Keywords(att).Aggressive; n > 0 {
		g.push(&StackItem{Kind: StackAggressive, Controller: a.Player, Source: att.ID, Amount: n})
	}
}

func handCard(p *Player, id int) *Instance {
	for _, inst := range p.Hand {
		if inst.ID == id {
			return inst
		}
	}
	return nil
}

// check returns nil if a is legal right now.
func (g *Game) check(a Action) error {
	if g.Outcome.Over {
		return illegal("match is over")
	}
	if g.Phase != PhaseMain {
		return illegal("manual actions require the Main Phase")
	}
	if a.Player != g.Active {
		return illegal("player %d is not the active player (FR-TRN-002)", a.Player)
	}
	if len(g.Stack) > 0 && a.Kind != ActGenerate && a.Kind != ActConvert {
		return illegal("the Resolution Stack is not empty (FR-QUE-001)")
	}
	p := g.Players[a.Player]

	switch a.Kind {
	case ActGenerate:
		as, err := g.readyAsset(a)
		if err != nil {
			return err
		}
		if !g.canYield(as, a.Resource) {
			return illegal("%s cannot generate that resource in %s (FR-CAP-001)", as, as.Zone)
		}
	case ActConvert:
		if p.Req < 2 {
			return illegal("conversion needs 2 Requisition")
		}
	case ActDeploy:
		inst := handCard(p, a.Card)
		if inst == nil {
			return illegal("card %d is not in hand", a.Card)
		}
		c := inst.Card
		if !c.Supported() {
			return illegal("%s needs rules the M1 kernel does not implement", c.Name)
		}
		if c.Type != TypeUnit && c.Type != TypeStructure {
			return illegal("%s cannot be deployed", c.Name)
		}
		if a.Zone == ZoneHeadquarters || !slices.Contains(c.DeployZones, a.Zone) {
			return illegal("%s cannot be deployed to %s", c.Name, a.Zone)
		}
		if p.Req < c.ReqCost || p.Tac < c.TacCost {
			return illegal("cannot pay for %s (needs R%d T%d)", c.Name, c.ReqCost, c.TacCost)
		}
	case ActMarch:
		as, err := g.readyAsset(a)
		if err != nil {
			return err
		}
		if as.Printed().Type != TypeUnit || (as.Zone != ZoneReserves && as.Zone != ZoneMainTheater) {
			return illegal("only Units march (FR-ZON-002)")
		}
	case ActAttack:
		att, err := g.readyAsset(a)
		if err != nil {
			return err
		}
		if !canDeclareAttack(att) {
			return illegal("%s cannot attack from %s (FR-CBT-001)", att, att.Zone)
		}
		def := g.Asset(a.Target)
		if def == nil || !g.CanAttack(att, def) {
			return illegal("illegal combat target (FR-CBT-001)")
		}
		if a.Secondary != 0 {
			sec := g.Asset(a.Secondary)
			if sec == nil || !g.CanCleave(att, def, sec) {
				return illegal("illegal Berserk secondary target (FR-KWD-002o)")
			}
		}
	case ActScratch:
		if p.HQ == nil || a.Asset != p.HQ.ID {
			return illegal("only your Headquarters can Deploy Scratch Company")
		}
		if p.HQ.Exhausted {
			return illegal("Headquarters is exhausted")
		}
	case ActEndTurn:
	default:
		return illegal("unknown action kind %d", a.Kind)
	}
	return nil
}

func (g *Game) readyAsset(a Action) (*Asset, error) {
	as := g.Asset(a.Asset)
	if as == nil || as.Controller != a.Player {
		return nil, illegal("asset %d is not yours on the battlefield", a.Asset)
	}
	if as.Exhausted {
		return nil, illegal("%s is exhausted", as)
	}
	return as, nil
}

// canYield implements the Generate Resources rows of FR-CAP-001.
func (g *Game) canYield(a *Asset, r Resource) bool {
	if a.Printed().Type != TypeUnit || g.Keywords(a).Versatile {
		return true
	}
	if a.Zone == ZoneReserves {
		return r == Requisition
	}
	return r == Tactics
}

func canDeclareAttack(a *Asset) bool {
	switch a.Printed().Type {
	case TypeUnit:
		return a.Zone == ZoneMainTheater
	case TypeHeadquarters:
		return a.Zone == ZoneHeadquarters
	}
	return false
}

// CanAttack implements defender eligibility, Ethereal and Guardian (FR-CBT-001).
func (g *Game) CanAttack(att, def *Asset) bool {
	if def.Controller == att.Controller {
		return false
	}
	if def.Zone != ZoneMainTheater && def.Zone != ZoneHeadquarters {
		return false // Reserves are never attackable
	}
	if g.Keywords(def).Ethereal {
		return false
	}
	if g.hasGuardian(def.Controller) {
		return def.Zone == ZoneMainTheater && g.Keywords(def).Guardian
	}
	return true
}

// hasGuardian reports whether p controls a non-Ethereal Guardian in the Main Theater.
func (g *Game) hasGuardian(p PlayerID) bool {
	for _, a := range g.TheaterOf(p) {
		kw := g.Keywords(a)
		if kw.Guardian && !kw.Ethereal {
			return true
		}
	}
	return false
}

// CanCleave reports whether sec is a legal Berserk secondary target.
func (g *Game) CanCleave(att, def, sec *Asset) bool {
	if att.Printed().Type != TypeUnit || g.Keywords(att).Berserk <= 0 {
		return false
	}
	if def.Zone != ZoneMainTheater || sec.ID == def.ID {
		return false // Berserk only cleaves when attacking a Main Theater asset
	}
	// OPEN(KIM-B01): FRD 01 says only "another legal opposing Main Theater
	// asset". Whether Ward, Ethereal or the Guardian restriction limit the
	// secondary target is undefined; the kernel applies none of them.
	return sec.Zone == ZoneMainTheater && sec.Controller != att.Controller
}

// LegalActions enumerates every legal action for the active player.
func (g *Game) LegalActions() []Action {
	if g.Outcome.Over || g.Phase != PhaseMain {
		return nil
	}
	pid := g.Active
	p := g.Players[pid]
	var out []Action
	try := func(a Action) {
		if g.check(a) == nil {
			out = append(out, a)
		}
	}

	var mine []*Asset
	for _, a := range g.Assets() {
		if a.Controller == pid {
			mine = append(mine, a)
		}
	}
	for _, a := range mine {
		try(Action{Kind: ActGenerate, Player: pid, Asset: a.ID, Resource: Requisition})
		try(Action{Kind: ActGenerate, Player: pid, Asset: a.ID, Resource: Tactics})
		try(Action{Kind: ActMarch, Player: pid, Asset: a.ID})
		if a.Exhausted || !canDeclareAttack(a) {
			continue
		}
		for _, def := range g.Assets() {
			if !g.CanAttack(a, def) {
				continue
			}
			try(Action{Kind: ActAttack, Player: pid, Asset: a.ID, Target: def.ID})
			for _, sec := range g.Theater {
				if g.CanCleave(a, def, sec) {
					try(Action{Kind: ActAttack, Player: pid, Asset: a.ID, Target: def.ID, Secondary: sec.ID})
				}
			}
		}
	}
	try(Action{Kind: ActConvert, Player: pid})
	for _, inst := range p.Hand {
		for _, z := range inst.Card.DeployZones {
			try(Action{Kind: ActDeploy, Player: pid, Card: inst.ID, Zone: z})
		}
	}
	if p.HQ != nil {
		try(Action{Kind: ActScratch, Player: pid, Asset: p.HQ.ID})
	}
	return append(out, Action{Kind: ActEndTurn, Player: pid})
}
