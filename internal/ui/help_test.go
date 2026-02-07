package ui

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "shipyard",
		Short: "Chart your project's version journey",
	}

	cmd := &cobra.Command{
		Use:     "version [command] [flags]",
		Aliases: []string{"bump", "sail"},
		Short:   "Sail to the next port",
		Long:    "Set sail with your cargo and reach the next version port.",
		Example: `  # Set sail for all vessels
  shipyard version

  # Preview the route without sailing
  shipyard version --preview`,
	}

	cmd.Flags().Bool("preview", false, "Show changes without applying them")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed output")
	cmd.Flags().StringP("package", "p", "", "Filter to specific packages")

	sub := &cobra.Command{
		Use:   "prerelease",
		Short: "Chart test waters before the main voyage",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(sub)

	root.PersistentFlags().BoolP("json", "j", false, "output in JSON format")
	root.AddCommand(cmd)

	return cmd
}

func TestHelpFunc_AllSections(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "Set sail with your cargo", "Should contain long description")
	assert.Contains(t, output, "USAGE", "Should contain USAGE section")
	assert.Contains(t, output, "ALIASES", "Should contain ALIASES section")
	assert.Contains(t, output, "EXAMPLES", "Should contain EXAMPLES section")
	assert.Contains(t, output, "COMMANDS", "Should contain COMMANDS section")
	assert.Contains(t, output, "FLAGS", "Should contain FLAGS section")
	assert.Contains(t, output, "GLOBAL FLAGS", "Should contain GLOBAL FLAGS section")
}

func TestHelpFunc_Description(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "Set sail with your cargo and reach the next version port.")
}

func TestHelpFunc_FallbackToShort(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	cmd := &cobra.Command{
		Use:   "subcmd",
		Short: "A short description",
	}
	root.AddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "A short description", "Should fall back to Short when Long is empty")
}

func TestHelpFunc_Aliases(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "version", "Should contain the command name in aliases")
	assert.Contains(t, output, "bump", "Should contain alias 'bump'")
	assert.Contains(t, output, "sail", "Should contain alias 'sail'")
}

func TestHelpFunc_Flags(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "--preview", "Should contain --preview flag")
	assert.Contains(t, output, "--verbose", "Should contain --verbose flag")
	assert.Contains(t, output, "--package", "Should contain --package flag")
	assert.Contains(t, output, "-v", "Should contain -v shorthand")
	assert.Contains(t, output, "-p", "Should contain -p shorthand")
}

func TestHelpFunc_GlobalFlags(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "--json", "Should contain global --json flag")
	assert.Contains(t, output, "-j", "Should contain global -j shorthand")
}

func TestHelpFunc_Subcommands(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "prerelease", "Should contain subcommand name")
	assert.Contains(t, output, "Chart test waters", "Should contain subcommand description")
}

func TestHelpFunc_Examples(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, "Set sail for all vessels", "Should contain example comment")
	assert.Contains(t, output, "--preview", "Should contain example flag")
}

func TestHelpFunc_NoSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "shipyard"}
	cmd := &cobra.Command{
		Use:   "simple",
		Short: "A simple command",
	}
	cmd.Flags().Bool("flag1", false, "A test flag")
	root.AddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.NotContains(t, output, "COMMANDS", "Should not contain COMMANDS section")
	assert.NotContains(t, output, "[command] --help", "Should not contain footer hint")
}

func TestHelpFunc_NoExample(t *testing.T) {
	root := &cobra.Command{Use: "shipyard"}
	cmd := &cobra.Command{
		Use:   "minimal",
		Short: "Minimal command",
	}
	root.AddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.NotContains(t, output, "EXAMPLES", "Should not contain EXAMPLES section when no examples")
}

func TestHelpFunc_NoAliases(t *testing.T) {
	root := &cobra.Command{Use: "shipyard"}
	cmd := &cobra.Command{
		Use:   "noalias",
		Short: "No aliases command",
	}
	root.AddCommand(cmd)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.NotContains(t, output, "ALIASES", "Should not contain ALIASES section when no aliases")
}

func TestHelpFunc_FooterHint(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	HelpFunc(cmd, nil)
	output := buf.String()

	assert.Contains(t, output, `Use "shipyard version [command] --help"`, "Should contain footer hint for commands with subcommands")
}
