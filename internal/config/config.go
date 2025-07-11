// Package config provides internal configuration management for Shipyard.
// This package handles loading, saving, and managing configuration files.
package config

import (
	"errors"

	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/spf13/viper"
)

var (
	// AppConfig holds the application-wide configuration
	AppConfig *viper.Viper = viper.New()
	// projectConfig holds the configuration for a specific project
	projectConfig *viper.Viper = viper.New()
)

// Re-export public types for backward compatibility
type RepoType = config.RepoType
type ProjectConfig = config.ProjectConfig
type Package = config.Package
type PackageEcosystem = config.PackageEcosystem
type ChangelogConfig = config.ChangelogConfig

// Re-export public constants for backward compatibility
const (
	RepositoryTypeMonorepo   = config.RepositoryTypeMonorepo
	RepositoryTypeSingleRepo = config.RepositoryTypeSingleRepo
)

func init() {
	// Set the default project configuration in the app configuration
	AppConfig.SetDefault("config", ".shipyard/config.yaml")
	AppConfig.SetDefault("log.level", "info")
	AppConfig.SetDefault("log.file", ".shipyard/logs/shipyard.log")
	AppConfig.SetDefault("verbose", false)
	AppConfig.AutomaticEnv() // read environment variables that match the config keys
	// Set the default project configuration
	projectConfig.SetDefault("extends", nil)
	projectConfig.SetDefault("type", RepositoryTypeMonorepo)
	projectConfig.SetDefault("repo", "github.com/example/example-repo")
	projectConfig.SetDefault("changelog.template", "keepachangelog")
}

// LoadProjectConfig loads the project configuration from the configured path
func LoadProjectConfig() (*ProjectConfig, error) {
	// Get the config path from app configuration
	configPath := AppConfig.GetString("config")

	// Use the public API to load the configuration
	return config.LoadFromFile(configPath)
}

// InitProjectConfig initializes a new project configuration
func InitProjectConfig(configData map[string]interface{}) error {
	// Create a temporary viper instance to convert the map to a proper config
	v := viper.New()
	v.MergeConfigMap(configData)

	var projectConfig config.ProjectConfig
	if err := v.Unmarshal(&projectConfig); err != nil {
		logger.Error("Failed to unmarshal project config", "error", err)
		return errors.New("failed to unmarshal project config: " + err.Error())
	}

	// Use the public API to save the configuration
	configPath := AppConfig.GetString("config")
	if err := config.SaveToFile(&projectConfig, configPath); err != nil {
		logger.Error("Failed to save project config", "error", err)
		return errors.New("failed to save project config: " + err.Error())
	}

	return nil
}
