// Package changelog provides template interfaces and implementations for generating changelogs.
package changelog

import (
	"fmt"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// Template represents a changelog template
type Template interface {
	// Generate generates a changelog from entries
	Generate(entries []ChangelogEntry, projectConfig *config.ProjectConfig) (string, error)
	// Name returns the template name
	Name() string
	// Description returns the template description
	Description() string
}

// GetTemplate returns a template by name
func GetTemplate(name string) (Template, error) {
	switch name {
	case "keepachangelog":
		return &KeepAChangelogTemplate{}, nil
	case "conventional":
		return &ConventionalTemplate{}, nil
	case "simple":
		return &SimpleTemplate{}, nil
	default:
		return nil, fmt.Errorf("unknown changelog template: %s", name)
	}
}

// GetTemplateInfo returns information about all available templates
func GetTemplateInfo() []TemplateInfo {
	return []TemplateInfo{
		{
			Name:        "keepachangelog",
			Description: "Keep a Changelog format (https://keepachangelog.com/)",
		},
		{
			Name:        "conventional",
			Description: "Conventional Commits format with automatic categorization",
		},
		{
			Name:        "simple",
			Description: "Simple, minimal changelog format",
		},
	}
}

// TemplateInfo contains information about a changelog template
type TemplateInfo struct {
	Name        string
	Description string
}
