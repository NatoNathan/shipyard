package template

import (
	"embed"
	"fmt"
	"strings"
)

// TemplateType represents the type/purpose of a template
type TemplateType string

const (
	TemplateTypeChangelog    TemplateType = "changelog"
	TemplateTypeTag          TemplateType = "tag"
	TemplateTypeRelease      TemplateType = "release"
	TemplateTypeReleaseNotes TemplateType = "releasenotes"
	TemplateTypeCommit       TemplateType = "commit"
)

//go:embed builtin/**/*.tmpl
var builtinTemplates embed.FS

// GetBuiltinTemplate retrieves a builtin template by type and name
func GetBuiltinTemplate(templateType TemplateType, name string) (string, error) {
	path := fmt.Sprintf("builtin/%s/%s.tmpl", templateType, name)

	content, err := builtinTemplates.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("builtin template not found: %s/%s", templateType, name)
	}

	return string(content), nil
}

// ListBuiltinTemplates lists available builtin templates for a given type
func ListBuiltinTemplates(templateType TemplateType) ([]string, error) {
	dir := fmt.Sprintf("builtin/%s", templateType)

	entries, err := builtinTemplates.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates for type %s: %w", templateType, err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			// Remove .tmpl extension
			name := strings.TrimSuffix(entry.Name(), ".tmpl")
			names = append(names, name)
		}
	}

	return names, nil
}

// GetBuiltinChangelogTemplate retrieves a builtin changelog template by name
func GetBuiltinChangelogTemplate(name string) (string, error) {
	return GetBuiltinTemplate(TemplateTypeChangelog, name)
}

// GetBuiltinTagTemplate retrieves a builtin tag template by name
func GetBuiltinTagTemplate(name string) (string, error) {
	return GetBuiltinTemplate(TemplateTypeTag, name)
}

// GetBuiltinReleaseTemplate retrieves a builtin release template by name
func GetBuiltinReleaseTemplate(name string) (string, error) {
	return GetBuiltinTemplate(TemplateTypeRelease, name)
}

// GetBuiltinReleaseNotesTemplate retrieves a builtin release notes template by name
func GetBuiltinReleaseNotesTemplate(name string) (string, error) {
	return GetBuiltinTemplate(TemplateTypeReleaseNotes, name)
}

// GetBuiltinCommitTemplate retrieves a builtin commit message template by name
func GetBuiltinCommitTemplate(name string) (string, error) {
	return GetBuiltinTemplate(TemplateTypeCommit, name)
}

// GetDefaultChangelogTemplate returns the default changelog template
func GetDefaultChangelogTemplate() (string, error) {
	return GetBuiltinChangelogTemplate("default")
}

// GetDefaultTagTemplate returns the default tag template
func GetDefaultTagTemplate() (string, error) {
	return GetBuiltinTagTemplate("default")
}

// GetDefaultReleaseTemplate returns the default release template
func GetDefaultReleaseTemplate() (string, error) {
	return GetBuiltinReleaseTemplate("date")
}

// GetDefaultReleaseNotesTemplate returns the default release notes template
func GetDefaultReleaseNotesTemplate() (string, error) {
	return GetBuiltinReleaseNotesTemplate("default")
}

// GetDefaultCommitTemplate returns the default commit message template
func GetDefaultCommitTemplate() (string, error) {
	return GetBuiltinCommitTemplate("default")
}


// GetAllBuiltinTemplates returns a map of all builtin templates (used for testing/docs)
// Returns map[type][name]content
func GetAllBuiltinTemplates() (map[TemplateType]map[string]string, error) {
	result := make(map[TemplateType]map[string]string)

	types := []TemplateType{
		TemplateTypeChangelog,
		TemplateTypeTag,
		TemplateTypeRelease,
		TemplateTypeReleaseNotes,
		TemplateTypeCommit,
	}

	for _, templateType := range types {
		names, err := ListBuiltinTemplates(templateType)
		if err != nil {
			return nil, err
		}

		result[templateType] = make(map[string]string)
		for _, name := range names {
			content, err := GetBuiltinTemplate(templateType, name)
			if err != nil {
				return nil, err
			}
			result[templateType][name] = content
		}
	}

	return result, nil
}
