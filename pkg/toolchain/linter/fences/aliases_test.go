package fences_test

import (
	"strings"
	"testing"
)

// Tests for T-0274 (directive #165, ADR 0007): cardtown lint fences rejects every type
// alias declaration in non-test Go files under pkg/, cmd/ and tools/. A move or rename
// updates every reference, so no old name survives as an alias.

const aliasRule = "a move or rename updates every reference instead (ADR 0007)"

// 41. An alias in a non-test file fails, naming its file, line and name.
func TestFences_Case41_TypeAliasFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/engine/alias.go", "package engine\n\nimport \"animus-mundi/pkg/engine/ability\"\n\ntype Keywords = ability.Keywords\n")
	out := expectFail(t, repo)
	want := "[ERROR] pkg/engine/alias.go:5: declares type alias Keywords; " + aliasRule
	if strings.Count(out, "[ERROR]") != 1 || !strings.Contains(out, want) {
		t.Fatalf("expected one error %q, got:\n%s", want, out)
	}
}

// 42. Each alias in a grouped type declaration fails on its own line; a defined type
// in the same group passes.
func TestFences_Case42_GroupedAliasesFail(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/group.go", "package sim\n\ntype (\n\tA = int\n\tB string\n\tC = []A\n)\n")
	out := expectFail(t, repo)
	if strings.Count(out, "[ERROR]") != 2 {
		t.Fatalf("expected two errors, got:\n%s", out)
	}
	for _, want := range []string{
		"[ERROR] pkg/sim/group.go:4: declares type alias A; ",
		"[ERROR] pkg/sim/group.go:6: declares type alias C; ",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q, got:\n%s", want, out)
		}
	}
}

// 43. An alias declared inside a function body fails too.
func TestFences_Case43_AliasInFunctionBodyFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/view/local.go", "package view\n\nfunc f() {\n\ttype n = int\n\tvar _ n\n}\n")
	out := expectFail(t, repo)
	if !strings.Contains(out, "[ERROR] pkg/view/local.go:4: declares type alias n; ") {
		t.Fatalf("expected local.go:4, got:\n%s", out)
	}
}

// 44. A generic alias fails.
func TestFences_Case44_GenericAliasFails(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/generic.go", "package sim\n\ntype List[T any] = []T\n")
	out := expectFail(t, repo)
	if !strings.Contains(out, "[ERROR] pkg/sim/generic.go:3: declares type alias List; ") {
		t.Fatalf("expected generic.go:3, got:\n%s", out)
	}
}

// 45. The alias check covers every non-test Go file under pkg/, cmd/ and tools/,
// including pkg/engine/cards and pkg/toolchain, which the card-name scan skips.
func TestFences_Case45_AliasScope(t *testing.T) {
	for _, rel := range []string{
		"pkg/engine/cards/a.go",
		"pkg/toolchain/x/a.go",
		"pkg/foo/bar/a.go",
		"cmd/cardtown/a.go",
		"tools/importer/a.go",
	} {
		repo := setupFixtureRepo(t)
		writePkgFile(t, repo, rel, "package a\n\ntype X = int\n")
		out := expectFail(t, repo)
		if !strings.Contains(out, "[ERROR] "+rel+":3: declares type alias X; ") {
			t.Fatalf("%s: expected an alias error, got:\n%s", rel, out)
		}
	}
}

// 46. Test files, files outside pkg/, cmd/ and tools/, and defined types pass, and the
// PASS line names the alias check.
func TestFences_Case46_AliasExemptions(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/engine/alias_test.go", "package engine\n\ntype X = int\n")
	writePkgFile(t, repo, "cmd/cardtown/a_test.go", "package main\n\ntype X = int\n")
	writePkgFile(t, repo, "scripts/a.go", "package scripts\n\ntype X = int\n")
	writePkgFile(t, repo, "pkg/engine/defined.go", "package engine\n\ntype Y int\n\ntype (\n\tZ struct{}\n\tW Y\n)\n")
	out := expectPass(t, repo)
	if !strings.Contains(out, "disallowed imports and type aliases.") {
		t.Fatalf("expected the PASS line to name type aliases, got:\n%s", out)
	}
}

// 47. A file that doesn't parse gets no alias error, only the import check's; an alias
// in another file is still reported.
func TestFences_Case47_UnparseableFileHasNoAliasError(t *testing.T) {
	repo := setupFixtureRepo(t)
	writePkgFile(t, repo, "pkg/sim/bad.go", "package sim\n\nimport (\n\t\"fmt\n)\n\ntype X = int\n")
	writePkgFile(t, repo, "pkg/sim/good.go", "package sim\n\ntype Y = int\n")
	out := expectFail(t, repo)
	if strings.Count(out, "[ERROR]") != 2 ||
		!strings.Contains(out, "[ERROR] pkg/sim/bad.go:4: cannot read imports") ||
		!strings.Contains(out, "[ERROR] pkg/sim/good.go:3: declares type alias Y; ") {
		t.Fatalf("expected bad.go's import error and good.go's alias error only, got:\n%s", out)
	}
}
