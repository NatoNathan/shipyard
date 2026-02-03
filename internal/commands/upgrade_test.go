package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeCommand_Contract(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	t.Run("command exists", func(t *testing.T) {
		cmd := NewUpgradeCommand(versionInfo)
		require.NotNil(t, cmd)
		assert.Equal(t, "upgrade", cmd.Use)
	})

	t.Run("has correct short description", func(t *testing.T) {
		cmd := NewUpgradeCommand(versionInfo)
		assert.Equal(t, "Upgrade shipyard to the latest version", cmd.Short)
	})

	t.Run("has long description", func(t *testing.T) {
		cmd := NewUpgradeCommand(versionInfo)
		assert.NotEmpty(t, cmd.Long)
		assert.Contains(t, cmd.Long, "automatically detects")
		assert.Contains(t, cmd.Long, "Homebrew")
		assert.Contains(t, cmd.Long, "npm")
		assert.Contains(t, cmd.Long, "Go install")
	})

	t.Run("has examples", func(t *testing.T) {
		cmd := NewUpgradeCommand(versionInfo)
		assert.Contains(t, cmd.Long, "Examples:")
		assert.Contains(t, cmd.Long, "shipyard upgrade")
		assert.Contains(t, cmd.Long, "--yes")
		assert.Contains(t, cmd.Long, "--dry-run")
		assert.Contains(t, cmd.Long, "--force")
	})
}

func TestUpgradeCommand_Flags(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	cmd := NewUpgradeCommand(versionInfo)

	t.Run("has --yes flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("yes")
		require.NotNil(t, flag)
		assert.Equal(t, "y", flag.Shorthand)
		assert.Equal(t, "false", flag.DefValue)
		assert.Contains(t, flag.Usage, "Skip confirmation")
	})

	t.Run("has --version flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("version")
		require.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
		assert.Contains(t, flag.Usage, "Upgrade to specific version")
		assert.Contains(t, flag.Usage, "latest")
	})

	t.Run("has --force flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("force")
		require.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
		assert.Contains(t, flag.Usage, "Force upgrade")
		assert.Contains(t, flag.Usage, "latest")
	})

	t.Run("has --dry-run flag", func(t *testing.T) {
		flag := cmd.Flags().Lookup("dry-run")
		require.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
		assert.Contains(t, flag.Usage, "Show upgrade plan")
		assert.Contains(t, flag.Usage, "without executing")
	})
}

func TestUpgradeCommand_Help(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	cmd := NewUpgradeCommand(versionInfo)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	helpOutput := buf.String()

	t.Run("shows usage", func(t *testing.T) {
		assert.Contains(t, helpOutput, "Usage:")
		assert.Contains(t, helpOutput, "shipyard upgrade")
	})

	t.Run("shows flags", func(t *testing.T) {
		assert.Contains(t, helpOutput, "Flags:")
		assert.Contains(t, helpOutput, "--yes")
		assert.Contains(t, helpOutput, "--version")
		assert.Contains(t, helpOutput, "--force")
		assert.Contains(t, helpOutput, "--dry-run")
	})

	t.Run("shows flags section", func(t *testing.T) {
		// Global flags only appear when attached to root command
		// Just verify we have a Flags section
		assert.Contains(t, helpOutput, "Flags:")
	})

	t.Run("shows examples", func(t *testing.T) {
		assert.Contains(t, helpOutput, "Examples:")
	})
}

func TestUpgradeCommand_FlagParsing(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "parses --yes flag",
			args: []string{"--yes"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				yes, err := cmd.Flags().GetBool("yes")
				require.NoError(t, err)
				assert.True(t, yes)
			},
		},
		{
			name: "parses -y shorthand",
			args: []string{"-y"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				yes, err := cmd.Flags().GetBool("yes")
				require.NoError(t, err)
				assert.True(t, yes)
			},
		},
		{
			name: "parses --version flag",
			args: []string{"--version", "v1.2.3"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				version, err := cmd.Flags().GetString("version")
				require.NoError(t, err)
				assert.Equal(t, "v1.2.3", version)
			},
		},
		{
			name: "parses --force flag",
			args: []string{"--force"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				force, err := cmd.Flags().GetBool("force")
				require.NoError(t, err)
				assert.True(t, force)
			},
		},
		{
			name: "parses --dry-run flag",
			args: []string{"--dry-run"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				dryRun, err := cmd.Flags().GetBool("dry-run")
				require.NoError(t, err)
				assert.True(t, dryRun)
			},
		},
		{
			name: "parses multiple flags",
			args: []string{"--yes", "--force", "--version", "v2.0.0"},
			validate: func(t *testing.T, cmd *cobra.Command) {
				yes, err := cmd.Flags().GetBool("yes")
				require.NoError(t, err)
				assert.True(t, yes)

				force, err := cmd.Flags().GetBool("force")
				require.NoError(t, err)
				assert.True(t, force)

				version, err := cmd.Flags().GetString("version")
				require.NoError(t, err)
				assert.Equal(t, "v2.0.0", version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewUpgradeCommand(versionInfo)
			cmd.SetArgs(tt.args)

			// Parse flags without executing
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			tt.validate(t, cmd)
		})
	}
}

func TestUpgradeCommand_NoArguments(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	cmd := NewUpgradeCommand(versionInfo)

	t.Run("accepts no arguments", func(t *testing.T) {
		// The command should not require any arguments
		// This is validated by the command structure
		assert.Nil(t, cmd.Args)
	})
}

func TestUpgradeOptions_Defaults(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	cmd := NewUpgradeCommand(versionInfo)

	t.Run("all flags default to false/empty", func(t *testing.T) {
		yes, err := cmd.Flags().GetBool("yes")
		require.NoError(t, err)
		assert.False(t, yes)

		force, err := cmd.Flags().GetBool("force")
		require.NoError(t, err)
		assert.False(t, force)

		dryRun, err := cmd.Flags().GetBool("dry-run")
		require.NoError(t, err)
		assert.False(t, dryRun)

		version, err := cmd.Flags().GetString("version")
		require.NoError(t, err)
		assert.Empty(t, version)
	})
}

func TestVersionInfo_Structure(t *testing.T) {
	t.Run("has required fields", func(t *testing.T) {
		info := VersionInfo{
			Version: "1.0.0",
			Commit:  "abc123",
			Date:    "2024-01-01",
		}

		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.Commit)
		assert.NotEmpty(t, info.Date)
	})
}


func TestUpgradeCommand_UsageMessage(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	cmd := NewUpgradeCommand(versionInfo)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Force usage display
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	usage := buf.String()

	t.Run("mentions all installation methods", func(t *testing.T) {
		methods := []string{"Homebrew", "npm", "Go install", "script"}
		for _, method := range methods {
			assert.Contains(t, strings.ToLower(usage), strings.ToLower(method),
				"Usage should mention %s installation method", method)
		}
	})

	t.Run("shows all flag descriptions", func(t *testing.T) {
		flagDescriptions := []string{
			"Skip confirmation",
			"Show upgrade plan",
			"Force upgrade",
			"specific version",
		}
		for _, desc := range flagDescriptions {
			assert.Contains(t, usage, desc,
				"Usage should describe: %s", desc)
		}
	})
}
