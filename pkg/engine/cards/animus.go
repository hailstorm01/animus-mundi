package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range animusDefs {
		Defs[name] = def
	}
}

// animusDefs holds ability definitions for Animus cards.
var animusDefs = map[string]*ability.Def{
	"Despoiler Larva": {
		Text: []string{"Whenever this unit attacks, Scratch 1."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookAttacks,
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
	"Despoiler Worm": {
		Text: []string{"When this unit is killed, all opponents draw 2 cards."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpDraw,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectOpponents,
						},
					},
				},
			},
		},
	},
	"Fungal Evangelist": {
		Text: []string{"When this unit deploys or is killed, Scratch 2."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpScratch,
							Amount: ability.Value{N: 2},
						},
					},
				},
			},
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpScratch,
							Amount: ability.Value{N: 2},
						},
					},
				},
			},
		},
	},
	"Portent Scrivener": {
		Text: []string{"When this unit is killed on your turn, draw a card."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Turn: ability.TurnYours,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpDraw,
							Amount: ability.Value{N: 1},
						},
					},
				},
			},
		},
	},
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
	"Titanic Spore Carrier": {
		Text: []string{"When this unit is killed, Scratch 1 and draw a card."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpScratch,
							Amount: ability.Value{N: 1},
						},
						{
							Op:     ability.OpDraw,
							Amount: ability.Value{N: 1},
						},
					},
				},
			},
		},
	},
}
