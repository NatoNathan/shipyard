package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expected    *Version
		expectError bool
	}{
		{
			name:       "standard version",
			versionStr: "1.2.3",
			expected:   &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:       "version with v prefix",
			versionStr: "v2.4.6",
			expected:   &Version{Major: 2, Minor: 4, Patch: 6},
		},
		{
			name:       "zero version",
			versionStr: "0.0.0",
			expected:   &Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:       "empty version",
			versionStr: "",
			expected:   &Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:       "latest version",
			versionStr: "latest",
			expected:   &Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:        "invalid format",
			versionStr:  "1.2",
			expectError: true,
		},
		{
			name:        "invalid major",
			versionStr:  "x.2.3",
			expectError: true,
		},
		{
			name:        "invalid minor",
			versionStr:  "1.x.3",
			expectError: true,
		},
		{
			name:        "invalid patch",
			versionStr:  "1.2.x",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := Parse(tt.versionStr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if version.Major != tt.expected.Major {
				t.Errorf("Expected major %d, got %d", tt.expected.Major, version.Major)
			}
			if version.Minor != tt.expected.Minor {
				t.Errorf("Expected minor %d, got %d", tt.expected.Minor, version.Minor)
			}
			if version.Patch != tt.expected.Patch {
				t.Errorf("Expected patch %d, got %d", tt.expected.Patch, version.Patch)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	// Test successful parsing
	version := MustParse("1.2.3")
	if version.Major != 1 || version.Minor != 2 || version.Patch != 3 {
		t.Errorf("Expected 1.2.3, got %s", version.String())
	}

	// Test panic on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic but didn't get one")
		}
	}()
	MustParse("invalid")
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{
			name:     "standard version",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "zero version",
			version:  &Version{Major: 0, Minor: 0, Patch: 0},
			expected: "0.0.0",
		},
		{
			name:     "major version only",
			version:  &Version{Major: 5, Minor: 0, Patch: 0},
			expected: "5.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Version
		v2       *Version
		expected int
	}{
		{
			name:     "equal versions",
			v1:       &Version{1, 2, 3},
			v2:       &Version{1, 2, 3},
			expected: 0,
		},
		{
			name:     "v1 major greater",
			v1:       &Version{2, 0, 0},
			v2:       &Version{1, 9, 9},
			expected: 1,
		},
		{
			name:     "v1 major less",
			v1:       &Version{1, 9, 9},
			v2:       &Version{2, 0, 0},
			expected: -1,
		},
		{
			name:     "v1 minor greater",
			v1:       &Version{1, 2, 0},
			v2:       &Version{1, 1, 9},
			expected: 1,
		},
		{
			name:     "v1 minor less",
			v1:       &Version{1, 1, 9},
			v2:       &Version{1, 2, 0},
			expected: -1,
		},
		{
			name:     "v1 patch greater",
			v1:       &Version{1, 2, 4},
			v2:       &Version{1, 2, 3},
			expected: 1,
		},
		{
			name:     "v1 patch less",
			v1:       &Version{1, 2, 3},
			v2:       &Version{1, 2, 4},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}

			// Test convenience methods
			switch tt.expected {
			case 0:
				if !tt.v1.Equals(tt.v2) {
					t.Error("Expected versions to be equal")
				}
				if tt.v1.LessThan(tt.v2) {
					t.Error("Expected v1 not to be less than v2")
				}
				if tt.v1.GreaterThan(tt.v2) {
					t.Error("Expected v1 not to be greater than v2")
				}
			case 1:
				if tt.v1.Equals(tt.v2) {
					t.Error("Expected versions not to be equal")
				}
				if tt.v1.LessThan(tt.v2) {
					t.Error("Expected v1 not to be less than v2")
				}
				if !tt.v1.GreaterThan(tt.v2) {
					t.Error("Expected v1 to be greater than v2")
				}
			case -1:
				if tt.v1.Equals(tt.v2) {
					t.Error("Expected versions not to be equal")
				}
				if !tt.v1.LessThan(tt.v2) {
					t.Error("Expected v1 to be less than v2")
				}
				if tt.v1.GreaterThan(tt.v2) {
					t.Error("Expected v1 not to be greater than v2")
				}
			}
		})
	}
}

func TestVersionBump(t *testing.T) {
	base := &Version{1, 2, 3}

	tests := []struct {
		name     string
		method   func(*Version) *Version
		expected *Version
	}{
		{
			name:     "bump major",
			method:   (*Version).BumpMajor,
			expected: &Version{2, 0, 0},
		},
		{
			name:     "bump minor",
			method:   (*Version).BumpMinor,
			expected: &Version{1, 3, 0},
		},
		{
			name:     "bump patch",
			method:   (*Version).BumpPatch,
			expected: &Version{1, 2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(base)

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %s, got %s", tt.expected.String(), result.String())
			}

			// Ensure original version is unchanged
			if !base.Equals(&Version{1, 2, 3}) {
				t.Error("Original version should not be modified")
			}
		})
	}
}

func TestVersionCopy(t *testing.T) {
	original := &Version{1, 2, 3}
	copy := original.Copy()

	if !original.Equals(copy) {
		t.Error("Copy should equal original")
	}

	// Modify copy and ensure original is unchanged
	copy.Major = 5
	if original.Major != 1 {
		t.Error("Original should not be modified when copy is changed")
	}
}

func TestNew(t *testing.T) {
	version := New(1, 2, 3)
	expected := &Version{1, 2, 3}

	if !version.Equals(expected) {
		t.Errorf("Expected %s, got %s", expected.String(), version.String())
	}
}

func TestZero(t *testing.T) {
	version := Zero()
	expected := &Version{0, 0, 0}

	if !version.Equals(expected) {
		t.Errorf("Expected %s, got %s", expected.String(), version.String())
	}
}
