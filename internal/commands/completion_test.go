package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommand_Bash(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run bash completion
	rootCmd.SetArgs([]string{"completion", "bash"})
	err := rootCmd.Execute()

	// Verify - just check it doesn't error
	require.NoError(t, err)
}

func TestCompletionCommand_Zsh(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run zsh completion
	rootCmd.SetArgs([]string{"completion", "zsh"})
	err := rootCmd.Execute()

	// Verify - just check it doesn't error
	require.NoError(t, err)
}

func TestCompletionCommand_Fish(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run fish completion
	rootCmd.SetArgs([]string{"completion", "fish"})
	err := rootCmd.Execute()

	// Verify - just check it doesn't error
	require.NoError(t, err)
}

func TestCompletionCommand_PowerShell(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run powershell completion
	rootCmd.SetArgs([]string{"completion", "powershell"})
	err := rootCmd.Execute()

	// Verify - just check it doesn't error
	require.NoError(t, err)
}

func TestCompletionCommand_InvalidShell(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run with invalid shell
	rootCmd.SetArgs([]string{"completion", "invalid-shell"})
	err := rootCmd.Execute()

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

func TestCompletionCommand_NoArgs(t *testing.T) {
	// Create root command
	rootCmd := &cobra.Command{Use: "shipyard"}
	completionCmd := NewCompletionCommand()
	rootCmd.AddCommand(completionCmd)

	// Run without arguments
	rootCmd.SetArgs([]string{"completion"})
	err := rootCmd.Execute()

	// Verify error
	assert.Error(t, err)
	// Check for either "requires" or "accepts" in error message
	errorMsg := strings.ToLower(err.Error())
	assert.True(t, strings.Contains(errorMsg, "requires") || strings.Contains(errorMsg, "accepts"),
		"Error should mention argument requirements")
}

func TestPackageCompletions(t *testing.T) {
	// Setup: Create a temp directory with shipyard config
	tempDir := t.TempDir()

	// Create config with multiple packages
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
			{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
			{Name: "utils", Path: "./utils", Ecosystem: config.EcosystemGo},
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Create a test command with package flag
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	var packageFlag string
	testCmd.Flags().StringVarP(&packageFlag, "package", "p", "", "package name")

	// Register completions - this should not error
	RegisterPackageCompletions(testCmd, "package")

	// Verify the flag exists
	flag := testCmd.Flags().Lookup("package")
	require.NotNil(t, flag, "package flag should exist")
}

func TestPackageCompletions_NoConfig(t *testing.T) {
	// Setup: Create a temp directory without shipyard config
	tempDir := t.TempDir()

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Create a test command with package flag
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	var packageFlag string
	testCmd.Flags().StringVarP(&packageFlag, "package", "p", "", "package name")

	// Register completions - should not error even without config
	RegisterPackageCompletions(testCmd, "package")

	// Verify the flag exists
	flag := testCmd.Flags().Lookup("package")
	require.NotNil(t, flag, "package flag should exist")
}
