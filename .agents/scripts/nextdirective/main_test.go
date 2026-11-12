package main

import "testing"

func TestParseLast(t *testing.T) {
	board := "# Board\n\n- **Integration branch:** `main`\n" +
		"- **Last directive:** `#143` (applied directives live in [x](y))\n" +
		"- **Last task:** `T-0217`\n"
	if n, ok := parseLast(board); !ok || n != 143 {
		t.Fatalf("parseLast = %d, %v; want 143, true", n, ok)
	}
	if _, ok := parseLast("- **Last task:** `T-0217`\n"); ok {
		t.Fatal("parseLast found a counter in a board without one")
	}
}
