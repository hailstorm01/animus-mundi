package cards

import "animus-mundi/pkg/engine/ability"

func init() {
	for name, def := range concordiaDefs {
		Defs[name] = def
	}
}

// concordiaDefs holds ability definitions for Concordia cards.
var concordiaDefs = map[string]*ability.Def{
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
}
