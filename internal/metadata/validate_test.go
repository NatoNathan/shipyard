package metadata

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestValidateMetadata_NoConfig tests validation when no metadata config exists
func TestValidateMetadata_NoConfig(t *testing.T) {
	cfg := &config.Config{}

	metadata := map[string]string{
		"author": "test@example.com",
		"issue":  "JIRA-123",
	}

	err := ValidateMetadata(cfg, metadata)
	assert.NoError(t, err, "Should allow any metadata when no config exists")
}

// TestValidateMetadata_RequiredFields tests validation of required fields
func TestValidateMetadata_RequiredFields(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{Name: "author", Required: true, Type: "string"},
				{Name: "issue", Required: false, Type: "string"},
			},
		},
	}

	// Missing required field
	metadata := map[string]string{
		"issue": "JIRA-123",
	}

	err := ValidateMetadata(cfg, metadata)
	assert.Error(t, err, "Should fail when required field is missing")
	assert.Contains(t, err.Error(), "author", "Error should mention missing field")
}

// TestValidateMetadata_Success tests successful validation
func TestValidateMetadata_Success(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{Name: "author", Required: true, Type: "string"},
				{Name: "issue", Required: false, Type: "string"},
			},
		},
	}

	metadata := map[string]string{
		"author": "test@example.com",
		"issue":  "JIRA-123",
	}

	err := ValidateMetadata(cfg, metadata)
	assert.NoError(t, err, "Should pass validation with all required fields")
}

// TestValidateMetadata_EmptyRequiredField tests validation of empty required fields
func TestValidateMetadata_EmptyRequiredField(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{Name: "author", Required: true, Type: "string"},
			},
		},
	}

	metadata := map[string]string{
		"author": "",
	}

	err := ValidateMetadata(cfg, metadata)
	assert.Error(t, err, "Should fail when required field is empty")
	assert.Contains(t, err.Error(), "author", "Error should mention the field")
}

// TestValidateMetadata_OptionalFields tests optional fields
func TestValidateMetadata_OptionalFields(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{Name: "author", Required: true, Type: "string"},
				{Name: "issue", Required: false, Type: "string"},
			},
		},
	}

	// Only required field provided
	metadata := map[string]string{
		"author": "test@example.com",
	}

	err := ValidateMetadata(cfg, metadata)
	assert.NoError(t, err, "Should pass validation without optional fields")
}
