package metadata

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetadata_StringPattern(t *testing.T) {
	min := 3
	max := 20

	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:      "email",
					Type:      "string",
					Required:  true,
					Pattern:   `^[^@]+@[^@]+\.[^@]+$`,
					MinLength: &min,
					MaxLength: &max,
				},
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid email", "test@example.com", false, ""},
		{"invalid pattern", "not-an-email", true, "does not match pattern"},
		{"too short (also fails pattern)", "ab", true, "does not match pattern"},
		{"too long", "verylongemailaddressx@example.com", true, "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]string{"email": tt.value}
			err := ValidateMetadata(cfg, metadata)
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

func TestValidateMetadata_IntegerRange(t *testing.T) {
	min := 1
	max := 10

	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:     "points",
					Type:     "int",
					Required: true,
					Min:      &min,
					Max:      &max,
				},
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid integer", "5", false, ""},
		{"at minimum", "1", false, ""},
		{"at maximum", "10", false, ""},
		{"below minimum", "0", true, "below minimum"},
		{"above maximum", "11", true, "above maximum"},
		{"invalid", "abc", true, "invalid integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]string{"points": tt.value}
			err := ValidateMetadata(cfg, metadata)
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

func TestValidateMetadata_ListItems(t *testing.T) {
	minItems := 1
	maxItems := 3

	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:     "tags",
					Type:     "list",
					Required: true,
					ItemType: "string",
					MinItems: &minItems,
					MaxItems: &maxItems,
				},
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid list", "tag1,tag2", false, ""},
		{"single item", "tag1", false, ""},
		{"three items", "tag1,tag2,tag3", false, ""},
		{"empty required field", "", true, "cannot be empty"},
		{"too many items", "tag1,tag2,tag3,tag4", true, "too many items"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]string{"tags": tt.value}
			err := ValidateMetadata(cfg, metadata)
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

func TestValidateMetadata_AllowedValues(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:          "priority",
					Type:          "string",
					Required:      true,
					AllowedValues: []string{"low", "medium", "high"},
				},
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid value", "high", false},
		{"invalid value", "urgent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]string{"priority": tt.value}
			err := ValidateMetadata(cfg, metadata)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not in allowed values")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMetadata_MapType(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:     "config",
					Type:     "map",
					Required: false,
				},
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid map", `{"key":"value"}`, false},
		{"empty map", `{}`, false},
		{"invalid JSON", `not json`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := map[string]string{"config": tt.value}
			err := ValidateMetadata(cfg, metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMetadata_DefaultValues(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:     "status",
					Type:     "string",
					Required: true,
					Default:  "pending",
				},
			},
		},
	}

	// Empty metadata should use default
	metadata := map[string]string{}
	err := ValidateMetadata(cfg, metadata)
	assert.NoError(t, err)
}

func TestValidateMetadata_MultipleFields(t *testing.T) {
	min := 1
	max := 10

	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{
					Name:     "author",
					Type:     "string",
					Required: true,
					Pattern:  `^[^@]+@[^@]+\.[^@]+$`,
				},
				{
					Name:     "priority",
					Type:     "string",
					Required: false,
					AllowedValues: []string{"low", "medium", "high"},
					Default:  "medium",
				},
				{
					Name:     "points",
					Type:     "int",
					Required: false,
					Min:      &min,
					Max:      &max,
				},
			},
		},
	}

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  bool
	}{
		{
			name: "all valid",
			metadata: map[string]string{
				"author":   "dev@example.com",
				"priority": "high",
				"points":   "5",
			},
			wantErr: false,
		},
		{
			name: "missing optional fields",
			metadata: map[string]string{
				"author": "dev@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			metadata: map[string]string{
				"author": "not-an-email",
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			metadata: map[string]string{
				"author":   "dev@example.com",
				"priority": "urgent",
			},
			wantErr: true,
		},
		{
			name: "invalid points",
			metadata: map[string]string{
				"author": "dev@example.com",
				"points": "20",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetadata(cfg, tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
