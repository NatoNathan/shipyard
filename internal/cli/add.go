package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/logger"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// isStdinPiped checks if stdin is being piped
func isStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// readStdinLine reads a single line from stdin
func readStdinLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no input available")
}

var AddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create", "new", "a"},
	Short:   "Add a new consignment",
	Long:    `Create a new consignment to describe changes made to packages. This is used to track changes for release management.`,
	Example: `
shipyard add
shipyard add --type patch
shipyard add --summary "Fixed bug"
shipyard add --package api --type patch
shipyard add --type patch --summary "Fixed bug"
shipyard add --package api --type patch --summary "Fixed bug"
echo "Fixed bug" | shipyard add --type patch
echo "patch" | shipyard add --summary "Fixed bug"
echo "1" | shipyard add --summary "Fixed bug"
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load project configuration
		projectConfig, err := config.LoadProjectConfig()
		if err != nil {
			logger.Error("Failed to load project configuration", "error", err)
			fmt.Println("Error: Unable to load project configuration.")
			fmt.Println("Make sure you're in a Shipyard project directory and run 'shipyard init' if needed.")
			os.Exit(1)
		}

		// Create consignment manager
		manager := consignment.NewManager(projectConfig)

		var (
			selectedPackages []string
			changeType       string
			summary          string
		)

		// Check if flags are provided for non-interactive mode
		packageFlag, _ := cmd.Flags().GetStringSlice("package")
		typeFlag, _ := cmd.Flags().GetString("type")
		summaryFlag, _ := cmd.Flags().GetString("summary")

		// Handle package selection based on repository type
		availablePackages := manager.GetAvailablePackages()
		if len(availablePackages) == 0 {
			fmt.Println("No packages configured in this project.")
			fmt.Println("Please run 'shipyard init' to configure your packages.")
			os.Exit(1)
		}

		// For single repo, we can infer the package if not provided
		if projectConfig.Type == config.RepositoryTypeSingleRepo && len(packageFlag) == 0 {
			packageFlag = []string{availablePackages[0].Name}
		}

		// Handle package selection with prompt builder approach
		if projectConfig.Type == config.RepositoryTypeMonorepo {
			if len(packageFlag) > 0 {
				// Packages provided via flag - validate them
				availablePackageNames := make(map[string]bool)
				for _, pkg := range availablePackages {
					availablePackageNames[pkg.Name] = true
				}

				for _, pkg := range packageFlag {
					if !availablePackageNames[pkg] {
						fmt.Printf("Error: Package '%s' not found in project configuration\n", pkg)
						fmt.Println("Available packages:")
						for _, availablePkg := range availablePackages {
							fmt.Printf("  - %s (%s)\n", availablePkg.Name, availablePkg.Path)
						}
						os.Exit(1)
					}
				}
				selectedPackages = packageFlag
			} else {
				// No packages provided via flag - prompt for selection
				packageOptions := make([]huh.Option[string], 0, len(availablePackages))
				for _, pkg := range availablePackages {
					packageOptions = append(packageOptions, huh.NewOption(fmt.Sprintf("%s (%s)", pkg.Name, pkg.Path), pkg.Name))
				}

				packageForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().
							Title("📦 Package Selection").
							Description("Select the packages that have been changed."),
						huh.NewMultiSelect[string]().
							Title("Which packages have changed?").
							Description("Select all packages that have been modified.").
							Options(packageOptions...).
							Value(&selectedPackages),
					),
				)

				if err := packageForm.Run(); err != nil {
					logger.Error("Failed to get package selection", "error", err)
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}

				if len(selectedPackages) == 0 {
					fmt.Println("No packages selected. Consignment creation cancelled.")
					return
				}
			}
		} else {
			// Single repo - use the configured package
			selectedPackages = []string{availablePackages[0].Name}
		}

		// Change type selection with prompt builder approach
		if typeFlag != "" {
			// Change type provided via flag - validate it
			changeType = typeFlag
			if err := manager.ValidateChangeType(changeType); err != nil {
				fmt.Printf("Error: %v\n", err)
				fmt.Println("Available change types:")
				for _, ct := range manager.GetAvailableChangeTypes() {
					displayName := ct.DisplayName
					if displayName == "" {
						displayName = ct.Name
					}
					fmt.Printf("  - %s (%s)\n", ct.Name, displayName)
				}
				os.Exit(1)
			}
		} else {
			// No change type provided via flag - prompt for selection
			if isStdinPiped() {
				// Handle piped input - expect change type name or number
				input, err := readStdinLine()
				if err != nil {
					fmt.Println("Error: No change type provided and unable to read from stdin")
					fmt.Println("Available change types:")
					for i, ct := range projectConfig.GetChangeTypes() {
						fmt.Printf("  %d. %s\n", i+1, ct.Name)
					}
					fmt.Println("Usage: echo '<type-name-or-number>' | shipyard add --summary 'your summary'")
					os.Exit(1)
				}

				// Try to parse as number first
				changeTypes := projectConfig.GetChangeTypes()
				if input >= "1" && input <= fmt.Sprintf("%d", len(changeTypes)) {
					// Input is a number (1-based)
					if num := int(input[0] - '0'); num >= 1 && num <= len(changeTypes) {
						changeType = changeTypes[num-1].Name
					}
				} else {
					// Input is a change type name
					changeType = input
				}

				// Validate the change type
				if err := manager.ValidateChangeType(changeType); err != nil {
					fmt.Printf("Error: %v\n", err)
					fmt.Println("Available change types:")
					for i, ct := range changeTypes {
						fmt.Printf("  %d. %s\n", i+1, ct.Name)
					}
					os.Exit(1)
				}
			} else {
				// Interactive mode
				changeTypeOptions := make([]huh.Option[string], 0, len(projectConfig.GetChangeTypes()))
				for _, ct := range projectConfig.GetChangeTypes() {
					optionText := ct.DisplayName
					if optionText == "" {
						optionText = ct.Name
					}
					changeTypeOptions = append(changeTypeOptions, huh.NewOption(optionText, ct.Name))
				}

				changeTypeForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().
							Title("📝 Change Type").
							Description("Select the type of change you've made."),
						huh.NewSelect[string]().
							Title("What type of change is this?").
							Description("Choose the appropriate change type.").
							Options(changeTypeOptions...).
							Value(&changeType),
					),
				)

				if err := changeTypeForm.Run(); err != nil {
					logger.Error("Failed to get change type", "error", err)
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// Summary input with prompt builder approach
		if summaryFlag != "" {
			// Summary provided via flag - validate it
			summary = summaryFlag
			if strings.TrimSpace(summary) == "" {
				fmt.Println("Error: --summary flag cannot be empty")
				os.Exit(1)
			}
		} else {
			// No summary provided via flag - prompt for input
			if isStdinPiped() {
				// Handle piped input - read summary from stdin
				input, err := readStdinLine()
				if err != nil {
					fmt.Println("Error: No summary provided and unable to read from stdin")
					fmt.Println("Usage: echo 'your summary' | shipyard add --type <type>")
					os.Exit(1)
				}

				summary = input
				if strings.TrimSpace(summary) == "" {
					fmt.Println("Error: Summary cannot be empty")
					os.Exit(1)
				}
			} else {
				// Interactive mode
				summaryForm := huh.NewForm(
					huh.NewGroup(
						huh.NewNote().
							Title("📋 Change Summary").
							Description("Provide a summary of the changes made."),
						huh.NewText().
							Title("Summary").
							Description("Describe the changes made to the selected packages.").
							Placeholder("e.g. Fixed bug in user authentication flow").
							Value(&summary),
					),
				)

				if err := summaryForm.Run(); err != nil {
					logger.Error("Failed to get change summary", "error", err)
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		// Create consignment using the manager
		createdConsignment, err := manager.CreateConsignment(selectedPackages, changeType, summary)
		if err != nil {
			logger.Error("Failed to create consignment", "error", err)
			fmt.Printf("Error: Unable to create consignment: %v\n", err)
			os.Exit(1)
		}

		// Success message
		fmt.Printf("\n🎉 Consignment created successfully!\n")
		fmt.Printf("📄 File: %s/%s.md\n", manager.GetConsignmentDir(), createdConsignment.ID)
		fmt.Printf("📦 Packages: %s\n", strings.Join(selectedPackages, ", "))
		fmt.Printf("🔄 Type: %s\n", changeType)
		fmt.Printf("📝 Summary: %s\n", summary)
		fmt.Printf("\n💡 Next steps:\n")
		fmt.Printf("   - Review the consignment file\n")
		fmt.Printf("   - Commit the consignment to your repository\n")
		fmt.Printf("   - Run 'shipyard version' to calculate new versions\n")

		logger.Info("Consignment created successfully",
			"id", createdConsignment.ID,
			"packages", selectedPackages,
			"type", changeType,
		)
	},
}

func init() {
	// Add flags for prompt builder consignment creation
	AddCmd.Flags().StringSliceP("package", "p", []string{}, "Package(s) to include in the consignment (optional - will prompt if not provided for monorepo)")
	AddCmd.Flags().StringP("type", "t", "", "Change type (optional - will prompt if not provided)")
	AddCmd.Flags().StringP("summary", "s", "", "Summary of the changes (optional - will prompt if not provided)")
	AddCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts (currently unused but reserved for future use)")
}
