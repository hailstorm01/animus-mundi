package sim

import (
	"slices"

	"animus-mundi/pkg/engine"
)

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
