package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range pathDefs {
		Defs[name] = def
	}
}

// pathDefs holds ability definitions for Path cards.
var pathDefs = map[string]*ability.Def{
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
