package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid minimal config",
			config: &Config{
				Packages: []Package{
					{Name: "test", Path: "."},
				},
			},
			wantErr: false,
		},
		{
			name: "no packages",
			config: &Config{
				Packages: []Package{},
			},
			wantErr: true,
			errMsg:  "at least one package",
		},
		{
			name: "duplicate package names",
			config: &Config{
				Packages: []Package{
					{Name: "test", Path: "."},
					{Name: "test", Path: "other"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name: "invalid package",
			config: &Config{
				Packages: []Package{
					{Name: "", Path: "."},
				},
			},
			wantErr: true,
		},
		{
			name: "valid helm appDependency",
			config: &Config{
				Packages: []Package{
					{Name: "myapp", Path: ".", Ecosystem: EcosystemGo},
					{
						Name:      "myapp-chart",
						Path:      "./charts",
						Ecosystem: EcosystemHelm,
						Options: map[string]interface{}{
							"appDependency": "myapp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid helm appDependency - package not found",
			config: &Config{
				Packages: []Package{
					{
						Name:      "myapp-chart",
						Path:      "./charts",
						Ecosystem: EcosystemHelm,
						Options: map[string]interface{}{
							"appDependency": "nonexistent",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "no such package exists",
		},
		{
			name: "helm without appDependency is valid",
			config: &Config{
				Packages: []Package{
					{Name: "myapp", Path: ".", Ecosystem: EcosystemGo},
					{Name: "myapp-chart", Path: "./charts", Ecosystem: EcosystemHelm},
				},
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetPackage(t *testing.T) {
	config := &Config{
		Packages: []Package{
			{Name: "core", Path: "./core"},
			{Name: "api", Path: "./api"},
		},
	}
	
	tests := []struct {
		name      string
		pkgName   string
		wantFound bool
	}{
		{
			name:      "existing package",
			pkgName:   "core",
			wantFound: true,
		},
		{
			name:      "non-existing package",
			pkgName:   "missing",
			wantFound: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, found := config.GetPackage(tt.pkgName)
			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.pkgName, pkg.Name)
			}
		})
	}
}

func TestConfig_Merge(t *testing.T) {
	base := &Config{
		Packages: []Package{
			{Name: "base", Path: "./base"},
		},
	}
	
	overlay := &Config{
		Packages: []Package{
			{Name: "overlay", Path: "./overlay"},
		},
	}
	
	merged := base.Merge(overlay)
	
	// Should have packages from both
	assert.Len(t, merged.Packages, 2)
	
	// Verify both packages are present
	_, foundBase := merged.GetPackage("base")
	_, foundOverlay := merged.GetPackage("overlay")
	assert.True(t, foundBase)
	assert.True(t, foundOverlay)
}

func TestConfig_Defaults(t *testing.T) {
	config := &Config{
		Packages: []Package{
			{Name: "test", Path: "."},
		},
	}
	
	// Apply defaults
	config = config.WithDefaults()
	
	// Check default values are set
	assert.NotEmpty(t, config.Consignments.Path)
	assert.NotEmpty(t, config.History.Path)
}

func TestWithDefaultsDeepCopy(t *testing.T) {
	original := &Config{
		Packages: []Package{{
			Name: "pkg1",
			Path: ".",
			Dependencies: []Dependency{{
				Package:     "pkg2",
				BumpMapping: map[string]string{"major": "minor"},
			}},
		}},
		Metadata: MetadataConfig{
			Fields: []MetadataField{
				{Name: "author", Required: true},
			},
		},
		PreRelease: PreReleaseConfig{
			Stages: []StageConfig{
				{Name: "alpha", Order: 1},
			},
		},
	}

	result := original.WithDefaults()

	// Modify result
	result.Packages[0].Name = "modified"
	result.Packages[0].Dependencies[0].BumpMapping["major"] = "patch"
	result.Metadata.Fields[0].Name = "changed"
	result.PreRelease.Stages[0].Name = "beta"

	// Original should be unchanged
	assert.Equal(t, "pkg1", original.Packages[0].Name)
	assert.Equal(t, "minor", original.Packages[0].Dependencies[0].BumpMapping["major"])
	assert.Equal(t, "author", original.Metadata.Fields[0].Name)
	assert.Equal(t, "alpha", original.PreRelease.Stages[0].Name)

	// Verify defaults were applied
	assert.Equal(t, ".shipyard/consignments", result.Consignments.Path)
	assert.Equal(t, ".shipyard/history.json", result.History.Path)
	assert.Equal(t, "linked", result.Packages[0].Dependencies[0].Strategy)
}

// TestPackage_IsTagOnly tests the IsTagOnly method
func TestPackage_IsTagOnly(t *testing.T) {
	tests := []struct {
		name         string
		versionFiles []string
		expected     bool
	}{
		{
			name:         "tag-only keyword present",
			versionFiles: []string{"tag-only"},
			expected:     true,
		},
		{
			name:         "tag-only with other files (still tag-only)",
			versionFiles: []string{"version.go", "tag-only"},
			expected:     true,
		},
		{
			name:         "normal version files",
			versionFiles: []string{"version.go", "go.mod"},
			expected:     false,
		},
		{
			name:         "empty version files",
			versionFiles: []string{},
			expected:     false,
		},
		{
			name:         "nil version files",
			versionFiles: nil,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := Package{
				Name:         "test",
				Path:         "./",
				Ecosystem:    "go",
				VersionFiles: tt.versionFiles,
			}
			assert.Equal(t, tt.expected, pkg.IsTagOnly())
		})
	}
}
