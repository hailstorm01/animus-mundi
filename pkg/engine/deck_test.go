package engine

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Deck-construction fixtures. They build synthetic cards only; ValidateDeck
// itself is the code under test (FR-SET-001).

func deckHQ(faction string) *Card {
	return &Card{Name: "Deck HQ " + faction, Type: TypeHeadquarters, Faction: faction, Rarity: "HQ",
		Power: 3, Health: 20, DeployZones: []Zone{ZoneHeadquarters}}
}

func deckCard(name, faction, rarity string) *Card {
	return &Card{Name: name, Type: TypeUnit, Faction: faction, Rarity: rarity, Power: 1, Health: 1,
		DeployZones: []Zone{ZoneReserves, ZoneMainTheater}}
}

// legalMain returns n distinct-named COMMON cards of the given faction.
func legalMain(n int, faction string) []*Card {
	out := make([]*Card, n)
	for i := range out {
		out[i] = deckCard(fmt.Sprintf("Deck Filler %d", i), faction, "COMMON")
	}
	return out
}

func copies(c *Card, n int) []*Card {
	out := make([]*Card, n)
	for i := range out {
		cp := *c
		out[i] = &cp
	}
	return out
}

// legalDeck is a 50-card ANIMUS deck with an ANIMUS Headquarters.
func legalDeck() DeckList {
	return DeckList{HQ: deckHQ("ANIMUS"), Main: legalMain(50, "ANIMUS")}
}

func wantLegal(t *testing.T, d DeckList, why string) {
	t.Helper()
	if err := ValidateDeck(d); err != nil {
		t.Errorf("FR-SET-001: %s should be legal, got error: %v", why, err)
	}
}

func wantIllegal(t *testing.T, d DeckList, why string, names ...string) error {
	t.Helper()
	err := ValidateDeck(d)
	if err == nil {
		t.Errorf("FR-SET-001: %s should be illegal, got nil", why)
		return nil
	}
	if !errors.Is(err, ErrIllegalDeck) {
		t.Errorf("FR-SET-001: %s: error must wrap ErrIllegalDeck, got: %v", why, err)
	}
	for _, n := range names {
		if !strings.Contains(err.Error(), n) {
			t.Errorf("FR-SET-001: %s: error must name %q, got: %v", why, n, err)
		}
	}
	return err
}

func TestFR_SET_001_LegalDeckAccepted(t *testing.T) {
	wantLegal(t, legalDeck(), "a 50-card faction deck with a matching HQ")
}

func TestFR_SET_001_FiftyCardsIsLegal(t *testing.T) {
	d := DeckList{HQ: deckHQ("PATH"), Main: legalMain(50, "PATH")}
	wantLegal(t, d, "a main deck of exactly 50 cards")
}

func TestFR_SET_001_FortyNineCardsIsIllegal(t *testing.T) {
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: legalMain(49, "ANIMUS")}
	wantIllegal(t, d, "a 49-card main deck (minimum is 50)")
}

func TestFR_SET_001_EmptyMainDeckIsIllegal(t *testing.T) {
	wantIllegal(t, DeckList{HQ: deckHQ("ANIMUS")}, "an empty main deck")
}

func TestFR_SET_001_NoMaximumDeckSize(t *testing.T) {
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: legalMain(200, "ANIMUS")}
	wantLegal(t, d, "a 200-card main deck (no upper maximum)")
}

func TestFR_SET_001_MissingHQIsIllegal(t *testing.T) {
	d := legalDeck()
	d.HQ = nil
	wantIllegal(t, d, "a deck with no Headquarters")
}

func TestFR_SET_001_HQMustBeHeadquartersCard(t *testing.T) {
	d := legalDeck()
	d.HQ = deckCard("Deck Not An HQ", "ANIMUS", "COMMON")
	wantIllegal(t, d, "a designated HQ that is not a Headquarters card", "Deck Not An HQ")
}

func TestFR_SET_001_NeutralHQIsIllegal(t *testing.T) {
	d := legalDeck()
	d.HQ = deckHQ("NEUTRAL")
	wantIllegal(t, d, "a Neutral Headquarters", d.HQ.Name)
}

func TestFR_SET_001_HQInMainDeckIsIllegal(t *testing.T) {
	d := legalDeck()
	extra := deckHQ("ANIMUS")
	extra.Name = "Deck Stray HQ"
	d.Main = append(d.Main, extra)
	wantIllegal(t, d, "a Headquarters card in the main deck", "Deck Stray HQ")
}

func TestFR_SET_001_OffFactionCardIsIllegal(t *testing.T) {
	d := legalDeck()
	d.Main[0] = deckCard("Deck Foreign Unit", "PATH", "COMMON")
	wantIllegal(t, d, "a PATH card in an ANIMUS deck", "Deck Foreign Unit")
}

func TestFR_SET_001_OffFactionIsIllegalForEveryPairing(t *testing.T) {
	factions := []string{"CONCORDIA", "SOULLESS", "PATH", "ANIMUS"}
	for _, hqF := range factions {
		for _, cardF := range factions {
			if hqF == cardF {
				continue
			}
			d := DeckList{HQ: deckHQ(hqF), Main: legalMain(50, hqF)}
			d.Main[0] = deckCard("Deck Foreign Unit", cardF, "COMMON")
			wantIllegal(t, d, fmt.Sprintf("a %s card in a %s deck", cardF, hqF), "Deck Foreign Unit")
		}
	}
}

func TestFR_SET_001_NeutralCardsLegalInAnyDeck(t *testing.T) {
	for _, hqF := range []string{"CONCORDIA", "SOULLESS", "PATH", "ANIMUS"} {
		d := DeckList{HQ: deckHQ(hqF), Main: legalMain(50, hqF)}
		d.Main[0] = deckCard("Deck Neutral Unit", "NEUTRAL", "COMMON")
		wantLegal(t, d, "a Neutral card in a "+hqF+" deck")
	}
}

func TestFR_SET_001_FactionComparedCaseInsensitively(t *testing.T) {
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: legalMain(50, "Animus")}
	wantLegal(t, d, "main-deck factions written in a different case than the HQ")
}

func TestFR_SET_001_ScratchCompanyInMainDeckIsIllegal(t *testing.T) {
	d := legalDeck()
	sc := deckCard("Scratch Company", "NEUTRAL", "COMMON")
	sc.Subtypes = []string{"Scratch Company"}
	d.Main[0] = sc
	wantIllegal(t, d, "a Scratch Company token in the main deck", "Scratch Company")
}

func TestFR_SET_001_CommonLimitIsFour(t *testing.T) {
	c := deckCard("Deck Common", "ANIMUS", "COMMON")
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: append(copies(c, 4), legalMain(46, "ANIMUS")...)}
	wantLegal(t, d, "4 copies of a Common")
	d.Main = append(copies(c, 5), legalMain(45, "ANIMUS")...)
	wantIllegal(t, d, "5 copies of a Common", "Deck Common")
}

func TestFR_SET_001_UncommonLimitIsThree(t *testing.T) {
	c := deckCard("Deck Uncommon", "ANIMUS", "UNCOMMON")
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: append(copies(c, 3), legalMain(47, "ANIMUS")...)}
	wantLegal(t, d, "3 copies of an Uncommon")
	d.Main = append(copies(c, 4), legalMain(46, "ANIMUS")...)
	wantIllegal(t, d, "4 copies of an Uncommon", "Deck Uncommon")
}

func TestFR_SET_001_RareLimitIsTwo(t *testing.T) {
	c := deckCard("Deck Rare", "ANIMUS", "RARE")
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: append(copies(c, 2), legalMain(48, "ANIMUS")...)}
	wantLegal(t, d, "2 copies of a Rare")
	d.Main = append(copies(c, 3), legalMain(47, "ANIMUS")...)
	wantIllegal(t, d, "3 copies of a Rare", "Deck Rare")
}

func TestFR_SET_001_MythicLimitIsOne(t *testing.T) {
	c := deckCard("Deck Mythic", "ANIMUS", "MYTHIC")
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: append(copies(c, 1), legalMain(49, "ANIMUS")...)}
	wantLegal(t, d, "1 copy of a Mythic")
	d.Main = append(copies(c, 2), legalMain(48, "ANIMUS")...)
	wantIllegal(t, d, "2 copies of a Mythic", "Deck Mythic")
}

func TestFR_SET_001_CopyLimitCountsByPrintedName(t *testing.T) {
	// Three distinct *Card values sharing a name count together.
	a := deckCard("Deck Twin", "ANIMUS", "RARE")
	b := deckCard("Deck Twin", "ANIMUS", "RARE")
	c := deckCard("Deck Twin", "ANIMUS", "RARE")
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: append([]*Card{a, b, c}, legalMain(47, "ANIMUS")...)}
	wantIllegal(t, d, "3 same-named Rares held in distinct card values", "Deck Twin")
}

func TestFR_SET_001_LimitsApplyPerName(t *testing.T) {
	// Different names do not pool: 2 Mythics with different names are legal.
	d := legalDeck()
	d.Main[0] = deckCard("Deck Mythic A", "ANIMUS", "MYTHIC")
	d.Main[1] = deckCard("Deck Mythic B", "ANIMUS", "MYTHIC")
	wantLegal(t, d, "two Mythics with different names")
}

func TestFR_SET_001_UnknownRarityIsIllegal(t *testing.T) {
	d := legalDeck()
	d.Main[0] = deckCard("Deck Odd Rarity", "ANIMUS", "LEGENDARY")
	wantIllegal(t, d, "a main-deck card whose rarity is not Common/Uncommon/Rare/Mythic", "Deck Odd Rarity")
}

func TestFR_SET_001_AllViolationsReportedTogether(t *testing.T) {
	over := deckCard("Deck Overstack", "ANIMUS", "MYTHIC")
	foreign := deckCard("Deck Foreign Unit", "SOULLESS", "COMMON")
	main := append(copies(over, 2), foreign)
	main = append(main, legalMain(10, "ANIMUS")...) // 13 cards: also under 50
	d := DeckList{HQ: deckHQ("ANIMUS"), Main: main}
	// A faction fault, a copy-limit fault and a size fault, all in one deck.
	wantIllegal(t, d, "a deck with several independent violations", "Deck Overstack", "Deck Foreign Unit")
}
