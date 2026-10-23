package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeArgs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"-ALL", []string{"-ALL"}, []string{"--all"}},
		{"--Repo-Root=x", []string{"--Repo-Root=x"}, []string{"--repo-root=x"}},
		{"-reporoot, C:/x", []string{"-reporoot", "C:/x"}, []string{"--repo-root", "C:/x"}},
		{"-DryRun", []string{"-DryRun"}, []string{"--dry-run"}},
		{"-updateExisting", []string{"-updateExisting"}, []string{"--update-existing"}},
		{"-SHOWBODIES", []string{"-SHOWBODIES"}, []string{"--show-bodies"}},
		{"-tddFile=a.md", []string{"-tddFile=a.md"}, []string{"--tdd-file=a.md"}},
		{"-commitMessage=Hi There", []string{"-commitMessage=Hi There"}, []string{"--msg=Hi There"}},
		{"-commit-message, m", []string{"-commit-message", "m"}, []string{"--msg", "m"}},
		{"-MSG", []string{"-MSG"}, []string{"--msg"}},
		{"-maxTurns=5", []string{"-maxTurns=5"}, []string{"--max-turns=5"}},
		{"-max-turns", []string{"-max-turns"}, []string{"--max-turns"}},
		{"-Seed=3, -P0, ANIMUS", []string{"-Seed=3", "-P0", "ANIMUS"}, []string{"--seed=3", "--p0", "ANIMUS"}},
		{"-h", []string{"-h"}, []string{"--h"}},
		{"-unknownFlag=V", []string{"-unknownFlag=V"}, []string{"--unknownflag=V"}},
		{"---x", []string{"---x"}, []string{"--x"}},
		{"-a=b=c", []string{"-a=b=c"}, []string{"--a=b=c"}},
		{"file.md, -", []string{"file.md", "-"}, []string{"file.md", "--"}},
		{"no arguments (nil)", nil, nil},
		{"no arguments (empty)", []string{}, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := normalizeArgs(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("normalizeArgs(%#v) = %#v, want %#v", c.in, got, c.want)
			}
		})
	}

	// Loop over every name the old switch listed.
	oldSwitchExpected := map[string]string{
		"all":             "all",
		"frd":             "frd",
		"tdd":             "tdd",
		"cards":           "cards",
		"sheets":          "sheets",
		"syntax":          "syntax",
		"fences":          "fences",
		"staged":          "staged",
		"files":           "files",
		"reporoot":        "repo-root",
		"repo-root":       "repo-root",
		"dryrun":          "dry-run",
		"dry-run":         "dry-run",
		"publish":         "publish",
		"updateexisting":  "update-existing",
		"update-existing": "update-existing",
		"showbodies":      "show-bodies",
		"show-bodies":     "show-bodies",
		"tddfile":         "tdd-file",
		"tdd-file":        "tdd-file",
		"source":          "source",
		"msg":             "msg",
		"commitmessage":   "msg",
		"commit-message":  "msg",
		"push":            "push",
		"section":         "section",
		"rules":           "rules",
		"deck":            "deck",
		"games":           "games",
		"seed":            "seed",
		"p0":              "p0",
		"p1":              "p1",
		"seat":            "seat",
		"policy":          "policy",
		"maxturns":        "max-turns",
		"max-turns":       "max-turns",
	}

	oldNames := strings.Fields("all frd tdd cards sheets syntax fences staged files reporoot repo-root dryrun dry-run publish updateexisting update-existing showbodies show-bodies tddfile tdd-file source msg commitmessage commit-message push section rules deck games seed p0 p1 seat policy maxturns max-turns")

	for _, name := range oldNames {
		expectedOutput, ok := oldSwitchExpected[name]
		if !ok {
			t.Fatalf("unexpected name in test setup: %s", name)
		}
		t.Run("old_switch_"+name, func(t *testing.T) {
			inputArg := "-" + strings.ToUpper(name) + "=v"
			want := []string{"--" + expectedOutput + "=v"}
			got := normalizeArgs([]string{inputArg})
			if !reflect.DeepEqual(got, want) {
				t.Errorf("normalizeArgs(%#v) = %#v, want %#v", []string{inputArg}, got, want)
			}
		})
	}
}
