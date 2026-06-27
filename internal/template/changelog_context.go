package template

import (
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// ChangelogContext is the root value passed to changelog (multi-version) templates.
// It provides named top-level fields for the package and version summary, plus the
// full sorted entry list under Entries.
type ChangelogContext struct {
	Package          string          // package name, from the newest entry
	LatestVersion    string          // most recent version string (stable or pre-release)
	LatestStable     string          // most recent non-pre-release version; empty if none
	LatestPreRelease string          // most recent pre-release version; empty if none
	Entries          []history.Entry // all entries, sorted newest-first
}

// newChangelogContext builds a ChangelogContext from a slice already sorted newest-first.
func newChangelogContext(sorted []history.Entry) ChangelogContext {
	ctx := ChangelogContext{Entries: sorted}
	if len(sorted) == 0 {
		return ctx
	}
	ctx.Package = sorted[0].Package
	ctx.LatestVersion = sorted[0].Version

	for _, e := range sorted {
		v, err := semver.Parse(e.Version)
		if err != nil {
			continue // skip malformed versions
		}
		if v.IsPreRelease() {
			if ctx.LatestPreRelease == "" {
				ctx.LatestPreRelease = e.Version
			}
		} else {
			if ctx.LatestStable == "" {
				ctx.LatestStable = e.Version
			}
		}
		if ctx.LatestStable != "" && ctx.LatestPreRelease != "" {
			break
		}
	}
	return ctx
}
