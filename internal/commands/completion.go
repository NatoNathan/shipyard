package commands

import (
	"os"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/spf13/cobra"
)

// NewCompletionCommand creates the completion command for shell completions.
func NewCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Teach your shell to speak Shipyard",
		Long: `Train your shell to understand the shipyard's language. Enables your navigator
(shell) to suggest commands, flags, and arguments as you chart your course.

Installation Instructions:

Bash:
  # Linux:
  $ shipyard completion bash > /etc/bash_completion.d/shipyard

  # macOS:
  $ shipyard completion bash > /usr/local/etc/bash_completion.d/shipyard

  # Or add to your ~/.bashrc:
  $ echo 'source <(shipyard completion bash)' >> ~/.bashrc

Zsh:
  # Generate completion file:
  $ shipyard completion zsh > "${fpath[1]}/_shipyard"

  # Or add to your ~/.zshrc:
  $ echo 'source <(shipyard completion zsh)' >> ~/.zshrc
  $ echo 'compdef _shipyard shipyard' >> ~/.zshrc

Fish:
  $ shipyard completion fish > ~/.config/fish/completions/shipyard.fish

PowerShell:
  # Add to your PowerShell profile:
  PS> shipyard completion powershell | Out-String | Invoke-Expression

  # Or save to profile:
  PS> shipyard completion powershell >> $PROFILE

After installing, restart your shell or source the completion file.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE:                  runCompletion,
	}

	return cmd
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}

	return nil
}

// RegisterPackageCompletions registers package name completions for a command flag.
// This enables tab-completion of package names from the Shipyard configuration.
func RegisterPackageCompletions(cmd *cobra.Command, flagName string) {
	_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Try to load configuration
		cwd, err := os.Getwd()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cfg, err := config.LoadFromDir(cwd)
		if err != nil {
			// If config not found, don't show error, just don't provide completions
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Extract package names
		var names []string
		for _, pkg := range cfg.Packages {
			names = append(names, pkg.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	})
}
