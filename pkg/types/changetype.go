package types

import "fmt"

// ChangeType represents the type of semantic version change
type ChangeType string

const (
	// ChangeTypePatch represents a patch-level change (0.0.X)
	ChangeTypePatch ChangeType = "patch"
	
	// ChangeTypeMinor represents a minor-level change (0.X.0)
	ChangeTypeMinor ChangeType = "minor"
	
	// ChangeTypeMajor represents a major-level change (X.0.0)
	ChangeTypeMajor ChangeType = "major"
)

// String returns the string representation of the change type
func (ct ChangeType) String() string {
	return string(ct)
}

// Validate checks if the change type is valid
func (ct ChangeType) Validate() error {
	switch ct {
	case ChangeTypePatch, ChangeTypeMinor, ChangeTypeMajor:
		return nil
	default:
		return fmt.Errorf("invalid change type: %s (must be patch, minor, or major)", ct)
	}
}

// Priority returns the numeric priority of the change type
// Higher values indicate more significant changes
// patch=1, minor=2, major=3
func (ct ChangeType) Priority() int {
	switch ct {
	case ChangeTypePatch:
		return 1
	case ChangeTypeMinor:
		return 2
	case ChangeTypeMajor:
		return 3
	default:
		return 0
	}
}

// ParseChangeType parses a string into a ChangeType
func ParseChangeType(s string) (ChangeType, error) {
	ct := ChangeType(s)
	if err := ct.Validate(); err != nil {
		return "", err
	}
	return ct, nil
}
