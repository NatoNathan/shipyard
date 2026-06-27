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

func TestPreReleaseConfig_StageNavigation(t *testing.T) {
	cfg := &PreReleaseConfig{
		Stages: []StageConfig{
			{Name: "rc", Order: 3},
			{Name: "alpha", Order: 1},
			{Name: "beta", Order: 2},
		},
	}

	t.Run("GetLowestOrderStage returns alpha", func(t *testing.T) {
		s, ok := cfg.GetLowestOrderStage()
		assert.True(t, ok)
		assert.Equal(t, "alpha", s.Name)
	})

	t.Run("GetLowestOrderStage empty config", func(t *testing.T) {
		_, ok := (&PreReleaseConfig{}).GetLowestOrderStage()
		assert.False(t, ok)
	})

	t.Run("GetStageByName found", func(t *testing.T) {
		s, ok := cfg.GetStageByName("beta")
		assert.True(t, ok)
		assert.Equal(t, 2, s.Order)
	})

	t.Run("GetStageByName not found", func(t *testing.T) {
		_, ok := cfg.GetStageByName("gamma")
		assert.False(t, ok)
	})

	t.Run("GetNextStage from alpha returns beta", func(t *testing.T) {
		s, ok := cfg.GetNextStage("alpha")
		assert.True(t, ok)
		assert.Equal(t, "beta", s.Name)
	})

	t.Run("GetNextStage from rc returns false (highest)", func(t *testing.T) {
		_, ok := cfg.GetNextStage("rc")
		assert.False(t, ok)
	})

	t.Run("GetNextStage unknown stage returns false", func(t *testing.T) {
		_, ok := cfg.GetNextStage("gamma")
		assert.False(t, ok)
	})

	t.Run("IsHighestStage rc is highest", func(t *testing.T) {
		assert.True(t, cfg.IsHighestStage("rc"))
	})

	t.Run("IsHighestStage alpha is not highest", func(t *testing.T) {
		assert.False(t, cfg.IsHighestStage("alpha"))
	})

	t.Run("IsHighestStage unknown returns false", func(t *testing.T) {
		assert.False(t, cfg.IsHighestStage("gamma"))
	})
}

func TestPreReleaseConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     PreReleaseConfig
		wantErr string
	}{
		{
			name: "no stages is valid",
			cfg:  PreReleaseConfig{},
		},
		{
			name: "valid stages",
			cfg: PreReleaseConfig{Stages: []StageConfig{
				{Name: "alpha", Order: 1},
				{Name: "beta", Order: 2},
			}},
		},
		{
			name: "duplicate stage name",
			cfg: PreReleaseConfig{Stages: []StageConfig{
				{Name: "alpha", Order: 1},
				{Name: "alpha", Order: 2},
			}},
			wantErr: "duplicate stage name: alpha",
		},
		{
			name: "duplicate stage order",
			cfg: PreReleaseConfig{Stages: []StageConfig{
				{Name: "alpha", Order: 1},
				{Name: "beta", Order: 1},
			}},
			wantErr: "duplicate stage order: 1",
		},
		{
			name: "empty stage name",
			cfg: PreReleaseConfig{Stages: []StageConfig{
				{Name: "", Order: 1},
			}},
			wantErr: "stage name is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewRemoteConfig(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantURL  string
		wantGit  string
		wantPath string
		wantRef  string
	}{
		{
			name:    "plain HTTP URL becomes URL field",
			input:   "https://example.com/shipyard.yaml",
			wantURL: "https://example.com/shipyard.yaml",
		},
		{
			name:    "SSH git URL",
			input:   "git@github.com:org/repo.git",
			wantGit: "git@github.com:org/repo.git",
			wantPath: "shipyard.yaml",
			wantRef:  "main",
		},
		{
			name:    "HTTPS git URL with fragment",
			input:   "https://github.com/org/repo.git#configs/shipyard.yaml@v1.2.0",
			wantGit: "https://github.com/org/repo.git",
			wantPath: "configs/shipyard.yaml",
			wantRef:  "v1.2.0",
		},
		{
			name:    "HTTPS .git URL without fragment defaults",
			input:   "https://github.com/org/repo.git",
			wantGit: "https://github.com/org/repo.git",
			wantPath: "shipyard.yaml",
			wantRef:  "main",
		},
		{
			name:    "plain string treated as URL",
			input:   "file:///local/shipyard.yaml",
			wantURL: "file:///local/shipyard.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rc := NewRemoteConfig(tc.input)
			assert.Equal(t, tc.wantURL, rc.URL)
			assert.Equal(t, tc.wantGit, rc.Git)
			assert.Equal(t, tc.wantPath, rc.Path)
			assert.Equal(t, tc.wantRef, rc.Ref)
		})
	}
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
