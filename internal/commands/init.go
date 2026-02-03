package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/detect"
	shipyarderrors "github.com/NatoNathan/shipyard/internal/errors"
	"github.com/NatoNathan/shipyard/internal/fileutil"
	"github.com/NatoNathan/shipyard/internal/git"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/internal/prompt"
	"github.com/NatoNathan/shipyard/internal/ui"
	"github.com/spf13/cobra"
)

// InitOptions contains options for the init command
type InitOptions struct {
	Force  bool
	Remote string
	Yes    bool  // Skip prompts and use defaults
	JSON   bool  // Output in JSON format
	Quiet  bool  // Suppress output
}

// InitCmd creates the init command
func InitCmd() *cobra.Command {
	var force bool
	var remote string
	var yes bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set sail - prepare your repository",
		Long: `Prepare your repository for the versioning voyage ahead. Sets up the shipyard
with cargo manifests, navigation charts, and the captain's log.

In interactive mode, you'll configure your fleet (packages) and choose between
sailing solo or commanding a flotilla (monorepo). Use --yes to set sail with
default configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// Extract global flags
			globalFlags := GetGlobalFlags(cmd)

			return runInit(cwd, InitOptions{
				Force:  force,
				Remote: remote,
				Yes:    yes,
				JSON:   globalFlags.JSON,
				Quiet:  globalFlags.Quiet,
			})
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "force re-initialization if already initialized")
	cmd.Flags().StringVarP(&remote, "remote", "r", "", "remote configuration URL to extend from")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip all prompts and accept defaults")

	return cmd
}

// runInit performs the actual initialization logic
func runInit(projectPath string, options InitOptions) error {
	log := logger.Get()

	// Step 1: Verify git repository
	log.Info("Verifying git repository...")
	isGitRepo, err := git.IsRepository(projectPath)
	if err != nil {
		return shipyarderrors.NewGitError("failed to check git repository", err)
	}
	if !isGitRepo {
		return shipyarderrors.NewGitError("not a git repository", nil)
	}

	// Step 2: Check for existing configuration
	shipyardDir := filepath.Join(projectPath, ".shipyard")
	configPath := filepath.Join(shipyardDir, "shipyard.yaml")

	if fileutil.PathExists(configPath) && !options.Force {
		return shipyarderrors.NewConfigError("shipyard already initialized (use --force to reinitialize)", nil)
	}

	log.Info("Initializing Shipyard...")

	// Step 3: Create directory structure
	if err := initializeDirectories(projectPath); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Step 4: Generate configuration
	cfg, err := generateConfiguration(projectPath, options)
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Step 5: Write configuration file
	if err := config.WriteConfig(cfg, configPath); err != nil {
		return shipyarderrors.NewConfigError("failed to write configuration", err)
	}

	// Step 6: Initialize history file
	historyPath := filepath.Join(shipyardDir, "history.json")
	if err := initializeHistoryFile(historyPath); err != nil {
		return fmt.Errorf("failed to initialize history file: %w", err)
	}

	// Output based on format flags
	if options.JSON {
		// JSON output
		jsonData := map[string]interface{}{
			"success":              true,
			"configPath":           configPath,
			"consignmentsDir":      filepath.Join(shipyardDir, "consignments"),
			"historyFile":          historyPath,
			"initialized":          true,
		}
		return PrintJSON(os.Stdout, jsonData)
	}

	if !options.Quiet {
		// Print success message with styled output
		fmt.Println()
		fmt.Println(ui.SuccessMessage("Shipyard initialized successfully"))
		fmt.Println()
		fmt.Println(ui.KeyValue("Configuration", configPath))
		fmt.Println(ui.KeyValue("Consignments directory", filepath.Join(shipyardDir, "consignments")))
		fmt.Println(ui.KeyValue("History file", historyPath))
		fmt.Println()
	}

	return nil
}

// initializeDirectories creates the required directory structure
func initializeDirectories(projectPath string) error {
	shipyardDir := filepath.Join(projectPath, ".shipyard")
	consignmentsDir := filepath.Join(shipyardDir, "consignments")

	// Create .shipyard directory
	if err := fileutil.EnsureDir(shipyardDir); err != nil {
		return fmt.Errorf("failed to create .shipyard directory: %w", err)
	}

	// Create consignments directory
	if err := fileutil.EnsureDir(consignmentsDir); err != nil {
		return fmt.Errorf("failed to create consignments directory: %w", err)
	}

	return nil
}

// generateConfiguration creates a configuration based on detected packages
func generateConfiguration(projectPath string, options InitOptions) (*config.Config, error) {
	log := logger.Get()

	cfg := &config.Config{
		Packages: []config.Package{},
		Templates: config.TemplateConfig{
			Changelog: &config.TemplateSource{
				Source: "builtin:default",
			},
			TagName: &config.TemplateSource{
				Source: "builtin:default",
			},
			ReleaseNotes: &config.TemplateSource{
				Source: "builtin:default",
			},
		},
		Consignments: config.ConsignmentConfig{
			Path: ".shipyard/consignments",
		},
		History: config.HistoryConfig{
			Path: ".shipyard/history.json",
		},
	}

	// Add remote config if provided
	if options.Remote != "" {
		cfg.Extends = []config.RemoteConfig{
			{
				URL: options.Remote,
			},
		}
	}

	// Auto-detect packages
	log.Debug("Detecting packages...")
	detectedPackages, err := detect.DetectPackages(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect packages: %w", err)
	}

	// Interactive mode (default) - prompt for repo type and packages
	if !options.Yes {
		return generateInteractiveConfig(cfg, detectedPackages, projectPath)
	}

	// Non-interactive mode (--yes flag) - use auto-detection
	if len(detectedPackages) > 0 {
		log.Infof("Detected %d package(s)", len(detectedPackages))
		cfg.Packages = detectedPackages
	} else {
		// No packages detected, create a default one
		log.Info("No packages detected, creating default package")
		defaultPkg := config.Package{
			Name:      "default",
			Path:      "./",
			Ecosystem: config.EcosystemGo,
		}
		cfg.Packages = append(cfg.Packages, defaultPkg)
	}

	return cfg, nil
}

// generateInteractiveConfig prompts user for all configuration options
func generateInteractiveConfig(cfg *config.Config, detectedPackages []config.Package, projectPath string) (*config.Config, error) {
	log := logger.Get()

	fmt.Println() // Spacing

	// Step 1: Ask if monorepo or single repo
	repoType, err := prompt.PromptRepoType()
	if err != nil {
		return nil, fmt.Errorf("repository type selection failed: %w", err)
	}

	fmt.Println() // Spacing

	switch repoType {
	case prompt.RepoTypeMonorepo:
		// Monorepo: Review detected packages
		if len(detectedPackages) > 0 {
			log.Infof("Detected %d package(s)", len(detectedPackages))
			selectedPackages, err := prompt.PromptReviewPackages(detectedPackages)
			if err != nil {
				return nil, fmt.Errorf("package review failed: %w", err)
			}
			cfg.Packages = selectedPackages
			log.Infof("Selected %d package(s)", len(selectedPackages))
		} else {
			log.Warn("No packages detected in monorepo")
			// Prompt to add manually
			addManual, err := prompt.PromptConfirm("Would you like to add a package manually?", true)
			if err != nil {
				return nil, err
			}
			if addManual {
				pkg, err := promptForPackageDetails("package-1")
				if err != nil {
					return nil, err
				}
				cfg.Packages = append(cfg.Packages, pkg)
			}
		}

	case prompt.RepoTypeSingle:
		// Single repo: Configure one package
		var pkg config.Package
		if len(detectedPackages) == 1 {
			// Use detected package as default
			log.Infof("Detected package: %s (%s)", detectedPackages[0].Name, detectedPackages[0].Ecosystem)
			confirm, err := prompt.PromptConfirm("Use detected package configuration?", true)
			if err != nil {
				return nil, err
			}
			if confirm {
				pkg = detectedPackages[0]
			} else {
				pkg, err = promptForPackageDetails("main")
				if err != nil {
					return nil, err
				}
			}
		} else {
			// Prompt for package details
			pkg, err = promptForPackageDetails("main")
			if err != nil {
				return nil, err
			}
		}
		cfg.Packages = []config.Package{pkg}
		log.Infof("Configured package: %s", pkg.Name)
	}

	return cfg, nil
}

// promptForPackageDetails prompts user for package configuration
func promptForPackageDetails(defaultName string) (config.Package, error) {
	// Prompt for package name
	name, err := prompt.PromptTextInput("Package name:", defaultName)
	if err != nil {
		return config.Package{}, err
	}

	// Prompt for path
	path, err := prompt.PromptTextInput("Package path:", "./")
	if err != nil {
		return config.Package{}, err
	}

	// For now, default to Go ecosystem (could add ecosystem selection)
	return config.Package{
		Name:      name,
		Path:      path,
		Ecosystem: config.EcosystemGo,
	}, nil
}

// initializeHistoryFile creates an empty history file
func initializeHistoryFile(historyPath string) error {
	// Create empty JSON array
	emptyHistory := []byte("[]")
	return fileutil.AtomicWrite(historyPath, emptyHistory, 0644)
}
