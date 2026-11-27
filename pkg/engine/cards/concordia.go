package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	register(concordiaDefs)
}

// concordiaDefs holds ability definitions for Concordia cards.
var concordiaDefs = map[string]*ability.Def{
	"Echoing Nightmare": {
		Text: []string{"When this unit deploys, return chosen opposing unit in the Main Theater to its owner's hand. If a unit is returned this way, move this unit to the Main Theater."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
								Kinds:      ability.KindUnit,
								Zone:       ability.ZoneMainTheater,
							},
						},
					},
					Steps: []ability.Step{
						{
							Op:      ability.OpReturnToHand,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
						{
							Op:      ability.OpMoveZone,
							To:      ability.ZoneMainTheater,
							Subject: ability.SubjectSelf,
							IfDone:  true,
						},
					},
				},
			},
		},
	},
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
	"Hurricane Outrigger": {
		Text: []string{"When this unit deploys, your Headquarters gets +1 power until the end of your turn."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:      ability.OpModifyStat,
							Amount:  ability.Value{N: 1},
							Subject: ability.SubjectAll,
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Kinds:      ability.KindHeadquarters,
							},
						},
					},
				},
			},
		},
	},
	"Icewind Regent": {
		Text: []string{"Exhaust: Return chosen unit to its owner's hand."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostExhaustSelf},
				},
				Effect: ability.Effect{
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
			},
		},
	},
	"Leyline Regrowth": {
		Text: []string{"Return up to one chosen unit card from your discard pile to your hand. Draw a card."},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Zone:       ability.ZoneDiscard,
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
	"Seismic Instability": {
		Text: []string{
			"Deal 3 damage to all assets in the Main Theater.",
			"Deal 4 damage to all structures in the Main Theater.",
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 3},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Zone: ability.ZoneMainTheater,
					},
				},
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 4},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Kinds: ability.KindStructure,
						Zone:  ability.ZoneMainTheater,
					},
				},
			},
		},
	},
	"Soul Whip": {
		Text: []string{"Chosen allied unit deals damage equal to its power to chosen opposing unit."},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
					},
				},
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
					Amount:  ability.Value{From: ability.ValueSourcePower},
					Subject: ability.SubjectTarget,
					Target:  1,
					Dealer:  1,
				},
			},
		},
	},
	"Stormbolt Ray": {
		Text: []string{
			"When this structure deploys, deal 3 damage to chosen asset.",
			"T2, exhaust: Return this structure to its owner's hand.",
		},
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
							Op:      ability.OpDealDamage,
							Amount:  ability.Value{N: 3},
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 2},
					{Kind: ability.CostExhaustSelf},
				},
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
	"Tempest City Ray": {
		Text: []string{"T2, exhaust: Move chosen opposing unit to its controller's Reserves."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 2},
					{Kind: ability.CostExhaustSelf},
				},
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
							Op:      ability.OpMoveZone,
							To:      ability.ZoneReserves,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Swirling Dustdevil": {
		Text:          []string{"This unit costs R1 less to deploy for each maneuver card in your discard pile."},
		CostReduction: ability.CostReduction{Req: 1, Per: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindManeuver, Zone: ability.ZoneDiscard}},
	},
	"Will Breaker": {
		Text: []string{"When this unit deploys, kill chosen opposing unit with Requisition cost 3 or less."},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceOpposing,
								Kinds:      ability.KindUnit,
								Stat:       ability.StatRequisition,
								AtMost:     3,
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
		},
	},
}
