package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range pathDefs {
		Defs[name] = def
	}
}

// pathDefs holds ability definitions for Path cards.
var pathDefs = map[string]*ability.Def{
	"Allshan, Chainbreaker Revenant": {
		Text: []string{
			"Unbowed: This unit gains 2 armor.",
			"Cornered: Whenever this unit attacks, it gets +3 power until end of turn.",
		},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDamaged,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpAddArmor,
							Subject: ability.SubjectSelf,
							Amount:  ability.Value{N: 2},
						},
					},
				},
			},
			{
				Hook:     ability.HookAttacks,
				Cornered: true,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpModifyStat,
							Subject: ability.SubjectSelf,
							Amount:  ability.Value{N: 3},
						},
					},
				},
			},
		},
	},
	"Chainbreaker Irregulars": {
		Text: []string{"Unbowed: Draw a card, then discard a card."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDamaged,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpDraw,
							Amount: ability.Value{N: 1},
						},
						{
							Op:     ability.OpDiscard,
							Amount: ability.Value{N: 1},
						},
					},
				},
			},
		},
	},
	"Chainbreaker Sappers": {
		Text: []string{"When this unit deploys, deal 1 damage to chosen structure or Headquarters."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Kinds: ability.KindStructure | ability.KindHeadquarters,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Devilhorn Captain": {
		Text: []string{"Unbowed: Deal 1 damage to chosen asset in the Main Theater."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDamaged,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Zone: ability.ZoneMainTheater,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Devilhorn Guerrillas": {
		Text: []string{"When this unit deploys, if it was deployed to the Main Theater, deal 2 damage to chosen opposing unit."},
		Triggers: []ability.Trigger{
			{
				Hook:   ability.HookDeploy,
				ToZone: ability.ZoneMainTheater,
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
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Hammerhand Brigands": {
		Text: []string{"Unbowed: Scratch 1."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDamaged,
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
	"Ironclad Colossus": {
		Text: []string{"Whenever this unit deploys or attacks, chosen unit gains 2 armor."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Kinds: ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpAddArmor,
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
							Filter: ability.Filter{
								Kinds: ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpAddArmor,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Necropolis Hunt": {
		Text: []string{"When this unit is killed, deal damage equal to its power to chosen non-Headquarters asset."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookKilled,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Kinds: ability.KindUnit | ability.KindStructure,
							},
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
		},
	},
	"Saboteur": {
		Text: []string{"When this unit deploys, deal 2 damage to chosen opposing non-Scratch unit and 1 damage to chosen allied unit."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
								Kinds:      ability.KindUnit,
								NonScratch: true,
							},
						},
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 2},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectTarget,
							Target:  1,
						},
					},
				},
			},
		},
	},
}
