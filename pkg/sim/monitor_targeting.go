package sim

import "animus-mundi/pkg/engine"

// targetPrompt is a Target decision waiting for the next non-decision event.
type targetPrompt struct{ seq, player int }

// battlefield lists every asset on the battlefield in s.
func battlefield(s engine.Snapshot) []engine.AssetSnap {
	var out []engine.AssetSnap
	eachAsset(s, func(a *engine.AssetSnap) bool {
		out = append(out, *a)
		return true
	})
	return out
}

// checkTargetDecision queues the placement check for a Target decision and
// checks Ward on the options now (FR-KWD-002h).
func (m *Monitors) checkTargetDecision(e engine.Event) {
	const mon = monTargeting
	if e.Decision == nil || e.Decision.Kind != "Target" {
		return
	}
	m.count(mon, "FR-TGT-001", "SVY-R3")
	m.targets = append(m.targets, targetPrompt{seq: e.Seq, player: e.Decision.Player})
	if m.last == nil {
		return
	}
	player := e.Decision.Player
	warded := map[int]bool{}
	for _, a := range battlefield(*m.last) {
		if a.Controller == player {
			continue
		}
		if c := m.card(a); c != nil && c.Keywords.Ward {
			warded[a.ID] = true
		}
	}
	if len(warded) == 0 {
		return
	}
	m.count(mon, "FR-KWD-002h")
	for _, ref := range e.Decision.Refs {
		if warded[ref] {
			m.fail(mon, e.Seq, "player %d was offered asset %d, an opponent's Ward asset", player, ref)
		}
	}
}

// endTargetPrompts checks the event that follows queued Target decisions: it
// must be the same player's place event, or the target was chosen while an
// item resolved (FR-TGT-001, SVY-R3).
func (m *Monitors) endTargetPrompts(e engine.Event) {
	for _, t := range m.targets {
		if e.Kind != "place" || e.Player != t.player {
			m.fail(monTargeting, t.seq, "player %d's Target decision is followed by a %s event (player %d), not the placement of its item", t.player, e.Kind, e.Player)
		}
	}
	m.targets = nil
}
