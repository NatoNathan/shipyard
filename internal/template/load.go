package template

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SourceType represents the type of template source
type SourceType int

const (
	SourceTypeBuiltin SourceType = iota
	SourceTypeFile
	SourceTypeGit
	SourceTypeHTTPS
	SourceTypeInline
)

// TemplateLoader handles loading templates from various sources
type TemplateLoader struct {
	baseDir   string
	cache     map[string]string
	authToken string
	timeout   time.Duration
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader() *TemplateLoader {
	return &TemplateLoader{
		cache:   make(map[string]string),
		timeout: 30 * time.Second,
	}
}

// SetBaseDir sets the base directory for resolving relative file paths
func (l *TemplateLoader) SetBaseDir(dir string) {
	l.baseDir = dir
}

// SetAuthToken sets the authentication token for remote sources
func (l *TemplateLoader) SetAuthToken(token string) {
	l.authToken = token
}

// GetAuthToken returns the current authentication token
func (l *TemplateLoader) GetAuthToken() string {
	return l.authToken
}

// SetTimeout sets the timeout for remote operations
func (l *TemplateLoader) SetTimeout(timeout time.Duration) {
	l.timeout = timeout
}

// Load loads a template from the specified source
// For builtin templates, expectedType specifies which type directory to search
func (l *TemplateLoader) Load(source string, expectedType ...TemplateType) (string, error) {
	// Determine expected type (default to empty for non-builtin)
	var templateType TemplateType
	if len(expectedType) > 0 {
		templateType = expectedType[0]
	}

	// Create cache key including type for builtins
	cacheKey := source
	if templateType != "" {
		cacheKey = fmt.Sprintf("%s:%s", templateType, source)
	}

	// Check cache first
	if cached, ok := l.cache[cacheKey]; ok {
		return cached, nil
	}

	sourceType, target := DetectSourceType(source)

	var content string
	var err error

	switch sourceType {
	case SourceTypeBuiltin:
		if templateType == "" {
			return "", fmt.Errorf("builtin templates require expectedType parameter")
		}
		content, err = l.loadBuiltin(target, templateType)
	case SourceTypeFile:
		content, err = l.loadFile(target)
	case SourceTypeHTTPS:
		content, err = l.loadHTTPS(target)
	case SourceTypeGit:
		content, err = l.loadGit(target)
	case SourceTypeInline:
		content = target // Inline content is the target itself
	default:
		return "", fmt.Errorf("unknown source type for: %s", source)
	}

	if err != nil {
		return "", err
	}

	// Cache the result
	l.cache[cacheKey] = content

	return content, nil
}

// LoadInline loads an inline template (no source prefix parsing)
func (l *TemplateLoader) LoadInline(content string) (string, error) {
	return content, nil
}

// DetectSourceType detects the template source type and extracts the target
func DetectSourceType(source string) (SourceType, string) {
	if source == "" {
		return SourceTypeInline, ""
	}

	// Check for explicit prefixes
	if strings.HasPrefix(source, "builtin:") {
		return SourceTypeBuiltin, strings.TrimPrefix(source, "builtin:")
	}

	if strings.HasPrefix(source, "file:") {
		return SourceTypeFile, strings.TrimPrefix(source, "file:")
	}

	if strings.HasPrefix(source, "git:") {
		return SourceTypeGit, strings.TrimPrefix(source, "git:")
	}

	if strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return SourceTypeHTTPS, source
	}

	// If contains newlines, treat as inline
	if strings.Contains(source, "\n") {
		return SourceTypeInline, source
	}

	// Otherwise, treat as file path (implicit file:)
	return SourceTypeFile, source
}

// loadBuiltin loads a built-in template of the specified type
func (l *TemplateLoader) loadBuiltin(name string, templateType TemplateType) (string, error) {
	// Remove builtin: prefix if present
	name = strings.TrimPrefix(name, "builtin:")

	// Load from the specified type directory
	return GetBuiltinTemplate(templateType, name)
}

// loadFile loads a template from a file
func (l *TemplateLoader) loadFile(path string) (string, error) {
	// If path is relative and base dir is set, resolve it
	if !filepath.IsAbs(path) && l.baseDir != "" {
		path = filepath.Join(l.baseDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	return string(content), nil
}

// loadHTTPS loads a template from an HTTPS URL
func (l *TemplateLoader) loadHTTPS(url string) (string, error) {
	client := &http.Client{
		Timeout: l.timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is set
	if l.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+l.authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch template: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}

// loadGit loads a template from a git repository
// Format: git:https://github.com/user/repo.git#path/to/template@branch
func (l *TemplateLoader) loadGit(source string) (string, error) {
	// Parse git source
	gitURL, path, ref := parseGitSource(source)

	if gitURL == "" {
		return "", fmt.Errorf("invalid git source format: %s", source)
	}

	// For now, return error indicating git support is not yet implemented
	// Full implementation would:
	// 1. Clone or fetch the repository
	// 2. Checkout the specified ref
	// 3. Read the file at the specified path
	return "", fmt.Errorf("git template loading not yet implemented (URL: %s, path: %s, ref: %s)", gitURL, path, ref)
}

// parseGitSource parses a git source string into components
// Format: https://github.com/user/repo.git#path/to/file@branch
func parseGitSource(source string) (gitURL, path, ref string) {
	// Split on # to separate URL from path+ref
	parts := strings.SplitN(source, "#", 2)
	gitURL = parts[0]

	if len(parts) < 2 {
		// No path specified
		return gitURL, "", "main"
	}

	// Split path and ref on @
	pathRef := parts[1]
	pathRefParts := strings.SplitN(pathRef, "@", 2)
	path = pathRefParts[0]

	if len(pathRefParts) > 1 {
		ref = pathRefParts[1]
	} else {
		ref = "main" // Default branch
	}

	return gitURL, path, ref
}

