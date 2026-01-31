package consignment

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Serialize converts a consignment to markdown with YAML frontmatter
func Serialize(cons *Consignment) (string, error) {
	// Create a struct for frontmatter (excludes Summary)
	type Frontmatter struct {
		ID         string                 `yaml:"id"`
		Timestamp  string                 `yaml:"timestamp"`
		Packages   []string               `yaml:"packages"`
		ChangeType string                 `yaml:"changeType"`
		Metadata   map[string]interface{} `yaml:"metadata,omitempty"`
	}

	frontmatter := Frontmatter{
		ID:         cons.ID,
		Timestamp:  cons.Timestamp.Format("2006-01-02T15:04:05Z"),
		Packages:   cons.Packages,
		ChangeType: string(cons.ChangeType),
		Metadata:   cons.Metadata,
	}

	// Marshal frontmatter to YAML
	yamlBytes, err := yaml.Marshal(&frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Build final markdown with frontmatter
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.Write(yamlBytes)
	builder.WriteString("---\n")
	builder.WriteString("\n")
	builder.WriteString(cons.Summary)
	builder.WriteString("\n")

	return builder.String(), nil
}
