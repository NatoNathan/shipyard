package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBuiltinTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		templateName string
		expectError  bool
		checkContent func(t *testing.T, content string)
	}{
		{
			name:         "changelog default",
			templateType: TemplateTypeChangelog,
			templateName: "default",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "# Changelog")
				assert.Contains(t, content, "{{ .Version }}")
			},
		},
		{
			name:         "changelog keepachangelog",
			templateType: TemplateTypeChangelog,
			templateName: "keepachangelog",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "# Changelog")
				assert.Contains(t, content, "Keep a Changelog")
			},
		},
		{
			name:         "tagname default",
			templateType: TemplateTypeTag,
			templateName: "default",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "v{{")
			},
		},
		{
			name:         "tagname go",
			templateType: TemplateTypeTag,
			templateName: "go",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "v{{")
			},
		},
		{
			name:         "tagname npm",
			templateType: TemplateTypeTag,
			templateName: "npm",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "@{{")
			},
		},
		{
			name:         "releasenotes default",
			templateType: TemplateTypeReleaseNotes,
			templateName: "default",
			expectError:  false,
			checkContent: func(t *testing.T, content string) {
				assert.Contains(t, content, "# Release")
				assert.Contains(t, content, "{{ .Version }}")
			},
		},
		{
			name:         "nonexistent template",
			templateType: TemplateTypeChangelog,
			templateName: "nonexistent",
			expectError:  true,
		},
		{
			name:         "wrong type - keepachangelog is not a tagname",
			templateType: TemplateTypeTag,
			templateName: "keepachangelog",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GetBuiltinTemplate(tt.templateType, tt.templateName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, content)
				if tt.checkContent != nil {
					tt.checkContent(t, content)
				}
			}
		})
	}
}

func TestListBuiltinTemplates(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		expectError  bool
		checkNames   func(t *testing.T, names []string)
	}{
		{
			name:         "list changelog templates",
			templateType: TemplateTypeChangelog,
			expectError:  false,
			checkNames: func(t *testing.T, names []string) {
				assert.Contains(t, names, "default")
				assert.Contains(t, names, "keepachangelog")
			},
		},
		{
			name:         "list tagname templates",
			templateType: TemplateTypeTag,
			expectError:  false,
			checkNames: func(t *testing.T, names []string) {
				assert.Contains(t, names, "default")
				assert.Contains(t, names, "go")
				assert.Contains(t, names, "npm")
				assert.Contains(t, names, "go-annotated")
				assert.Contains(t, names, "detailed-annotated")
			},
		},
		{
			name:         "list release templates",
			templateType: TemplateTypeRelease,
			expectError:  false,
			checkNames: func(t *testing.T, names []string) {
				assert.Contains(t, names, "date")
				assert.Contains(t, names, "versions")
			},
		},
		{
			name:         "list releasenotes templates",
			templateType: TemplateTypeReleaseNotes,
			expectError:  false,
			checkNames: func(t *testing.T, names []string) {
				assert.Contains(t, names, "default")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := ListBuiltinTemplates(tt.templateType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, names)
				if tt.checkNames != nil {
					tt.checkNames(t, names)
				}
			}
		})
	}
}

func TestGetAllBuiltinTemplates(t *testing.T) {
	allTemplates, err := GetAllBuiltinTemplates()
	require.NoError(t, err)

	// Should have all four types
	assert.Contains(t, allTemplates, TemplateTypeChangelog)
	assert.Contains(t, allTemplates, TemplateTypeTag)
	assert.Contains(t, allTemplates, TemplateTypeRelease)
	assert.Contains(t, allTemplates, TemplateTypeReleaseNotes)

	// Changelog should have default and keepachangelog
	assert.Contains(t, allTemplates[TemplateTypeChangelog], "default")
	assert.Contains(t, allTemplates[TemplateTypeChangelog], "keepachangelog")

	// Tag should have default, go, npm, and annotated templates
	assert.Contains(t, allTemplates[TemplateTypeTag], "default")
	assert.Contains(t, allTemplates[TemplateTypeTag], "go")
	assert.Contains(t, allTemplates[TemplateTypeTag], "npm")
	assert.Contains(t, allTemplates[TemplateTypeTag], "go-annotated")
	assert.Contains(t, allTemplates[TemplateTypeTag], "detailed-annotated")

	// Release should have date and versions
	assert.Contains(t, allTemplates[TemplateTypeRelease], "date")
	assert.Contains(t, allTemplates[TemplateTypeRelease], "versions")

	// Releasenotes should have default
	assert.Contains(t, allTemplates[TemplateTypeReleaseNotes], "default")

	// All templates should be non-empty
	for templateType, templates := range allTemplates {
		for name, content := range templates {
			assert.NotEmpty(t, content, "template %s/%s should not be empty", templateType, name)
		}
	}
}

func TestBuiltinTemplateHelpers(t *testing.T) {
	t.Run("GetBuiltinChangelogTemplate", func(t *testing.T) {
		content, err := GetBuiltinChangelogTemplate("default")
		require.NoError(t, err)
		assert.Contains(t, content, "# Changelog")
	})

	t.Run("GetBuiltinTagTemplate", func(t *testing.T) {
		content, err := GetBuiltinTagTemplate("default")
		require.NoError(t, err)
		assert.Contains(t, content, "v{{")
	})

	t.Run("GetBuiltinReleaseNotesTemplate", func(t *testing.T) {
		content, err := GetBuiltinReleaseNotesTemplate("default")
		require.NoError(t, err)
		assert.Contains(t, content, "# Release")
	})

	t.Run("GetDefaultChangelogTemplate", func(t *testing.T) {
		content, err := GetDefaultChangelogTemplate()
		require.NoError(t, err)
		assert.Contains(t, content, "# Changelog")
	})

	t.Run("GetDefaultTagTemplate", func(t *testing.T) {
		content, err := GetDefaultTagTemplate()
		require.NoError(t, err)
		assert.Contains(t, content, "v{{")
	})

	t.Run("GetDefaultReleaseNotesTemplate", func(t *testing.T) {
		content, err := GetDefaultReleaseNotesTemplate()
		require.NoError(t, err)
		assert.Contains(t, content, "# Release")
	})
}

func TestTemplateLoader_WithType(t *testing.T) {
	loader := NewTemplateLoader()

	t.Run("load changelog with type", func(t *testing.T) {
		content, err := loader.Load("builtin:default", TemplateTypeChangelog)
		require.NoError(t, err)
		assert.Contains(t, content, "# Changelog")
	})

	t.Run("load keepachangelog", func(t *testing.T) {
		content, err := loader.Load("builtin:keepachangelog", TemplateTypeChangelog)
		require.NoError(t, err)
		assert.Contains(t, content, "Keep a Changelog")
	})

	t.Run("load tagname go", func(t *testing.T) {
		content, err := loader.Load("builtin:go", TemplateTypeTag)
		require.NoError(t, err)
		assert.Contains(t, content, "v{{")
	})

	t.Run("load tagname npm", func(t *testing.T) {
		content, err := loader.Load("builtin:npm", TemplateTypeTag)
		require.NoError(t, err)
		assert.Contains(t, content, "@{{")
	})

	t.Run("fail when type mismatch - keepachangelog is not a tagname", func(t *testing.T) {
		_, err := loader.Load("builtin:keepachangelog", TemplateTypeTag)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "builtin template not found")
	})

	t.Run("fail when no type specified for builtin", func(t *testing.T) {
		_, err := loader.Load("builtin:default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "require expectedType")
	})
}
