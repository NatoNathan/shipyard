// Package semver provides semantic versioning functionality for Shipyard.
// It implements parsing, comparison, and manipulation of semantic versions
// according to the Semantic Versioning specification (https://semver.org/).
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version with major, minor, and patch components
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string in the format "major.minor.patch"
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares this version with another version.
// Returns:
//
//	-1 if this version is less than the other
//	 0 if this version equals the other
//	 1 if this version is greater than the other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	return 0
}

// Equals returns true if this version equals the other version
func (v *Version) Equals(other *Version) bool {
	return v.Compare(other) == 0
}

// LessThan returns true if this version is less than the other version
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// GreaterThan returns true if this version is greater than the other version
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// Copy returns a copy of this version
func (v *Version) Copy() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

// BumpMajor increments the major version and resets minor and patch to 0
func (v *Version) BumpMajor() *Version {
	return &Version{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
	}
}

// BumpMinor increments the minor version and resets patch to 0
func (v *Version) BumpMinor() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
	}
}

// BumpPatch increments the patch version
func (v *Version) BumpPatch() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

// Parse parses a version string into a Version struct
func Parse(versionStr string) (*Version, error) {
	versionStr = strings.TrimSpace(versionStr)
	if versionStr == "" || versionStr == "latest" {
		return &Version{Major: 0, Minor: 0, Patch: 0}, nil
	}

	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	parts := strings.Split(versionStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected major.minor.patch)", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

// MustParse parses a version string and panics if it's invalid
func MustParse(versionStr string) *Version {
	version, err := Parse(versionStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse version %s: %v", versionStr, err))
	}
	return version
}

// New creates a new Version with the given major, minor, and patch values
func New(major, minor, patch int) *Version {
	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

// Zero returns a zero version (0.0.0)
func Zero() *Version {
	return &Version{Major: 0, Minor: 0, Patch: 0}
}
