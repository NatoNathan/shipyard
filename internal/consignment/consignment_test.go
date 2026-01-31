package consignment

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	timestamp := time.Now()
	id1, err1 := GenerateID(timestamp)
	assert.NoError(t, err1)

	id2, err2 := GenerateID(timestamp)
	assert.NoError(t, err2)

	// IDs should be unique
	assert.NotEqual(t, id1, id2)

	// Should match format: YYYYMMDD-HHMMSS-xxxxxx
	assert.Len(t, id1, 22) // 8 + 1 + 6 + 1 + 6
	assert.Contains(t, id1, "-")
}

func TestConsignment_Validate(t *testing.T) {
	tests := []struct {
		name        string
		consignment *Consignment
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid consignment",
			consignment: &Consignment{
				ID:         "20260130-143022-abc123",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fixed a bug",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			consignment: &Consignment{
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fixed a bug",
			},
			wantErr: true,
			errMsg:  "ID is required",
		},
		{
			name: "no packages",
			consignment: &Consignment{
				ID:         "20260130-143022-abc123",
				Timestamp:  time.Now(),
				Packages:   []string{},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fixed a bug",
			},
			wantErr: true,
			errMsg:  "at least one package",
		},
		{
			name: "invalid change type",
			consignment: &Consignment{
				ID:         "20260130-143022-abc123",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeType("invalid"),
				Summary:    "Fixed a bug",
			},
			wantErr: true,
		},
		{
			name: "empty summary",
			consignment: &Consignment{
				ID:         "20260130-143022-abc123",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "",
			},
			wantErr: true,
			errMsg:  "summary is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.consignment.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConsignment_AffectsPackage(t *testing.T) {
	c := &Consignment{
		Packages: []string{"core", "api"},
	}
	
	assert.True(t, c.AffectsPackage("core"))
	assert.True(t, c.AffectsPackage("api"))
	assert.False(t, c.AffectsPackage("other"))
}

func TestNew(t *testing.T) {
	packages := []string{"core"}
	changeType := types.ChangeTypePatch
	summary := "Test summary"
	metadata := map[string]interface{}{
		"author": "test@example.com",
	}

	c, err := New(packages, changeType, summary, metadata)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NotEmpty(t, c.ID)
	assert.False(t, c.Timestamp.IsZero())
	assert.Equal(t, packages, c.Packages)
	assert.Equal(t, changeType, c.ChangeType)
	assert.Equal(t, summary, c.Summary)
	assert.Equal(t, metadata, c.Metadata)

	// Should be valid
	err = c.Validate()
	assert.NoError(t, err)
}
