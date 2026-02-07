package commands

import (
	"fmt"
	"os"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewConfigShowCommand creates the config show command
func NewConfigShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display the resolved configuration",
		Long: `Display the current shipyard configuration with all defaults applied.

Outputs as YAML by default, or JSON with the --json flag.

Examples:
  # Show config as YAML
  shipyard config show

  # Show config as JSON
  shipyard config show --json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			globalFlags := GetGlobalFlags(cmd)
			return runConfigShow(globalFlags)
		},
	}

	return cmd
}

func runConfigShow(flags GlobalFlags) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return runConfigShowWithDir(cwd, flags)
}

func runConfigShowWithDir(projectPath string, flags GlobalFlags) error {
	cfg, err := config.LoadFromDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	resolved := cfg.WithDefaults()

	if flags.JSON {
		return PrintJSON(os.Stdout, resolved)
	}

	data, err := yaml.Marshal(resolved)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	fmt.Print(string(data))
	return nil
}
