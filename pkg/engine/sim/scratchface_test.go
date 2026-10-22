package sim

import (
	"slices"
	"testing"

	"animus-mundi/pkg/engine"
)

// Tests for T-0034 Acceptance Case 1: the exported Scratch Company face (FR-ZON-005).

func TestScratchFace_Characteristics(t *testing.T) {
	c := engine.ScratchFace()
	if c.Name != "Scratch Company" || c.Type != engine.TypeUnit {
		t.Errorf("face is %q %v, want a Unit named Scratch Company", c.Name, c.Type)
	}
	if c.Power != 1 || c.Health != 1 || c.ReqYield != 1 || c.TacYield != 1 {
		t.Errorf("face is %d/%d yielding %d/%d, want 1/1 yielding 1/1", c.Power, c.Health, c.ReqYield, c.TacYield)
	}
	if want := []string{"Infantry", "Scratch Company"}; !slices.Equal(c.Subtypes, want) {
		t.Errorf("subtypes = %v, want %v", c.Subtypes, want)
	}
	if c.Keywords != (engine.Keywords{}) {
		t.Errorf("keywords = %+v, want none", c.Keywords)
	}
}

func TestScratchFace_ReturnsACopy(t *testing.T) {
	c := engine.ScratchFace()
	c.Name = "Changed"
	c.Power = 9
	c.Keywords.Ward = true
	p := (&engine.Asset{FaceDown: true}).Printed()
	if p.Name != "Scratch Company" || p.Power != 1 || p.Keywords.Ward {
		t.Errorf("a face-down asset's Printed() is %q %d ward %t after the copy changed, want it untouched", p.Name, p.Power, p.Keywords.Ward)
	}
}
