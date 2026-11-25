package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range concordiaDefs {
		Defs[name] = def
	}
}

// concordiaDefs holds ability definitions for Concordia cards.
var concordiaDefs = map[string]*ability.Def{
	"Growing Legion": {
		Text: []string{"Scratch 1."},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:     ability.OpScratch,
					Amount: ability.Value{N: 1},
				},
			},
		},
	},
}
