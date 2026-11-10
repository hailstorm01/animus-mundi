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
		},
		Cornered: ability.CorneredKeywords{Aggressive: 3},
	},
	"Ambush Tactics": {
		Text: []string{
			"Choose two:",
			"- Kill chosen unit.",
			"- Return chosen unit to its owner's hand.",
			"- Chosen unit deals damage equal to its power to another chosen unit.",
			"- Put chosen unit on top of its owner's deck.",
		},
		Effect: ability.Effect{
			ChooseN: 2,
			Modes: []ability.Effect{
				{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{Kinds: ability.KindUnit},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpKill,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
				{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{Kinds: ability.KindUnit},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpReturnToHand,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
				{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{Kinds: ability.KindUnit},
						},
						{
							Filter:  ability.Filter{Kinds: ability.KindUnit},
							Another: 1,
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{From: ability.ValueSourcePower},
							Subject: ability.SubjectTarget,
							Target:  1,
							Dealer:  1,
						},
					},
				},
				{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{Kinds: ability.KindUnit},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpPutOnDeck,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
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
	"Boneyard Amalgam": {
		Text:          []string{"This unit costs R1 less to deploy for each unit card in your discard pile."},
		CostReduction: ability.CostReduction{Req: 1, Per: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit, Zone: ability.ZoneDiscard}},
	},
	"Boneyard Scavenger": {
		Text: []string{"T1, exhaust, kill this unit: Return chosen maneuver card from your discard pile to your hand."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 1},
					{Kind: ability.CostExhaustSelf},
					{Kind: ability.CostKillSelf},
				},
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Zone:       ability.ZoneDiscard,
								Kinds:      ability.KindManeuver,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpReturnToHand,
							Subject: ability.SubjectTarget,
							Target:  0,
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
	"Chainbreaker Phalanx": {
		Text: []string{"Whenever this unit moves to the Main Theater, you may move chosen allied unit from your Reserves to the Main Theater. That unit gains 2 armor."},
		Triggers: []ability.Trigger{
			{
				Hook:     ability.HookMoves,
				ToZone:   ability.ZoneMainTheater,
				Optional: true,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindUnit,
								Zone:       ability.ZoneReserves,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpMoveZone,
							To:      ability.ZoneMainTheater,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
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
	"Chainbreaker's Tenacity": {
		Text: []string{"Chosen allied unit gets +1 power until end of turn. It gains 2 armor."},
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
					Op:      ability.OpModifyStat,
					Amount:  ability.Value{N: 1},
					Subject: ability.SubjectTarget,
					Target:  0,
				},
				{
					Op:      ability.OpAddArmor,
					Amount:  ability.Value{N: 2},
					Subject: ability.SubjectTarget,
					Target:  0,
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
	"Devilhorn Sanction": {
		Text: []string{"Kill chosen opposing unit."},
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
					Subject: ability.SubjectTarget,
					Target:  0,
				},
			},
		},
	},
	"Fortify": {
		Text: []string{"Chosen allied unit gains 3 armor."},
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
					Amount:  ability.Value{N: 3},
					Subject: ability.SubjectTarget,
					Target:  0,
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
	"Improvised Orders": {
		Text: []string{"Draw 3 cards, then discard a card."},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:     ability.OpDraw,
					Amount: ability.Value{N: 3},
				},
				{
					Op:     ability.OpDiscard,
					Amount: ability.Value{N: 1},
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
	"Macabre Dancers": {
		Text: []string{
			"Unbowed: Heal 2.",
			"Cornered: Resistant 1.",
		},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDamaged,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpHeal,
							Subject: ability.SubjectSelf,
							Amount:  ability.Value{N: 2},
						},
					},
				},
			},
		},
		Cornered: ability.CorneredKeywords{Resistant: 1},
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
	"Pincer Tactics": {
		Text: []string{"Deal 1 damage to each of up to 2 chosen allied units. Deal 2 damage to chosen non-Scratch asset."},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
					},
					UpTo: 2,
				},
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
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 2},
					Subject: ability.SubjectTarget,
					Target:  1,
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
	"Ruin Partisans": {
		Text: []string{"T2, exhaust, kill this unit: Return chosen unit card from your discard pile to your hand."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 2},
					{Kind: ability.CostExhaustSelf},
					{Kind: ability.CostKillSelf},
				},
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Zone:       ability.ZoneDiscard,
								Kinds:      ability.KindUnit,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpReturnToHand,
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
	"Steel Rain": {
		Text: []string{
			"As an additional cost to enact this maneuver, discard up to 3 cards.",
			"For each card discarded this way, deal 2 damage to chosen asset.",
			"Then, if you have no cards in hand, draw 2 cards.",
		},
		ExtraCosts: []ability.Cost{
			{Kind: ability.CostDiscard, UpTo: 3},
		},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					PerDiscard: true,
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
					Op:          ability.OpDraw,
					Amount:      ability.Value{N: 2},
					IfHandEmpty: true,
				},
			},
		},
	},
	"Tactical Withdrawal": {
		Text: []string{"Return up to one chosen allied unit to its owner's hand. Draw a card."},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
					},
					UpTo: 1,
				},
			},
			Steps: []ability.Step{
				{
					Op:      ability.OpReturnToHand,
					Subject: ability.SubjectTarget,
					Target:  0,
				},
				{
					Op:     ability.OpDraw,
					Amount: ability.Value{N: 1},
				},
			},
		},
	},
}
