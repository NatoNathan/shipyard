// Package changelog provides Keep a Changelog template implementation.
package changelog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// KeepAChangelogTemplate implements the Keep a Changelog format
// Based on https://keepachangelog.com/en/1.0.0/
type KeepAChangelogTemplate struct{}

func (t *KeepAChangelogTemplate) Name() string {
	return "keepachangelog"
}

func (t *KeepAChangelogTemplate) Description() string {
	return "Keep a Changelog format (https://keepachangelog.com/)"
}

func (t *KeepAChangelogTemplate) Generate(entries []ChangelogEntry, projectConfig *config.ProjectConfig) (string, error) {
	var buf strings.Builder

	// Header
	buf.WriteString("# Changelog\n\n")
	buf.WriteString("All notable changes to this project will be documented in this file.\n\n")
	buf.WriteString("The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n")
	buf.WriteString("and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n")

	// Custom header template if provided
	if projectConfig.Changelog.HeaderTemplate != "" {
		buf.WriteString(projectConfig.Changelog.HeaderTemplate)
		buf.WriteString("\n\n")
	}

	// Generate entries
	for _, entry := range entries {
		if err := t.generateEntry(&buf, entry, projectConfig); err != nil {
			return "", fmt.Errorf("failed to generate changelog entry: %w", err)
		}
	}

	// Custom footer template if provided
	if projectConfig.Changelog.FooterTemplate != "" {
		buf.WriteString("\n")
		buf.WriteString(projectConfig.Changelog.FooterTemplate)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func (t *KeepAChangelogTemplate) generateEntry(buf *strings.Builder, entry ChangelogEntry, projectConfig *config.ProjectConfig) error {
	// Version header
	if projectConfig.Type == config.RepositoryTypeMonorepo {
		buf.WriteString(fmt.Sprintf("## [%s] - %s - %s\n\n",
			entry.Version,
			entry.PackageName,
			entry.Date.Format("2006-01-02")))
	} else {
		buf.WriteString(fmt.Sprintf("## [%s] - %s\n\n",
			entry.Version,
			entry.Date.Format("2006-01-02")))
	}

	// Define the order of sections according to Keep a Changelog
	sectionOrder := []string{
		"Breaking Changes",
		"Added",
		"Changed",
		"Deprecated",
		"Removed",
		"Fixed",
		"Security",
	}

	// Generate sections in order
	for _, section := range sectionOrder {
		if changes, exists := entry.Changes[section]; exists && len(changes) > 0 {
			buf.WriteString(fmt.Sprintf("### %s\n\n", section))

			// Sort changes by summary for consistent output
			sort.Slice(changes, func(i, j int) bool {
				return changes[i].Summary < changes[j].Summary
			})

			for _, change := range changes {
				if projectConfig.Type == config.RepositoryTypeMonorepo {
					buf.WriteString(fmt.Sprintf("- **%s**: %s\n", change.PackageName, change.Summary))
				} else {
					buf.WriteString(fmt.Sprintf("- %s\n", change.Summary))
				}
			}
			buf.WriteString("\n")
		}
	}

	return nil
}

// ConventionalTemplate implements the Conventional Commits changelog format
type ConventionalTemplate struct{}

func (t *ConventionalTemplate) Name() string {
	return "conventional"
}

func (t *ConventionalTemplate) Description() string {
	return "Conventional Commits format with automatic categorization"
}

func (t *ConventionalTemplate) Generate(entries []ChangelogEntry, projectConfig *config.ProjectConfig) (string, error) {
	var buf strings.Builder

	// Header
	buf.WriteString("# Changelog\n\n")
	if projectConfig.Changelog.HeaderTemplate != "" {
		buf.WriteString(projectConfig.Changelog.HeaderTemplate)
		buf.WriteString("\n\n")
	}

	// Generate entries
	for _, entry := range entries {
		if err := t.generateEntry(&buf, entry, projectConfig); err != nil {
			return "", fmt.Errorf("failed to generate changelog entry: %w", err)
		}
	}

	// Footer
	if projectConfig.Changelog.FooterTemplate != "" {
		buf.WriteString("\n")
		buf.WriteString(projectConfig.Changelog.FooterTemplate)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func (t *ConventionalTemplate) generateEntry(buf *strings.Builder, entry ChangelogEntry, projectConfig *config.ProjectConfig) error {
	// Version header
	if projectConfig.Type == config.RepositoryTypeMonorepo {
		buf.WriteString(fmt.Sprintf("## %s (%s) - %s\n\n",
			entry.Version,
			entry.PackageName,
			entry.Date.Format("2006-01-02")))
	} else {
		buf.WriteString(fmt.Sprintf("## %s (%s)\n\n",
			entry.Version,
			entry.Date.Format("2006-01-02")))
	}

	// Map sections to conventional commit types
	sectionOrder := []string{
		"Breaking Changes", // BREAKING CHANGE
		"Added",            // feat
		"Fixed",            // fix
		"Changed",          // refactor, perf, style
		"Removed",          // remove
		"Security",         // security
	}

	// Generate sections
	for _, section := range sectionOrder {
		if changes, exists := entry.Changes[section]; exists && len(changes) > 0 {
			buf.WriteString(fmt.Sprintf("### %s\n\n", section))

			for _, change := range changes {
				if projectConfig.Type == config.RepositoryTypeMonorepo {
					buf.WriteString(fmt.Sprintf("- **%s**: %s\n", change.PackageName, change.Summary))
				} else {
					buf.WriteString(fmt.Sprintf("- %s\n", change.Summary))
				}
			}
			buf.WriteString("\n")
		}
	}

	return nil
}

// SimpleTemplate implements a simple, minimal changelog format
type SimpleTemplate struct{}

func (t *SimpleTemplate) Name() string {
	return "simple"
}

func (t *SimpleTemplate) Description() string {
	return "Simple, minimal changelog format"
}

func (t *SimpleTemplate) Generate(entries []ChangelogEntry, projectConfig *config.ProjectConfig) (string, error) {
	var buf strings.Builder

	// Header
	buf.WriteString("# Changelog\n\n")
	if projectConfig.Changelog.HeaderTemplate != "" {
		buf.WriteString(projectConfig.Changelog.HeaderTemplate)
		buf.WriteString("\n\n")
	}

	// Generate entries
	for _, entry := range entries {
		if err := t.generateEntry(&buf, entry, projectConfig); err != nil {
			return "", fmt.Errorf("failed to generate changelog entry: %w", err)
		}
	}

	// Footer
	if projectConfig.Changelog.FooterTemplate != "" {
		buf.WriteString("\n")
		buf.WriteString(projectConfig.Changelog.FooterTemplate)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func (t *SimpleTemplate) generateEntry(buf *strings.Builder, entry ChangelogEntry, projectConfig *config.ProjectConfig) error {
	// Version header
	if projectConfig.Type == config.RepositoryTypeMonorepo {
		buf.WriteString(fmt.Sprintf("## %s - %s (%s)\n\n",
			entry.Version,
			entry.PackageName,
			entry.Date.Format("2006-01-02")))
	} else {
		buf.WriteString(fmt.Sprintf("## %s (%s)\n\n",
			entry.Version,
			entry.Date.Format("2006-01-02")))
	}

	// Simple list of all changes
	var allChanges []ChangelogChange
	for _, changes := range entry.Changes {
		allChanges = append(allChanges, changes...)
	}

	// Sort changes by summary
	sort.Slice(allChanges, func(i, j int) bool {
		return allChanges[i].Summary < allChanges[j].Summary
	})

	for _, change := range allChanges {
		if projectConfig.Type == config.RepositoryTypeMonorepo {
			buf.WriteString(fmt.Sprintf("- **%s**: %s\n", change.PackageName, change.Summary))
		} else {
			buf.WriteString(fmt.Sprintf("- %s\n", change.Summary))
		}
	}
	buf.WriteString("\n")

	return nil
}
