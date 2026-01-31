package metadata

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMetadataValue_String(t *testing.T) {
	field := config.MetadataField{Type: "string"}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "hello"},
		{"with spaces", "hello world", "hello world"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMetadataValue(field, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMetadataValue_Integer(t *testing.T) {
	field := config.MetadataField{Type: "int"}

	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"positive integer", "42", 42, false},
		{"negative integer", "-10", -10, false},
		{"zero", "0", 0, false},
		{"with spaces", "  123  ", 123, false},
		{"invalid", "not a number", 0, true},
		{"float", "3.14", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMetadataValue(field, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseMetadataValue_List(t *testing.T) {
	tests := []struct {
		name     string
		field    config.MetadataField
		input    string
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "comma-separated strings",
			field:    config.MetadataField{Type: "list"},
			input:    "a,b,c",
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "comma-separated with spaces",
			field:    config.MetadataField{Type: "list"},
			input:    "a, b, c",
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "JSON array of strings",
			field:    config.MetadataField{Type: "list"},
			input:    `["a","b","c"]`,
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "empty list",
			field:    config.MetadataField{Type: "list"},
			input:    "",
			expected: []interface{}{},
		},
		{
			name:     "comma-separated integers",
			field:    config.MetadataField{Type: "list", ItemType: "int"},
			input:    "1,2,3",
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "JSON array of integers",
			field:    config.MetadataField{Type: "list", ItemType: "int"},
			input:    `[1,2,3]`,
			expected: []interface{}{1, 2, 3},
		},
		{
			name:    "invalid JSON array",
			field:   config.MetadataField{Type: "list"},
			input:   `["a","b"`,
			wantErr: true,
		},
		{
			name:    "invalid integer in list",
			field:   config.MetadataField{Type: "list", ItemType: "int"},
			input:   "1,abc,3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMetadataValue(tt.field, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseMetadataValue_Map(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple map",
			input:    `{"key":"value"}`,
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "map with multiple keys",
			input:    `{"name":"test","count":3}`,
			expected: map[string]interface{}{"name": "test", "count": float64(3)},
		},
		{
			name:    "non-JSON input",
			input:   "not json",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   `{"key":"value"`,
			wantErr: true,
		},
	}

	field := config.MetadataField{Type: "map"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMetadataValue(field, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseMetadataValue_DefaultString(t *testing.T) {
	// When type is empty, default to string
	field := config.MetadataField{}
	result, err := ParseMetadataValue(field, "test")
	require.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestParseMetadataValue_UnsupportedType(t *testing.T) {
	field := config.MetadataField{Type: "unknown"}
	_, err := ParseMetadataValue(field, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported type")
}
