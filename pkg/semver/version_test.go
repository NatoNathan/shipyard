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
			v1:   Version{1, 2, 3},
			v2:   Version{1, 2, 3},
			want: 0,
		},
		{
			name: "v1 > v2 by major",
			v1:   Version{2, 0, 0},
			v2:   Version{1, 9, 9},
			want: 1,
		},
		{
			name: "v1 < v2 by major",
			v1:   Version{1, 9, 9},
			v2:   Version{2, 0, 0},
			want: -1,
		},
		{
			name: "v1 > v2 by minor",
			v1:   Version{1, 2, 0},
			v2:   Version{1, 1, 9},
			want: 1,
		},
		{
			name: "v1 < v2 by minor",
			v1:   Version{1, 1, 9},
			v2:   Version{1, 2, 0},
			want: -1,
		},
		{
			name: "v1 > v2 by patch",
			v1:   Version{1, 2, 4},
			v2:   Version{1, 2, 3},
			want: 1,
		},
		{
			name: "v1 < v2 by patch",
			v1:   Version{1, 2, 3},
			v2:   Version{1, 2, 4},
			want: -1,
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
			version:    Version{1, 2, 3},
			changeType: "patch",
			want:       Version{1, 2, 4},
		},
		{
			name:       "bump minor resets patch",
			version:    Version{1, 2, 3},
			changeType: "minor",
			want:       Version{1, 3, 0},
		},
		{
			name:       "bump major resets minor and patch",
			version:    Version{1, 2, 3},
			changeType: "major",
			want:       Version{2, 0, 0},
		},
		{
			name:       "bump from zero",
			version:    Version{0, 0, 0},
			changeType: "patch",
			want:       Version{0, 0, 1},
		},
		{
			name:       "invalid change type",
			version:    Version{1, 2, 3},
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
