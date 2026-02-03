package commands

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// GlobalFlags represents the global flags that can be inherited by all commands
type GlobalFlags struct {
	JSON    bool
	Quiet   bool
	Verbose bool
}

// GetGlobalFlags extracts global flags from the root command
func GetGlobalFlags(cmd *cobra.Command) GlobalFlags {
	flags := GlobalFlags{}

	// Traverse up to root command to get global flags
	rootCmd := cmd.Root()
	if rootCmd != nil {
		if flag := rootCmd.PersistentFlags().Lookup("json"); flag != nil {
			flags.JSON, _ = rootCmd.PersistentFlags().GetBool("json")
		}
		if flag := rootCmd.PersistentFlags().Lookup("quiet"); flag != nil {
			flags.Quiet, _ = rootCmd.PersistentFlags().GetBool("quiet")
		}
		if flag := rootCmd.PersistentFlags().Lookup("verbose"); flag != nil {
			flags.Verbose, _ = rootCmd.PersistentFlags().GetBool("verbose")
		}
	}

	return flags
}

// PrintJSON outputs data as formatted JSON
func PrintJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// PrintSuccess outputs a success message respecting the quiet flag
func PrintSuccess(w io.Writer, msg string, quiet bool) {
	if !quiet {
		_, _ = fmt.Fprintln(w, msg)
	}
}

// PrintKeyValue outputs a key-value pair respecting the quiet flag
func PrintKeyValue(w io.Writer, key, value string, quiet bool) {
	if !quiet {
		_, _ = fmt.Fprintf(w, "%s: %s\n", key, value)
	}
}
