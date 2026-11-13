// Free deploys: deploying a card from a discard pile for free, and looking at the top cards of a deck to deploy some for free and put the rest on the bottom (FR-TRN-003a, FR-MOD-001 §4).

package engine

import (
	"fmt"
	"slices"
	"strings"

	"animus-mundi/pkg/engine/ability"
)

// stepDeployFree puts each card of stepCards, each still in its owner's
// discard pile, onto the battlefield for free under the ability's controller and
// returns how many it deployed (FR-TRN-003a, FR-CST-001). Only a supported Unit
// or Structure card goes. The zone is st.To if it names one (FR-TRN-003a 2),
// else the card's printed zone limited to Reserves and the Main Theater; with
// both printed the controller chooses, Reserves listed first, and a timeout
// takes Reserves (SVY-R43). A card printing neither is skipped. It enters
// through deployCard, so deploy triggers fire and it is not a move (SVY-R1).
func (g *Game) stepDeployFree(r *resolution, st ability.Step) int {
	n := 0
	for _, inst := range g.stepCards(r, st) {
		c := inst.Card
		zones := freeDeployZones(c, st.To)
		if len(zones) == 0 {
			continue
		}
		zone := zones[0]
		if len(zones) == 2 {
			d := Decision{Player: r.it.Controller, Kind: DecideZone, Prompt: "Choose the zone to deploy " + c.Name + " into", Min: 1, Max: 1,
				Options: []Option{{Label: zones[0].String()}, {Label: zones[1].String()}}}
			if picks := g.choose(d); len(picks) > 0 && picks[0] == 1 {
				zone = zones[1]
			}
		}
		g.takeCard(inst)
		a := g.deployCard(inst, r.it.Controller, zone)
		if st.Subject == ability.SubjectTarget {
			if r.became == nil {
				r.became = map[int][]*Asset{}
			}
			r.became[st.Target] = append(r.became[st.Target], a)
		}
		n++
	}
	return n
}

// freeDeployZones returns the zones a free deploy of c may enter, in listed
// order (FR-TRN-003a, SVY-R43): nil unless c is a supported Unit or Structure;
// [to] if to names a battlefield zone (FR-TRN-003a 2); else the printed
// Reserves, then the Main Theater, whichever c.DeployZones contains
// (FR-TRN-003a 1).
func freeDeployZones(c *Card, to ability.Zone) []Zone {
	if (c.Type != TypeUnit && c.Type != TypeStructure) || !c.Supported() {
		return nil
	}
	if zone, ok := battlefieldZone(to); ok {
		return []Zone{zone}
	}
	var zones []Zone
	for _, z := range []Zone{ZoneReserves, ZoneMainTheater} {
		if slices.Contains(c.DeployZones, z) {
			zones = append(zones, z)
		}
	}
	return zones
}

// lookOutcome is one option of a look at the top cards: the looked positions
// deployed (look order), their zones, and the looked positions put on the
// bottom, first named ends highest.
type lookOutcome struct {
	deploy []int
	zones  []Zone
	bottom []int
}

// stepLookTop has the controller look at the top st.Amount cards of their deck
// and choose, in one hidden decision, which are deployed for free (up to
// st.UpTo, or exactly one when UpTo is 0 and a card is eligible), their zones
// and the order the rest go to the bottom (FR-MOD-001 §4, SVY-R41-SVY-R43). A
// card is eligible if it meets st.Filter's card-level criteria and has a free
// deploy zone. Each option is a whole outcome; the first is the timeout
// (SVY-R41, SVY-R42). Chosen cards enter through deployCard, so deploy
// triggers fire (SVY-R1). It returns how many cards it deployed.
func (g *Game) stepLookTop(r *resolution, st ability.Step) int {
	ctl := r.it.Controller
	p := g.Players[ctl]
	n := min(g.value(r.it.Source, st.Amount, nil), len(p.Deck))
	if n <= 0 {
		return 0
	}
	looked := slices.Clone(p.Deck[:n])
	var eligible []int
	for i, inst := range looked {
		if g.cardMeets(ctl, r.it.Inst, st.Filter, inst) && len(freeDeployZones(inst.Card, st.To)) > 0 {
			eligible = append(eligible, i)
		}
	}
	var sets [][]int
	switch {
	case st.UpTo > 0:
		for size := 0; size <= min(st.UpTo, len(eligible)); size++ {
			sets = append(sets, combinations(eligible, size)...)
		}
	case len(eligible) > 0:
		sets = combinations(eligible, 1)
	default:
		sets = [][]int{nil}
	}
	var outs []lookOutcome
	var opts []Option
	for _, set := range sets {
		var rest []int
		for i := range looked {
			if !slices.Contains(set, i) {
				rest = append(rest, i)
			}
		}
		for _, zs := range zoneAssignments(looked, set, st.To) {
			for _, ord := range permutations(rest) {
				o := lookOutcome{set, zs, ord}
				outs = append(outs, o)
				opts = append(opts, Option{Label: lookLabel(looked, o)})
			}
		}
	}
	d := Decision{Player: ctl, Kind: DecideLook, Prompt: fmt.Sprintf("Look at the top %d cards of your deck", n), Options: opts, Min: 1, Max: 1, Hidden: true}
	pick := 0
	if picks := g.choose(d); len(picks) > 0 {
		pick = picks[0]
	}
	o := outs[pick]
	for _, inst := range looked {
		g.takeCard(inst)
	}
	for _, i := range o.bottom {
		g.putCard(looked[i], ZoneDeck, true)
	}
	for k, i := range o.deploy {
		g.deployCard(looked[i], ctl, o.zones[k])
	}
	return len(o.deploy)
}

// combinations returns every size-element subset of items, lexicographic by position.
func combinations(items []int, size int) [][]int {
	if size == 0 {
		return [][]int{nil}
	}
	var out [][]int
	for i := 0; i+size <= len(items); i++ {
		for _, tail := range combinations(items[i+1:], size-1) {
			out = append(out, append([]int{items[i]}, tail...))
		}
	}
	return out
}

// permutations returns every order of items, lexicographic by position.
func permutations(items []int) [][]int {
	if len(items) == 0 {
		return [][]int{nil}
	}
	var out [][]int
	for i, it := range items {
		rest := append(slices.Clone(items[:i]), items[i+1:]...)
		for _, tail := range permutations(rest) {
			out = append(out, append([]int{it}, tail...))
		}
	}
	return out
}

// zoneAssignments returns every zone choice for the looked cards at set, the
// first card varying slowest.
func zoneAssignments(looked []*Instance, set []int, to ability.Zone) [][]Zone {
	out := [][]Zone{nil}
	for _, i := range set {
		var next [][]Zone
		for _, prefix := range out {
			for _, z := range freeDeployZones(looked[i].Card, to) {
				next = append(next, append(slices.Clone(prefix), z))
			}
		}
		out = next
	}
	return out
}

// lookLabel names one look outcome for the play screen and the log.
func lookLabel(looked []*Instance, o lookOutcome) string {
	var deploy, bottom []string
	for k, i := range o.deploy {
		deploy = append(deploy, looked[i].Card.Name+" to "+o.zones[k].String())
	}
	for _, i := range o.bottom {
		bottom = append(bottom, looked[i].Card.Name)
	}
	label := "Deploy nothing"
	if len(deploy) > 0 {
		label = "Deploy " + strings.Join(deploy, ", ")
	}
	if len(bottom) == 0 {
		return label + "; bottom: nothing"
	}
	return label + "; bottom: " + strings.Join(bottom, ", ")
}
