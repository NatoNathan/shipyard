// Package templates provides template rendering functionality for Shipyard.
package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// TemplateData represents the data available to templates
type TemplateData struct {
	// Repository information
	RepoURL   string
	RepoOwner string
	RepoName  string
	RepoType  string // "monorepo" or "single-repo"

	// Package information
	Package  string   // Current package name
	Packages []string // All package names

	// Shipment information (version being released)
	Version     string            // Current version being shipped
	Versions    map[string]string // All package versions in this shipment
	PrevVersion string            // Previous shipped version

	// Consignment information (changes being shipped)
	ChangeType           string   // The change type (patch, minor, major, etc.)
	ChangeTypes          []string // All change types in this shipment
	ConsignmentSummary   string   // Summary of consignment changes
	ConsignmentSummaries []string // All consignment summaries in this shipment

	// Date information
	Date         string // Current date in YYYY-MM-DD format
	DateTime     string // Current date and time
	Timestamp    int64  // Unix timestamp
	ShipmentDate string // Date when shipment is being created

	// Git information
	GitHash   string // Current git commit hash
	GitBranch string // Current git branch

	// Custom fields
	Custom map[string]interface{} // Custom template variables
}

// TagTemplateData represents data specifically for git tag templates
type TagTemplateData struct {
	TemplateData
	TagName string // The tag name being created
}

// CommitTemplateData represents data specifically for git commit templates
type CommitTemplateData struct {
	TemplateData
	CommitType   string   // Type of commit (release, etc.)
	FilesChanged []string // List of files being committed
}

// ChangelogTemplateData represents data specifically for changelog templates
type ChangelogTemplateData struct {
	TemplateData
	ShipmentEntries       []ChangelogEntry             // All shipment entries in the changelog
	CurrentShipment       *ChangelogEntry              // Current shipment entry being rendered
	ConsignmentsBySection map[string][]ChangelogChange // Consignments grouped by section
}

// ChangelogEntry represents a single shipment entry in the changelog
type ChangelogEntry struct {
	Version      string
	Date         string                       // Formatted date
	DateTime     string                       // Full datetime
	ShipmentDate string                       // Date when shipment was created
	Changes      map[string][]ChangelogChange // section -> list of consignment changes
	PackageName  string                       // for monorepo projects
	ShipmentID   string                       // ID of the shipment
}

// ChangelogChange represents a single change in a changelog entry
type ChangelogChange struct {
	Summary       string // Summary of the consignment change
	ChangeType    string // e.g., "patch", "minor", "major"
	Section       string // e.g., "Added", "Fixed", "Changed"
	PackageName   string // for monorepo projects
	ConsignmentID string // ID of the original consignment
}

// DefaultTagTemplate returns the default git tag template
func DefaultTagTemplate(repoType config.RepoType) string {
	if repoType == config.RepositoryTypeMonorepo {
		return "{{.Package}}/v{{.Version}}"
	}
	return "v{{.Version}}"
}

// DefaultCommitTemplate returns the default git commit message template
func DefaultCommitTemplate(repoType config.RepoType) string {
	if repoType == config.RepositoryTypeMonorepo {
		return "chore: release {{range $pkg, $ver := .Versions}}{{$pkg}} v{{$ver}} {{end}}"
	}
	return "chore: release v{{.Version}}"
}

// RenderChangelogTemplate renders a changelog using the specified template
func (e *TemplateEngine) RenderChangelogTemplate(entries []ChangelogEntry, templateName string) (string, error) {
	// Get the template content
	templateContent, err := e.getTemplateContent(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template content: %w", err)
	}

	// Create template data
	data := &ChangelogTemplateData{
		TemplateData:    *e.createBaseTemplateData(nil),
		ShipmentEntries: entries,
	}

	// Render the template
	return e.renderTemplate("changelog", templateContent, data)
}

// getTemplateContent gets template content from various sources (built-in, file, URL)
func (e *TemplateEngine) getTemplateContent(templateName string) (string, error) {
	// Check if it's a built-in template
	if builtInTemplate := DefaultChangelogTemplate(templateName); builtInTemplate != "" {
		return builtInTemplate, nil
	}

	// Check if it's a file path
	if strings.Contains(templateName, "/") || strings.Contains(templateName, "\\") {
		return e.loadTemplateFromFile(templateName)
	}

	// Check if it's a URL
	if strings.HasPrefix(templateName, "http://") || strings.HasPrefix(templateName, "https://") {
		return e.loadTemplateFromURL(templateName)
	}

	// Check if it's a GitHub reference (github:owner/repo/path)
	if strings.HasPrefix(templateName, "github:") {
		return e.loadTemplateFromGitHub(templateName)
	}

	return "", fmt.Errorf("unknown template: %s", templateName)
}

// loadTemplateFromFile loads a template from a local file
func (e *TemplateEngine) loadTemplateFromFile(filePath string) (string, error) {
	// Handle relative paths
	if !filepath.IsAbs(filePath) {
		// Make relative to .shipyard directory
		filePath = filepath.Join(".shipyard", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", filePath, err)
	}

	return string(content), nil
}

// loadTemplateFromURL loads a template from a URL with local caching
func (e *TemplateEngine) loadTemplateFromURL(url string) (string, error) {
	// Create cache directory
	cacheDir := filepath.Join(".shipyard", "cache", "templates")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache file name from URL
	cacheFile := filepath.Join(cacheDir, hashURL(url)+".tmpl")

	// Check if cached version exists and is recent (within 24 hours)
	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			content, err := os.ReadFile(cacheFile)
			if err == nil {
				return string(content), nil
			}
		}
	}

	// Download template (placeholder for actual HTTP implementation)
	return "", fmt.Errorf("HTTP template loading not implemented yet for URL: %s", url)
}

// loadTemplateFromGitHub loads a template from GitHub with local caching
func (e *TemplateEngine) loadTemplateFromGitHub(githubRef string) (string, error) {
	// Parse github:owner/repo/path format
	parts := strings.SplitN(strings.TrimPrefix(githubRef, "github:"), "/", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid GitHub reference format: %s (expected github:owner/repo/path)", githubRef)
	}

	owner, repo, path := parts[0], parts[1], parts[2]

	// Create cache directory
	cacheDir := filepath.Join(".shipyard", "cache", "templates", owner, repo)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache file path
	cacheFile := filepath.Join(cacheDir, strings.ReplaceAll(path, "/", "_"))

	// Check if cached version exists and is recent (within 24 hours)
	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			content, err := os.ReadFile(cacheFile)
			if err == nil {
				return string(content), nil
			}
		}
	}

	// Download from GitHub (placeholder for actual GitHub API implementation)
	return "", fmt.Errorf("GitHub template loading not implemented yet for reference: %s", githubRef)
}

// hashURL creates a hash of a URL for cache file naming
func hashURL(url string) string {
	// Simple hash for demo - in production, use a proper hash function
	hash := 0
	for _, char := range url {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// DefaultChangelogTemplate returns the default changelog template based on template name
func DefaultChangelogTemplate(templateName string) string {
	switch templateName {
	case "keepachangelog":
		return keepAChangelogTemplate
	case "conventional":
		return conventionalTemplate
	case "simple":
		return simpleTemplate
	default:
		return keepAChangelogTemplate
	}
}

// Built-in changelog templates
const keepAChangelogTemplate = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

{{range .ShipmentEntries}}{{if eq $.RepoType "monorepo"}}## [{{.Version}}] - {{.PackageName}} - {{.Date}}
{{else}}## [{{.Version}}] - {{.Date}}
{{end}}
{{$sections := list "Breaking Changes" "Added" "Changed" "Deprecated" "Removed" "Fixed" "Security"}}{{range $section := $sections}}{{if index .Changes $section}}### {{$section}}

{{range index .Changes $section}}{{if eq $.RepoType "monorepo"}}+ **{{.PackageName}}**: {{.Summary}}
{{else}}+ {{.Summary}}
{{end}}{{end}}
{{end}}{{end}}
{{end}}`

const conventionalTemplate = `# Changelog

{{range .ShipmentEntries}}{{if eq $.RepoType "monorepo"}}## {{.Version}} ({{.PackageName}}) - {{.Date}}
{{else}}## {{.Version}} ({{.Date}})
{{end}}
{{$sections := list "Breaking Changes" "Added" "Fixed" "Changed" "Removed" "Security"}}{{range $section := $sections}}{{if index .Changes $section}}### {{$section}}

{{range index .Changes $section}}{{if eq $.RepoType "monorepo"}}+ **{{.PackageName}}**: {{.Summary}}
{{else}}+ {{.Summary}}
{{end}}{{end}}
{{end}}{{end}}
{{end}}`

const simpleTemplate = `# Changelog

{{range .ShipmentEntries}}{{if eq $.RepoType "monorepo"}}## {{.Version}} - {{.PackageName}} ({{.Date}})
{{else}}## {{.Version}} ({{.Date}})
{{end}}
{{range $type, $changes := .Changes}}{{range $changes}}{{if eq $.RepoType "monorepo"}}+ **{{.PackageName}}**: {{.Summary}}
{{else}}+ {{.Summary}}
{{end}}{{end}}{{end}}
{{end}}`

// TemplateEngine provides template rendering functionality
type TemplateEngine struct {
	projectConfig *config.ProjectConfig
	funcMap       template.FuncMap
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(projectConfig *config.ProjectConfig) *TemplateEngine {
	engine := &TemplateEngine{
		projectConfig: projectConfig,
		funcMap:       createFuncMap(),
	}
	return engine
}

// createFuncMap creates the function map for templates
func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    func(s string) string { return strings.ToUpper(s[:1]) + strings.ToLower(s[1:]) }, // Simple title case replacement
		"trim":     strings.TrimSpace,
		"replace":  strings.ReplaceAll,
		"contains": strings.Contains,
		"split":    strings.Split,
		"join":     strings.Join,
		"list": func(args ...interface{}) []interface{} {
			return args
		},
		"now": time.Now,
		"date": func(format string) string {
			return time.Now().Format(format)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}
}

// RenderTagTemplate renders a git tag template
func (e *TemplateEngine) RenderTagTemplate(packageName string, version *semver.Version, customData map[string]interface{}) (string, error) {
	// Use custom template or default
	templateStr := e.projectConfig.Git.TagTemplate
	if templateStr == "" {
		templateStr = DefaultTagTemplate(e.projectConfig.Type)
	}

	data := e.createTagTemplateData(packageName, version, customData)
	return e.renderTemplate("tag", templateStr, data)
}

// RenderCommitTemplate renders a git commit message template
func (e *TemplateEngine) RenderCommitTemplate(versions map[string]*semver.Version, summaries []string, customData map[string]interface{}) (string, error) {
	// Use custom template or default
	templateStr := e.projectConfig.Git.CommitTemplate
	if templateStr == "" {
		templateStr = DefaultCommitTemplate(e.projectConfig.Type)
	}

	data := e.createCommitTemplateData(versions, summaries, customData)
	return e.renderTemplate("commit", templateStr, data)
}

// createTagTemplateData creates template data for tag rendering
func (e *TemplateEngine) createTagTemplateData(packageName string, version *semver.Version, customData map[string]interface{}) *TagTemplateData {
	baseData := e.createBaseTemplateData(customData)
	baseData.Package = packageName
	baseData.Version = version.String()

	return &TagTemplateData{
		TemplateData: *baseData,
		TagName:      "", // Will be filled after rendering
	}
}

// createCommitTemplateData creates template data for commit message rendering
func (e *TemplateEngine) createCommitTemplateData(versions map[string]*semver.Version, summaries []string, customData map[string]interface{}) *CommitTemplateData {
	baseData := e.createBaseTemplateData(customData)

	// Convert versions map
	baseData.Versions = make(map[string]string)
	var changeTypes []string
	for pkg, ver := range versions {
		baseData.Versions[pkg] = ver.String()
	}

	baseData.ConsignmentSummaries = summaries
	baseData.ChangeTypes = changeTypes

	// For single repo, set the primary version
	if e.projectConfig.Type == config.RepositoryTypeSingleRepo {
		for _, ver := range versions {
			baseData.Version = ver.String()
			break
		}
	}

	return &CommitTemplateData{
		TemplateData: *baseData,
		CommitType:   "release",
		FilesChanged: []string{}, // Could be populated if needed
	}
}

// createBaseTemplateData creates the base template data
func (e *TemplateEngine) createBaseTemplateData(customData map[string]interface{}) *TemplateData {
	now := time.Now()

	// Parse repo URL for owner/name
	repoOwner, repoName := parseRepoURL(e.projectConfig.Repo)

	data := &TemplateData{
		RepoURL:   e.projectConfig.Repo,
		RepoOwner: repoOwner,
		RepoName:  repoName,
		RepoType:  string(e.projectConfig.Type),
		Packages:  e.projectConfig.GetPackageNames(),
		Date:      now.Format("2006-01-02"),
		DateTime:  now.Format("2006-01-02 15:04:05"),
		Timestamp: now.Unix(),
		Custom:    customData,
	}

	if customData == nil {
		data.Custom = make(map[string]interface{})
	}

	return data
}

// renderTemplate renders a template with the given data
func (e *TemplateEngine) renderTemplate(name, templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New(name).Funcs(e.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// parseRepoURL extracts owner and name from a repository URL
func parseRepoURL(repoURL string) (owner, name string) {
	// Handle different URL formats
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "http://")
	repoURL = strings.TrimPrefix(repoURL, "git@")
	repoURL = strings.ReplaceAll(repoURL, ":", "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	parts := strings.Split(repoURL, "/")
	if len(parts) >= 3 {
		// github.com/owner/repo format
		owner = parts[len(parts)-2]
		name = parts[len(parts)-1]
	} else if len(parts) == 2 {
		// owner/repo format
		owner = parts[0]
		name = parts[1]
	}

	return owner, name
}

// ValidateTemplate validates a template string
func (e *TemplateEngine) ValidateTemplate(templateStr string) error {
	_, err := template.New("validation").Funcs(e.funcMap).Parse(templateStr)
	return err
}
