package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Help output styles — reuses the color palette from output.go
var (
	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)

	helpProgramStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	helpCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13"))

	helpPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	helpDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	helpFlagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	helpFlagDefaultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	helpCommentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	helpAliasStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	helpFooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

// HelpFunc is a custom help function for cobra commands.
// It renders styled, colorized help output using lipgloss.
func HelpFunc(cmd *cobra.Command, args []string) {
	w := cmd.OutOrStdout()
	writeHelp(w, cmd)
}

func writeHelp(w io.Writer, cmd *cobra.Command) {
	// Description (Long or Short)
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}
	if desc != "" {
		fmt.Fprintln(w, helpDescriptionStyle.Render(desc))
		fmt.Fprintln(w)
	}

	// USAGE
	fmt.Fprintln(w, helpSectionStyle.Render("USAGE"))
	fmt.Fprintln(w, "  "+renderUsageLine(cmd))
	fmt.Fprintln(w)

	// ALIASES
	if len(cmd.Aliases) > 0 {
		fmt.Fprintln(w, helpSectionStyle.Render("ALIASES"))
		allNames := append([]string{cmd.Name()}, cmd.Aliases...)
		styledNames := make([]string, len(allNames))
		for i, name := range allNames {
			styledNames[i] = helpAliasStyle.Render(name)
		}
		fmt.Fprintln(w, "  "+strings.Join(styledNames, ", "))
		fmt.Fprintln(w)
	}

	// EXAMPLES
	if cmd.Example != "" {
		fmt.Fprintln(w, helpSectionStyle.Render("EXAMPLES"))
		fmt.Fprintln(w, renderExamples(cmd))
		fmt.Fprintln(w)
	}

	// COMMANDS (subcommands)
	if hasVisibleSubcommands(cmd) {
		fmt.Fprintln(w, helpSectionStyle.Render("COMMANDS"))
		fmt.Fprint(w, renderCommands(cmd))
		fmt.Fprintln(w)
	}

	// FLAGS (local)
	localFlags := cmd.LocalNonPersistentFlags()
	if hasVisibleFlags(localFlags) {
		fmt.Fprintln(w, helpSectionStyle.Render("FLAGS"))
		fmt.Fprint(w, renderFlags(localFlags))
		fmt.Fprintln(w)
	}

	// GLOBAL FLAGS (inherited)
	inheritedFlags := cmd.InheritedFlags()
	if hasVisibleFlags(inheritedFlags) {
		fmt.Fprintln(w, helpSectionStyle.Render("GLOBAL FLAGS"))
		fmt.Fprint(w, renderFlags(inheritedFlags))
		fmt.Fprintln(w)
	}

	// Footer hint
	if hasVisibleSubcommands(cmd) {
		cmdPath := cmd.CommandPath()
		footer := fmt.Sprintf(`Use "%s [command] --help" for more information about a command.`, cmdPath)
		fmt.Fprintln(w, helpFooterStyle.Render(footer))
	}
}

// renderUsageLine colorizes the command usage line.
func renderUsageLine(cmd *cobra.Command) string {
	useLine := cmd.UseLine()

	// The usage line uses standard syntax conventions:
	//   [ ] = optional    { } = required choice    | = mutually exclusive    ... = repeatable
	// We colorize: program name → cyan, flags → green, placeholders/syntax → gray, subcommands → magenta
	parts := strings.Fields(useLine)
	if len(parts) == 0 {
		return useLine
	}

	var styled []string
	for i, part := range parts {
		if i == 0 {
			// Program/root command name
			styled = append(styled, helpProgramStyle.Render(part))
		} else if isUsagePlaceholder(part) {
			styled = append(styled, helpPlaceholderStyle.Render(part))
		} else if isUsageFlag(part) {
			styled = append(styled, helpFlagStyle.Render(part))
		} else if part == "|" || part == "..." {
			styled = append(styled, helpPlaceholderStyle.Render(part))
		} else {
			// Subcommand name or argument
			styled = append(styled, helpCommandStyle.Render(part))
		}
	}
	return strings.Join(styled, " ")
}

// isUsagePlaceholder returns true for bracket/brace-enclosed tokens like [flags], {bash|zsh}, [--preview].
func isUsagePlaceholder(s string) bool {
	if len(s) < 2 {
		return false
	}
	return (s[0] == '[' && s[len(s)-1] == ']') ||
		(s[0] == '{' && s[len(s)-1] == '}') ||
		(s[0] == '[' && strings.HasSuffix(s, "]...")) ||
		(s[0] == '{' && strings.HasSuffix(s, "}..."))
}

// isUsageFlag returns true for flag-like tokens in the usage line (e.g. [-f], [--preview]).
func isUsageFlag(s string) bool {
	inner := strings.TrimPrefix(s, "[")
	inner = strings.TrimSuffix(inner, "]")
	inner = strings.TrimSuffix(inner, "]...")
	return strings.HasPrefix(inner, "-")
}

// renderExamples syntax-highlights example lines.
func renderExamples(cmd *cobra.Command) string {
	lines := strings.Split(cmd.Example, "\n")

	// Determine the root command name for highlighting
	rootName := cmd.Root().Name()

	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			result = append(result, "")
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			// Comment line — render entirely in gray
			result = append(result, "  "+helpCommentStyle.Render(trimmed))
			continue
		}

		// Command line — tokenize and colorize
		result = append(result, "  "+colorizeExampleLine(trimmed, rootName))
	}

	return strings.Join(result, "\n")
}

// colorizeExampleLine colorizes a single example command line.
func colorizeExampleLine(line, rootName string) string {
	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return line
	}

	var styled []string
	for _, token := range tokens {
		switch {
		case token == rootName:
			styled = append(styled, helpProgramStyle.Render(token))
		case strings.HasPrefix(token, "--") || (strings.HasPrefix(token, "-") && len(token) == 2):
			styled = append(styled, helpFlagStyle.Render(token))
		default:
			styled = append(styled, token)
		}
	}
	return strings.Join(styled, " ")
}

// renderFlags renders a flag set with styled flag names and descriptions.
func renderFlags(flags *pflag.FlagSet) string {
	type flagEntry struct {
		nameStr string
		usage   string
		defVal  string
	}

	var entries []flagEntry
	maxWidth := 0

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}

		var nameStr string
		if f.Shorthand != "" {
			nameStr = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
		} else {
			nameStr = fmt.Sprintf("    --%s", f.Name)
		}

		// Append type for non-bool flags
		typeName := f.Value.Type()
		if typeName != "bool" {
			nameStr += " " + typeName
		}

		if len(nameStr) > maxWidth {
			maxWidth = len(nameStr)
		}

		entries = append(entries, flagEntry{
			nameStr: nameStr,
			usage:   f.Usage,
			defVal:  f.DefValue,
		})
	})

	var lines []string
	for _, e := range entries {
		padding := strings.Repeat(" ", maxWidth-len(e.nameStr)+4)
		line := "  " + helpFlagStyle.Render(e.nameStr) + padding + helpDescriptionStyle.Render(e.usage)

		// Show default value for non-zero, non-bool, non-empty defaults
		if e.defVal != "" && e.defVal != "false" && e.defVal != "0" && e.defVal != "[]" {
			line += " " + helpFlagDefaultStyle.Render(fmt.Sprintf("(default %s)", e.defVal))
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// renderCommands renders subcommands with styled names and descriptions.
func renderCommands(cmd *cobra.Command) string {
	commands := cmd.Commands()

	// Find max command name length for alignment
	maxWidth := 0
	for _, sub := range commands {
		if sub.IsAvailableCommand() || sub.Name() == "help" {
			if len(sub.Name()) > maxWidth {
				maxWidth = len(sub.Name())
			}
		}
	}

	var lines []string
	for _, sub := range commands {
		if !sub.IsAvailableCommand() && sub.Name() != "help" {
			continue
		}
		padding := strings.Repeat(" ", maxWidth-len(sub.Name())+4)
		line := "  " + helpCommandStyle.Render(sub.Name()) + padding + helpDescriptionStyle.Render(sub.Short)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// hasVisibleFlags returns true if the flag set has any non-hidden flags.
func hasVisibleFlags(flags *pflag.FlagSet) bool {
	hasFlags := false
	flags.VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			hasFlags = true
		}
	})
	return hasFlags
}

// hasVisibleSubcommands returns true if the command has visible subcommands.
func hasVisibleSubcommands(cmd *cobra.Command) bool {
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() || sub.Name() == "help" {
			return true
		}
	}
	return false
}
