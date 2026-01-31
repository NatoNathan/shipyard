package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version with major, minor, and patch numbers
type Version struct {
	Major int
	Minor int
	Patch int
}

// Parse parses a semantic version string into a Version struct
// Accepts formats: "1.2.3" or "v1.2.3"
func Parse(s string) (Version, error) {
	// Remove optional 'v' prefix
	s = strings.TrimPrefix(s, "v")
	
	if s == "" {
		return Version{}, fmt.Errorf("empty version string")
	}
	
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s (expected major.minor.patch)", s)
	}
	
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}
	
	if major < 0 || minor < 0 || patch < 0 {
		return Version{}, fmt.Errorf("version numbers must be non-negative")
	}
	
	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String returns the string representation of the version (e.g., "1.2.3")
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares this version with another version
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
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
	
	return 0
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

// Bump returns a new version with the specified change type applied
// changeType should be one of: "patch", "minor", "major"
func (v Version) Bump(changeType string) (Version, error) {
	switch changeType {
	case "patch":
		return Version{
			Major: v.Major,
			Minor: v.Minor,
			Patch: v.Patch + 1,
		}, nil
	case "minor":
		return Version{
			Major: v.Major,
			Minor: v.Minor + 1,
			Patch: 0,
		}, nil
	case "major":
		return Version{
			Major: v.Major + 1,
			Minor: 0,
			Patch: 0,
		}, nil
	default:
		return Version{}, fmt.Errorf("invalid change type: %s (must be patch, minor, or major)", changeType)
	}
}
