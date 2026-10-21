package engine

import "fmt"

// CheckInvariants verifies properties that must hold between actions.
func (g *Game) CheckInvariants() error {
	seen := make(map[int]string)
	note := func(inst *Instance, where string) error {
		if prev, dup := seen[inst.ID]; dup {
			return fmt.Errorf("card instance %d is in both %s and %s", inst.ID, prev, where)
		}
		seen[inst.ID] = where
		return nil
	}
	for _, p := range g.Players {
		if p.Req < 0 || p.Tac < 0 {
			return fmt.Errorf("player %d has a negative resource pool (R%d T%d)", p.ID, p.Req, p.Tac)
		}
		if p.DrewFromEmpty && !g.Outcome.Over {
			return fmt.Errorf("player %d drew from an empty deck but no death sweep settled it", p.ID)
		}
		for _, zone := range []struct {
			name  string
			cards []*Instance
		}{{"deck", p.Deck}, {"hand", p.Hand}, {"discard", p.Discard}} {
			for _, inst := range zone.cards {
				if err := note(inst, fmt.Sprintf("player %d %s", p.ID, zone.name)); err != nil {
					return err
				}
			}
		}
	}
	for _, a := range g.Assets() {
		if err := note(a.Inst, a.Zone.String()); err != nil {
			return err
		}
		if a.Armor < 0 || a.Damage < 0 {
			return fmt.Errorf("%s has negative armor or damage", a)
		}
		if len(g.Stack) == 0 && g.IsLethal(a) {
			return fmt.Errorf("%s is lethal but survived the death sweep", a)
		}
	}
	for _, p := range g.Players {
		if p.HQ != nil && (p.HQ.Zone != ZoneHeadquarters || p.HQ.Controller != p.ID) {
			return fmt.Errorf("%s is player %d's HQ but records zone %s", p.HQ, p.ID, p.HQ.Zone)
		}
		for _, a := range p.Reserves {
			if a.Zone != ZoneReserves || a.Controller != p.ID {
				return fmt.Errorf("%s is in player %d's Reserves but records zone %s, controller %d", a, p.ID, a.Zone, a.Controller)
			}
		}
	}
	for _, a := range g.Theater {
		if a.Zone != ZoneMainTheater {
			return fmt.Errorf("%s is in the Main Theater but records zone %s", a, a.Zone)
		}
	}
	if !g.Outcome.Over && len(g.Stack) != 0 {
		return fmt.Errorf("stack not empty between actions")
	}
	if len(g.pending) != 0 {
		return fmt.Errorf("%d fired triggers left in the pending buffer between actions", len(g.pending))
	}
	return nil
}
