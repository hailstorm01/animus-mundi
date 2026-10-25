package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range headquartersDefs {
		Defs[name] = def
	}
}

// headquartersDefs holds ability definitions for Headquarters cards.
var headquartersDefs = map[string]*ability.Def{
	"Animoceps Rager": {
		Text: []string{"Whenever this Headquarters is dealt damage, Scratch 1. This ability triggers only once each turn."},
		Triggers: []ability.Trigger{
			{
				Hook:       ability.HookDamaged,
				MaxPerTurn: 1,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpScratch,
							Amount: ability.Value{N: 1},
						},
					},
				},
			},
		},
	},
}
