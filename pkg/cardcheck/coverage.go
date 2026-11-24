// Package cardcheck holds checks over the whole card set that are not game
// rules (M3 design note, Boundaries): the executable-coverage checker and the
// definition-to-text checker, which this package's tests run against
// data/cards in every go tool task check. It imports engine, ability and
// cards, and nothing imports it.
package cardcheck

import (
	"fmt"
	"sort"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/engine/ability"
)

// checkExecutable compares every card the kernel cannot execute against allow
// (cards.NotYetExecutable). Each such card must be listed, each listed name
// must be a card that still cannot run, and allow must be sorted with no name
// twice. Card tasks change only cards.NotYetExecutable, never this checker.
func checkExecutable(all map[string]*engine.Card, allow []string) []string {
	var problems []string
	listed := make(map[string]bool)
	for i, name := range allow {
		if listed[name] {
			problems = append(problems, fmt.Sprintf("NotYetExecutable lists %q more than once", name))
		} else if i > 0 && name < allow[i-1] {
			problems = append(problems, fmt.Sprintf("NotYetExecutable is not sorted at %q", name))
		}
		listed[name] = true
		c, ok := all[name]
		switch {
		case !ok:
			problems = append(problems, fmt.Sprintf("NotYetExecutable names %q, which matches no card", name))
		case c.Supported():
			problems = append(problems, fmt.Sprintf("%q is executable now: remove it from NotYetExecutable", name))
		}
	}
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !all[name].Supported() && !listed[name] {
			problems = append(problems, fmt.Sprintf("%q cannot run: it must become executable or go on NotYetExecutable", name))
		}
	}
	return problems
}

// checkDefs reports each definition whose card is not loaded or whose Text is
// empty.
func checkDefs(all map[string]*engine.Card, defs map[string]*ability.Def) []string {
	var problems []string
	names := make([]string, 0, len(defs))
	for name := range defs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := all[name]; !ok {
			problems = append(problems, fmt.Sprintf("def %q matches no loaded card", name))
		} else if len(defs[name].Text) == 0 {
			problems = append(problems, fmt.Sprintf("def %q has empty Text", name))
		}
	}
	return problems
}
