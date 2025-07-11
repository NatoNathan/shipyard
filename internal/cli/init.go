package cli

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Shipyard project",
	Long:  "This command sets up a new Shipyard project with the necessary configuration files and directories.",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			repoType          config.RepoType
			repoPath          string
			changelogTemplate string
			packages          []PackageConfig
			addAnother        bool = true
		)

		// Initial form for repository configuration
		repoForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("`Shipyard Init`").
					Description("Welcome to Shipyard! Let's set up your project."),
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
				huh.NewNote().
					Title("Package Configuration").
					Description("You can configure multiple packages for a monorepo or a single package for a single repo. Shipyard will help you manage them."),
			),
		)

		if err := repoForm.Run(); err != nil {
			logger.Error("Failed to initialize project", "error", err)
			return
		}

		// Configure packages based on repository type
		if repoType == config.RepositoryTypeMonorepo {
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
		result = append(result, config.Package{
			Name:      pkg.Name,
			Path:      pkg.Path,
			Ecosystem: pkg.Ecosystem,
			Manifest:  pkg.Manifest,
		})
	}
	return result
}

// convertPackageForConfig converts a single PackageConfig to config.Package
func convertPackageForConfig(pkg PackageConfig) config.Package {
	return config.Package{
		Name:      pkg.Name,
		Path:      pkg.Path,
		Ecosystem: pkg.Ecosystem,
		Manifest:  pkg.Manifest,
	}
}
