//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/fileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/spf13/viper"
)

func TestConfigLoad_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.yaml")
	
	yamlContent := `
packages:
  - name: core
    path: ./core
    ecosystem: go
  - name: api
    path: ./api
    ecosystem: go
    dependencies:
      - package: core
        strategy: linked
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	// Load using viper
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	// Validate loaded config
	assert.Len(t, cfg.Packages, 2)
	assert.Equal(t, "core", cfg.Packages[0].Name)
	assert.Equal(t, "api", cfg.Packages[1].Name)
	assert.Len(t, cfg.Packages[1].Dependencies, 1)
	assert.Equal(t, "core", cfg.Packages[1].Dependencies[0].Package)
}

func TestConfigLoad_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.json")
	
	jsonContent := `{
  "packages": [
    {
      "name": "core",
      "path": "./core",
      "ecosystem": "go"
    }
  ]
}`
	
	err := os.WriteFile(configPath, []byte(jsonContent), 0644)
	require.NoError(t, err)
	
	// Load using viper
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	assert.Len(t, cfg.Packages, 1)
	assert.Equal(t, "core", cfg.Packages[0].Name)
}

func TestConfigLoad_TOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.toml")
	
	tomlContent := `
[[packages]]
name = "core"
path = "./core"
ecosystem = "go"
`
	
	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
	require.NoError(t, err)
	
	// Load using viper
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	assert.Len(t, cfg.Packages, 1)
	assert.Equal(t, "core", cfg.Packages[0].Name)
}

func TestConfigLoad_WithTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.yaml")
	
	yamlContent := `
packages:
  - name: core
    path: ./core

templates:
  changelog:
    source: "builtin:default"
  tagName:
    inline: "v{{.Version}}"
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	assert.NotNil(t, cfg.Templates.Changelog)
	assert.Equal(t, "builtin:default", cfg.Templates.Changelog.Source)
	assert.NotNil(t, cfg.Templates.TagName)
	assert.Equal(t, "v{{.Version}}", cfg.Templates.TagName.Inline)
}

func TestConfigLoad_WithMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.yaml")
	
	yamlContent := `
packages:
  - name: core
    path: ./core

metadata:
  fields:
    - name: author
      required: true
      type: string
    - name: issue
      required: false
      type: string
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	require.NoError(t, err)
	
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	assert.Len(t, cfg.Metadata.Fields, 2)
	assert.Equal(t, "author", cfg.Metadata.Fields[0].Name)
	assert.True(t, cfg.Metadata.Fields[0].Required)
}

func TestConfigLoad_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "shipyard.yaml")

	err := fileutil.WriteYAMLFile(configPath, map[string]interface{}{
		"packages": []map[string]interface{}{
			{"name": "core", "path": "./core"},
		},
	}, 0644)
	require.NoError(t, err)
	
	v := viper.New()
	v.SetConfigFile(configPath)
	err = v.ReadInConfig()
	require.NoError(t, err)
	
	var cfg config.Config
	err = v.Unmarshal(&cfg)
	require.NoError(t, err)
	
	err = cfg.Validate()
	assert.NoError(t, err)
}
