// Effects: resolving a stack item's effect, its steps in order and their values (FR-TGT-003). Untargeted picks are targeting.go's.

package engine

import (
	"fmt"
	"slices"
	"strings"

	"animus-mundi/pkg/engine/ability"
)

// resolution is one resolving ability as its steps see it: the stack item and,
// for each target slot, the chosen entities still legal after the second
// verification stage, in pick order (board directive #51): chosen holds an
// asset slot's assets, cards a pile slot's cards (the other is empty). Then how many things
// each step did so far, one entry per step (IfDone, SVY-R14).
type resolution struct {
	it     *StackItem
	chosen [][]*Asset
	cards  [][]*Instance
	done   []int
	// replaced is the asset whose event a replacement replaces: SubjectReplaced (replace.go).
	replaced *Asset
	// became is, per pile slot, the assets that cards deployed by an earlier
	// step of this resolution became ("it", FR-TRN-003a).
	became map[int][]*Asset
}

// resolveAbility runs the stack item's effect, a triggered or activated
// ability's or an enacted Maneuver's, step by step in order. It resolves even if
// its source has left the battlefield: untargeted steps verify nothing.
func (g *Game) resolveAbility(it *StackItem) {
	// Second verification stage (FR-TGT-003), entity by entity: one that is
	// gone or no longer legal is dropped, and the slot keeps the rest (partial
	// invalidation).
	r := &resolution{it: it, chosen: make([][]*Asset, len(it.Chosen)), cards: make([][]*Instance, len(it.Chosen))}
	for i, ids := range it.Chosen {
		for _, id := range ids {
			if it.PileSlot(i) {
				if c := g.pileTarget(it.Controller, it.Inst, it.Effect.Targets[i].Filter, id); c != nil {
					r.cards[i] = append(r.cards[i], c)
				}
				continue
			}
			if a := g.Asset(id); a != nil && g.legalTarget(it.Controller, it.Source, it.Locked, it.Effect.Targets[i].Filter, a) {
				r.chosen[i] = append(r.chosen[i], a)
			}
		}
	}
	g.runSteps(r)
}

// runSteps runs the steps of r's item in order, recording how many things each
// did. A resolving ability and an inline replacement (replace.go) share it.
func (g *Game) runSteps(r *resolution) {
	for i, st := range r.it.Effect.Steps {
		n := 0
		if g.stepRuns(r, i, st) {
			n = g.runStep(r, st)
		}
		r.done = append(r.done, n)
	}
}

// stepRuns reports whether step i may run: every gate it sets must hold. A
// skipped step counts as 0 in r.done. Each gate is read as the step is reached.
func (g *Game) stepRuns(r *resolution, i int, st ability.Step) bool {
	// "If you do": the step just before it must have done something.
	if st.IfDone && (i == 0 || r.done[i-1] == 0) {
		return false
	}
	// "If you have no cards in hand": the controller's hand, never an opponent's.
	if st.IfHandEmpty && len(g.Players[r.it.Controller].Hand) != 0 {
		return false
	}
	return true
}

// runStep runs one step and returns how many things it did.
func (g *Game) runStep(r *resolution, st ability.Step) int {
	switch st.Op {
	case ability.OpDealDamage:
		return g.stepDealDamage(r, st)
	case ability.OpHeal:
		return g.stepHeal(r, st)
	case ability.OpAddArmor:
		return g.stepAddArmor(r, st)
	case ability.OpScratch:
		return g.stepScratch(r, st)
	case ability.OpDraw:
		return g.stepDraw(r, st)
	case ability.OpDiscard:
		return g.stepDiscard(r, st)
	case ability.OpModifyStat:
		return g.stepModifyStat(r, st)
	case ability.OpKill:
		return g.stepKill(r, st)
	case ability.OpReturnToHand, ability.OpPutOnDeck:
		return g.stepRelocate(r, st)
	case ability.OpMoveZone:
		return g.stepMoveZone(r, st)
	case ability.OpSearchDeck:
		return g.stepSearchDeck(r, st)
	case ability.OpDeployFree:
		return g.stepDeployFree(r, st)
	case ability.OpMorphScratch:
		return g.stepMorphScratch(r, st)
	case ability.OpGrantKeyword:
		return g.stepGrantKeyword(r, st)
	case ability.OpLookTop:
		return g.stepLookTop(r, st)
	}
	return 0
}

// stepDeployFree puts each card of stepCards, still in its owner's discard
// pile, onto the battlefield for free under the ability's controller and
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
		if g.DiscardCard(inst.ID) != inst || len(zones) == 0 {
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

// stepSearchDeck has each step player choose a card from their deck, shuffle
// the rest and put the chosen card into their hand (FR-DEK-001). The decision
// is hidden and must pick one (SVY-R37); an empty deck is unsatisfiable, so
// nothing is asked, but it is still shuffled and the shuffle logged. A search
// is not a draw (FR-WIN-001). It returns how many cards moved to a hand.
func (g *Game) stepSearchDeck(r *resolution, st ability.Step) int {
	n := 0
	for _, pid := range stepPlayers(r.it, st) {
		p := g.Players[pid]
		d := Decision{Player: pid, Kind: DecideSearch, Prompt: "Search your deck for a card", Min: 1, Max: 1, Hidden: true}
		for _, inst := range p.Deck {
			d.Options = append(d.Options, Option{Label: inst.Card.Name, Ref: inst.ID})
		}
		var found *Instance
		if picks := g.choose(d); len(picks) > 0 {
			found = p.Deck[picks[0]]
			g.takeCard(found)
		}
		g.shuffle(p)
		g.logEvent(Event{Kind: EventShuffle, Player: int(pid), Zone: ZoneDeck.String()}, "player %d shuffles their deck", pid)
		if found != nil {
			g.putCard(found, ZoneHand, false)
			n++
		}
	}
	return n
}

// stepMoveZone moves each unit of stepAssets to st.To in its controller's
// battlefield and returns how many moved. A Structure, a Headquarters and a
// unit already there are skipped and don't count (FR-ZON-002). The unit keeps
// its state (FR-ZON-004); moving is not deploying, so no deploy trigger fires
// (SVY-R1). The event carries no Card, as stepRelocate's leave doesn't.
func (g *Game) stepMoveZone(r *resolution, st ability.Step) int {
	to, ok := battlefieldZone(st.To)
	if !ok {
		return 0
	}
	n := 0
	for _, a := range g.stepAssets(r, st) {
		if a.Printed().Type != TypeUnit || a.Zone == to {
			continue
		}
		g.moveOnBattlefield(a, to)
		g.logEvent(Event{Kind: EventMove, Player: -1, Target: a.ID, Zone: to.String()}, "%s moves to %s", a, to)
		n++
	}
	return n
}

// stepPlayers returns the players a player step acts on: the opponent for
// SubjectOpponents, else the item's controller.
func stepPlayers(it *StackItem, st ability.Step) []PlayerID {
	if st.Subject == ability.SubjectOpponents {
		return []PlayerID{it.Controller.Opponent()}
	}
	return []PlayerID{it.Controller}
}

// stepDealDamage deals one instance per asset, all before the item's death
// sweep, so the deaths they cause are simultaneous (FR-REP-004, FR-LAY-002).
// Like every step it returns how many things it did: here, the assets it hit.
// With a chosen dealer (Step.Dealer) that asset deals the damage, at its current
// power; with none left the step deals nothing and logs no event (SVY-R27).
func (g *Game) stepDealDamage(r *resolution, st ability.Step) int {
	src, dealer := g.itemSource(r.it), (*Asset)(nil)
	if slot, ok := st.DealerSlot(); ok {
		ds := g.slotAssets(r, slot)
		if len(ds) == 0 {
			return 0
		}
		dealer = ds[0]
		src = dealer
	}
	n := g.value(r.it.Source, st.Amount, dealer)
	as := g.stepAssets(r, st)
	for _, a := range as {
		g.dealDamage(src, a, n, false)
	}
	return len(as)
}

// stepHeal removes damage from each asset the step acts on. Its event's Amount is
// the damage actually removed.
func (g *Game) stepHeal(r *resolution, st ability.Step) int {
	n := g.value(r.it.Source, st.Amount, nil)
	as := g.stepAssets(r, st)
	for _, a := range as {
		before := a.Damage
		a.Damage = max(0, a.Damage-n)
		g.logEvent(Event{Kind: EventHeal, Player: -1, Target: a.ID, Amount: before - a.Damage}, "%s heals %d", a, before-a.Damage)
	}
	return len(as)
}

// stepAddArmor gives each asset the step acts on Armor.
func (g *Game) stepAddArmor(r *resolution, st ability.Step) int {
	n := g.value(r.it.Source, st.Amount, nil)
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.Armor += n
		g.logEvent(Event{Kind: EventArmor, Player: -1, Target: a.ID, Amount: n}, "%s gains %d Armor", a, n)
	}
	return len(as)
}

// stepScratch has each step player scratch. It returns the cards that left the
// step players' decks.
func (g *Game) stepScratch(r *resolution, st ability.Step) int {
	return g.deckDrop(r, st, g.scratch)
}

// stepDraw has each step player draw. It returns the cards that left the step
// players' decks: a draw from an empty deck counts nothing.
func (g *Game) stepDraw(r *resolution, st ability.Step) int {
	return g.deckDrop(r, st, g.draw)
}

// deckDrop runs take for each step player and returns how many cards left their decks.
func (g *Game) deckDrop(r *resolution, st ability.Step, take func(PlayerID, int)) int {
	total := 0
	for _, pid := range stepPlayers(r.it, st) {
		before := len(g.Players[pid].Deck)
		take(pid, g.value(r.it.Source, st.Amount, nil))
		total += before - len(g.Players[pid].Deck)
	}
	return total
}

// stepModifyStat gives each Subject +Amount power until end of turn: a Layer 5
// modifier, which expires in Sub-Phase 4c (FR-LAY-001, FR-LAY-003). Power clamps at 0.
func (g *Game) stepModifyStat(r *resolution, st ability.Step) int {
	as := g.stepAssets(r, st)
	for _, a := range as {
		n := g.value(r.it.Source, st.Amount, nil)
		a.Modifiers = append(a.Modifiers, Modifier{Layer: 5, Power: n, Until: UntilEndOfTurn})
		g.logEvent(Event{Kind: EventPower, Player: -1, Target: a.ID, Amount: n}, "%s gets %+d power until end of turn", a, n)
	}
	return len(as)
}

// stepKill kills the step assets as one moment (kill). It returns how many it killed.
func (g *Game) stepKill(r *resolution, st ability.Step) int {
	return g.kill(g.stepAssets(r, st))
}

// stepDiscard has each step player discard Amount cards of their choice from
// their own hand (discardFromHand). It returns how many cards they discarded.
func (g *Game) stepDiscard(r *resolution, st ability.Step) int {
	n, total := g.value(r.it.Source, st.Amount, nil), 0
	for _, pid := range stepPlayers(r.it, st) {
		total += len(g.discardFromHand(pid, n, false))
	}
	return total
}

// relocateTo maps a relocate step to where its cards go: its owner's hand, or
// its owner's deck (the bottom if st.Bottom).
func relocateTo(st ability.Step) (to Zone, bottom bool) {
	if st.Op == ability.OpPutOnDeck {
		return ZoneDeck, st.Bottom
	}
	return ZoneHand, false
}

// stepRelocate moves each asset of stepAssets, then each card of stepCards, to
// its owner's hand or deck, and returns how many moved. An asset leaves the
// battlefield with its last-known state (FR-ZON-004); the event carries no
// Card, as a face-down asset's card is private in a hand or deck. A card must
// still be in a discard pile, which is public: its event names it there, before
// the move can hide it.
func (g *Game) stepRelocate(r *resolution, st ability.Step) int {
	to, bottom := relocateTo(st)
	text := "%s returns to its owner's hand"
	if to == ZoneDeck {
		text = "%s is put into its owner's deck"
	}
	n := 0
	for _, a := range g.stepAssets(r, st) {
		g.logEvent(Event{Kind: EventLeave, Player: -1, Target: a.ID, Zone: to.String()}, text, a)
		g.leaveBattlefield(a, g.effectiveState(a))
		g.putCard(a.Inst, to, bottom)
		n++
	}
	for _, inst := range g.stepCards(r, st) {
		if g.DiscardCard(inst.ID) != inst {
			continue
		}
		g.logEvent(Event{Kind: EventCardMove, Player: -1, Card: inst.Card.Name, Inst: inst.ID, Zone: to.String()}, text, inst.Card.Name)
		g.takeCard(inst)
		g.putCard(inst, to, bottom)
		n++
	}
	return n
}

// stepCards returns the cards a card step acts on. SubjectSelf is a death
// trigger's own card: a Tracked item whose exact instance is still in its
// owner's discard pile (FR-ZON-004). SubjectPick with a pile filter is the
// cards picked from discard piles, in pick order. SubjectTarget is the cards
// chosen for its pile slot that are still legal. Anything else has none.
func (g *Game) stepCards(r *resolution, st ability.Step) []*Instance {
	switch st.Subject {
	case ability.SubjectSelf:
		it := r.it
		if it.Tracked && it.Inst != nil && slices.Contains(g.Players[it.Inst.Owner].Discard, it.Inst) {
			return []*Instance{it.Inst}
		}
	case ability.SubjectTarget:
		if st.Target >= 0 && st.Target < len(r.cards) {
			return r.cards[st.Target]
		}
	case ability.SubjectPick:
		if st.Filter.InPile() {
			return g.pickCards(r, st)
		}
	}
	return nil
}

// slotAssets returns slot i's chosen assets that are still on the battlefield,
// in pick order, or nil for an index out of range (FR-ZON-004, SVY-R27).
func (g *Game) slotAssets(r *resolution, i int) []*Asset {
	if i < 0 || i >= len(r.chosen) {
		return nil
	}
	var out []*Asset
	for _, a := range append(slices.Clone(r.chosen[i]), r.became[i]...) {
		if g.Asset(a.ID) == a {
			out = append(out, a)
		}
	}
	return out
}

// value reads v as a step resolves (SVY-R10), for the ability whose source asset
// has the ID source. With a nil dealer, source power is the source's effective
// power if it is on the battlefield, else the effective power recorded as it left
// (FR-ZON-004), else 0. With a chosen dealer it is that asset's current power,
// never a last-known value (SVY-R27).
func (g *Game) value(source int, v ability.Value, dealer *Asset) int {
	if v.From != ability.ValueSourcePower {
		return v.N
	}
	if dealer != nil {
		return g.Power(dealer)
	}
	if a := g.Asset(source); a != nil {
		return g.Power(a)
	}
	if k := g.LastKnownEffective(source); k != nil {
		return k.Power
	}
	return 0
}

// itemSource returns the item's source asset, or its last-known record once it has
// left the battlefield (FR-ZON-004), so a damage event still names and attributes
// it (FR-REP-004 Stage 5) and an attack still knows its attacker's controller
// (FR-CBT-002 Step 3). It is nil only for a source that was never an asset.
func (g *Game) itemSource(it *StackItem) *Asset {
	if a := g.Asset(it.Source); a != nil {
		return a
	}
	return g.LastKnown(it.Source)
}

// stepAssets returns the assets an asset step acts on: the entities chosen for
// its slot that are still legal and still on the battlefield, in pick order (a
// target an earlier step or mode removed is skipped: FR-ZON-004, FR-TGT-003
// partial invalidation), or for SubjectAll every
// battlefield asset matching st.Filter now, oldest first, or for SubjectSelf the
// source if it is still on the battlefield. "All" is not a
// target, so Ward excludes nothing and nothing is asked (FR-TGT-002,
// FR-KWD-002h, SVY-R3).
func (g *Game) stepAssets(r *resolution, st ability.Step) []*Asset {
	switch st.Subject {
	case ability.SubjectTarget:
		return g.slotAssets(r, st.Target)
	case ability.SubjectReplaced:
		if a := r.replaced; a != nil && g.Asset(a.ID) == a {
			return []*Asset{a}
		}
	case ability.SubjectSelf:
		// The source itself, if it is still on the battlefield; once it has
		// left, the step fizzles (FR-TGT-003).
		if a := g.Asset(r.it.Source); a != nil {
			return []*Asset{a}
		}
	case ability.SubjectAll:
		var out []*Asset
		for _, a := range g.assetsByID() {
			if g.matchesFilterAt(r.it.Controller, r.it.Source, r.it.Locked, st.Filter, a) {
				out = append(out, a)
			}
		}
		return out
	case ability.SubjectPick:
		if !st.Filter.InPile() {
			return g.pickAssets(r, st)
		}
	}
	return nil
}

// stepMorphScratch turns each unit of stepAssets face down in place and
// returns how many it turned (FR-ENT-103). Printed() then reads the Scratch
// face; nothing else changes. A Structure, a Headquarters and a unit already
// face down are skipped. The event is logged first, so its text names the card
// while it is still public.
func (g *Game) stepMorphScratch(r *resolution, st ability.Step) int {
	n := 0
	for _, a := range g.stepAssets(r, st) {
		if a.Printed().Type != TypeUnit || a.FaceDown {
			continue
		}
		g.logEvent(Event{Kind: EventFaceDown, Player: -1, Target: a.ID}, "%s turns face down", a)
		a.FaceDown = true
		n++
	}
	return n
}

// stepGrantKeyword gives each asset of stepAssets the step's keyword until end
// of turn, as a Layer 3 modifier (FR-LAY-001, FR-LAY-003), and returns how many
// it granted. A keyword with no flag in keywordDefs (KeywordNone, or Aggressive,
// whose trigger comes only from the printed keyword) grants nothing.
func (g *Game) stepGrantKeyword(r *resolution, st ability.Step) int {
	name := st.Keyword.String()
	var granted Keywords
	found := false
	for _, d := range keywordDefs {
		if name != "" && d.name == name && d.flag != nil {
			*d.flag(&granted) = true
			found = true
		}
	}
	if !found {
		return 0
	}
	as := g.stepAssets(r, st)
	for _, a := range as {
		a.Modifiers = append(a.Modifiers, Modifier{Layer: 3, Keywords: granted, Until: UntilEndOfTurn})
		g.logEvent(Event{Kind: EventKeyword, Player: -1, Target: a.ID}, "%s gains %s until end of turn", a, name)
	}
	return len(as)
}
