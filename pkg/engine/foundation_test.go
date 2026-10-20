package engine

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Foundation tests: FRD 01 requirements the kernel implemented before any
// TestFR_* test traced to them (T-0015).

// --- FR-SET-003: First Player Determination ---

func setupWith(t *testing.T, seed uint64, playFirst func(PlayerID) bool) *Game {
	t.Helper()
	g, err := NewGame(Config{
		Seed:      seed,
		Decks:     [2]DeckList{{HQ: hq(3, 20), Main: filler(20)}, {HQ: hq(3, 20), Main: filler(20)}},
		PlayFirst: playFirst,
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestFR_SET_003_DeciderIsRandomAcrossSeeds(t *testing.T) {
	seen := map[PlayerID]bool{}
	for seed := uint64(1); seed <= 64; seed++ {
		setupWith(t, seed, func(d PlayerID) bool { seen[d] = true; return true })
	}
	if !seen[0] || !seen[1] {
		t.Errorf("deciders seen over 64 seeds = %v, want both players to be chosen as decider sometimes", seen)
	}
}

func TestFR_SET_003_DeciderPlayFirstTakesTurn1(t *testing.T) {
	for seed := uint64(1); seed <= 32; seed++ {
		var decider PlayerID = -1
		g := setupWith(t, seed, func(d PlayerID) bool { decider = d; return true })
		if decider != 0 && decider != 1 {
			t.Fatalf("seed %d: PlayFirst was never asked of a decider", seed)
		}
		if g.FirstPlayer != decider || g.Active != decider || g.Turn != 1 {
			t.Errorf("seed %d: decider %d chose Play First; first=%d active=%d turn=%d, want the decider to take Turn 1",
				seed, decider, g.FirstPlayer, g.Active, g.Turn)
		}
	}
}

func TestFR_SET_003_DeciderDrawFirstGivesOpponentTurn1(t *testing.T) {
	for seed := uint64(1); seed <= 32; seed++ {
		var decider PlayerID = -1
		g := setupWith(t, seed, func(d PlayerID) bool { decider = d; return false })
		if decider != 0 && decider != 1 {
			t.Fatalf("seed %d: PlayFirst was never asked of a decider", seed)
		}
		opp := 1 - decider
		if g.FirstPlayer != opp || g.Active != opp || g.Turn != 1 {
			t.Errorf("seed %d: decider %d chose Draw First; first=%d active=%d turn=%d, want the opponent %d to take Turn 1",
				seed, decider, g.FirstPlayer, g.Active, g.Turn, opp)
		}
	}
}

func TestFR_SET_003_TimeoutDefaultsToPlayFirst(t *testing.T) {
	// A nil PlayFirst models the 15-second timeout. The decider is not
	// observable then, so assert the consequence: the random decider takes
	// Turn 1, which means both seats go first across seeds.
	first := map[PlayerID]bool{}
	for seed := uint64(1); seed <= 64; seed++ {
		g := setupWith(t, seed, nil)
		if g.Active != g.FirstPlayer || g.Turn != 1 {
			t.Fatalf("seed %d: active=%d first=%d turn=%d, want the first player active on Turn 1", seed, g.Active, g.FirstPlayer, g.Turn)
		}
		first[g.FirstPlayer] = true
	}
	if !first[0] || !first[1] {
		t.Errorf("Turn 1 players on timeout = %v, want the random decider (either seat) to Play First by default", first)
	}
}

// Left out for FR-SET-003: "prior to drawing opening hands". The order of the
// decider prompt against the opening draw is not observable through Config or
// the public API (the callback receives only the decider).

// --- FR-CRD-101: Printed Card Characteristics ---

// csvCard returns the data/cards row for name, keyed by header.
func csvCard(t *testing.T, name string) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "data", "cards", "*.csv"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no card CSVs found: %v", err)
	}
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := csv.NewReader(fh).ReadAll()
		fh.Close()
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, r := range rows[1:] {
			if r[0] != name {
				continue
			}
			m := map[string]string{}
			for i, h := range rows[0] {
				m[h] = r[i]
			}
			return m
		}
	}
	t.Fatalf("card %q not in data/cards", name)
	return nil
}

func csvInt(t *testing.T, row map[string]string, key string) int {
	t.Helper()
	n, err := strconv.Atoi(row[key])
	if err != nil {
		t.Fatalf("%s=%q: %v", key, row[key], err)
	}
	return n
}

func csvZones(t *testing.T, s string) []Zone {
	t.Helper()
	var out []Zone
	for _, tok := range strings.Split(s, ",") {
		switch strings.TrimSpace(tok) {
		case "RESERVES":
			out = append(out, ZoneReserves)
		case "MAIN_THEATER":
			out = append(out, ZoneMainTheater)
		case "HEADQUARTERS":
			out = append(out, ZoneHeadquarters)
		default:
			t.Fatalf("unknown deployment zone %q", tok)
		}
	}
	return out
}

func sameZones(a, b []Zone) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for _, y := range b {
			found = found || x == y
		}
		if !found {
			return false
		}
	}
	return true
}

// checkCommon verifies the characteristics every card type carries.
func checkCommon(t *testing.T, c *Card, row map[string]string, wantType CardType) {
	t.Helper()
	if c.Name != row["name"] {
		t.Errorf("name = %q, want %q", c.Name, row["name"])
	}
	if c.Type != wantType {
		t.Errorf("%s type = %v, want %v", c.Name, c.Type, wantType)
	}
	if !strings.EqualFold(c.Faction, row["faction"]) {
		t.Errorf("%s faction = %q, want %q", c.Name, c.Faction, row["faction"])
	}
	if got := csvInt(t, row, "requisitionCost"); c.ReqCost != got {
		t.Errorf("%s Requisition cost = %d, want %d", c.Name, c.ReqCost, got)
	}
	if got := csvInt(t, row, "tacticsCost"); c.TacCost != got {
		t.Errorf("%s Tactics cost = %d, want %d", c.Name, c.TacCost, got)
	}
}

func checkSubtypes(t *testing.T, c *Card, csvSubtypes string) {
	t.Helper()
	var want []string
	if csvSubtypes != "--" && csvSubtypes != "" {
		for _, s := range strings.Split(csvSubtypes, ",") {
			want = append(want, strings.TrimSpace(s))
		}
	}
	if len(c.Subtypes) != len(want) {
		t.Errorf("%s subtypes = %v, want %v", c.Name, c.Subtypes, want)
		return
	}
	for _, w := range want {
		if !c.HasSubtype(w) {
			t.Errorf("%s subtypes = %v, want to include %q", c.Name, c.Subtypes, w)
		}
	}
}

func checkBody(t *testing.T, c *Card, row map[string]string) {
	t.Helper()
	if got := csvInt(t, row, "power"); c.Power != got {
		t.Errorf("%s Combat Power = %d, want %d", c.Name, c.Power, got)
	}
	if got := csvInt(t, row, "health"); c.Health != got {
		t.Errorf("%s Maximum Health = %d, want %d", c.Name, c.Health, got)
	}
	if got := csvInt(t, row, "requisitionYield"); c.ReqYield != got {
		t.Errorf("%s Requisition Yield = %d, want %d", c.Name, c.ReqYield, got)
	}
	if got := csvInt(t, row, "tacticsYield"); c.TacYield != got {
		t.Errorf("%s Tactics Yield = %d, want %d", c.Name, c.TacYield, got)
	}
	if want := csvZones(t, row["deploymentZones"]); !sameZones(c.DeployZones, want) {
		t.Errorf("%s deployment zones = %v, want %v", c.Name, c.DeployZones, want)
	}
}

func TestFR_CRD_101_UnitCarriesCSVCharacteristics(t *testing.T) {
	s := newScene(t)
	for _, name := range []string{"Shifting Impossibility", "Devilhorn Guerrillas"} {
		c := s.card(name)
		row := csvCard(t, name)
		checkCommon(t, c, row, TypeUnit)
		checkSubtypes(t, c, row["subtypes"])
		checkBody(t, c, row)
		if !strings.EqualFold(c.Rarity, row["rarity"]) {
			t.Errorf("%s rarity = %q, want %q", name, c.Rarity, row["rarity"])
		}
	}
	// Pin that the chosen row really has asymmetric values, so a swapped
	// column would fail rather than coincide.
	si := csvCard(t, "Shifting Impossibility")
	if si["requisitionCost"] == si["tacticsCost"] || si["power"] == si["health"] {
		t.Fatalf("test card Shifting Impossibility no longer has distinct cost/stat columns: %v", si)
	}
}

func TestFR_CRD_101_StructureCarriesCSVCharacteristics(t *testing.T) {
	s := newScene(t)
	name := "Cathedral of Self"
	c := s.card(name)
	row := csvCard(t, name)
	checkCommon(t, c, row, TypeStructure)
	checkSubtypes(t, c, row["subtypes"])
	checkBody(t, c, row)
	if !strings.EqualFold(c.Rarity, row["rarity"]) {
		t.Errorf("%s rarity = %q, want %q", name, c.Rarity, row["rarity"])
	}
	if row["power"] == row["health"] || row["requisitionCost"] == row["tacticsCost"] {
		t.Fatalf("test card %s no longer has distinct columns: %v", name, row)
	}
}

func TestFR_CRD_101_ManeuverCarriesCSVCharacteristics(t *testing.T) {
	// A Maneuver has no Combat Power, Maximum Health, yields or deployment
	// zones to check ("not applicable"); only name, type, faction, costs and
	// rarity are pinned.
	s := newScene(t)
	name := "Disorienting Pulse"
	c := s.card(name)
	row := csvCard(t, name)
	checkCommon(t, c, row, TypeManeuver)
	if !strings.EqualFold(c.Rarity, row["rarity"]) {
		t.Errorf("%s rarity = %q, want %q", name, c.Rarity, row["rarity"])
	}
	if row["requisitionCost"] == row["tacticsCost"] {
		t.Fatalf("test card %s no longer has distinct cost columns: %v", name, row)
	}
}

func TestFR_CRD_101_HeadquartersCarriesCSVCharacteristics(t *testing.T) {
	s := newScene(t)
	name := "Command Barge"
	c := s.card(name)
	row := csvCard(t, name)
	checkCommon(t, c, row, TypeHeadquarters)
	checkSubtypes(t, c, row["subtypes"]) // "--" means zero subtypes
	checkBody(t, c, row)
	if !strings.EqualFold(c.Rarity, "Headquarters") {
		t.Errorf("%s rarity = %q, want %q (FR-CRD-101 rarity legal values)", name, c.Rarity, "Headquarters")
	}
	if row["power"] == row["health"] || row["requisitionYield"] == row["tacticsYield"] {
		t.Fatalf("test card %s no longer has distinct columns: %v", name, row)
	}
}

// --- FR-ZON-003: Assets, Cards, Ownership, Control & Readiness ---

func TestFR_ZON_003_BattlefieldZonesHoldAssetsOthersHoldCards(t *testing.T) {
	for _, z := range []Zone{ZoneHeadquarters, ZoneReserves, ZoneMainTheater} {
		if !z.IsBattlefield() {
			t.Errorf("%v.IsBattlefield() = false, want true: its occupants are Assets", z)
		}
	}
	for _, z := range []Zone{ZoneHand, ZoneDeck, ZoneDiscard, ZoneStack} {
		if z.IsBattlefield() {
			t.Errorf("%v.IsBattlefield() = true, want false: its occupants are Cards", z)
		}
	}
	s := newScene(t)
	for p := PlayerID(0); p <= 1; p++ {
		if s.g.Players[p].HQ == nil || s.g.Players[p].HQ.Zone != ZoneHeadquarters || !s.alive(s.g.Players[p].HQ) {
			t.Errorf("player %d Headquarters should be an Asset in the Headquarters Zone", p)
		}
	}
}

// deployToReserves deploys a free unit from p's hand and returns its asset.
func deployToReserves(s *scene, p PlayerID) *Asset {
	s.t.Helper()
	inst := s.give(p, unit("Recruit", 1, 1, Keywords{}))
	s.do(Action{Kind: ActDeploy, Player: p, Card: inst.ID, Zone: ZoneReserves})
	for _, x := range s.g.Players[p].Reserves {
		if x.Inst.ID == inst.ID {
			return x
		}
	}
	s.t.Fatal("deployed unit is not in its deployer's Reserves")
	return nil
}

func TestFR_ZON_003_DeployerControlsAndDeckOwnerOwns(t *testing.T) {
	s := newScene(t)
	a := deployToReserves(s, 0)
	if a.Controller != 0 {
		t.Errorf("controller = %d, want 0 (the player who deployed it)", a.Controller)
	}
	if a.Inst.Owner != 0 {
		t.Errorf("owner = %d, want 0 (the player whose deck it began in)", a.Inst.Owner)
	}
}

func TestFR_ZON_003_OwnerUnchangedByMovingBetweenZones(t *testing.T) {
	s := newScene(t)
	a := deployToReserves(s, 0)
	s.do(Action{Kind: ActMarch, Player: 0, Asset: a.ID})
	m := s.g.Asset(a.ID)
	if m == nil || m.Zone != ZoneMainTheater {
		t.Fatalf("marched unit should be in the Main Theater, got %v", m)
	}
	if m.Inst.Owner != 0 || m.Controller != 0 {
		t.Errorf("after March owner=%d controller=%d, want 0 and 0", m.Inst.Owner, m.Controller)
	}
}

func TestFR_ZON_003_ReadiedAssetCanPayExhaustCostOnce(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Worker", 1, 1, Keywords{}))
	if u.Exhausted {
		t.Fatal("a newly placed asset should be Readied")
	}
	before := s.g.Players[0].Req
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: u.ID, Resource: Requisition})
	if !u.Exhausted {
		t.Errorf("asset should be Exhausted after paying its exhaust cost")
	}
	if got := s.g.Players[0].Req; got != before+u.Inst.Card.ReqYield {
		t.Errorf("Requisition = %d, want %d (Readied asset pays exhaust, adds its Requisition Yield)", got, before+u.Inst.Card.ReqYield)
	}
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: u.ID, Resource: Requisition})
}

func TestFR_ZON_003_ExhaustedAssetCannotPayExhaustCost(t *testing.T) {
	s := newScene(t)
	u := s.put(0, ZoneReserves, unit("Tired Worker", 1, 1, Keywords{}))
	u.Exhausted = true
	before := s.g.Players[0].Req
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: u.ID, Resource: Requisition})
	if got := s.g.Players[0].Req; got != before {
		t.Errorf("Requisition = %d, want unchanged %d: an Exhausted asset cannot pay exhaust costs", got, before)
	}
}

// Not repeated here: the owner's-discard rule is pinned under FR-ZON-002, and
// owner-versus-controller on real cards by T-0014's card tests. A
// control-changing effect cannot be built through the public API without
// internal fields, so control diverging from ownership is left out.

// --- FR-ZON-001: Zone Definitions & Uncapped Capacity ---

func TestFR_ZON_001_HandOverSixSurvivesEndOfTurnWithoutDiscard(t *testing.T) {
	s := newScene(t)
	for i := 0; i < 15; i++ {
		s.give(0, unit("Extra", 1, 1, Keywords{}))
	}
	hand := len(s.g.Players[0].Hand)
	if hand <= 6 {
		t.Fatalf("setup: hand = %d, want far more than 6", hand)
	}
	discard := len(s.g.Players[0].Discard)
	s.endTurn()
	if s.g.Active != 1 {
		t.Fatalf("active = %d after ending Turn 1, want 1", s.g.Active)
	}
	if got := len(s.g.Players[0].Hand); got != hand {
		t.Errorf("hand = %d after end of turn, want %d (no hand-size limit, no end-of-turn discard)", got, hand)
	}
	if got := len(s.g.Players[0].Discard); got != discard {
		t.Errorf("discard = %d after end of turn, want %d (nothing discarded)", got, discard)
	}
}

func TestFR_ZON_001_ReservesAndMainTheaterAcceptManyAssets(t *testing.T) {
	s := newScene(t)
	const n = 12
	for i := 0; i < n; i++ {
		r := s.give(0, unit("Reservist", 1, 1, Keywords{}))
		s.do(Action{Kind: ActDeploy, Player: 0, Card: r.ID, Zone: ZoneReserves})
		m := s.give(0, unit("Trooper", 1, 1, Keywords{}))
		s.do(Action{Kind: ActDeploy, Player: 0, Card: m.ID, Zone: ZoneMainTheater})
	}
	if got := len(s.g.Players[0].Reserves); got != n {
		t.Errorf("Reserves holds %d assets, want %d (no capacity limit)", got, n)
	}
	if got := len(s.g.TheaterOf(0)); got != n {
		t.Errorf("Main Theater holds %d of player 0's assets, want %d (no capacity limit)", got, n)
	}
}

func TestFR_ZON_001_ReservesGenerateRequisitionTheaterGeneratesTactics(t *testing.T) {
	s := newScene(t)
	r := s.put(0, ZoneReserves, unit("Reservist", 1, 1, Keywords{}))
	m := s.put(0, ZoneMainTheater, unit("Trooper", 1, 1, Keywords{}))
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: r.ID, Resource: Tactics})
	s.reject(Action{Kind: ActGenerate, Player: 0, Asset: m.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: r.ID, Resource: Requisition})
	s.do(Action{Kind: ActGenerate, Player: 0, Asset: m.ID, Resource: Tactics})
	if s.g.Players[0].Req != 1 || s.g.Players[0].Tac != 1 {
		t.Errorf("pools Req=%d Tac=%d, want 1 and 1 (Reserves yield Requisition, Main Theater yields Tactics)",
			s.g.Players[0].Req, s.g.Players[0].Tac)
	}
}

func TestFR_ZON_001_ReservesAssetsAreProtectedFromAttacks(t *testing.T) {
	s := newScene(t)
	att := s.put(0, ZoneMainTheater, unit("Raider", 2, 2, Keywords{}))
	safe := s.put(1, ZoneReserves, unit("Backliner", 1, 1, Keywords{}))
	if s.g.CanAttack(att, safe) {
		t.Errorf("an asset in Reserves must be protected from combat attacks")
	}
}
