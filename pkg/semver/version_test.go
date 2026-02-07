package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "valid standard version",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "valid v-prefixed version",
			input: "v1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "pre-release alpha",
			input: "1.2.3-alpha.1",
			want:  Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
		},
		{
			name:  "pre-release beta with v prefix",
			input: "v2.0.0-beta.3",
			want:  Version{Major: 2, Minor: 0, Patch: 0, PreRelease: "beta.3"},
		},
		{
			name:  "pre-release rc",
			input: "1.0.0-rc.1",
			want:  Version{Major: 1, Minor: 0, Patch: 0, PreRelease: "rc.1"},
		},
		{
			name:  "pre-release snapshot with timestamp",
			input: "1.2.0-snapshot.20260204-153045",
			want:  Version{Major: 1, Minor: 2, Patch: 0, PreRelease: "snapshot.20260204-153045"},
		},
		{
			name:    "invalid format",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			input:   "a.b.c",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "standard version",
			version: Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "zero version",
			version: Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
		{
			name:    "large numbers",
			version: Version{Major: 100, Minor: 200, Patch: 300},
			want:    "100.200.300",
		},
		{
			name:    "with pre-release",
			version: Version{Major: 1, Minor: 2, Patch: 0, PreRelease: "alpha.1"},
			want:    "1.2.0-alpha.1",
		},
		{
			name:    "with snapshot pre-release",
			version: Version{Major: 1, Minor: 2, Patch: 0, PreRelease: "snapshot.20260204-153045"},
			want:    "1.2.0-snapshot.20260204-153045",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want int
	}{
		{
			name: "equal versions",
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "v1 > v2 by major",
			v1:   Version{Major: 2, Minor: 0, Patch: 0},
			v2:   Version{Major: 1, Minor: 9, Patch: 9},
			want: 1,
		},
		{
			name: "v1 < v2 by major",
			v1:   Version{Major: 1, Minor: 9, Patch: 9},
			v2:   Version{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "v1 > v2 by minor",
			v1:   Version{Major: 1, Minor: 2, Patch: 0},
			v2:   Version{Major: 1, Minor: 1, Patch: 9},
			want: 1,
		},
		{
			name: "v1 < v2 by minor",
			v1:   Version{Major: 1, Minor: 1, Patch: 9},
			v2:   Version{Major: 1, Minor: 2, Patch: 0},
			want: -1,
		},
		{
			name: "v1 > v2 by patch",
			v1:   Version{Major: 1, Minor: 2, Patch: 4},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			name: "v1 < v2 by patch",
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 4},
			want: -1,
		},
		{
			name: "release > pre-release",
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			want: 1,
		},
		{
			name: "pre-release < release",
			v1:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: -1,
		},
		{
			name: "equal pre-release",
			v1:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.1"},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.1"},
			want: 0,
		},
		{
			name: "alpha < beta lexicographically",
			v1:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.1"},
			want: -1,
		},
		{
			name: "beta > alpha lexicographically",
			v1:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.1"},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Compare(tt.v2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBump(t *testing.T) {
	tests := []struct {
		name       string
		version    Version
		changeType string
		want       Version
		wantErr    bool
	}{
		{
			name:       "bump patch",
			version:    Version{Major: 1, Minor: 2, Patch: 3},
			changeType: "patch",
			want:       Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:       "bump minor resets patch",
			version:    Version{Major: 1, Minor: 2, Patch: 3},
			changeType: "minor",
			want:       Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name:       "bump major resets minor and patch",
			version:    Version{Major: 1, Minor: 2, Patch: 3},
			changeType: "major",
			want:       Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:       "bump from zero",
			version:    Version{Major: 0, Minor: 0, Patch: 0},
			changeType: "patch",
			want:       Version{Major: 0, Minor: 0, Patch: 1},
		},
		{
			name:       "bump from pre-release strips pre-release",
			version:    Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"},
			changeType: "patch",
			want:       Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:       "bump minor from pre-release strips pre-release",
			version:    Version{Major: 1, Minor: 2, Patch: 0, PreRelease: "beta.3"},
			changeType: "minor",
			want:       Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name:       "invalid change type",
			version:    Version{Major: 1, Minor: 2, Patch: 3},
			changeType: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.version.Bump(tt.changeType)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBaseVersion(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"}
	base := v.BaseVersion()
	assert.Equal(t, Version{Major: 1, Minor: 2, Patch: 3}, base)
	assert.Empty(t, base.PreRelease)
}

func TestWithPreRelease(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	withPR := v.WithPreRelease("beta.2")
	assert.Equal(t, Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta.2"}, withPR)
	// Original unchanged
	assert.Empty(t, v.PreRelease)
}

func TestIsPreRelease(t *testing.T) {
	assert.False(t, Version{Major: 1, Minor: 2, Patch: 3}.IsPreRelease())
	assert.True(t, Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"}.IsPreRelease())
}

func TestMustParsePreRelease(t *testing.T) {
	v := MustParse("1.2.3-rc.1")
	assert.Equal(t, 1, v.Major)
	assert.Equal(t, 2, v.Minor)
	assert.Equal(t, 3, v.Patch)
	assert.Equal(t, "rc.1", v.PreRelease)
}

// --- Task 1.1: Spec-compliant pre-release comparison tests ---

func TestComparePreReleaseSpecCompliant(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "numeric vs alphanumeric: alpha.1 < alpha.beta",
			v1:   "1.0.0-alpha.1",
			v2:   "1.0.0-alpha.beta",
			want: -1,
		},
		{
			name: "numeric comparison not string: beta.2 < beta.11",
			v1:   "1.0.0-beta.2",
			v2:   "1.0.0-beta.11",
			want: -1,
		},
		{
			name: "fewer segments = lower precedence: alpha < alpha.1",
			v1:   "1.0.0-alpha",
			v2:   "1.0.0-alpha.1",
			want: -1,
		},
		{
			name: "full spec example chain: alpha < alpha.1",
			v1:   "1.0.0-alpha",
			v2:   "1.0.0-alpha.1",
			want: -1,
		},
		{
			name: "full spec example chain: alpha.1 < alpha.beta",
			v1:   "1.0.0-alpha.1",
			v2:   "1.0.0-alpha.beta",
			want: -1,
		},
		{
			name: "full spec example chain: alpha.beta < beta",
			v1:   "1.0.0-alpha.beta",
			v2:   "1.0.0-beta",
			want: -1,
		},
		{
			name: "full spec example chain: beta < beta.2",
			v1:   "1.0.0-beta",
			v2:   "1.0.0-beta.2",
			want: -1,
		},
		{
			name: "full spec example chain: beta.2 < beta.11",
			v1:   "1.0.0-beta.2",
			v2:   "1.0.0-beta.11",
			want: -1,
		},
		{
			name: "full spec example chain: beta.11 < rc.1",
			v1:   "1.0.0-beta.11",
			v2:   "1.0.0-rc.1",
			want: -1,
		},
		{
			name: "full spec example chain: rc.1 < release",
			v1:   "1.0.0-rc.1",
			v2:   "1.0.0",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := MustParse(tt.v1)
			v2 := MustParse(tt.v2)
			got := v1.Compare(v2)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Task 1.2: Build metadata tests ---

func TestParseBuildMetadata(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantPR        string
		wantBM        string
		wantErr       bool
	}{
		{
			name:   "version with build metadata only",
			input:  "1.2.3+build.123",
			wantPR: "",
			wantBM: "build.123",
		},
		{
			name:   "version with pre-release and build metadata",
			input:  "1.2.3-alpha+build",
			wantPR: "alpha",
			wantBM: "build",
		},
		{
			name:   "version with complex build metadata",
			input:  "1.2.3-beta.1+20130313144700",
			wantPR: "beta.1",
			wantBM: "20130313144700",
		},
		{
			name:   "version without metadata",
			input:  "1.2.3",
			wantPR: "",
			wantBM: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPR, v.PreRelease)
			assert.Equal(t, tt.wantBM, v.BuildMetadata)
		})
	}
}

func TestStringBuildMetadata(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "with build metadata only",
			version: Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "build.123"},
			want:    "1.2.3+build.123",
		},
		{
			name:    "with pre-release and build metadata",
			version: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha", BuildMetadata: "build"},
			want:    "1.2.3-alpha+build",
		},
		{
			name:    "without build metadata",
			version: Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.version.String())
		})
	}
}

func TestCompareIgnoresBuildMetadata(t *testing.T) {
	v1 := Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "build1"}
	v2 := Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "build2"}
	assert.Equal(t, 0, v1.Compare(v2))

	v3 := Version{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, 0, v1.Compare(v3))
}

func TestWithBuildMetadata(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha"}
	withBM := v.WithBuildMetadata("build.456")
	assert.Equal(t, Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha", BuildMetadata: "build.456"}, withBM)
	// Original unchanged
	assert.Empty(t, v.BuildMetadata)
}

func TestBaseVersionDropsBuildMetadata(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha", BuildMetadata: "build.1"}
	base := v.BaseVersion()
	assert.Empty(t, base.PreRelease)
	assert.Empty(t, base.BuildMetadata)
}

func TestWithPreReleaseDropsBuildMetadata(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "build.1"}
	withPR := v.WithPreRelease("beta.2")
	assert.Equal(t, "beta.2", withPR.PreRelease)
	assert.Empty(t, withPR.BuildMetadata)
}

// --- Task 1.3: Leading zeros rejection tests ---

func TestParseRejectsLeadingZeros(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "leading zero in major",
			input:   "01.0.0",
			wantErr: true,
		},
		{
			name:    "leading zero in minor",
			input:   "0.01.0",
			wantErr: true,
		},
		{
			name:    "leading zero in patch",
			input:   "0.0.01",
			wantErr: true,
		},
		{
			name:    "single zero is fine",
			input:   "0.0.0",
			wantErr: false,
		},
		{
			name:    "leading zero in numeric pre-release identifier",
			input:   "1.2.3-alpha.01",
			wantErr: true,
		},
		{
			name:    "valid numeric pre-release identifier",
			input:   "1.2.3-alpha.1",
			wantErr: false,
		},
		{
			name:    "leading zero in alphanumeric pre-release is fine",
			input:   "1.2.3-01alpha",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
