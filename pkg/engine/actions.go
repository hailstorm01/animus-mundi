// Actions: the legal-action surface, Apply and its precondition checks.

package engine

import (
	"errors"
	"fmt"
	"slices"

	"animus-mundi/pkg/engine/ability"
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
	ActEnact                      // Maneuver from Hand, stack-pushed (FR-MNV-001, FR-TRN-003)
	ActActivate                   // printed ability of a battlefield asset, stack-pushed (FR-ACT-001)
	ActEndTurn
)

func (k ActionKind) String() string {
	return [...]string{"Generate", "Convert", "Deploy", "March", "Attack", "Scratch", "Enact", "Activate", "EndTurn"}[k]
}

// Action is one manual declaration. Fields not used by Kind are ignored.
type Action struct {
	Kind     ActionKind
	Player   PlayerID
	Asset    int      // acting asset: Generate, March, Attack, Scratch, Activate
	Resource Resource // Generate
	Card     int      // Deploy, Enact: Instance.ID in hand
	Zone     Zone     // Deploy: declared battlefield zone
	Target   int      // Attack: the defender
	Ability  int      // Activate: index into the asset's Def.Activated
}

// Apply validates and executes a, then resolves the stack until it is empty.
func (g *Game) Apply(a Action) error {
	if err := g.check(a); err != nil {
		return err
	}
	g.emitAction(a)
	p := g.Players[a.Player]
	switch a.Kind {
	case ActGenerate:
		g.generate(a)
	case ActConvert:
		p.convert()
	case ActDeploy:
		inst := g.playFromHand(p, a.Card)
		g.declare(deployPlacement(a.Player, inst, a.Zone), ability.Effect{}, StackDeploy)
	case ActEnact:
		inst := g.playFromHand(p, a.Card)
		g.declare(maneuverPlacement(a.Player, inst), inst.Card.Ability.Effect, StackManeuver)
	case ActActivate:
		as := g.Asset(a.Asset)
		act := as.Printed().Ability.Activated[a.Ability]
		g.declare(activatedPlacement(a.Player, as, act), act.Effect, StackAbility)
	case ActMarch:
		as := g.Asset(a.Asset)
		g.exhaust(as)
		g.push(&StackItem{Kind: StackMarch, Controller: a.Player, Source: as.ID, Zone: as.Zone})
	case ActAttack:
		g.declareAttack(a)
	case ActScratch:
		g.exhaust(p.HQ)
		g.push(&StackItem{Kind: StackScratch, Controller: a.Player, Source: p.HQ.ID})
	case ActEndTurn:
		g.endTurn()
		g.settled()
		return nil
	}
	g.resolveStack()
	g.settled()
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
		if !p.canPay(ConvertReq, 0) {
			return illegal("conversion needs %d Requisition", ConvertReq)
		}
	case ActDeploy:
		inst, err := g.handPlay(p, a.Card)
		if err != nil {
			return err
		}
		c := inst.Card
		if c.Type != TypeUnit && c.Type != TypeStructure {
			return illegal("%s cannot be deployed", c.Name)
		}
		if a.Zone == ZoneHeadquarters || !slices.Contains(c.DeployZones, a.Zone) {
			return illegal("%s cannot be deployed to %s", c.Name, a.Zone)
		}
		if !g.placeable(deployPlacement(a.Player, inst, a.Zone), ability.Effect{}) {
			return illegal("%s's additional costs cannot be paid (FR-TRN-003a)", c.Name)
		}
	case ActEnact:
		inst, err := g.handPlay(p, a.Card)
		if err != nil {
			return err
		}
		if inst.Card.Type != TypeManeuver {
			return illegal("%s is not a Maneuver", inst.Card.Name)
		}
		if !g.placeable(maneuverPlacement(a.Player, inst), inst.Card.Ability.Effect) {
			return illegal("%s has no legal choice for a required target or enough legal modes (SVY-R29)", inst.Card.Name)
		}
	case ActActivate:
		as, err := g.ownAsset(a)
		if err != nil {
			return err
		}
		def := as.Printed().Ability
		if def == nil || a.Ability < 0 || a.Ability >= len(def.Activated) {
			return illegal("%s has no activated ability %d", as, a.Ability)
		}
		act := def.Activated[a.Ability]
		if !g.placeable(activatedPlacement(a.Player, as, act), act.Effect) {
			return illegal("a cost of %s ability %d cannot be paid, or a required target or mode has no legal choice (FR-ACT-001, SVY-R36)", as, a.Ability)
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
		return g.checkAttack(a)
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

// handPlay returns the card with instance ID id if p may play it from hand:
// it is in hand, the kernel supports it and p can pay its cost. Deploy and
// Enact both start here and then check only what is their own.
func (g *Game) handPlay(p *Player, id int) (*Instance, error) {
	inst := handCard(p, id)
	if inst == nil {
		return nil, illegal("card %d is not in hand", id)
	}
	c := inst.Card
	if !c.Supported() {
		return nil, illegal("%s needs rules the kernel does not implement yet", c.Name)
	}
	req, tac := g.EffectiveCost(inst)
	if !p.canPay(req, tac) {
		return nil, illegal("cannot pay for %s (needs R%d T%d)", c.Name, req, tac)
	}
	return inst, nil
}

// playFromHand finds the card handPlay approved, pays its cost and takes it from
// hand: what Deploy and Enact do before each goes on in its own way.
func (g *Game) playFromHand(p *Player, id int) *Instance {
	inst := handCard(p, id)
	g.payCost(p, inst)
	g.takeCard(inst)
	return inst
}

// ownAsset returns the battlefield asset a names, if a.Player controls it.
func (g *Game) ownAsset(a Action) (*Asset, error) {
	as := g.Asset(a.Asset)
	if as == nil || as.Controller != a.Player {
		return nil, illegal("asset %d is not yours on the battlefield", a.Asset)
	}
	return as, nil
}

func (g *Game) readyAsset(a Action) (*Asset, error) {
	as, err := g.ownAsset(a)
	if err != nil {
		return nil, err
	}
	if as.Exhausted {
		return nil, illegal("%s is exhausted", as)
	}
	return as, nil
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
		}
	}
	try(Action{Kind: ActConvert, Player: pid})
	for _, inst := range p.Hand {
		for _, z := range inst.Card.DeployZones {
			try(Action{Kind: ActDeploy, Player: pid, Card: inst.ID, Zone: z})
		}
	}
	for _, inst := range p.Hand {
		try(Action{Kind: ActEnact, Player: pid, Card: inst.ID})
	}
	for _, a := range mine {
		if def := a.Printed().Ability; def != nil {
			for i := range def.Activated {
				try(Action{Kind: ActActivate, Player: pid, Asset: a.ID, Ability: i})
			}
		}
	}
	if p.HQ != nil {
		try(Action{Kind: ActScratch, Player: pid, Asset: p.HQ.ID})
	}
	return append(out, Action{Kind: ActEndTurn, Player: pid})
}
