package engine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func writeDeckFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	p := filepath.Join(dir, filename)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write deck file %s: %v", p, err)
	}
	return p
}

func TestLoadDeck_Case1_ValidFile(t *testing.T) {
	cards := map[string]*Card{
		"Commander": deckHQ("ANIMUS"),
		"Trooper":   deckCard("Trooper", "ANIMUS", "COMMON"),
		"Scout":     deckCard("Scout", "ANIMUS", "COMMON"),
	}
	content := "HQ: Commander\n2 Trooper\n1 Scout\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	d, err := LoadDeck(p, cards)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.HQ != cards["Commander"] {
		t.Errorf("expected HQ Commander, got %v", d.HQ)
	}
	if len(d.Main) != 3 {
		t.Fatalf("expected 3 cards in Main, got %d", len(d.Main))
	}
	if d.Main[0] != cards["Trooper"] || d.Main[1] != cards["Trooper"] || d.Main[2] != cards["Scout"] {
		t.Errorf("unexpected Main order or content: %v, %v, %v", d.Main[0].Name, d.Main[1].Name, d.Main[2].Name)
	}
}

func TestLoadDeck_Case2_CommentsBlankLinesWhitespace(t *testing.T) {
	cmd := deckHQ("ANIMUS")
	cmd.Name = "Commander"
	cards := map[string]*Card{
		"Commander": cmd,
		"Trooper":   deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	content := "\n  # this is a comment\n\n  HQ: Commander  \n\n# another comment\n  2 Trooper  \n\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	d, err := LoadDeck(p, cards)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.HQ == nil || d.HQ.Name != "Commander" {
		t.Errorf("expected HQ Commander, got %v", d.HQ)
	}
	if len(d.Main) != 2 || d.Main[0].Name != "Trooper" || d.Main[1].Name != "Trooper" {
		t.Errorf("unexpected Main: %v", d.Main)
	}
}

func TestLoadDeck_Case3_UnknownCardName(t *testing.T) {
	cards := map[string]*Card{
		"Commander": deckHQ("ANIMUS"),
	}
	content := "HQ: Commander\n\n1 Ghost Fighter\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	_, err := LoadDeck(p, cards)
	if err == nil {
		t.Fatal("expected error for unknown card, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Ghost Fighter") {
		t.Errorf("expected error to name 'Ghost Fighter', got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "3") {
		t.Errorf("expected error to mention line number 3, got: %v", errMsg)
	}
}

func TestLoadDeck_Case4_HQLineCount(t *testing.T) {
	cards := map[string]*Card{
		"Commander":  deckHQ("ANIMUS"),
		"Commander2": deckHQ("ANIMUS"),
		"Trooper":    deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	t.Run("no HQ line", func(t *testing.T) {
		p := writeDeckFile(t, t.TempDir(), "deck.txt", "2 Trooper\n")
		_, err := LoadDeck(p, cards)
		if err == nil {
			t.Fatal("expected error when no HQ line, got nil")
		}
	})
	t.Run("two HQ lines", func(t *testing.T) {
		p := writeDeckFile(t, t.TempDir(), "deck.txt", "HQ: Commander\nHQ: Commander2\n2 Trooper\n")
		_, err := LoadDeck(p, cards)
		if err == nil {
			t.Fatal("expected error when two HQ lines, got nil")
		}
	})
}

func TestLoadDeck_Case5_InvalidCount(t *testing.T) {
	cards := map[string]*Card{
		"Commander": deckHQ("ANIMUS"),
		"Trooper":   deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	tests := []struct {
		name    string
		content string
		badLine int
	}{
		{"zero count", "HQ: Commander\n0 Trooper\n", 2},
		{"negative count", "HQ: Commander\n-1 Trooper\n", 2},
		{"word count", "HQ: Commander\ntwo Trooper\n", 2},
		{"no count", "HQ: Commander\nTrooper\n", 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := writeDeckFile(t, t.TempDir(), "deck.txt", tc.content)
			_, err := LoadDeck(p, cards)
			if err == nil {
				t.Fatalf("%s: expected error, got nil", tc.name)
			}
			lineStr := strconv.Itoa(tc.badLine)
			if !strings.Contains(err.Error(), lineStr) {
				t.Errorf("%s: error should mention line number %s, got: %v", tc.name, lineStr, err)
			}
		})
	}
}

func TestLoadDeck_Case6_DuplicateCard(t *testing.T) {
	cards := map[string]*Card{
		"Commander": deckHQ("ANIMUS"),
		"Trooper":   deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	content := "HQ: Commander\n2 Trooper\n1 Trooper\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	_, err := LoadDeck(p, cards)
	if err == nil {
		t.Fatal("expected error for duplicate card on multiple lines, got nil")
	}
	if !strings.Contains(err.Error(), "Trooper") {
		t.Errorf("error should name the duplicate card 'Trooper', got: %v", err)
	}
}

func TestLoadDeck_Case7_ResolvableIllegalDeck(t *testing.T) {
	cards := map[string]*Card{
		"Deck HQ ANIMUS": deckHQ("ANIMUS"),
		"Trooper":        deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	content := "HQ: Deck HQ ANIMUS\n49 Trooper\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	d, err := LoadDeck(p, cards)
	if err != nil {
		t.Fatalf("LoadDeck should succeed on resolvable illegal deck, got error: %v", err)
	}
	verr := ValidateDeck(d)
	if verr == nil {
		t.Fatal("ValidateDeck should fail for 49-card deck, got nil")
	}
	if !errors.Is(verr, ErrIllegalDeck) {
		t.Errorf("ValidateDeck error should wrap ErrIllegalDeck, got: %v", verr)
	}
}

func TestLoadDeck_Case8_MissingFile(t *testing.T) {
	cards := map[string]*Card{}
	missingPath := filepath.Join(t.TempDir(), "no-such-file.txt")
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LoadDeck panicked on missing file: %v", r)
		}
	}()
	_, err := LoadDeck(missingPath, cards)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadDeck_Case9_ExactNames(t *testing.T) {
	cards := map[string]*Card{
		"Despoiler Larva": deckCard("Despoiler Larva", "ANIMUS", "COMMON"),
		"HQ":              deckHQ("ANIMUS"),
	}
	content := "HQ: HQ\n1 despoiler larva\n"
	p := writeDeckFile(t, t.TempDir(), "deck.txt", content)
	_, err := LoadDeck(p, cards)
	if err == nil {
		t.Fatal("expected error for casing mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "despoiler larva") {
		t.Errorf("error should name 'despoiler larva', got: %v", err)
	}
	if !strings.Contains(err.Error(), "2") {
		t.Errorf("error should name line number 2, got: %v", err)
	}
}

func TestLoadDeck_Case10_AnimusStarter(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	path := filepath.Join("..", "..", "data", "decks", "animus-starter.txt")
	d, err := LoadDeck(path, cards)
	if err != nil {
		t.Fatalf("failed to load animus-starter.txt: %v", err)
	}
	if d.HQ == nil || d.HQ.Name != "Animoceps Rager" {
		t.Errorf("expected HQ Animoceps Rager, got %v", d.HQ)
	}
	if len(d.Main) != 50 {
		t.Errorf("expected 50 main cards, got %d", len(d.Main))
	}
	names := map[string]bool{}
	for _, c := range d.Main {
		names[c.Name] = true
	}
	if len(names) != 21 {
		t.Errorf("expected 21 distinct names, got %d", len(names))
	}
	if err := ValidateDeck(d); err != nil {
		t.Errorf("expected ValidateDeck to return nil, got: %v", err)
	}
}

func TestLoadDeck_Case11_PathStarter(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	path := filepath.Join("..", "..", "data", "decks", "path-starter.txt")
	d, err := LoadDeck(path, cards)
	if err != nil {
		t.Fatalf("failed to load path-starter.txt: %v", err)
	}
	if d.HQ == nil || d.HQ.Name != "Eternal Watchman" {
		t.Errorf("expected HQ Eternal Watchman, got %v", d.HQ)
	}
	if len(d.Main) != 50 {
		t.Errorf("expected 50 main cards, got %d", len(d.Main))
	}
	names := map[string]bool{}
	for _, c := range d.Main {
		names[c.Name] = true
	}
	if len(names) != 23 {
		t.Errorf("expected 23 distinct names, got %d", len(names))
	}
	if err := ValidateDeck(d); err != nil {
		t.Errorf("expected ValidateDeck to return nil, got: %v", err)
	}
}

func TestLoadDeck_Case12_DesignNoteCountsMatch(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	wantAnimus := map[string]int{
		"Despoiler Larva":        4,
		"Portent Scrivener":      4,
		"Shifting Impossibility": 2,
		"Animoceps Shambler":     3,
		"Sinuous Growth":         4,
		"Cliff of Portents":      1,
		"Spoil Forest Creeper":   4,
		"Stormfront of Ideals":   2,
		"Unceasing Tide":         3,
		"Veil Render":            1,
		"Infested Apostle":       4,
		"Spoil Forest Amalgam":   1,
		"Spoil Forest Colossus":  2,
		"Despoiler Worm":         1,
		"Craterfall Ancient":     2,
		"Despoiler's Emergence":  1,
		"Forced Mutation":        2,
		"Rapturous Sermon":       2,
		"Animoceps Spores":       3,
		"Rhapsody's Embrace":     2,
		"Funeral Fungus":         2,
	}
	wantPath := map[string]int{
		"Chainbreaker Sneak":             4,
		"Eerie Remnant":                  3,
		"Hammerhand Brigands":            3,
		"Broken Horn Armorer":            2,
		"Macabre Dancers":                2,
		"Devilhorn Captain":              3,
		"Chainbreaker Irregulars":        4,
		"Necropolis Hunt":                1,
		"Saboteur":                       3,
		"Boneyard Scavenger":             2,
		"Chainbreaker Phalanx":           2,
		"Devilhorn Guerrillas":           3,
		"Allshan, Chainbreaker Revenant": 1,
		"Hammerhand Brigade":             2,
		"Arkha, Sun's Fury":              1,
		"Boneyard Amalgam":               2,
		"Allshan's Battle Trance":        1,
		"Ambush Tactics":                 2,
		"Steel Rain":                     2,
		"Devilhorn Sanction":             2,
		"Improvised Orders":              1,
		"Reaver's Rush":                  2,
		"Devilhorn Scouting Party":       2,
	}

	animusDeck, err := LoadDeck(filepath.Join("..", "..", "data", "decks", "animus-starter.txt"), cards)
	if err != nil {
		t.Fatalf("failed to load animus starter deck: %v", err)
	}
	animusCounts := map[string]int{}
	for _, c := range animusDeck.Main {
		animusCounts[c.Name]++
	}
	if !reflect.DeepEqual(animusCounts, wantAnimus) {
		t.Errorf("Animus counts do not match design note table.\nGot:  %v\nWant: %v", animusCounts, wantAnimus)
	}

	pathDeck, err := LoadDeck(filepath.Join("..", "..", "data", "decks", "path-starter.txt"), cards)
	if err != nil {
		t.Fatalf("failed to load path starter deck: %v", err)
	}
	pathCounts := map[string]int{}
	for _, c := range pathDeck.Main {
		pathCounts[c.Name]++
	}
	if !reflect.DeepEqual(pathCounts, wantPath) {
		t.Errorf("Path counts do not match design note table.\nGot:  %v\nWant: %v", pathCounts, wantPath)
	}
}

func TestSimDecks_Case13_BuildsValidDecks(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	decksDir := filepath.Join("..", "..", "data", "decks")
	decks, err := SimDecks(cards, decksDir, "starter", [2]string{"ANIMUS", "PATH"})
	if err != nil {
		t.Fatalf("SimDecks failed: %v", err)
	}
	if decks[0].HQ == nil || decks[0].HQ.Name != "Animoceps Rager" {
		t.Errorf("expected P0 HQ Animoceps Rager, got %v", decks[0].HQ)
	}
	if len(decks[0].Main) != 50 {
		t.Errorf("expected P0 50 main cards, got %d", len(decks[0].Main))
	}
	if decks[1].HQ == nil || decks[1].HQ.Name != "Eternal Watchman" {
		t.Errorf("expected P1 HQ Eternal Watchman, got %v", decks[1].HQ)
	}
	if len(decks[1].Main) != 50 {
		t.Errorf("expected P1 50 main cards, got %d", len(decks[1].Main))
	}
}

func TestSimDecks_Case14_MissingFileNamesPath(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	decksDir := filepath.Join("..", "..", "data", "decks")
	_, err = SimDecks(cards, decksDir, "starter", [2]string{"SOULLESS", "PATH"})
	if err == nil {
		t.Fatal("expected error for missing soulless-starter.txt, got nil")
	}
	if !strings.Contains(err.Error(), "soulless-starter.txt") {
		t.Errorf("expected error to name missing file 'soulless-starter.txt', got: %v", err)
	}
}

func TestSimDecks_Case15_InvalidDeckFails(t *testing.T) {
	cards := map[string]*Card{
		"Deck HQ ANIMUS": deckHQ("ANIMUS"),
		"Trooper":        deckCard("Trooper", "ANIMUS", "COMMON"),
	}
	dir := t.TempDir()
	writeDeckFile(t, dir, "animus-invalid.txt", "HQ: Deck HQ ANIMUS\n10 Trooper\n")
	_, err := SimDecks(cards, dir, "invalid", [2]string{"ANIMUS", "ANIMUS"})
	if err == nil {
		t.Fatal("expected error when ValidateDeck fails on loaded deck, got nil")
	}
	if !errors.Is(err, ErrIllegalDeck) {
		t.Errorf("expected error to wrap ErrIllegalDeck, got: %v", err)
	}
}

func TestSimDecks_Case16_FactionDeckStressDecksUnchanged(t *testing.T) {
	cards, err := LoadCards(filepath.Join("..", "..", "data", "cards"))
	if err != nil {
		t.Fatalf("failed to load cards: %v", err)
	}
	for _, f := range []string{"SOULLESS", "CONCORDIA"} {
		d, err := FactionDeck(cards, f)
		if err != nil {
			t.Fatalf("FactionDeck(%s) failed: %v", f, err)
		}
		if d.HQ == nil {
			t.Errorf("expected HQ for %s", f)
		}
		if len(d.Main) != 50 {
			t.Errorf("expected 50 cards for %s stress deck, got %d", f, len(d.Main))
		}
	}
}
