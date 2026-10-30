package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range pathDefs {
		Defs[name] = def
	}
}

func cornered(t ability.Trigger) ability.Trigger { t.Cornered = true; return t }

// pathDefs holds ability definitions for Path cards.
var pathDefs = map[string]*ability.Def{
	"Allshan, Chainbreaker Revenant": {
		Text: []string{
			"Unbowed: This unit gains 2 armor.",
			"Cornered: Aggressive 3.",
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
			cornered(ability.Aggressive(3)),
		},
	},
	"Arkha, Sun's Fury": {
		Text: []string{"Whenever this unit attacks, you may put a unit card from your discard pile on the bottom of your deck. If you do, deal 3 damage to chosen opposing Headquarters."},
		Triggers: []ability.Trigger{
			{
				Hook:     ability.HookAttacks,
				Optional: true,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
								Kinds:      ability.KindHeadquarters,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpPutOnDeck,
							Bottom:  true,
							Subject: ability.SubjectPick,
							UpTo:    1,
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindUnit,
								Zone:       ability.ZoneDiscard,
							},
						},
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 3},
							Subject: ability.SubjectTarget,
							Target:  0,
							IfDone:  true,
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
	"Rogue Fleshshaper": {
		Text: []string{
			"When this unit deploys, choose one:",
			"- Chosen asset heals 1.",
			"- Deal 1 damage to chosen non-Scratch asset.",
		},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					ChooseN: 1,
					Modes: []ability.Effect{
						{
							Targets: []ability.Target{
								{
									Filter: ability.Filter{},
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
						{
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
									Amount:  ability.Value{N: 1},
									Subject: ability.SubjectTarget,
									Target:  0,
								},
							},
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
								Scratch:    ability.ScratchExcluded,
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
