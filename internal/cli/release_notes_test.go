package cli

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/shipment"
	"github.com/NatoNathan/shipyard/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateReleaseNotesForShipment(t *testing.T) {
	// Create a test project config
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name:      "test-package",
			Path:      ".",
			Ecosystem: config.EcosystemGo,
		},
		Changelog: config.ChangelogConfig{
			Template: "simple",
		},
		ChangeTypes: []config.ChangeTypeConfig{
			{
				Name:        "minor",
				Section:     "Added",
				DisplayName: "Added",
			},
			{
				Name:        "patch",
				Section:     "Fixed",
				DisplayName: "Fixed",
			},
		},
	}

	// Create template engine
	templateEngine := templates.NewTemplateEngine(projectConfig)

	// Create test consignments
	consignments := []*shipment.Consignment{
		{
			ID: "test-consignment-1",
			Packages: map[string]string{
				"test-package": "minor",
			},
			Summary: "Add new feature",
			Created: time.Now(),
		},
		{
			ID: "test-consignment-2",
			Packages: map[string]string{
				"test-package": "patch",
			},
			Summary: "Fix bug",
			Created: time.Now(),
		},
	}

	// Create test shipment
	testShipment := &shipment.Shipment{
		ID:   "test-shipment",
		Date: time.Date(2025, 7, 17, 12, 0, 0, 0, time.UTC),
		Versions: map[string]*semver.Version{
			"test-package": semver.New(1, 2, 3),
		},
		Consignments: consignments,
		Template:     "simple",
	}

	// Test version
	version := semver.New(1, 2, 3)

	// Generate release notes
	releaseNotes, err := generateReleaseNotesForShipment(testShipment, "test-package", version, templateEngine, projectConfig)

	// Assertions
	require.NoError(t, err)
	assert.Contains(t, releaseNotes, "1.2.3")
	assert.Contains(t, releaseNotes, "2025-07-17")
	assert.Contains(t, releaseNotes, "Add new feature")
	assert.Contains(t, releaseNotes, "Fix bug")
}

func TestMapChangeTypeToSection(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		ChangeTypes: []config.ChangeTypeConfig{
			{
				Name:        "feature",
				Section:     "Added",
				DisplayName: "Added Features",
			},
			{
				Name:        "bugfix",
				Section:     "Fixed",
				DisplayName: "Bug Fixes",
			},
			{
				Name:        "improvement",
				DisplayName: "Improvements",
			},
			{
				Name: "other",
			},
		},
	}

	tests := []struct {
		name            string
		changeType      string
		expectedSection string
	}{
		{
			name:            "configured change type with section",
			changeType:      "feature",
			expectedSection: "Added",
		},
		{
			name:            "configured change type with display name but no section",
			changeType:      "improvement",
			expectedSection: "Improvements",
		},
		{
			name:            "configured change type with name only",
			changeType:      "other",
			expectedSection: "other",
		},
		{
			name:            "unknown change type",
			changeType:      "unknown",
			expectedSection: "Changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapChangeTypeToSection(tt.changeType, projectConfig)
			assert.Equal(t, tt.expectedSection, result)
		})
	}
}
