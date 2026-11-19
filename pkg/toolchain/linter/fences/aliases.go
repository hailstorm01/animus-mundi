package fences

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

// aliasRoots are the repo-relative directories whose non-test Go files may declare no
// type alias: a move or rename updates every reference instead (ADR 0007).
var aliasRoots = []string{"pkg", "cmd", "tools"}

// checkAliases reports every type alias declaration (type X = Y), at any depth, in the
// non-test Go files under aliasRoots. A file that doesn't parse is skipped: the import
// check and the compiler report it.
func checkAliases(repoRoot string) ([]violationItem, error) {
	var items []violationItem
	for _, root := range aliasRoots {
		files, err := goFilesUnder(repoRoot, root)
		if err != nil {
			return nil, err
		}
		for _, rel := range files {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, filepath.Join(repoRoot, filepath.FromSlash(rel)), nil, parser.SkipObjectResolution)
			if err != nil {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok && ts.Assign.IsValid() {
					items = append(items, violationItem{
						file: rel,
						line: fset.Position(ts.Pos()).Line,
						msg:  fmt.Sprintf("declares type alias %s; a move or rename updates every reference instead (ADR 0007)", ts.Name.Name),
					})
				}
				return true
			})
		}
	}
	return items, nil
}
