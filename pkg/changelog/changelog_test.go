package changelog

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

func TestGenerateChangelog(t *testing.T) {
	// Setup test config
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Repo: "github.com/test/test-package",
		Package: config.Package{
			Name: "test-package",
		},
		ChangeTypes: []config.ChangeTypeConfig{
			{
				Name:        "feature",
				DisplayName: "Features",
				Section:     "Added",
				SemverBump:  "minor",
			},
			{
				Name:        "bugfix",
				DisplayName: "Bug Fixes",
				Section:     "Fixed",
				SemverBump:  "patch",
			},
		},
		Changelog: config.ChangelogConfig{
			Template: "simple", // Use simple template to avoid complex sections
		},
	}

	// Create generator
	generator, err := NewGenerator(projectConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create test consignments
	consignments := []*consignment.Consignment{
		{
			ID:      "test-1",
			Summary: "Add new feature",
			Packages: map[string]string{
				"test-package": "feature",
			},
			Created: time.Now(),
		},
		{
			ID:      "test-2",
			Summary: "Fix bug in feature",
			Packages: map[string]string{
				"test-package": "bugfix",
			},
			Created: time.Now(),
		},
	}

	// Create version map
	version, _ := semver.Parse("1.1.0")
	versions := map[string]*semver.Version{
		"test-package": version,
	}

	// Generate changelog
	changelog, err := generator.GenerateChangelog(consignments, versions)
	if err != nil {
		t.Fatalf("Failed to generate changelog: %v", err)
	}

	// Verify changelog contains expected content
	if len(changelog) == 0 {
		t.Error("Generated changelog is empty")
	}

	// Check that the version is present
	if !contains(changelog, "1.1.0") {
		t.Error("Changelog should contain version 1.1.0")
	}

	// Check that consignment summaries are present
	if !contains(changelog, "Add new feature") {
		t.Error("Changelog should contain 'Add new feature'")
	}

	if !contains(changelog, "Fix bug in feature") {
		t.Error("Changelog should contain 'Fix bug in feature'")
	}
}

func TestGenerateChangelogsForPackages(t *testing.T) {
	// Setup test config for monorepo
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeMonorepo,
		Repo: "github.com/test/monorepo",
		Packages: []config.Package{
			{Name: "package-a"},
			{Name: "package-b"},
		},
		ChangeTypes: []config.ChangeTypeConfig{
			{
				Name:        "feature",
				DisplayName: "Features",
				Section:     "Added",
				SemverBump:  "minor",
			},
			{
				Name:        "bugfix",
				DisplayName: "Bug Fixes",
				Section:     "Fixed",
				SemverBump:  "patch",
			},
		},
		Changelog: config.ChangelogConfig{
			Template: "simple",
		},
	}

	// Create generator
	generator, err := NewGenerator(projectConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create test consignments for different packages
	consignments := []*consignment.Consignment{
		{
			ID:      "test-1",
			Summary: "Add feature to package A",
			Packages: map[string]string{
				"package-a": "feature",
			},
			Created: time.Now(),
		},
		{
			ID:      "test-2",
			Summary: "Fix bug in package B",
			Packages: map[string]string{
				"package-b": "bugfix",
			},
			Created: time.Now(),
		},
		{
			ID:      "test-3",
			Summary: "Add feature to both packages",
			Packages: map[string]string{
				"package-a": "feature",
				"package-b": "feature",
			},
			Created: time.Now(),
		},
	}

	// Create version map for both packages
	versionA, _ := semver.Parse("1.1.0")
	versionB, _ := semver.Parse("2.0.0")
	versions := map[string]*semver.Version{
		"package-a": versionA,
		"package-b": versionB,
	}

	// Generate separate changelogs for each package
	packageChangelogs, err := generator.GenerateChangelogsForPackages(consignments, versions)
	if err != nil {
		t.Fatalf("Failed to generate package changelogs: %v", err)
	}

	// Verify we got changelogs for both packages
	if len(packageChangelogs) != 2 {
		t.Errorf("Expected 2 package changelogs, got %d", len(packageChangelogs))
	}

	// Verify package-a changelog
	changelogA, existsA := packageChangelogs["package-a"]
	if !existsA {
		t.Error("Missing changelog for package-a")
	} else {
		if !contains(changelogA, "1.1.0") {
			t.Error("Package-a changelog should contain version 1.1.0")
		}
		if !contains(changelogA, "Add feature to package A") {
			t.Error("Package-a changelog should contain 'Add feature to package A'")
		}
		if !contains(changelogA, "Add feature to both packages") {
			t.Error("Package-a changelog should contain 'Add feature to both packages'")
		}
		// Should NOT contain package-b specific change
		if contains(changelogA, "Fix bug in package B") {
			t.Error("Package-a changelog should NOT contain 'Fix bug in package B'")
		}
	}

	// Verify package-b changelog
	changelogB, existsB := packageChangelogs["package-b"]
	if !existsB {
		t.Error("Missing changelog for package-b")
	} else {
		if !contains(changelogB, "2.0.0") {
			t.Error("Package-b changelog should contain version 2.0.0")
		}
		if !contains(changelogB, "Fix bug in package B") {
			t.Error("Package-b changelog should contain 'Fix bug in package B'")
		}
		if !contains(changelogB, "Add feature to both packages") {
			t.Error("Package-b changelog should contain 'Add feature to both packages'")
		}
		// Should NOT contain package-a specific change
		if contains(changelogB, "Add feature to package A") {
			t.Error("Package-b changelog should NOT contain 'Add feature to package A'")
		}
	}
}

func TestMapChangeTypeToSection(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		ChangeTypes: []config.ChangeTypeConfig{
			{
				Name:        "feature",
				DisplayName: "Features",
				Section:     "Added",
				SemverBump:  "minor",
			},
			{
				Name:        "bugfix",
				DisplayName: "Bug Fixes",
				Section:     "Fixed",
				SemverBump:  "patch",
			},
		},
	}

	generator, _ := NewGenerator(projectConfig)

	tests := []struct {
		changeType      string
		expectedSection string
	}{
		{"feature", "Added"},
		{"bugfix", "Fixed"},
		{"unknown", "Changed"},
	}

	for _, test := range tests {
		section := generator.mapChangeTypeToSection(test.changeType)
		if section != test.expectedSection {
			t.Errorf("mapChangeTypeToSection(%s) = %s, want %s", test.changeType, section, test.expectedSection)
		}
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		template    string
		shouldError bool
	}{
		{"keepachangelog", false},
		{"conventional", false},
		{"simple", false},
		{"invalid", true},
	}

	for _, test := range tests {
		err := ValidateTemplate(test.template)
		if test.shouldError && err == nil {
			t.Errorf("ValidateTemplate(%s) should have returned an error", test.template)
		}
		if !test.shouldError && err != nil {
			t.Errorf("ValidateTemplate(%s) should not have returned an error: %v", test.template, err)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
