package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range soullessDefs {
		Defs[name] = def
	}
}

// soullessDefs holds ability definitions for Soulless cards.
var soullessDefs = map[string]*ability.Def{
	"Bring It Down": {
		Text: []string{"Deal 2 damage to chosen non-Scratch asset."},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Scratch: ability.ScratchExcluded,
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
	"Colossus Platform": {
		Text: []string{"Whenever this Headquarters attacks, chosen allied unit gains 1 armor."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookAttacks,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpAddArmor,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Conscript Rifles": {
		Text: []string{"When this unit deploys, deal 1 damage to chosen non-Scratch Company unit."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Kinds:   ability.KindUnit,
								Scratch: ability.ScratchExcluded,
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
	"Culverin Brigade": {
		Text: []string{"When this unit deploys, deal 1 damage to each of up to 2 chosen opposing assets."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
							},
							UpTo: 2,
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
}
