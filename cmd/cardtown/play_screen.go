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
			fmt.Fprintf(out, "  [%d] %s (R%d T%d, %d/%d)%s\n", i+1, c.Card, c.ReqCost, c.TacCost, c.Power, c.Health, notExec)
		}
	} else {
		fmt.Fprintln(out, "Your Hand: (empty)")
	}
	fmt.Fprintf(out, "Your Deck: %d cards\n", v.You.DeckCount)

	// Stack
	if len(v.Stack) > 0 {
		fmt.Fprintln(out, "Resolution Stack:")
		for _, item := range v.Stack {
			fmt.Fprintf(out, "  %s (Player %d, Source %d)\n", item.Kind, item.Controller, item.Source)
		}
	} else {
		fmt.Fprintln(out, "Resolution Stack: (empty)")
	}

	// Discards
	fmt.Fprintf(out, "Discards: You %d | Opponent %d\n\n", len(v.You.Discard), len(v.Opponent.Discard))
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
	return fmt.Sprintf("%s#%d (%d/%d, %s%s)", a.Card, a.ID, a.Power, a.MaxHealth, status, statStr)
}

func actionLabel(v view.View, a engine.Action) string {
	findAsset := func(id int) string {
		if v.You.HQ != nil && v.You.HQ.ID == id {
			return fmt.Sprintf("%s#%d", v.You.HQ.Card, id)
		}
		for _, as := range v.You.Reserves {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		if v.Opponent.HQ != nil && v.Opponent.HQ.ID == id {
			return fmt.Sprintf("%s#%d", v.Opponent.HQ.Card, id)
		}
		for _, as := range v.Opponent.Reserves {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		for _, as := range v.Theater {
			if as.ID == id {
				return fmt.Sprintf("%s#%d", as.Card, id)
			}
		}
		return fmt.Sprintf("Asset#%d", id)
	}

	findHandCard := func(instID int) string {
		for _, c := range v.You.Hand {
			if c.Inst == instID {
				return c.Card
			}
		}
		return fmt.Sprintf("Card#%d", instID)
	}

	switch a.Kind {
	case engine.ActGenerate:
		res := "Requisition"
		if a.Resource == engine.Tactics {
			res = "Tactics"
		}
		return fmt.Sprintf("Generate: %s -> %s", findAsset(a.Asset), res)
	case engine.ActConvert:
		return "Convert: 2 Requisition -> 1 Tactics"
	case engine.ActDeploy:
		return fmt.Sprintf("Deploy: %s to %s", findHandCard(a.Card), a.Zone)
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
