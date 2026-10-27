package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range animusDefs {
		Defs[name] = def
	}
}

// animusDefs holds ability definitions for Animus cards.
var animusDefs = map[string]*ability.Def{
	"Animoceps Shambler": {
		Text: []string{"Whenever this unit exhausts, chosen allied asset in the Main Theater heals 1."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookExhausted,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Zone:       ability.ZoneMainTheater,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpHeal,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Craterfall Ancient": {
		Text: []string{
			"Whenever this unit attacks, you may deal damage equal to its power to each of up to 2 other chosen assets.",
			"When this unit is killed, deal 4 damage to all assets.",
		},
		Triggers: []ability.Trigger{
			{
				Hook:     ability.HookAttacks,
				Optional: true,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{Other: true},
							UpTo:   2,
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{From: ability.ValueSourcePower},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 4},
							Subject: ability.SubjectAll,
							Filter:  ability.Filter{},
						},
					},
				},
			},
		},
	},
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
	"Infested Apostle": {
		Text: []string{"When this unit deploys, you may kill an allied unit. If you do, deal 2 damage to chosen opposing unit."},
		Triggers: []ability.Trigger{
			{
				Hook:     ability.HookDeploy,
				Optional: true,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
								Kinds:      ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpKill,
							Subject: ability.SubjectPick,
							UpTo:    1,
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindUnit,
							},
						},
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
							IfDone:  true,
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
	"Spoil Forest Colossus": {
		Text: []string{"Whenever this unit deploys or attacks, chosen asset heals 2."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpHeal,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
			{
				Hook: ability.HookAttacks,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpHeal,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
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
	"Stormfront of Ideals": {
		Text: []string{"When this unit is killed, put it on the bottom of its owner's deck."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpPutOnDeck,
							Subject: ability.SubjectSelf,
							Bottom:  true,
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
	"Unceasing Tide": {
		Text: []string{"When this unit is killed, return it to its owner's hand."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpReturnToHand,
							Subject: ability.SubjectSelf,
						},
					},
				},
			},
		},
	},
}
