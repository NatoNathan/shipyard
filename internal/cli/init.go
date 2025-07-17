package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/git"
	"github.com/NatoNathan/shipyard/pkg/handlers"
	"github.com/charmbracelet/huh"
	gitPkg "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Shipyard project",
	Long:  "This command sets up a new Shipyard project with the necessary configuration files and directories. It can automatically detect your repository URL and scan for packages.",
	Run: func(cmd *cobra.Command, args []string) {
		skipAutoScan, _ := cmd.Flags().GetBool("skip-auto-scan")

		var (
			repoType          config.RepoType
			repoPath          string
			changelogTemplate string
			packages          []PackageConfig
			useAutoScan       bool = !skipAutoScan
		)

		// Welcome screen - only show if auto-scan wasn't skipped via flag
		if !skipAutoScan {
			welcomeForm := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("`Shipyard Init`").
						Description("Welcome to Shipyard! Let's set up your project with automatic detection."),
					huh.NewConfirm().
						Title("Auto-scan for packages and repository?").
						Description("Shipyard can automatically detect your repository URL and scan for packages. Would you like to use auto-detection?").
						Value(&useAutoScan),
				),
			)

			if err := welcomeForm.Run(); err != nil {
				logger.Error("Failed to get user input", "error", err)
				return
			}
		} else {
			// If auto-scan was skipped via flag, show a simple welcome
			fmt.Println("🚢 Shipyard Init")
			fmt.Println("Setting up your project with manual configuration...")
		}

		// Auto-detection phase
		if useAutoScan {
			logger.Info("Auto-detecting repository and package information...")

			// Try to detect repository info
			if detectedRepo, err := autoDetectRepoInfo(); err == nil {
				repoPath = detectedRepo
				logger.Info("Detected repository", "repo", repoPath)
			} else {
				logger.Warn("Could not auto-detect repository", "error", err.Error())
			}

			// Try to scan for packages
			if detectedPackages, err := autoScanPackages(); err == nil && len(detectedPackages) > 0 {
				logger.Info("Found packages", "count", len(detectedPackages))
				for _, pkg := range detectedPackages {
					packages = append(packages, *pkg)
					logger.Info("Found package", "name", pkg.Name, "ecosystem", pkg.Ecosystem, "path", pkg.Path)
				}

				// Determine repository type based on detected packages
				repoType = determineRepoType(detectedPackages)
				logger.Info("Determined repository type", "type", repoType)
			} else {
				if err != nil {
					logger.Warn("Could not auto-scan packages", "error", err.Error())
				} else {
					logger.Info("No packages found automatically")
				}
			}
		}

		// Repository configuration form - with pre-filled values if detected
		repoForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Repository Configuration").
					Description("Please review and update the repository configuration."),
				huh.NewInput().
					Title("Repository Path").
					Value(&repoPath).
					Placeholder("e.g., github.com/your-org/your-repo").
					Description("Enter the path to your repository. This can be a GitHub, GitLab, or Bitbucket URL."),
				huh.NewSelect[config.RepoType]().
					Value(&repoType).
					Title("Select Repository Type").
					Description("Choose the type of repository for your project.").
					Options(
						huh.NewOption("Monorepo", config.RepositoryTypeMonorepo),
						huh.NewOption("Single Repo", config.RepositoryTypeSingleRepo),
					),
				huh.NewInput().
					Title("Changelog Template").
					Value(&changelogTemplate).
					Placeholder("e.g., keepachangelog").
					Description("Enter the template for your changelog. This can be a predefined template like 'keepachangelog' or a custom one."),
			),
		)

		if err := repoForm.Run(); err != nil {
			logger.Error("Failed to initialize project", "error", err)
			return
		}

		// Package configuration
		if len(packages) == 0 || !useAutoScan {
			// Manual package configuration
			if repoType == config.RepositoryTypeMonorepo {
				// TODO: add an option to scan for supported package ecosystems
				var addAnother bool = true
				// Loop to add packages for monorepo
				for addAnother {
					var pkgConfig PackageConfig

					// Package configuration form
					packageForm := huh.NewForm(
						packageGroup(&pkgConfig),
					)

					if err := packageForm.Run(); err != nil {
						logger.Error("Failed to configure package", "error", err)
						return
					}

					packages = append(packages, pkgConfig)

					// Ask if user wants to add another package
					continueForm := huh.NewForm(
						huh.NewGroup(
							huh.NewConfirm().
								Title("Add Another Package?").
								Description("Would you like to add another package to your monorepo?").
								Value(&addAnother),
						),
					)

					if err := continueForm.Run(); err != nil {
						logger.Error("Failed to get user input", "error", err)
						return
					}
				}
			} else {
				// Single repo - configure only one package
				var pkgConfig PackageConfig

				packageForm := huh.NewForm(
					packageGroup(&pkgConfig),
				)

				if err := packageForm.Run(); err != nil {
					logger.Error("Failed to configure package", "error", err)
					return
				}

				packages = append(packages, pkgConfig)
			}
		} else {
			// Review auto-detected packages
			var confirmPackages bool
			var packagesDesc strings.Builder
			packagesDesc.WriteString("The following packages were detected:\n")
			for _, pkg := range packages {
				packagesDesc.WriteString(fmt.Sprintf("• %s (%s) at %s\n", pkg.Name, pkg.Ecosystem, pkg.Path))
			}
			packagesDesc.WriteString("\nWould you like to use these packages?")

			reviewForm := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().
						Title("Package Review").
						Description(packagesDesc.String()),
					huh.NewConfirm().
						Title("Use detected packages?").
						Description("Accept the auto-detected packages or configure manually?").
						Value(&confirmPackages),
				),
			)

			if err := reviewForm.Run(); err != nil {
				logger.Error("Failed to get user input", "error", err)
				return
			}

			if !confirmPackages {
				// User wants to configure manually
				packages = []PackageConfig{}
				if repoType == config.RepositoryTypeMonorepo {
					var addAnother bool = true
					// Loop to add packages for monorepo
					for addAnother {
						var pkgConfig PackageConfig

						packageForm := huh.NewForm(
							packageGroup(&pkgConfig),
						)

						if err := packageForm.Run(); err != nil {
							logger.Error("Failed to configure package", "error", err)
							return
						}

						packages = append(packages, pkgConfig)

						continueForm := huh.NewForm(
							huh.NewGroup(
								huh.NewConfirm().
									Title("Add Another Package?").
									Description("Would you like to add another package to your monorepo?").
									Value(&addAnother),
							),
						)

						if err := continueForm.Run(); err != nil {
							logger.Error("Failed to get user input", "error", err)
							return
						}
					}
				} else {
					// Single repo - configure only one package
					var pkgConfig PackageConfig

					packageForm := huh.NewForm(
						packageGroup(&pkgConfig),
					)

					if err := packageForm.Run(); err != nil {
						logger.Error("Failed to configure package", "error", err)
						return
					}

					packages = append(packages, pkgConfig)
				}
			}
		}

		// Build the configuration map
		configMap := map[string]interface{}{
			"type": repoType,
			"repo": repoPath,
			"changelog": map[string]interface{}{
				"template": changelogTemplate,
			},
		}

		// Add packages based on repository type
		if repoType == config.RepositoryTypeMonorepo {
			configMap["packages"] = convertPackagesForConfig(packages)
		} else {
			// For single repo, use the package structure
			configMap["package"] = convertPackageForConfig(packages[0])
		}

		// Initialize the project configuration
		if err := config.InitProjectConfig(configMap); err != nil {
			logger.Error("Failed to initialize project configuration", "error", err)
			return
		}

		logger.Info("Successfully initialized Shipyard project", "type", repoType, "repo", repoPath)
		logger.Info("Configuration saved to", "path", config.AppConfig.GetString("config"))

		// Provide next steps to the user
		fmt.Println("\n🎉 Shipyard project initialized successfully!")
		fmt.Printf("📁 Configuration saved to: %s\n", config.AppConfig.GetString("config"))
		fmt.Printf("📝 Repository type: %s\n", repoType)
		fmt.Printf("🔗 Repository: %s\n", repoPath)
		fmt.Printf("📋 Changelog template: %s\n", changelogTemplate)

		if repoType == config.RepositoryTypeMonorepo {
			fmt.Printf("📦 Packages configured: %d\n", len(packages))
			for _, pkg := range packages {
				fmt.Printf("   - %s (%s) at %s\n", pkg.Name, pkg.Ecosystem, pkg.Path)
			}
		} else {
			fmt.Printf("📦 Package: %s (%s) at %s\n", packages[0].Name, packages[0].Ecosystem, packages[0].Path)
		}

		fmt.Println("\n🚀 Next steps:")
		fmt.Println("   - Review the generated configuration file")
		fmt.Println("   - Run 'shipyard --help' to see available commands")
		fmt.Println("   - Start managing your releases with Shipyard!")
	},
}

// autoDetectRepoInfo attempts to detect repository information from git
func autoDetectRepoInfo() (string, error) {
	// Check if we're in a git repository
	if !git.IsGitRepository(".") {
		return "", fmt.Errorf("not in a git repository")
	}

	// Open git repository directly using go-git
	repo, err := gitPkg.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get remote origin URL
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("no origin remote found: %w", err)
	}

	// Get the remote URL and clean it up
	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("no remote URLs found")
	}

	remoteURL := urls[0]

	// Convert SSH URLs to HTTPS format and extract repo path
	repoPath := extractRepoPath(remoteURL)
	if repoPath == "" {
		return "", fmt.Errorf("could not extract repository path from URL: %s", remoteURL)
	}

	return repoPath, nil
}

// extractRepoPath extracts the repository path from various URL formats
func extractRepoPath(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) >= 2 {
			hostAndPath := strings.Split(parts[0], "@")
			if len(hostAndPath) >= 2 {
				host := hostAndPath[1]
				path := strings.Join(parts[1:], ":")
				return host + "/" + path
			}
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
		return url
	}

	// Handle HTTP format: http://github.com/owner/repo
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
		return url
	}

	return url
}

// autoScanPackages scans the current directory for packages
func autoScanPackages() ([]*PackageConfig, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Scan for packages using the handlers
	packages, err := handlers.ScanForPackages(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for packages: %w", err)
	}

	// Convert to our internal package config type
	var packageConfigs []*PackageConfig
	for _, pkg := range packages {
		// Calculate relative path from current directory
		relPath, err := filepath.Rel(cwd, pkg.Path)
		if err != nil {
			relPath = pkg.Path
		}
		// Use "." for root directory packages
		if relPath == "" {
			relPath = "."
		}

		packageConfigs = append(packageConfigs, &PackageConfig{
			Name:      pkg.Name,
			Path:      relPath,
			Ecosystem: string(pkg.Ecosystem),
			Manifest:  pkg.Manifest,
		})
	}

	return packageConfigs, nil
}

// determineRepoType determines whether this should be a monorepo or single repo
func determineRepoType(packages []*PackageConfig) config.RepoType {
	if len(packages) > 1 {
		return config.RepositoryTypeMonorepo
	}

	// Check if single package is in root directory
	if len(packages) == 1 && packages[0].Path == "." {
		return config.RepositoryTypeSingleRepo
	}

	// Default to monorepo for packages in subdirectories
	return config.RepositoryTypeMonorepo
}

type PackageConfig struct {
	Name      string
	Path      string
	Ecosystem string
	Manifest  string
}

func packageGroup(pkg *PackageConfig) *huh.Group {
	return huh.NewGroup(
		huh.NewNote().
			Title("Package Configuration").
			Description("Configure your package settings."),
		huh.NewInput().
			Title("Package Name").
			Value(&pkg.Name).
			Placeholder("e.g., api, frontend").
			Description("Enter the name of your package."),
		huh.NewInput().
			Title("Package Path").
			Value(&pkg.Path).
			Placeholder("e.g., packages/api, packages/frontend").
			Description("Enter the path to your package."),
		huh.NewSelect[string]().
			Value(&pkg.Ecosystem).
			Title("Ecosystem").
			Description("Select the ecosystem for your package.").
			Options(
				huh.NewOption("NPM", "npm"),
				huh.NewOption("Go", "go"),
				huh.NewOption("Helm", "helm"),
				huh.NewOption("Python", "python"),
				huh.NewOption("Docker", "docker"),
			),
		huh.NewInput().
			Title("Manifest Path").
			Value(&pkg.Manifest).
			Placeholder("e.g., packages/api/package.json, packages/frontend/Chart.yaml").
			Description("Enter the path to your package manifest"),
	)
}

// convertPackagesForConfig converts PackageConfig slice to config.Package slice
func convertPackagesForConfig(packages []PackageConfig) []config.Package {
	var result []config.Package
	for _, pkg := range packages {
		result = append(result, convertPackageForConfig(pkg))
	}
	return result
}

func init() {
	InitCmd.Flags().Bool("skip-auto-scan", false, "Skip automatic repository and package detection")
}

// convertPackageForConfig converts a single PackageConfig to config.Package
func convertPackageForConfig(pkg PackageConfig) config.Package {
	return config.Package{
		Name:      pkg.Name,
		Path:      pkg.Path,
		Ecosystem: config.PackageEcosystem(pkg.Ecosystem),
		Manifest:  pkg.Manifest,
	}
}
