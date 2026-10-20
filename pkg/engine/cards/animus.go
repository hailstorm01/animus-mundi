package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range animusDefs {
		Defs[name] = def
	}
}

// animusDefs holds ability definitions for Animus cards.
var animusDefs = map[string]*ability.Def{
	"Spoil Forest Creeper": {
		Text: []string{"When this unit deploys, Scratch 1."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
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
