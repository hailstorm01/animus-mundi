package frd

import (
	"strings"
	"testing"
)

var rulesFixture = []string{
	"#### FR-CBT-002: 3-Step Combat Sequence",
	"- **FR-KWD-002c: Stalwart**: Caps combat damage at 1.",
	"- **FR-KWD-002o: Berserk N**: Deals N secondary damage.",
}

func TestCheckOwnershipFlagsDuplicateDefinitions(t *testing.T) {
	problems, _ := CheckOwnership(map[string][]string{
		RulesFRD:           rulesFixture,
		"05_Engine_FRD.md": {"#### FR-CBT-002: Combat again", "See `FR-KWD-002o`."},
	})
	if len(problems) != 1 || !strings.Contains(problems[0], "FR-CBT-002") {
		t.Fatalf("problems = %v, want one duplicate-definition problem for FR-CBT-002", problems)
	}
}

func TestCheckOwnershipFlagsKeywordRestatementOutsideFRD01(t *testing.T) {
	problems, _ := CheckOwnership(map[string][]string{
		RulesFRD:           rulesFixture,
		"06_Server_FRD.md": {"| **Berserk N** | cleaves only on a kill |"},
	})
	if len(problems) != 1 || !strings.Contains(problems[0], `"Berserk"`) {
		t.Fatalf("problems = %v, want one Berserk restatement", problems)
	}
}

func TestCheckOwnershipSkipsFrozenFRDsAndTheMatrix(t *testing.T) {
	problems, frozen := CheckOwnership(map[string][]string{
		RulesFRD:           rulesFixture,
		"02_Data_FRD.md":   {"> **Status: FROZEN as design intent.**", "| **Stalwart** | KEYWORD_GRANT |"},
		KeywordMatrix:      {"- **Stalwart**: quoted to pose a question"},
		"07_Client_FRD.md": {"Units with **Stalwart** show a shield icon."},
	})
	if len(problems) != 0 || frozen != 1 {
		t.Fatalf("problems=%v frozen=%d, want none and 1", problems, frozen)
	}
}
