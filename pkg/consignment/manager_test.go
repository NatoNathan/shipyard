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

func TestCalculateNextVersionWithConsignments(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary package.json for testing
	packagePath := filepath.Join(tempDir, "test-package")
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		t.Fatal(err)
	}

	packageJSON := `{
		"name": "test-package",
		"version": "1.0.0"
	}`

	if err := os.WriteFile(filepath.Join(packagePath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .shipyard directory for history
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name:      "test-package",
			Path:      packagePath,
			Ecosystem: config.EcosystemNPM,
			Manifest:  filepath.Join(packagePath, "package.json"),
		},
	}

	consignmentDir := filepath.Join(tempDir, "consignments")
	manager := NewManagerWithDir(projectConfig, consignmentDir)

	// Create consignments with different change types
	tests := []struct {
		name         string
		changeType   ChangeType
		expectedNext string
	}{
		{
			name:         "patch change",
			changeType:   Patch,
			expectedNext: "1.0.1",
		},
		{
			name:         "minor change",
			changeType:   Minor,
			expectedNext: "1.1.0",
		},
		{
			name:         "major change",
			changeType:   Major,
			expectedNext: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing consignments and history
			os.RemoveAll(consignmentDir)
			os.RemoveAll(filepath.Join(shipyardDir, "shipment-history.json"))

			// Create a consignment with the specified change type
			_, err := manager.CreateConsignment([]string{"test-package"}, tt.changeType, "Test change")
			if err != nil {
				t.Fatalf("Failed to create consignment: %v", err)
			}

			// Calculate the next version
			version, err := manager.CalculateNextVersion("test-package")
			if err != nil {
				t.Fatalf("Failed to calculate next version: %v", err)
			}

			if version.String() != tt.expectedNext {
				t.Errorf("Expected next version %s, got %s", tt.expectedNext, version.String())
			}
		})
	}
}

func TestCalculateNextVersionWithMultipleConsignments(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary package.json for testing
	packagePath := filepath.Join(tempDir, "test-package")
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		t.Fatal(err)
	}

	packageJSON := `{
		"name": "test-package",
		"version": "1.0.0"
	}`

	if err := os.WriteFile(filepath.Join(packagePath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .shipyard directory for history
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name:      "test-package",
			Path:      packagePath,
			Ecosystem: config.EcosystemNPM,
			Manifest:  filepath.Join(packagePath, "package.json"),
		},
	}

	consignmentDir := filepath.Join(tempDir, "consignments")
	manager := NewManagerWithDir(projectConfig, consignmentDir)

	// Create multiple consignments
	consignments := []struct {
		changeType ChangeType
		summary    string
	}{
		{Patch, "Fix bug 1"},
		{Minor, "Add new feature"},
		{Patch, "Fix bug 2"},
	}

	for _, c := range consignments {
		_, err := manager.CreateConsignment([]string{"test-package"}, c.changeType, c.summary)
		if err != nil {
			t.Fatalf("Failed to create consignment: %v", err)
		}
	}

	// Calculate the next version - should be minor bump (highest change type)
	version, err := manager.CalculateNextVersion("test-package")
	if err != nil {
		t.Fatalf("Failed to calculate next version: %v", err)
	}

	expected := "1.1.0"
	if version.String() != expected {
		t.Errorf("Expected next version %s, got %s", expected, version.String())
	}
}

func TestCalculateAllVersions(t *testing.T) {
	tempDir := t.TempDir()

	// Create test packages
	packages := []struct {
		name      string
		ecosystem config.PackageEcosystem
		content   string
		filename  string
	}{
		{
			name:      "frontend",
			ecosystem: config.EcosystemNPM,
			content:   `{"name": "frontend", "version": "1.0.0"}`,
			filename:  "package.json",
		},
		{
			name:      "backend",
			ecosystem: config.EcosystemGo,
			content:   "module backend\n\ngo 1.21",
			filename:  "go.mod",
		},
	}

	// Create .shipyard directory for history
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		t.Fatal(err)
	}

	var configPackages []config.Package
	for _, pkg := range packages {
		pkgPath := filepath.Join(tempDir, pkg.name)
		if err := os.MkdirAll(pkgPath, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(pkgPath, pkg.filename), []byte(pkg.content), 0644); err != nil {
			t.Fatal(err)
		}

		// For Go packages, create a .version file
		if pkg.ecosystem == config.EcosystemGo {
			if err := os.WriteFile(filepath.Join(pkgPath, ".version"), []byte("1.0.0"), 0644); err != nil {
				t.Fatal(err)
			}
		}

		configPackages = append(configPackages, config.Package{
			Name:      pkg.name,
			Path:      pkgPath,
			Ecosystem: pkg.ecosystem,
			Manifest:  filepath.Join(pkgPath, pkg.filename),
		})
	}

	projectConfig := &config.ProjectConfig{
		Type:     config.RepositoryTypeMonorepo,
		Packages: configPackages,
	}

	consignmentDir := filepath.Join(tempDir, "consignments")
	manager := NewManagerWithDir(projectConfig, consignmentDir)

	// Create consignments for each package
	_, err := manager.CreateConsignment([]string{"frontend"}, Minor, "Frontend feature")
	if err != nil {
		t.Fatalf("Failed to create frontend consignment: %v", err)
	}

	_, err = manager.CreateConsignment([]string{"backend"}, Patch, "Backend fix")
	if err != nil {
		t.Fatalf("Failed to create backend consignment: %v", err)
	}

	// Calculate all versions
	versions, err := manager.CalculateAllVersions()
	if err != nil {
		t.Fatalf("Failed to calculate all versions: %v", err)
	}

	// Check that we got versions for all packages
	if len(versions) != len(packages) {
		t.Errorf("Expected %d versions, got %d", len(packages), len(versions))
	}

	// Check specific versions
	if versions["frontend"].String() != "1.1.0" {
		t.Errorf("Expected frontend version 1.1.0, got %s", versions["frontend"].String())
	}

	if versions["backend"].String() != "1.0.1" {
		t.Errorf("Expected backend version 1.0.1, got %s", versions["backend"].String())
	}
}

func TestClearConsignments(t *testing.T) {
	tempDir := t.TempDir()

	projectConfig := &config.ProjectConfig{
		Type:    config.RepositoryTypeSingleRepo,
		Package: config.Package{Name: "app", Path: "."},
	}

	manager := NewManagerWithDir(projectConfig, tempDir)

	// Create some consignments
	_, err := manager.CreateConsignment([]string{"app"}, Patch, "Change 1")
	if err != nil {
		t.Fatalf("Failed to create consignment: %v", err)
	}

	_, err = manager.CreateConsignment([]string{"app"}, Minor, "Change 2")
	if err != nil {
		t.Fatalf("Failed to create consignment: %v", err)
	}

	// Verify consignments exist
	consignments, err := manager.GetConsignmens()
	if err != nil {
		t.Fatalf("Failed to get consignments: %v", err)
	}
	if len(consignments) != 2 {
		t.Errorf("Expected 2 consignments, got %d", len(consignments))
	}

	// Clear consignments
	if err := manager.ClearConsignments(); err != nil {
		t.Fatalf("Failed to clear consignments: %v", err)
	}

	// Verify consignments are cleared
	consignments, err = manager.GetConsignmens()
	if err != nil {
		t.Fatalf("Failed to get consignments after clear: %v", err)
	}
	if len(consignments) != 0 {
		t.Errorf("Expected 0 consignments after clear, got %d", len(consignments))
	}
}

func TestApplyConsignments(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary package.json for testing
	packagePath := filepath.Join(tempDir, "test-package")
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		t.Fatal(err)
	}

	packageJSON := `{
		"name": "test-package",
		"version": "1.0.0"
	}`

	if err := os.WriteFile(filepath.Join(packagePath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .shipyard directory for history
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name:      "test-package",
			Path:      packagePath,
			Ecosystem: config.EcosystemNPM,
			Manifest:  filepath.Join(packagePath, "package.json"),
		},
		Changelog: config.ChangelogConfig{
			Template: "default",
		},
	}

	consignmentDir := filepath.Join(tempDir, "consignments")
	manager := NewManagerWithDir(projectConfig, consignmentDir)

	// Create consignments
	_, err := manager.CreateConsignment([]string{"test-package"}, Minor, "Add feature")
	if err != nil {
		t.Fatalf("Failed to create consignment: %v", err)
	}

	// Apply consignments
	versions, err := manager.ApplyConsignments()
	if err != nil {
		t.Fatalf("Failed to apply consignments: %v", err)
	}

	// Check that versions were calculated correctly
	if versions["test-package"].String() != "1.1.0" {
		t.Errorf("Expected version 1.1.0, got %s", versions["test-package"].String())
	}

	// Check that consignments were cleared
	consignments, err := manager.GetConsignmens()
	if err != nil {
		t.Fatalf("Failed to get consignments after apply: %v", err)
	}
	if len(consignments) != 0 {
		t.Errorf("Expected 0 consignments after apply, got %d", len(consignments))
	}

	// Check that the package.json was updated
	updatedContent, err := os.ReadFile(filepath.Join(packagePath, "package.json"))
	if err != nil {
		t.Fatalf("Failed to read updated package.json: %v", err)
	}

	if !strings.Contains(string(updatedContent), `"version": "1.1.0"`) {
		t.Errorf("package.json was not updated correctly. Content: %s", string(updatedContent))
	}
}

func TestCalculateNextVersionWithHistory(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary package.json for testing
	packagePath := filepath.Join(tempDir, "test-package")
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		t.Fatal(err)
	}

	packageJSON := `{
		"name": "test-package",
		"version": "1.0.0"
	}`

	if err := os.WriteFile(filepath.Join(packagePath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .shipyard directory for history
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	if err := os.MkdirAll(shipyardDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name:      "test-package",
			Path:      packagePath,
			Ecosystem: config.EcosystemNPM,
			Manifest:  filepath.Join(packagePath, "package.json"),
		},
		Changelog: config.ChangelogConfig{
			Template: "default",
		},
	}

	consignmentDir := filepath.Join(tempDir, "consignments")
	manager := NewManagerWithDir(projectConfig, consignmentDir)

	// Create and apply first consignment to build history
	_, err := manager.CreateConsignment([]string{"test-package"}, Minor, "First feature")
	if err != nil {
		t.Fatalf("Failed to create first consignment: %v", err)
	}

	// Apply consignments to create history
	versions, err := manager.ApplyConsignments()
	if err != nil {
		t.Fatalf("Failed to apply first consignment: %v", err)
	}

	// Verify first version is correct
	if versions["test-package"].String() != "1.1.0" {
		t.Errorf("Expected first version 1.1.0, got %s", versions["test-package"].String())
	}

	// Create second consignment
	_, err = manager.CreateConsignment([]string{"test-package"}, Patch, "Bug fix")
	if err != nil {
		t.Fatalf("Failed to create second consignment: %v", err)
	}

	// Calculate next version - should be based on history version (1.1.0) not manifest version (1.0.0)
	nextVersion, err := manager.CalculateNextVersion("test-package")
	if err != nil {
		t.Fatalf("Failed to calculate next version: %v", err)
	}

	expected := "1.1.1"
	if nextVersion.String() != expected {
		t.Errorf("Expected next version %s (based on history), got %s", expected, nextVersion.String())
	}

	// Apply second consignment
	versions, err = manager.ApplyConsignments()
	if err != nil {
		t.Fatalf("Failed to apply second consignment: %v", err)
	}

	// Verify second version is correct
	if versions["test-package"].String() != "1.1.1" {
		t.Errorf("Expected second version 1.1.1, got %s", versions["test-package"].String())
	}

	// Create third consignment
	_, err = manager.CreateConsignment([]string{"test-package"}, Major, "Breaking change")
	if err != nil {
		t.Fatalf("Failed to create third consignment: %v", err)
	}

	// Calculate next version - should be based on latest history version (1.1.1)
	nextVersion, err = manager.CalculateNextVersion("test-package")
	if err != nil {
		t.Fatalf("Failed to calculate next version: %v", err)
	}

	expected = "2.0.0"
	if nextVersion.String() != expected {
		t.Errorf("Expected next version %s (based on history), got %s", expected, nextVersion.String())
	}
}
