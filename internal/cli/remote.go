package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/logger"
	pkgconfig "github.com/NatoNathan/shipyard/pkg/config"
	"github.com/spf13/cobra"
)

var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote configuration sources",
	Long: `Manage remote configuration sources for Shipyard.

Remote configurations allow teams to extend a shared configuration from:
- HTTP/HTTPS URLs: https://example.com/config.yaml
- GitHub repositories: github:owner/repo/path/to/config.yaml[@ref] (supports SSH and HTTPS)
- Git repositories: git+https://github.com/owner/repo.git/path/to/config.yaml[@ref] (HTTPS)
- Git repositories: git+git@github.com:owner/repo.git/path/to/config.yaml[@ref] (SSH)

Authentication is supported for private repositories via SSH keys, tokens, and git credentials.
Remote configurations are cached locally for performance and can be refreshed as needed.`,
}

var remoteFetchCmd = &cobra.Command{
	Use:   "fetch <remote-url>",
	Short: "Fetch and display a remote configuration",
	Long: `Fetch a remote configuration and display its contents.

Examples:
  # Fetch from HTTP URL
  shipyard remote fetch https://example.com/shipyard/config.yaml

  # Fetch from GitHub repository (main branch) - tries SSH first, then HTTPS
  shipyard remote fetch github:myorg/shared-configs/shipyard.yaml

  # Fetch from GitHub repository (specific branch)
  shipyard remote fetch github:myorg/shared-configs/shipyard.yaml@develop

  # Fetch from Git repository (HTTPS)
  shipyard remote fetch git+https://github.com/myorg/configs.git/shipyard/config.yaml
  
  # Fetch from Git repository (SSH)
  shipyard remote fetch git+git@github.com:myorg/configs.git/shipyard/config.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteURL := args[0]

		forceFresh, _ := cmd.Flags().GetBool("fresh")

		logger.Info("Fetching remote configuration", "url", remoteURL, "fresh", forceFresh)

		config, err := pkgconfig.LoadRemoteConfig(remoteURL, forceFresh)
		if err != nil {
			logger.Error("Failed to fetch remote configuration", "error", err)
			fmt.Printf("Error: Failed to fetch remote configuration: %v\n", err)
			os.Exit(1)
		}

		// Display the configuration
		fmt.Printf("Remote Configuration from: %s\n", remoteURL)
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("Type: %s\n", config.Type)
		fmt.Printf("Repository: %s\n", config.Repo)
		fmt.Printf("Changelog Template: %s\n", config.Changelog.Template)

		if config.Type == pkgconfig.RepositoryTypeMonorepo {
			fmt.Printf("Packages (%d):\n", len(config.Packages))
			for _, pkg := range config.Packages {
				fmt.Printf("  - %s (%s) at %s\n", pkg.Name, pkg.Ecosystem, pkg.Path)
			}
		} else {
			fmt.Printf("Package: %s (%s) at %s\n", config.Package.Name, config.Package.Ecosystem, config.Package.Path)
		}

		if len(config.ChangeTypes) > 0 {
			fmt.Printf("Change Types (%d):\n", len(config.ChangeTypes))
			for _, ct := range config.ChangeTypes {
				fmt.Printf("  - %s: %s (%s)\n", ct.Name, ct.DisplayName, ct.SemverBump)
			}
		}

		logger.Info("Remote configuration fetched successfully")
	},
}

var remoteCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage remote configuration cache",
	Long:  "Manage the local cache for remote configurations",
}

var remoteCacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached remote configurations",
	Long:  "List all cached remote configurations with their metadata",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Listing cached remote configurations")

		cached, err := pkgconfig.ListCachedRemoteConfigs()
		if err != nil {
			logger.Error("Failed to list cached configurations", "error", err)
			fmt.Printf("Error: Failed to list cached configurations: %v\n", err)
			os.Exit(1)
		}

		if len(cached) == 0 {
			fmt.Println("No cached remote configurations found.")
			return
		}

		fmt.Printf("Cached Remote Configurations (%d):\n", len(cached))
		fmt.Println("========================================")

		for i, cache := range cached {
			fmt.Printf("%d. URL: %s\n", i+1, cache.URL)
			fmt.Printf("   Hash: %s\n", cache.Hash[:12]+"...")
			fmt.Printf("   Last Fetched: %s\n", cache.LastFetched.Format(time.RFC3339))
			fmt.Printf("   TTL: %d minutes\n", cache.TTL)

			// Check if expired
			if cache.TTL > 0 {
				expireTime := cache.LastFetched.Add(time.Duration(cache.TTL) * time.Minute)
				if time.Now().After(expireTime) {
					fmt.Printf("   Status: EXPIRED\n")
				} else {
					remaining := time.Until(expireTime)
					fmt.Printf("   Status: Valid (expires in %s)\n", remaining.Truncate(time.Minute))
				}
			} else {
				fmt.Printf("   Status: No expiry\n")
			}
			fmt.Println()
		}

		logger.Info("Listed cached remote configurations", "count", len(cached))
	},
}

var remoteCacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the remote configuration cache",
	Long:  "Clear all cached remote configurations",
	Run: func(cmd *cobra.Command, args []string) {
		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Print("Are you sure you want to clear all cached remote configurations? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		logger.Info("Clearing remote configuration cache")

		err := pkgconfig.ClearRemoteConfigCache()
		if err != nil {
			logger.Error("Failed to clear cache", "error", err)
			fmt.Printf("Error: Failed to clear cache: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Remote configuration cache cleared successfully.")
		logger.Info("Remote configuration cache cleared")
	},
}

var remoteValidateCmd = &cobra.Command{
	Use:   "validate <remote-url>",
	Short: "Validate a remote configuration URL",
	Long: `Validate that a remote configuration URL is accessible and contains valid configuration.

This command checks:
- URL format validity
- Remote accessibility
- Configuration file validity
- Schema compliance`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteURL := args[0]

		logger.Info("Validating remote configuration", "url", remoteURL)

		// Check URL format
		if !pkgconfig.IsValidRemoteURL(remoteURL) {
			fmt.Printf("Error: Invalid remote URL format: %s\n", remoteURL)
			fmt.Println("Supported formats:")
			fmt.Println("  - HTTP/HTTPS: https://example.com/config.yaml")
			fmt.Println("  - GitHub: github:owner/repo/path/to/config.yaml[@ref]")
			fmt.Println("  - Git: git+https://github.com/owner/repo.git/path/to/config.yaml[@ref]")
			os.Exit(1)
		}

		fmt.Printf("✓ URL format is valid: %s\n", remoteURL)

		// Try to fetch the configuration
		config, err := pkgconfig.LoadRemoteConfig(remoteURL, true) // Force fresh to test actual accessibility
		if err != nil {
			logger.Error("Failed to fetch remote configuration", "error", err)
			fmt.Printf("✗ Failed to fetch configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Remote configuration is accessible\n")

		// Validate the configuration
		if err := validateRemoteConfigForCLI(config); err != nil {
			logger.Error("Remote configuration is invalid", "error", err)
			fmt.Printf("✗ Configuration validation failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Configuration is valid\n")

		// Show basic info
		fmt.Printf("\nConfiguration Summary:\n")
		fmt.Printf("  Type: %s\n", config.Type)
		fmt.Printf("  Repository: %s\n", config.Repo)

		if config.Type == pkgconfig.RepositoryTypeMonorepo {
			fmt.Printf("  Packages: %d\n", len(config.Packages))
		} else {
			fmt.Printf("  Package: %s\n", config.Package.Name)
		}

		fmt.Printf("\n✓ Remote configuration is valid and ready to use!\n")
		logger.Info("Remote configuration validation successful")
	},
}

var remoteTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage remote templates",
	Long:  "Manage remote template files used in configurations",
}

var remoteTemplateFetchCmd = &cobra.Command{
	Use:   "fetch <template-url>",
	Short: "Fetch and display a remote template",
	Long: `Fetch a remote template file and display its contents.

Examples:
  # Fetch from HTTP URL
  shipyard remote template fetch https://example.com/templates/changelog.md

  # Fetch from GitHub repository (tries SSH first, then HTTPS)
  shipyard remote template fetch github:myorg/configs/templates/changelog.md

  # Fetch from Git repository (SSH)
  shipyard remote template fetch git+git@github.com:myorg/configs.git/templates/changelog.md`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateURL := args[0]

		forceFresh, _ := cmd.Flags().GetBool("fresh")

		logger.Info("Fetching remote template", "url", templateURL, "fresh", forceFresh)

		content, err := pkgconfig.LoadRemoteTemplate(templateURL, forceFresh)
		if err != nil {
			logger.Error("Failed to fetch remote template", "error", err)
			fmt.Printf("Error: Failed to fetch remote template: %v\n", err)
			os.Exit(1)
		}

		// Display the template content
		fmt.Printf("Remote Template from: %s\n", templateURL)
		fmt.Printf("========================================\n")
		fmt.Printf("%s\n", content)

		logger.Info("Remote template fetched successfully")
	},
}

func init() {
	// Add subcommands to remote command
	RemoteCmd.AddCommand(remoteFetchCmd)
	RemoteCmd.AddCommand(remoteCacheCmd)
	RemoteCmd.AddCommand(remoteValidateCmd)
	RemoteCmd.AddCommand(remoteTemplateCmd)

	// Add cache subcommands
	remoteCacheCmd.AddCommand(remoteCacheListCmd)
	remoteCacheCmd.AddCommand(remoteCacheClearCmd)

	// Add template subcommands
	remoteTemplateCmd.AddCommand(remoteTemplateFetchCmd)

	// Add flags
	remoteFetchCmd.Flags().BoolP("fresh", "f", false, "Force fresh fetch (ignore cache)")
	remoteTemplateFetchCmd.Flags().BoolP("fresh", "f", false, "Force fresh fetch (ignore cache)")
	remoteCacheClearCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

// validateRemoteConfigForCLI validates a remote config for CLI display
// Uses more lenient validation than local configs
func validateRemoteConfigForCLI(config *pkgconfig.ProjectConfig) error {
	// Basic validation - remote configs don't need to have packages
	if config.Type == "" {
		return fmt.Errorf("type: repository type is required")
	}

	if config.Type != pkgconfig.RepositoryTypeMonorepo && config.Type != pkgconfig.RepositoryTypeSingleRepo {
		return fmt.Errorf("type: repository type must be 'monorepo' or 'single-repo'")
	}

	// Remote configs should NOT contain packages - they should be base configs
	if config.Type == pkgconfig.RepositoryTypeMonorepo && len(config.Packages) > 0 {
		return fmt.Errorf("packages: remote base configurations should not contain packages - packages should be defined in the extending project")
	}

	if config.Type == pkgconfig.RepositoryTypeSingleRepo && config.Package.Name != "" {
		return fmt.Errorf("package: remote base configurations should not contain package definition - package should be defined in the extending project")
	}

	// Don't require repo URL in remote configs as they're meant to be extended

	return nil
}
