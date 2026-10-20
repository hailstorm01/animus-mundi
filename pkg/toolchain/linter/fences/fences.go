package fences

import "errors"

// Exemptions maps card names that FRD 01 also defines as game objects to their FR ID.
var Exemptions = map[string]string{
	"Scratch Company": "FR-ZON-005",
}

// Report holds fence violations found in engine code.
type Report struct {
	Violations []string
}

// Analyze scans engine code for card names and invalid markers.
func Analyze(repoRoot string) (*Report, error) {
	return nil, errors.New("not implemented")
}

// Lint runs the fences check and prints the results.
func Lint(repoRoot string, _ []string) (bool, error) {
	return false, errors.New("not implemented")
}
