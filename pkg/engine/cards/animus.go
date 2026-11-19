package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range animusDefs {
		Defs[name] = def
	}
}

// animusDefs holds ability definitions for Animus cards.
var animusDefs = map[string]*ability.Def{
	"Animoceps Node": {
		Text: []string{"Allied units in the same zone have +1 power."},
		Statics: []ability.Static{
			{
				Kind: ability.StaticAura,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					ZoneRef:    ability.ZoneRefSame,
				},
				Power: 1,
			},
		},
	},
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
	"Animoceps Spore Cloud": {
		Text: []string{"Redirect all damage that would be dealt to allied units in the same zone to this structure."},
		Replacements: []ability.Replacement{
			{
				Event: ability.WouldBeDealtDamage,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					ZoneRef:    ability.ZoneRefSame,
				},
			},
		},
	},
	"Animoceps Spores": {
		Text: []string{
			"Scratch 2, then return an allied Scratch Company unit to its owner's hand.",
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:     ability.OpScratch,
					Amount: ability.Value{N: 2},
				},
				{
					Op:      ability.OpReturnToHand,
					Subject: ability.SubjectPick,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Scratch:    ability.ScratchOnly,
					},
				},
			},
		},
	},
	"Animoceps Stalk": {
		Text: []string{"Other allied assets in the same zone have +1 power."},
		Statics: []ability.Static{
			{
				Kind: ability.StaticAura,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Other:      true,
					ZoneRef:    ability.ZoneRefSame,
				},
				Power: 1,
			},
		},
	},
	"Cliff of Portents": {
		Text: []string{
			"R1, T1, exhaust: Search your deck for a card, put it into your hand, then shuffle your deck.",
		},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Req: 1, Tac: 1},
					{Kind: ability.CostExhaustSelf},
				},
				Effect: ability.Effect{
					Steps: []ability.Step{
						{Op: ability.OpSearchDeck},
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
	"Conversion Spike": {
		Text: []string{
			"Deal 1 damage to chosen opposing unit, and chosen allied unit or structure heals 1.",
		},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceOpposing,
						Kinds:      ability.KindUnit,
					},
				},
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit | ability.KindStructure,
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
					Op:      ability.OpHeal,
					Amount:  ability.Value{N: 1},
					Subject: ability.SubjectTarget,
					Target:  1,
				},
			},
		},
	},
	"Deafening Aria": {
		Text: []string{
			"Deal 1 damage to all opposing units in the Main Theater.",
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 1},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceOpposing,
						Kinds:      ability.KindUnit,
						Zone:       ability.ZoneMainTheater,
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
	"Despoiler's Emergence": {
		Text: []string{
			"Deal 4 damage to all opposing assets in the Main Theater.",
			"Deal 2 damage to all opposing non-Scratch assets in opponents' Reserves.",
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 4},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceOpposing,
						Zone:       ability.ZoneMainTheater,
					},
				},
				{
					Op:      ability.OpDealDamage,
					Amount:  ability.Value{N: 2},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceOpposing,
						Scratch:    ability.ScratchExcluded,
						Zone:       ability.ZoneReserves,
					},
				},
			},
		},
	},
	"Disorienting Pulse": {
		Text: []string{"Return chosen opposing unit in the Main Theater to its owner's hand."},
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
			},
		},
	},
	"Drifting Concept": {
		Text: []string{
			"When this unit deploys, chosen opponent reveals their hand. You choose a card from it. That player discards that card.",
		},
		Triggers: []ability.Trigger{
			{
				Hook: ability.HookDeploy,
				Effect: ability.Effect{
					Steps: []ability.Step{
						{
							Op:                ability.OpDiscard,
							Amount:            ability.Value{N: 1},
							Subject:           ability.SubjectOpponents,
							ControllerChooses: true,
						},
					},
				},
			},
		},
	},
	"Ensnaring Mist": {
		Text: []string{
			"Move chosen unit to its controller's Reserves.",
			"Chosen unit gets -1 power until end of turn.",
		},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Kinds: ability.KindUnit,
					},
				},
				{
					Filter: ability.Filter{
						Kinds: ability.KindUnit,
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
				{
					Op:      ability.OpModifyStat,
					Amount:  ability.Value{N: -1},
					Subject: ability.SubjectTarget,
					Target:  1,
				},
			},
		},
	},
	"Feed the Despoiler": {
		Text: []string{
			"As an additional cost to enact this maneuver, kill an allied unit.",
			"Draw 2 cards.",
		},
		ExtraCosts: []ability.Cost{
			{
				Kind:   ability.CostKill,
				Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit},
			},
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:     ability.OpDraw,
					Amount: ability.Value{N: 2},
				},
			},
		},
	},
	"Forced Mutation": {
		Text: []string{"Turn chosen opposing unit face down. It becomes a Scratch Company unit."},
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
					Op:      ability.OpMorphScratch,
					Subject: ability.SubjectTarget,
					Target:  0,
				},
			},
		},
	},
	"Funeral Fungus": {
		Text: []string{
			"As an additional cost to enact this maneuver, exhaust a readied allied unit.",
			"Return chosen unit card from your discard pile to your hand.",
		},
		ExtraCosts: []ability.Cost{
			{
				Kind:   ability.CostExhaust,
				Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit},
			},
		},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Zone:       ability.ZoneDiscard,
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
	"Hulking Incongruity": {
		Text: []string{"Allied Horror units in the same zone have Resistant 1."},
		Statics: []ability.Static{
			{
				Kind: ability.StaticAura,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					Subtypes:   []string{"Horror"},
					ZoneRef:    ability.ZoneRefSame,
				},
				Keywords: ability.Keywords{Resistant: 1},
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
	"Melodic Pummeler": {
		Text: []string{"Allied Horror units in the same zone have +1 power."},
		Statics: []ability.Static{
			{
				Kind: ability.StaticAura,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					Subtypes:   []string{"Horror"},
					ZoneRef:    ability.ZoneRefSame,
				},
				Power: 1,
			},
		},
	},
	"Mutation Swell": {
		Text: []string{"As an additional cost to enact this maneuver, kill 2 or 4 allied Scratch Company units. For every 2 units killed this way, deploy a different chosen unit card from your discard pile for free."},
		ExtraCosts: []ability.Cost{
			{
				Kind: ability.CostKill,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					Scratch:    ability.ScratchOnly,
				},
				Counts: []int{2, 4},
			},
		},
		Effect: ability.Effect{
			Targets: []ability.Target{
				{
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Zone:       ability.ZoneDiscard,
					},
					PerKilled: 2,
					Different: true,
				},
			},
			Steps: []ability.Step{
				{
					Op:      ability.OpDeployFree,
					Subject: ability.SubjectTarget,
					Target:  0,
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
	"Rapturous Sermon": {
		Text: []string{
			"Move all allied Scratch Company units to the Main Theater. Allied Scratch Company units get +2 power until end of turn.",
		},
		Effect: ability.Effect{
			Steps: []ability.Step{
				{
					Op:      ability.OpMoveZone,
					To:      ability.ZoneMainTheater,
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Scratch:    ability.ScratchOnly,
					},
				},
				{
					Op:      ability.OpModifyStat,
					Amount:  ability.Value{N: 2},
					Subject: ability.SubjectAll,
					Filter: ability.Filter{
						Allegiance: ability.AllegianceAllied,
						Kinds:      ability.KindUnit,
						Scratch:    ability.ScratchOnly,
					},
				},
			},
		},
	},
	"Rhapsody's Embrace": {
		Text: []string{
			"As an additional cost to enact this maneuver, kill an allied unit.",
			"Kill chosen opposing unit.",
		},
		ExtraCosts: []ability.Cost{
			{
				Kind:   ability.CostKill,
				Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit},
			},
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
					Op:      ability.OpKill,
					Subject: ability.SubjectTarget,
					Target:  0,
				},
			},
		},
	},
	"Rhapsody's Hollow": {
		Text: []string{
			"T3, exhaust, kill this structure: Gain control of chosen opposing unit in the Main Theater.",
		},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 3},
					{Kind: ability.CostExhaustSelf},
					{Kind: ability.CostKillSelf},
				},
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
							Op:      ability.OpGainControl,
							Subject: ability.SubjectTarget,
							Target:  0,
						},
					},
				},
			},
		},
	},
	"Shifting Impossibility": {
		Text: []string{"If an allied Scratch Company unit would be killed, return it to its owner's hand and deal 1 damage to this unit instead."},
		Replacements: []ability.Replacement{
			{
				Event: ability.WouldBeDestroyed,
				Filter: ability.Filter{
					Allegiance: ability.AllegianceAllied,
					Kinds:      ability.KindUnit,
					Scratch:    ability.ScratchOnly,
				},
				Steps: []ability.Step{
					{Op: ability.OpReturnToHand, Subject: ability.SubjectReplaced},
					{Op: ability.OpDealDamage, Amount: ability.Value{N: 1}, Subject: ability.SubjectSelf},
				},
			},
		},
	},
	"Sinuous Growth": {
		Text: []string{"T1, kill this unit: Return chosen Fungus or Plant card from your discard pile to your hand."},
		Activated: []ability.Activated{
			{
				Costs: []ability.Cost{
					{Kind: ability.CostResources, Tac: 1},
					{Kind: ability.CostKillSelf},
				},
				Effect: ability.Effect{
					Targets: []ability.Target{
						{
							Filter: ability.Filter{
								Allegiance: ability.AllegianceAllied,
								Zone:       ability.ZoneDiscard,
								Subtypes:   []string{"Fungus", "Plant"},
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
	"Veil Render": {
		Text: []string{"You may deploy this card from your discard pile by killing an allied unit in addition to paying its deployment costs."},
		Statics: []ability.Static{
			{
				Kind:       ability.StaticDeployPermission,
				ActiveZone: ability.ZoneDiscard,
				Costs: []ability.Cost{
					{
						Kind:   ability.CostKill,
						Filter: ability.Filter{Allegiance: ability.AllegianceAllied, Kinds: ability.KindUnit},
					},
				},
			},
		},
	},
}
