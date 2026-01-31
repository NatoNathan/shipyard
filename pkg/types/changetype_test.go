package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		want       string
	}{
		{
			name:       "patch",
			changeType: ChangeTypePatch,
			want:       "patch",
		},
		{
			name:       "minor",
			changeType: ChangeTypeMinor,
			want:       "minor",
		},
		{
			name:       "major",
			changeType: ChangeTypeMajor,
			want:       "major",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.changeType.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseChangeType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ChangeType
		wantErr bool
	}{
		{
			name:  "valid patch",
			input: "patch",
			want:  ChangeTypePatch,
		},
		{
			name:  "valid minor",
			input: "minor",
			want:  ChangeTypeMinor,
		},
		{
			name:  "valid major",
			input: "major",
			want:  ChangeTypeMajor,
		},
		{
			name:    "invalid",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "uppercase",
			input:   "PATCH",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChangeType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChangeType_Validate(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		wantErr    bool
	}{
		{
			name:       "valid patch",
			changeType: ChangeTypePatch,
			wantErr:    false,
		},
		{
			name:       "valid minor",
			changeType: ChangeTypeMinor,
			wantErr:    false,
		},
		{
			name:       "valid major",
			changeType: ChangeTypeMajor,
			wantErr:    false,
		},
		{
			name:       "invalid empty",
			changeType: ChangeType(""),
			wantErr:    true,
		},
		{
			name:       "invalid custom",
			changeType: ChangeType("custom"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.changeType.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChangeType_Priority(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		want       int
	}{
		{
			name:       "patch has lowest priority",
			changeType: ChangeTypePatch,
			want:       1,
		},
		{
			name:       "minor has medium priority",
			changeType: ChangeTypeMinor,
			want:       2,
		},
		{
			name:       "major has highest priority",
			changeType: ChangeTypeMajor,
			want:       3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.changeType.Priority()
			assert.Equal(t, tt.want, got)
		})
	}
}
