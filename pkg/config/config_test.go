package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectConfig_GetPackages(t *testing.T) {
	tests := []struct {
		name          string
		config        *ProjectConfig
		expectedCount int
		expectedNames []string
	}{
		{
			name: "monorepo with multiple packages",
			config: &ProjectConfig{
				Type: RepositoryTypeMonorepo,
				Packages: []Package{
					{Name: "api", Path: "packages/api"},
					{Name: "frontend", Path: "packages/frontend"},
				},
			},
			expectedCount: 2,
			expectedNames: []string{"api", "frontend"},
		},
		{
			name: "single repo",
			config: &ProjectConfig{
				Type:    RepositoryTypeSingleRepo,
				Package: Package{Name: "app", Path: "."},
			},
			expectedCount: 1,
			expectedNames: []string{"app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages := tt.config.GetPackages()

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

func TestProjectConfig_GetPackageByName(t *testing.T) {
	config := &ProjectConfig{
		Type: RepositoryTypeMonorepo,
		Packages: []Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
	}

	// Test finding existing package
	pkg := config.GetPackageByName("api")
	if pkg == nil {
		t.Error("Expected to find package 'api', but got nil")
	} else if pkg.Name != "api" {
		t.Errorf("Expected package name 'api', got %s", pkg.Name)
	}

	// Test finding non-existing package
	pkg = config.GetPackageByName("nonexistent")
	if pkg != nil {
		t.Error("Expected nil for non-existent package, but got package")
	}
}

func TestProjectConfig_HasPackage(t *testing.T) {
	config := &ProjectConfig{
		Type: RepositoryTypeMonorepo,
		Packages: []Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
	}

	if !config.HasPackage("api") {
		t.Error("Expected HasPackage('api') to return true")
	}

	if config.HasPackage("nonexistent") {
		t.Error("Expected HasPackage('nonexistent') to return false")
	}
}

func TestProjectConfig_GetPackageNames(t *testing.T) {
	config := &ProjectConfig{
		Type: RepositoryTypeMonorepo,
		Packages: []Package{
			{Name: "api", Path: "packages/api"},
			{Name: "frontend", Path: "packages/frontend"},
		},
	}

	names := config.GetPackageNames()
	expectedNames := []string{"api", "frontend"}

	if len(names) != len(expectedNames) {
		t.Errorf("Expected %d package names, got %d", len(expectedNames), len(names))
	}

	for i, expectedName := range expectedNames {
		if i >= len(names) {
			t.Errorf("Missing package name %s", expectedName)
			continue
		}
		if names[i] != expectedName {
			t.Errorf("Expected package name %s, got %s", expectedName, names[i])
		}
	}
}

func TestProjectConfig_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		config  *ProjectConfig
		wantErr bool
	}{
		{
			name: "valid monorepo config",
			config: &ProjectConfig{
				Type: RepositoryTypeMonorepo,
				Repo: "github.com/example/repo",
				Packages: []Package{
					{Name: "api", Path: "packages/api", Ecosystem: "go"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid single repo config",
			config: &ProjectConfig{
				Type:    RepositoryTypeSingleRepo,
				Repo:    "github.com/example/repo",
				Package: Package{Name: "app", Path: ".", Ecosystem: "go"},
			},
			wantErr: false,
		},
		{
			name: "missing repo type",
			config: &ProjectConfig{
				Repo: "github.com/example/repo",
				Packages: []Package{
					{Name: "api", Path: "packages/api", Ecosystem: "go"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid repo type",
			config: &ProjectConfig{
				Type: "invalid",
				Repo: "github.com/example/repo",
				Packages: []Package{
					{Name: "api", Path: "packages/api", Ecosystem: "go"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing repo URL",
			config: &ProjectConfig{
				Type: RepositoryTypeMonorepo,
				Packages: []Package{
					{Name: "api", Path: "packages/api", Ecosystem: "go"},
				},
			},
			wantErr: true,
		},
		{
			name: "monorepo with no packages",
			config: &ProjectConfig{
				Type:     RepositoryTypeMonorepo,
				Repo:     "github.com/example/repo",
				Packages: []Package{},
			},
			wantErr: true,
		},
		{
			name: "invalid package",
			config: &ProjectConfig{
				Type: RepositoryTypeMonorepo,
				Repo: "github.com/example/repo",
				Packages: []Package{
					{Name: "", Path: "packages/api", Ecosystem: "go"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectConfig.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackage_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		pkg     Package
		wantErr bool
	}{
		{
			name: "valid package",
			pkg: Package{
				Name:      "api",
				Path:      "packages/api",
				Ecosystem: "go",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			pkg: Package{
				Path:      "packages/api",
				Ecosystem: "go",
			},
			wantErr: true,
		},
		{
			name: "missing path",
			pkg: Package{
				Name:      "api",
				Ecosystem: "go",
			},
			wantErr: true,
		},
		{
			name: "missing ecosystem",
			pkg: Package{
				Name: "api",
				Path: "packages/api",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pkg.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Package.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultProjectConfig(t *testing.T) {
	config := DefaultProjectConfig()

	if config == nil {
		t.Error("DefaultProjectConfig() returned nil")
		return
	}

	if config.Type != RepositoryTypeMonorepo {
		t.Errorf("Expected default type to be %s, got %s", RepositoryTypeMonorepo, config.Type)
	}

	if config.Repo == "" {
		t.Error("Expected default repo to be non-empty")
	}

	if config.Changelog.Template == "" {
		t.Error("Expected default changelog template to be non-empty")
	}
}

func TestNewMonorepoConfig(t *testing.T) {
	repo := "github.com/example/repo"
	packages := []Package{
		{Name: "api", Path: "packages/api", Ecosystem: "go"},
		{Name: "frontend", Path: "packages/frontend", Ecosystem: "npm"},
	}

	config := NewMonorepoConfig(repo, packages)

	if config.Type != RepositoryTypeMonorepo {
		t.Errorf("Expected type to be %s, got %s", RepositoryTypeMonorepo, config.Type)
	}

	if config.Repo != repo {
		t.Errorf("Expected repo to be %s, got %s", repo, config.Repo)
	}

	if len(config.Packages) != len(packages) {
		t.Errorf("Expected %d packages, got %d", len(packages), len(config.Packages))
	}
}

func TestNewSingleRepoConfig(t *testing.T) {
	repo := "github.com/example/repo"
	pkg := Package{Name: "app", Path: ".", Ecosystem: "go"}

	config := NewSingleRepoConfig(repo, pkg)

	if config.Type != RepositoryTypeSingleRepo {
		t.Errorf("Expected type to be %s, got %s", RepositoryTypeSingleRepo, config.Type)
	}

	if config.Repo != repo {
		t.Errorf("Expected repo to be %s, got %s", repo, config.Repo)
	}

	if config.Package.Name != pkg.Name {
		t.Errorf("Expected package name to be %s, got %s", pkg.Name, config.Package.Name)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "test_field: test message"
	if err.Error() != expected {
		t.Errorf("Expected error message %s, got %s", expected, err.Error())
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create test config content
	configContent := `type: monorepo
repo: github.com/test/repo
changelog:
  template: keepachangelog
packages:
  - name: api
    path: packages/api
    ecosystem: go
    manifest: go.mod
  - name: frontend
    path: packages/frontend
    ecosystem: npm
    manifest: package.json
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config
	if config.Type != RepositoryTypeMonorepo {
		t.Errorf("Expected type %s, got %s", RepositoryTypeMonorepo, config.Type)
	}

	if config.Repo != "github.com/test/repo" {
		t.Errorf("Expected repo github.com/test/repo, got %s", config.Repo)
	}

	if len(config.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(config.Packages))
	}

	if config.Packages[0].Name != "api" {
		t.Errorf("Expected first package name 'api', got %s", config.Packages[0].Name)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	_, err := LoadFromFile("/non/existent/file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadFromFile_EmptyPath(t *testing.T) {
	_, err := LoadFromFile("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestSaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// Create a test config
	config := &ProjectConfig{
		Type: RepositoryTypeMonorepo,
		Repo: "github.com/test/save-repo",
		Changelog: ChangelogConfig{
			Template: "keepachangelog",
		},
		Packages: []Package{
			{Name: "api", Path: "packages/api", Ecosystem: "go", Manifest: "go.mod"},
			{Name: "web", Path: "packages/web", Ecosystem: "npm", Manifest: "package.json"},
		},
	}

	// Save the config
	err := SaveToFile(config, configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load it back and verify
	loadedConfig, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Type != config.Type {
		t.Errorf("Expected type %s, got %s", config.Type, loadedConfig.Type)
	}

	if loadedConfig.Repo != config.Repo {
		t.Errorf("Expected repo %s, got %s", config.Repo, loadedConfig.Repo)
	}

	if len(loadedConfig.Packages) != len(config.Packages) {
		t.Errorf("Expected %d packages, got %d", len(config.Packages), len(loadedConfig.Packages))
	}
}

func TestSaveToFile_InvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	// Create an invalid config (missing required fields)
	config := &ProjectConfig{
		Type: "invalid-type",
		Repo: "", // Empty repo
	}

	err := SaveToFile(config, configFile)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestSaveToFile_NilConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "nil-config.yaml")

	err := SaveToFile(nil, configFile)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestSaveToFile_EmptyPath(t *testing.T) {
	config := DefaultProjectConfig()

	err := SaveToFile(config, "")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestLoadFromDefaultPath(t *testing.T) {
	// This test assumes we're not in a shipyard project directory
	_, err := LoadFromDefaultPath()
	if err == nil {
		t.Error("Expected error when loading from default path in non-shipyard directory")
	}
}

func TestSaveAndLoadDefaultPath(t *testing.T) {
	// Save current directory and create a temp dir
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// Change to temp directory
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create a test config
	config := DefaultProjectConfig()
	config.Repo = "github.com/test/default-path"

	// Save to default path
	err = SaveToDefaultPath(config)
	if err != nil {
		t.Fatalf("Failed to save to default path: %v", err)
	}

	// Load from default path
	loadedConfig, err := LoadFromDefaultPath()
	if err != nil {
		t.Fatalf("Failed to load from default path: %v", err)
	}

	if loadedConfig.Repo != config.Repo {
		t.Errorf("Expected repo %s, got %s", config.Repo, loadedConfig.Repo)
	}
}

func TestSingleRepoConfigSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "single-repo.yaml")

	// Create a single-repo config
	pkg := Package{Name: "app", Path: ".", Ecosystem: "go", Manifest: "go.mod"}
	config := NewSingleRepoConfig("github.com/test/single-app", pkg)

	// Save and load
	err := SaveToFile(config, configFile)
	if err != nil {
		t.Fatalf("Failed to save single-repo config: %v", err)
	}

	loadedConfig, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to load single-repo config: %v", err)
	}

	if loadedConfig.Type != RepositoryTypeSingleRepo {
		t.Errorf("Expected type %s, got %s", RepositoryTypeSingleRepo, loadedConfig.Type)
	}

	if loadedConfig.Package.Name != pkg.Name {
		t.Errorf("Expected package name %s, got %s", pkg.Name, loadedConfig.Package.Name)
	}
}
