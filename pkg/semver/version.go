package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version with major, minor, and patch numbers,
// an optional pre-release identifier (e.g., "alpha.1", "beta.2", "snapshot.20260204-153045"),
// and optional build metadata (e.g., "build.123").
type Version struct {
	Major         int
	Minor         int
	Patch         int
	PreRelease    string
	BuildMetadata string
}

// Parse parses a semantic version string into a Version struct
// Accepts formats: "1.2.3", "v1.2.3", "1.2.3-alpha.1", "v1.2.3-beta.2"
func Parse(s string) (Version, error) {
	// Remove optional 'v' prefix
	s = strings.TrimPrefix(s, "v")

	if s == "" {
		return Version{}, fmt.Errorf("empty version string")
	}

	// Split off build metadata first (everything after first "+")
	var buildMetadata string
	if idx := strings.Index(s, "+"); idx != -1 {
		buildMetadata = s[idx+1:]
		s = s[:idx]
	}

	// Split off pre-release suffix (everything after first "-")
	var preRelease string
	if idx := strings.Index(s, "-"); idx != -1 {
		preRelease = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s (expected major.minor.patch)", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	if len(parts[0]) > 1 && parts[0][0] == '0' {
		return Version{}, fmt.Errorf("leading zeros not allowed in major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	if len(parts[1]) > 1 && parts[1][0] == '0' {
		return Version{}, fmt.Errorf("leading zeros not allowed in minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}
	if len(parts[2]) > 1 && parts[2][0] == '0' {
		return Version{}, fmt.Errorf("leading zeros not allowed in patch version: %s", parts[2])
	}

	if major < 0 || minor < 0 || patch < 0 {
		return Version{}, fmt.Errorf("version numbers must be non-negative")
	}

	// Validate pre-release: numeric identifiers must not have leading zeros
	if preRelease != "" {
		for _, seg := range strings.Split(preRelease, ".") {
			if _, err := strconv.Atoi(seg); err == nil {
				// It's a purely numeric segment — check for leading zeros
				if len(seg) > 1 && seg[0] == '0' {
					return Version{}, fmt.Errorf("leading zeros not allowed in numeric pre-release identifier: %s", seg)
				}
			}
		}
	}

	return Version{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		PreRelease:    preRelease,
		BuildMetadata: buildMetadata,
	}, nil
}

// String returns the string representation of the version (e.g., "1.2.3", "1.2.3-alpha.1", "1.2.3+build.123")
func (v Version) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		base += "-" + v.PreRelease
	}
	if v.BuildMetadata != "" {
		base += "+" + v.BuildMetadata
	}
	return base
}

// Compare compares this version with another version
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
// Per semver spec: pre-release versions have lower precedence than the associated normal version.
// When both have pre-release identifiers, they are compared lexicographically.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	// Same major.minor.patch — compare pre-release
	// No pre-release > has pre-release (release is higher than pre-release)
	if v.PreRelease == "" && other.PreRelease == "" {
		return 0
	}
	if v.PreRelease == "" && other.PreRelease != "" {
		return 1 // v is release, other is pre-release
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return -1 // v is pre-release, other is release
	}

	// Both have pre-release: compare per SemVer 2.0.0 section 11
	return comparePreRelease(v.PreRelease, other.PreRelease)
}

// comparePreRelease compares two pre-release strings per SemVer 2.0.0 section 11.
// Identifiers are split on "." and compared segment by segment:
//   - Numeric identifiers are compared as integers
//   - Alphanumeric identifiers are compared lexicographically
//   - Numeric identifiers always have lower precedence than alphanumeric
//   - A shorter set of identifiers has lower precedence when all preceding are equal
func comparePreRelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		cmp := comparePreReleaseSegment(aParts[i], bParts[i])
		if cmp != 0 {
			return cmp
		}
	}

	// All compared segments are equal; fewer segments = lower precedence
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}
	return 0
}

// comparePreReleaseSegment compares two individual pre-release segments.
func comparePreReleaseSegment(a, b string) int {
	aNum, aIsNum := isNumeric(a)
	bNum, bIsNum := isNumeric(b)

	switch {
	case aIsNum && bIsNum:
		// Both numeric: compare as integers
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	case aIsNum && !bIsNum:
		// Numeric has lower precedence than alphanumeric
		return -1
	case !aIsNum && bIsNum:
		return 1
	default:
		// Both alphanumeric: compare lexicographically
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
}

// isNumeric checks if a string is a purely numeric identifier and returns its value.
func isNumeric(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// MustParse parses a version string and panics if it fails
// Use only in tests or when the version string is known to be valid
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("MustParse: %v", err))
	}
	return v
}

// Bump returns a new version with the specified change type applied.
// The returned version always has an empty PreRelease (clean bump).
// changeType should be one of: "patch", "minor", "major"
func (v Version) Bump(changeType string) (Version, error) {
	// Bump is always based on the base version (without pre-release)
	base := v.BaseVersion()
	switch changeType {
	case "patch":
		return Version{
			Major: base.Major,
			Minor: base.Minor,
			Patch: base.Patch + 1,
		}, nil
	case "minor":
		return Version{
			Major: base.Major,
			Minor: base.Minor + 1,
			Patch: 0,
		}, nil
	case "major":
		return Version{
			Major: base.Major + 1,
			Minor: 0,
			Patch: 0,
		}, nil
	default:
		return Version{}, fmt.Errorf("invalid change type: %s (must be patch, minor, or major)", changeType)
	}
}

// BaseVersion returns a copy of this version without the PreRelease identifier
func (v Version) BaseVersion() Version {
	return Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

// WithPreRelease returns a copy of this version with the given pre-release identifier
func (v Version) WithPreRelease(preRelease string) Version {
	return Version{
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch,
		PreRelease: preRelease,
	}
}

// WithBuildMetadata returns a copy of this version with the given build metadata
func (v Version) WithBuildMetadata(metadata string) Version {
	return Version{
		Major:         v.Major,
		Minor:         v.Minor,
		Patch:         v.Patch,
		PreRelease:    v.PreRelease,
		BuildMetadata: metadata,
	}
}

// IsPreRelease returns true if this version has a pre-release identifier
func (v Version) IsPreRelease() bool {
	return v.PreRelease != ""
}
