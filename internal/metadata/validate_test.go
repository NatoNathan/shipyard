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

// TestValidateListItems tests per-item validation of list fields
func TestValidateListItems(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	t.Run("string items against allowed values", func(t *testing.T) {
		field := config.MetadataField{
			Name: "labels", Type: "list",
			AllowedValues: []string{"bug", "feature", "docs"},
		}

		assert.NoError(t, validateList(field, []interface{}{"bug", "docs"}))

		err := validateList(field, []interface{}{"bug", "chore"})
		assert.Error(t, err, "Should reject item not in allowed values")
		assert.Contains(t, err.Error(), "chore")
	})

	t.Run("string item type enforced", func(t *testing.T) {
		field := config.MetadataField{Name: "labels", Type: "list", ItemType: "string"}

		err := validateList(field, []interface{}{"ok", 42})
		assert.Error(t, err, "Should reject non-string item")
		assert.Contains(t, err.Error(), "item 1")
	})

	t.Run("int items with min and max", func(t *testing.T) {
		field := config.MetadataField{
			Name: "ports", Type: "list", ItemType: "int",
			Min: intPtr(1), Max: intPtr(100),
		}

		assert.NoError(t, validateList(field, []interface{}{1, 50, 100}))
		assert.Error(t, validateList(field, []interface{}{0}), "Should reject item below min")
		assert.Error(t, validateList(field, []interface{}{101}), "Should reject item above max")
		assert.Error(t, validateList(field, []interface{}{"nope"}), "Should reject non-int item")
	})

	t.Run("unsupported item type", func(t *testing.T) {
		field := config.MetadataField{Name: "things", Type: "list", ItemType: "bool"}
		assert.Error(t, validateList(field, []interface{}{true}))
	})
}

// TestValidateMetadata_ListItemsEndToEnd validates list items through ValidateMetadata
func TestValidateMetadata_ListItemsEndToEnd(t *testing.T) {
	cfg := &config.Config{
		Metadata: config.MetadataConfig{
			Fields: []config.MetadataField{
				{Name: "labels", Type: "list", AllowedValues: []string{"bug", "feature"}},
			},
		},
	}

	assert.NoError(t, ValidateMetadata(cfg, map[string]string{"labels": "bug,feature"}))

	err := ValidateMetadata(cfg, map[string]string{"labels": "bug,invalid"})
	assert.Error(t, err, "Should reject list containing disallowed value")
}
