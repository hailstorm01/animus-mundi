// Build information: this binary's VCS revision, recorded in match records.

package engine

import "runtime/debug"

// BuildInfo is this binary's VCS revision and dirty flag from
// runtime/debug.ReadBuildInfo, "" and false when absent.
func BuildInfo() (revision string, modified bool) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}
	return revision, modified
}
