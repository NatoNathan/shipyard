package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/fileutil"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Load loads the configuration from a file using Viper
func Load(configPath string) (*Config, error) {
	v := viper.New()
	
	// Set config file
	v.SetConfigFile(configPath)
	
	// Read config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Apply defaults
	result := cfg.WithDefaults()
	
	// Validate
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return result, nil
}

// LoadFromDir loads the configuration from a directory
// It looks for shipyard.yaml, shipyard.yml, shipyard.json, or shipyard.toml
// First checks .shipyard/ subdirectory, then the root directory
func LoadFromDir(dir string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("shipyard")
	// Check .shipyard/ subdirectory first (standard location)
	v.AddConfigPath(filepath.Join(dir, ".shipyard"))
	// Also check root directory (alternative location)
	v.AddConfigPath(dir)
	v.SetConfigType("yaml") // Will auto-detect format
	
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", dir, err)
	}
	
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	result := cfg.WithDefaults()
	
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return result, nil
}

// FindConfig searches for a shipyard config file in the current directory
// and parent directories up to the repository root
func FindConfig(startDir string) (string, error) {
	dir := startDir
	
	// Possible config filenames
	names := []string{
		"shipyard.yaml",
		"shipyard.yml",
		"shipyard.json",
		"shipyard.toml",
	}
	
	// Search up to root
	for {
		for _, name := range names {
			configPath := filepath.Join(dir, ".shipyard", name)
			if fileExists(configPath) {
				return configPath, nil
			}
		}
		
		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf("shipyard config not found in %s or parent directories", startDir)
}

func fileExists(path string) bool {
	return fileutil.PathExists(path)
}

// WriteConfig writes the configuration to a YAML file
func WriteConfig(cfg *Config, configPath string) error {
	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write atomically
	if err := fileutil.AtomicWrite(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
