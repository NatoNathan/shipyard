// setViperConfigFromPath sets the config name, type, and path for a viper instance based on a full file path
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/spf13/viper"
)

var (
	// AppConfig holds the application-wide configuration
	AppConfig *viper.Viper = viper.New()
	// projectConfig holds the configuration for a specific project
	projectConfig *viper.Viper = viper.New()
)

type RepoType string

// enum RepositoryType
const (
	RepositoryTypeMonorepo   RepoType = "monorepo"
	RepositoryTypeSingleRepo RepoType = "single-repo"
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

type ProjectConfig struct {
	Type RepoType `mapstructure:"type"` // "monorepo" or "single-repo"
	Repo string   `mapstructure:"repo"` // e.g., "github.com/NatoNathan/shipyard"

	Changelog struct {
		Template       string `mapstructure:"template"`
		HeaderTemplate string `mapstructure:"header_template"`
		FooterTemplate string `mapstructure:"footer_template"`
	} `mapstructure:"changelog"`

	Packages []Package                `mapstructure:"packages"`
	Package  `mapstructure:"package"` // for single-repo projects
}

type Package struct {
	Name      string `mapstructure:"name"`      // e.g., "api", "frontend"
	Path      string `mapstructure:"path"`      // e.g., "packages/api", "packages/frontend"
	Manifest  string `mapstructure:"manifest"`  // e.g., "packages/api/package.json"
	Ecosystem string `mapstructure:"ecosystem"` // e.g., "npm", "go", "python"
}

func LoadProjectConfig() (*ProjectConfig, error) {
	// load the configuration file from the path set in AppConfig
	configPath := AppConfig.GetString("config")
	setViperConfigFromPath(projectConfig, configPath)
	projectConfig.AutomaticEnv() // read environment variables that match the config keys

	if err := projectConfig.ReadInConfig(); err != nil {
		return nil, err
	}

	if extends := projectConfig.GetString("extends"); extends != "" {
		baseConfig, err := fetchBaseConfig(extends)
		if err != nil {
			return nil, err
		}
		projectConfig.MergeConfigMap(baseConfig.AllSettings())
	}

	var config ProjectConfig
	if err := projectConfig.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func InitProjectConfig(config map[string]interface{}) error {
	projectConfig.MergeConfigMap(config)
	configPath := AppConfig.GetString("config")
	setViperConfigFromPath(projectConfig, configPath)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create config directory", "error", err)
		return errors.New("failed to create config directory: " + err.Error())
	}

	if err := projectConfig.SafeWriteConfig(); err != nil {
		logger.Error("Failed to write project config", "error", err)
		return errors.New("failed to write project config: " + err.Error())
	}
	return nil
}

// Fetches the base configuration from the specified path
// if github:<owner>/<repo>/(<path>)(@<ref>) is specified, it fetches the config from the remote repository
// if https://<url> is specified, it fetches the config from the remote URL
// if <path> is specified, it fetches the config from the local file system
func fetchBaseConfig(extends string) (*viper.Viper, error) {
	if extends == "" {
		return nil, errors.New("no base config to fetch")
	}
	if strings.HasPrefix(extends, "github:") {
		// Note: We can't use logger here as it might not be initialized yet
		panic("github: not implemented yet")
	} else if strings.HasPrefix(extends, "https://") {
		// Note: We can't use logger here as it might not be initialized yet
		panic("https:// not implemented yet")
	}
	// local file system
	v := viper.New()
	setViperConfigFromPath(v, extends)
	if _, err := os.Stat(extends); os.IsNotExist(err) {
		return nil, errors.New("base config file does not exist: " + extends)
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

func setViperConfigFromPath(v *viper.Viper, path string) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	ext = strings.TrimPrefix(ext, ".")

	v.SetConfigName(name)
	v.SetConfigType(ext)
	v.AddConfigPath(dir)
}
