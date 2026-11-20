// Command: play, the screen: what the human sees, drawn only from view.PlayerView.

package main

import (
	"fmt"
	"io"
	"strings"

	"animus-mundi/pkg/engine"
	"animus-mundi/pkg/view"
)

func printScreen(out io.Writer, v view.View, newLogs []string) {
	if len(newLogs) > 0 {
		for _, line := range newLogs {
			fmt.Fprintln(out, line)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "=== Turn %d (%s) ===\n", v.Turn, v.Phase)
	fmt.Fprintf(out, "Pools: You (R%d T%d) | Opponent (R%d T%d)\n", v.You.Req, v.You.Tac, v.Opponent.Req, v.Opponent.Tac)

	// Opponent
	if v.Opponent.HQ != nil {
		fmt.Fprintf(out, "Opponent HQ: %s\n", formatAsset(*v.Opponent.HQ))
	} else {
		fmt.Fprintln(out, "Opponent HQ: (destroyed)")
	}
	if len(v.Opponent.Reserves) > 0 {
		fmt.Fprintln(out, "Opponent Reserves:")
		for _, a := range v.Opponent.Reserves {
			fmt.Fprintf(out, "  %s\n", formatAsset(a))
		}
	} else {
		fmt.Fprintln(out, "Opponent Reserves: (none)")
	}
	fmt.Fprintf(out, "Opponent Hand: %d cards | Deck: %d cards\n", v.Opponent.HandCount, v.Opponent.DeckCount)

	// Main Theater
	if len(v.Theater) > 0 {
		fmt.Fprintln(out, "Main Theater:")
		for _, a := range v.Theater {
			fmt.Fprintf(out, "  %s (Player %d)\n", formatAsset(a), a.Controller)
		}
	} else {
		fmt.Fprintln(out, "Main Theater: (empty)")
	}

	// You
	if v.You.HQ != nil {
		fmt.Fprintf(out, "Your HQ: %s\n", formatAsset(*v.You.HQ))
	} else {
		fmt.Fprintln(out, "Your HQ: (destroyed)")
	}
	if len(v.You.Reserves) > 0 {
		fmt.Fprintln(out, "Your Reserves:")
		for _, a := range v.You.Reserves {
			fmt.Fprintf(out, "  %s\n", formatAsset(a))
		}
	} else {
		fmt.Fprintln(out, "Your Reserves: (none)")
	}
	if len(v.You.Hand) > 0 {
		fmt.Fprintln(out, "Your Hand:")
		for i, c := range v.You.Hand {
			notExec := ""
			if !c.Executable {
				notExec = " [not executable yet]"
			}
			fmt.Fprintf(out, "  [%d] %s (R%d T%d, %d/%d)%s\n", i+1, c.Card, c.ReqNow, c.TacNow, c.Power, c.Health, notExec)
		}
	} else {
		fmt.Fprintln(out, "Your Hand: (empty)")
	}
	fmt.Fprintf(out, "Your Deck: %d cards\n", v.You.DeckCount)

	// Stack
	if len(v.Stack) > 0 {
		fmt.Fprintln(out, "Resolution Stack:")
		for _, item := range v.Stack {
			fmt.Fprintf(out, "  %s\n", stackLine(v, item))
		}
	} else {
		fmt.Fprintln(out, "Resolution Stack: (empty)")
	}

	// Discards
	fmt.Fprintf(out, "Discards: You %d | Opponent %d\n\n", len(v.You.Discard), len(v.Opponent.Discard))
}

// stackLine is the stack listing's line for item: what it is (stackItemLine),
// then, after " -> ", the targets it chose that are still there
// (chosenTargetsText). Chosen targets are public (FR-TGT-001).
func stackLine(v view.View, item view.StackView) string {
	if targets := chosenTargetsText(v, item); targets != "" {
		return stackItemLine(item) + " -> " + targets
	}
	return stackItemLine(item)
}

// stackItemLine is what item is: its Label, the engine's one wording of it
// (Game.StackLabel), with a capital first letter, then the declared zone of a
// Deploy, then its controller, and the source asset of anything but a Deploy
// or a Maneuver. An item with no Label, and an ability with no card, print
// their Kind.
func stackItemLine(item view.StackView) string {
	if item.Label == "" || item.Kind == "Ability" && item.Card == "" {
		return fmt.Sprintf("%s (Player %d, Source %d)", item.Kind, item.Controller, item.Source)
	}
	label := strings.ToUpper(item.Label[:1]) + item.Label[1:]
	switch item.Kind {
	case "Deploy":
		return fmt.Sprintf("%s to %s (Player %d)", label, item.Zone, item.Controller)
	case "Maneuver":
		return fmt.Sprintf("%s (Player %d)", label, item.Controller)
	}
	return fmt.Sprintf("%s (Player %d, Source %d)", label, item.Controller, item.Source)
}

// chosenTargetsText names item's chosen targets that are still where they were
// chosen: the slots in order, separated by "; ", and each slot's picks in pick
// order, separated by ", ". An asset is named as the view shows it (assetRef;
// a face-down asset by the Scratch face), and a pile slot's cards are followed
// by " (discard pile)". A chosen asset no longer on the battlefield is left
// out, and so is a slot with nothing left. It is "" when nothing is left.
func chosenTargetsText(v view.View, item view.StackView) string {
	var slots []string
	for i := 0; i < max(len(item.Chosen), len(item.ChosenCards)); i++ {
		var picks []string
		if i < len(item.Chosen) {
			for _, id := range item.Chosen[i] {
				if a := v.Asset(id); a != nil {
					picks = append(picks, assetRef(*a))
				}
			}
		}
		if len(picks) > 0 {
			slots = append(slots, strings.Join(picks, ", "))
		} else if i < len(item.ChosenCards) && len(item.ChosenCards[i]) > 0 {
			slots = append(slots, strings.Join(item.ChosenCards[i], ", ")+" (discard pile)")
		}
	}
	return strings.Join(slots, "; ")
}

// assetRef is how the screen names a battlefield asset: its card as the view
// shows it, and its ID.
func assetRef(a view.AssetView) string {
	return fmt.Sprintf("%s#%d", a.Card, a.ID)
}

func formatAsset(a view.AssetView) string {
	status := "Ready"
	if a.Exhausted {
		status = "Exhausted"
	}
	var extra []string
	if a.Armor > 0 {
		extra = append(extra, fmt.Sprintf("Armor %d", a.Armor))
	}
	if a.Damage > 0 {
		extra = append(extra, fmt.Sprintf("Dmg %d", a.Damage))
	}
	statStr := ""
	if len(extra) > 0 {
		statStr = ", " + strings.Join(extra, ", ")
	}
	return fmt.Sprintf("%s (%d/%d, %s, Yield R%d T%d%s)", assetRef(a), a.Power, a.MaxHealth, status, a.ReqYield, a.TacYield, statStr)
}

func actionLabel(v view.View, a engine.Action) string {
	findAsset := func(id int) string {
		if as := v.Asset(id); as != nil {
			return assetRef(*as)
		}
		return fmt.Sprintf("Asset#%d", id)
	}

	findCard := func(pile []view.CardView, instID int) string {
		for _, c := range pile {
			if c.Inst == instID {
				return c.Card
			}
		}
		return fmt.Sprintf("Card#%d", instID)
	}

	switch a.Kind {
	case engine.ActGenerate:
		return fmt.Sprintf("Generate: %s -> %s", findAsset(a.Asset), a.Resource)
	case engine.ActConvert:
		return fmt.Sprintf("Convert: %d Requisition -> %d Tactics", engine.ConvertReq, engine.ConvertTac)
	case engine.ActDeploy:
		if a.From == engine.ZoneDiscard {
			return fmt.Sprintf("Deploy: %s from discard pile to %s", findCard(v.You.Discard, a.Card), a.Zone)
		}
		return fmt.Sprintf("Deploy: %s to %s", findCard(v.You.Hand, a.Card), a.Zone)
	case engine.ActEnact:
		return fmt.Sprintf("Enact: %s", findCard(v.You.Hand, a.Card))
	case engine.ActActivate:
		if as := v.Asset(a.Asset); as != nil {
			return fmt.Sprintf("Activate: %s: %s", findAsset(a.Asset), as.Text)
		}
		return fmt.Sprintf("Activate: %s", findAsset(a.Asset))
	case engine.ActMarch:
		return fmt.Sprintf("March: %s", findAsset(a.Asset))
	case engine.ActAttack:
		return fmt.Sprintf("Attack: %s -> %s", findAsset(a.Asset), findAsset(a.Target))
	case engine.ActScratch:
		return fmt.Sprintf("Deploy Scratch Company: %s", findAsset(a.Asset))
	case engine.ActEndTurn:
		return "End Turn"
	default:
		return a.Kind.String()
	}
}
