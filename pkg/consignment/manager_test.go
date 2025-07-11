package consignment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
)

func TestNewManager(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeMonorepo,
		Packages: []config.Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
	}

	manager := NewManager(projectConfig)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	expectedDir := filepath.Join(".shipyard", "consignments")
	if manager.GetConsignmentDir() != expectedDir {
		t.Errorf("Expected consignment dir %s, got %s", expectedDir, manager.GetConsignmentDir())
	}
}

func TestNewManagerWithDir(t *testing.T) {
	projectConfig := &config.ProjectConfig{
		Type:    config.RepositoryTypeSingleRepo,
		Package: config.Package{Name: "app", Path: "."},
	}

	customDir := "/tmp/test-consignments"
	manager := NewManagerWithDir(projectConfig, customDir)

	if manager.GetConsignmentDir() != customDir {
		t.Errorf("Expected consignment dir %s, got %s", customDir, manager.GetConsignmentDir())
	}
}

func TestGetAvailablePackages(t *testing.T) {
	tests := []struct {
		name          string
		projectConfig *config.ProjectConfig
		expectedCount int
		expectedNames []string
	}{
		{
			name: "monorepo with multiple packages",
			projectConfig: &config.ProjectConfig{
				Type: config.RepositoryTypeMonorepo,
				Packages: []config.Package{
					{Name: "api", Path: "packages/api"},
					{Name: "frontend", Path: "packages/frontend"},
				},
			},
			expectedCount: 2,
			expectedNames: []string{"api", "frontend"},
		},
		{
			name: "single repo",
			projectConfig: &config.ProjectConfig{
				Type:    config.RepositoryTypeSingleRepo,
				Package: config.Package{Name: "app", Path: "."},
			},
			expectedCount: 1,
			expectedNames: []string{"app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.projectConfig)
			packages := manager.GetAvailablePackages()

			if len(packages) != tt.expectedCount {
				t.Errorf("Expected %d packages, got %d", tt.expectedCount, len(packages))
			}

			for i, expectedName := range tt.expectedNames {
				if i >= len(packages) {
					t.Errorf("Missing package %s", expectedName)
					continue
				}
				if packages[i].Name != expectedName {
					t.Errorf("Expected package name %s, got %s", expectedName, packages[i].Name)
				}
			}
		})
	}
}

func TestCreateConsignment(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeMonorepo,
		Packages: []config.Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
	}

	manager := NewManagerWithDir(projectConfig, tempDir)

	tests := []struct {
		name        string
		packages    []string
		changeType  ChangeType
		summary     string
		expectError bool
	}{
		{
			name:        "valid consignment",
			packages:    []string{"api"},
			changeType:  Minor,
			summary:     "Added new feature",
			expectError: false,
		},
		{
			name:        "multiple packages",
			packages:    []string{"api", "frontend"},
			changeType:  Patch,
			summary:     "Fixed bug in both packages",
			expectError: false,
		},
		{
			name:        "invalid package",
			packages:    []string{"nonexistent"},
			changeType:  Major,
			summary:     "Breaking change",
			expectError: true,
		},
		{
			name:        "empty summary",
			packages:    []string{"api"},
			changeType:  Patch,
			summary:     "",
			expectError: true,
		},
		{
			name:        "no packages",
			packages:    []string{},
			changeType:  Patch,
			summary:     "Some change",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consignment, err := manager.CreateConsignment(tt.packages, tt.changeType, tt.summary)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if consignment == nil {
				t.Error("Expected consignment but got nil")
				return
			}

			// Verify consignment properties
			if consignment.ID == "" {
				t.Error("Consignment ID should not be empty")
			}

			if consignment.Summary != tt.summary {
				t.Errorf("Expected summary %s, got %s", tt.summary, consignment.Summary)
			}

			if len(consignment.Packages) != len(tt.packages) {
				t.Errorf("Expected %d packages, got %d", len(tt.packages), len(consignment.Packages))
			}

			for _, pkg := range tt.packages {
				if changeType, exists := consignment.Packages[pkg]; !exists {
					t.Errorf("Package %s not found in consignment", pkg)
				} else if changeType != string(tt.changeType) {
					t.Errorf("Expected change type %s for package %s, got %s", tt.changeType, pkg, changeType)
				}
			}

			// Verify file was created
			filename := filepath.Join(tempDir, consignment.ID+".md")
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Errorf("Consignment file %s was not created", filename)
			}

			// Verify file content
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read consignment file: %v", err)
			}

			contentStr := string(content)
			if !contains(contentStr, tt.summary) {
				t.Errorf("File content does not contain summary: %s", contentStr)
			}
		})
	}
}

func TestEnsureConsignmentDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-consignments")

	projectConfig := &config.ProjectConfig{
		Type:    config.RepositoryTypeSingleRepo,
		Package: config.Package{Name: "app", Path: "."},
	}

	manager := NewManagerWithDir(projectConfig, testDir)

	// Directory should not exist initially
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Test directory should not exist initially")
	}

	// Create the directory
	if err := manager.EnsureConsignmentDir(); err != nil {
		t.Errorf("Failed to create consignment directory: %v", err)
	}

	// Directory should exist now
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Consignment directory was not created")
	}
}

func TestGenerateConsignmentID(t *testing.T) {
	// Test that IDs are generated and unique
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := generateConsignmentID()

		if id == "" {
			t.Error("Generated ID should not be empty")
		}

		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}

		ids[id] = true
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestConsignmentCreationTime(t *testing.T) {
	tempDir := t.TempDir()

	projectConfig := &config.ProjectConfig{
		Type:    config.RepositoryTypeSingleRepo,
		Package: config.Package{Name: "app", Path: "."},
	}

	manager := NewManagerWithDir(projectConfig, tempDir)

	before := time.Now()
	consignment, err := manager.CreateConsignment([]string{"app"}, Patch, "Test change")
	after := time.Now()

	if err != nil {
		t.Errorf("Failed to create consignment: %v", err)
		return
	}

	if consignment.Created.Before(before) || consignment.Created.After(after) {
		t.Errorf("Consignment creation time %v is not between %v and %v",
			consignment.Created, before, after)
	}
}
