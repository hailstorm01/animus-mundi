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
	"Dolus Macrocannon": {
		Text: []string{"R1, exhaust: Deal 2 damage to chosen non-Scratch asset."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Req: 1},
					{Kind: ability.CostExhaustSelf},
				},
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
		},
	},
	"Dolus Refiner": {
		Text: []string{"Exhaust: Draw 2 cards, then discard 2 cards."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostExhaustSelf},
				},
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:     ability.OpDraw,
							Amount: ability.Value{N: 2},
						},
						{
							Op:     ability.OpDiscard,
							Amount: ability.Value{N: 2},
						},
					},
				},
			},
		},
	},
	"Earthshaker Colossus": {
		Text: []string{"Whenever this unit deploys or attacks, deal 2 damage to chosen non-Scratch asset."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
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
			{
				Hook: ability.HookAttacks,
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
		},
	},
	"Fire for Effect": {
		Text: []string{
			"Opposing assets in the Main Theater gain Resistant -1 until end of turn.",
			"Deal 1 damage to chosen asset in the Main Theater.",
		},
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
					Op:      ability.OpGrantKeyword,
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceOpposing,
						Zone:       ability.ZoneMainTheater,
					},
					Keywords: ability.Keywords{Resistant: -1},
				},
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 1},
					Subject: ability.SubjectTarget,
					Target:  0,
				},
			},
		},
	},
	"Hellfire Armorers": {
		Text: []string{"Allied units in the same zone have Aggressive 1."},
		Statics: []ability.Static{
			{
				Kind: ability.StaticAura,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					ZoneRef:    ability.ZoneRefSame,
				},
				Keywords: ability.Keywords{Aggressive: 1},
			},
		},
	},
}
