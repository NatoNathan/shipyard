package changelog

import (
	"strings"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

func TestKeepAChangelogTemplate(t *testing.T) {
	template := &KeepAChangelogTemplate{}

	// Test template metadata
	if template.Name() != "keepachangelog" {
		t.Errorf("Expected template name 'keepachangelog', got %s", template.Name())
	}

	if template.Description() == "" {
		t.Error("Expected template description to be non-empty")
	}

	// Test changelog generation
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
		Changelog: config.ChangelogConfig{
			Template: "keepachangelog",
		},
	}

	entries := []ChangelogEntry{
		{
			Version: "1.1.0",
			Date:    time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			Changes: map[string][]ChangelogChange{
				"Added": {
					{
						Summary:     "New feature implementation",
						PackageName: "test-package",
					},
				},
				"Fixed": {
					{
						Summary:     "Bug fix for authentication",
						PackageName: "test-package",
					},
				},
			},
			PackageName: "test-package",
		},
	}

	changelog, err := template.Generate(entries, projectConfig)
	if err != nil {
		t.Fatalf("Failed to generate changelog: %v", err)
	}

	// Verify changelog content
	if !strings.Contains(changelog, "# Changelog") {
		t.Error("Expected changelog to contain header")
	}

	if !strings.Contains(changelog, "## [1.1.0] - 2023-12-01") {
		t.Error("Expected changelog to contain version header")
	}

	if !strings.Contains(changelog, "### Added") {
		t.Error("Expected changelog to contain 'Added' section")
	}

	if !strings.Contains(changelog, "### Fixed") {
		t.Error("Expected changelog to contain 'Fixed' section")
	}

	if !strings.Contains(changelog, "New feature implementation") {
		t.Error("Expected changelog to contain change summary")
	}

	if !strings.Contains(changelog, "Bug fix for authentication") {
		t.Error("Expected changelog to contain bug fix summary")
	}
}

func TestKeepAChangelogTemplateMonorepo(t *testing.T) {
	template := &KeepAChangelogTemplate{}

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeMonorepo,
		Packages: []config.Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
		Changelog: config.ChangelogConfig{
			Template: "keepachangelog",
		},
	}

	entries := []ChangelogEntry{
		{
			Version: "1.1.0",
			Date:    time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			Changes: map[string][]ChangelogChange{
				"Added": {
					{
						Summary:     "New API endpoint",
						PackageName: "api",
					},
				},
			},
			PackageName: "api",
		},
	}

	changelog, err := template.Generate(entries, projectConfig)
	if err != nil {
		t.Fatalf("Failed to generate changelog: %v", err)
	}

	// Verify monorepo-specific content
	if !strings.Contains(changelog, "## [1.1.0] - api - 2023-12-01") {
		t.Error("Expected changelog to contain package name in version header")
	}

	if !strings.Contains(changelog, "**api**: New API endpoint") {
		t.Error("Expected changelog to contain package name in change entry")
	}
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectError  bool
	}{
		{
			name:         "valid keepachangelog template",
			templateName: "keepachangelog",
			expectError:  false,
		},
		{
			name:         "valid conventional template",
			templateName: "conventional",
			expectError:  false,
		},
		{
			name:         "valid simple template",
			templateName: "simple",
			expectError:  false,
		},
		{
			name:         "invalid template",
			templateName: "nonexistent",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := GetTemplate(tt.templateName)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if template != nil {
					t.Error("Expected nil template but got one")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if template == nil {
					t.Error("Expected template but got nil")
				}
				if template.Name() != tt.templateName {
					t.Errorf("Expected template name %s, got %s", tt.templateName, template.Name())
				}
			}
		})
	}
}

func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()

	if len(templates) == 0 {
		t.Error("Expected at least one template")
	}

	expectedTemplates := []string{"keepachangelog", "conventional", "simple"}
	for _, expected := range expectedTemplates {
		found := false
		for _, template := range templates {
			if template == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found in available templates", expected)
		}
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectError  bool
	}{
		{
			name:         "valid template",
			templateName: "keepachangelog",
			expectError:  false,
		},
		{
			name:         "invalid template",
			templateName: "nonexistent",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.templateName)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGenerateChangelog(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
		Changelog: config.ChangelogConfig{
			Template: "keepachangelog",
		},
	}

	generator, err := NewGenerator(projectConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create test consignments
	consignments := []*consignment.Consignment{
		{
			ID: "test1",
			Packages: map[string]string{
				"test-package": "minor",
			},
			Summary: "Added new feature",
			Created: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID: "test2",
			Packages: map[string]string{
				"test-package": "patch",
			},
			Summary: "Fixed bug",
			Created: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	// Create test versions
	version, _ := semver.Parse("1.1.0")
	versions := map[string]*semver.Version{
		"test-package": version,
	}

	changelog, err := generator.GenerateChangelog(consignments, versions)
	if err != nil {
		t.Fatalf("Failed to generate changelog: %v", err)
	}

	// Verify changelog content
	if !strings.Contains(changelog, "# Changelog") {
		t.Error("Expected changelog to contain header")
	}

	if !strings.Contains(changelog, "Added new feature") {
		t.Error("Expected changelog to contain first change summary")
	}

	if !strings.Contains(changelog, "Fixed bug") {
		t.Error("Expected changelog to contain second change summary")
	}
}
