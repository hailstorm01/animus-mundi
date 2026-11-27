package cards

import (
	"fmt"
	"strings"
	"testing"

	"animus-mundi/pkg/engine/ability"
)

// TestDefs_RegisterRejectsRepeat registers one name that is not a card twice:
// the second register panics and names it (directive 192).
func TestDefs_RegisterRejectsRepeat(t *testing.T) {
	const name = "Not A Card (TestDefs_RegisterRejectsRepeat)"
	if _, ok := Defs[name]; ok {
		t.Fatalf("Defs already holds %q before the test", name)
	}
	t.Cleanup(func() { delete(Defs, name) })

	register(map[string]*ability.Def{name: {}})
	if _, ok := Defs[name]; !ok {
		t.Fatalf("register did not add %q to Defs", name)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("registering %q twice did not panic", name)
		}
		if msg := fmt.Sprint(r); !strings.Contains(msg, name) {
			t.Fatalf("panic %q does not name %q", msg, name)
		}
	}()
	register(map[string]*ability.Def{name: {}})
}
