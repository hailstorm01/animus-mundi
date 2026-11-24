package board

import "testing"

func TestLastDirective(t *testing.T) {
	text := "# Board\n\n- **Integration branch:** `main`\n" +
		"- **Last directive:** `#143` (applied directives live in [x](y))\n" +
		"- **Last task:** `T-0217`\n"
	if n, ok := LastDirective(text); !ok || n != 143 {
		t.Fatalf("LastDirective = %d, %v; want 143, true", n, ok)
	}
	if _, ok := LastDirective("- **Last task:** `T-0217`\n"); ok {
		t.Fatal("LastDirective found a counter in a board without one")
	}
}

func TestCurrent(t *testing.T) {
	counting := map[string]bool{"m2": true, "m3": true, "m10": true}
	only := func(names ...string) func(string) bool {
		return func(m string) bool {
			for _, n := range names {
				if m == n {
					return counting[m]
				}
			}
			return false
		}
	}
	for _, c := range []struct {
		name    string
		folders []string
		counts  func(string) bool
		want    string
		ok      bool
	}{
		// M2 closed and dropped its counter; M3 carries it.
		{"closed board skipped", []string{"m2", "m3"}, only("m3"), "m3", true},
		// A folder for the next milestone without a counter yet doesn't count.
		{"next board not started", []string{"m2", "m3", "m4"}, only("m2", "m3"), "m3", true},
		// m10 sorts after m9 by number, not before m2 by text.
		{"numeric order", []string{"m10", "m2", "m9"}, only("m2", "m10"), "m10", true},
		{"other folders ignored", []string{"README.md", "archive", "m3"}, only("m3", "archive"), "m3", true},
		{"no counter anywhere", []string{"m2", "m3"}, only(), "", false},
	} {
		got, ok := current(c.folders, c.counts)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: current = %q, %t; want %q, %t", c.name, got, ok, c.want, c.ok)
		}
	}
}
