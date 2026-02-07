package consignment

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

// ReadConsignment reads and parses a consignment file from the given path
// Returns a Consignment struct with parsed YAML frontmatter and markdown body
func ReadConsignment(path string) (*Consignment, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read consignment file: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("consignment file is empty: %s", path)
	}

	// Parse markdown with frontmatter using goldmark
	md := goldmark.New(
		goldmark.WithExtensions(meta.Meta),
	)

	var buf bytes.Buffer
	context := parser.NewContext()

	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		return nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Extract frontmatter metadata
	metaData := meta.Get(context)
	if metaData == nil {
		return nil, fmt.Errorf("no frontmatter found in consignment file: %s", path)
	}

	// Parse frontmatter into Consignment struct
	var c Consignment

	// Marshal metadata back to YAML for proper type handling
	yamlData, err := yaml.Marshal(metaData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Unmarshal into structured Consignment
	if err := yaml.Unmarshal(yamlData, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consignment: %w", err)
	}

	// Validate required fields
	if c.ID == "" {
		return nil, fmt.Errorf("missing required field: id")
	}
	if len(c.Packages) == 0 {
		return nil, fmt.Errorf("missing required field: packages")
	}
	if c.ChangeType == "" {
		return nil, fmt.Errorf("missing required field: changeType")
	}
	if c.Timestamp.IsZero() {
		return nil, fmt.Errorf("missing or invalid required field: timestamp")
	}

	// Validate changeType enum
	validTypes := map[types.ChangeType]bool{
		types.ChangeTypePatch: true,
		types.ChangeTypeMinor: true,
		types.ChangeTypeMajor: true,
	}
	if !validTypes[c.ChangeType] {
		return nil, fmt.Errorf("invalid changeType: %s (must be patch, minor, or major)", c.ChangeType)
	}

	// Extract markdown body (everything after frontmatter)
	body := extractMarkdownBody(string(content))
	c.Summary = strings.TrimSpace(body)

	if c.Summary == "" {
		return nil, fmt.Errorf("consignment summary cannot be empty")
	}

	return &c, nil
}

// ParseError represents a failure to parse a single consignment file
type ParseError struct {
	File    string
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s: %s", e.File, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ReadAllConsignments reads all consignment files from a directory
// Returns a slice of Consignment structs sorted by timestamp (oldest first)
// Parse errors are logged to stderr but do not cause the function to fail
func ReadAllConsignments(consignmentDir string) ([]*Consignment, error) {
	consignments, parseErrors, err := ReadAllConsignmentsWithErrors(consignmentDir)
	if err != nil {
		return nil, err
	}

	for _, pe := range parseErrors {
		fmt.Fprintf(os.Stderr, "Warning: skipping invalid consignment %s: %v\n", pe.File, pe.Err)
	}

	return consignments, nil
}

// ReadAllConsignmentsWithErrors reads all consignments and returns both successful
// parses and any parse errors (instead of failing on first error)
func ReadAllConsignmentsWithErrors(dir string) ([]*Consignment, []ParseError, error) {
	return readAllConsignmentsInternal(dir, nil)
}

// ReadAllConsignmentsFiltered reads consignments and filters by package names
// If packageFilter is nil or empty, returns all consignments
func ReadAllConsignmentsFiltered(consignmentDir string, packageFilter []string) ([]*Consignment, error) {
	consignments, parseErrors, err := readAllConsignmentsInternal(consignmentDir, packageFilter)
	if err != nil {
		return nil, err
	}

	for _, pe := range parseErrors {
		fmt.Fprintf(os.Stderr, "Warning: skipping invalid consignment %s: %v\n", pe.File, pe.Err)
	}

	return consignments, nil
}

// readAllConsignmentsInternal is the shared implementation for reading consignments
func readAllConsignmentsInternal(consignmentDir string, packageFilter []string) ([]*Consignment, []ParseError, error) {
	// Check if directory exists
	if _, err := os.Stat(consignmentDir); os.IsNotExist(err) {
		return []*Consignment{}, nil, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(consignmentDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read consignment directory: %w", err)
	}

	var consignments []*Consignment
	var parseErrors []ParseError

	// Process each markdown file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .md files
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(consignmentDir, entry.Name())

		// Read and parse consignment
		c, err := ReadConsignment(filePath)
		if err != nil {
			parseErrors = append(parseErrors, ParseError{
				File:    entry.Name(),
				Message: err.Error(),
				Err:     err,
			})
			continue
		}

		// Apply package filter if specified
		if len(packageFilter) > 0 {
			if !containsAnyPackage(c.Packages, packageFilter) {
				continue
			}
		}

		consignments = append(consignments, c)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(consignments, func(i, j int) bool {
		return consignments[i].Timestamp.Before(consignments[j].Timestamp)
	})

	return consignments, parseErrors, nil
}

// extractMarkdownBody extracts the markdown content after the YAML frontmatter
func extractMarkdownBody(content string) string {
	// Find the end of frontmatter (second occurrence of ---)
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return content // No frontmatter, return as-is
	}

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			// Return everything after the closing ---
			body := strings.Join(lines[i+1:], "\n")
			return strings.TrimSpace(body)
		}
	}

	return content // Malformed frontmatter, return as-is
}

// containsAnyPackage checks if any package in the consignment matches the filter
func containsAnyPackage(consignmentPackages []string, filter []string) bool {
	filterSet := make(map[string]bool)
	for _, pkg := range filter {
		filterSet[pkg] = true
	}

	for _, pkg := range consignmentPackages {
		if filterSet[pkg] {
			return true
		}
	}

	return false
}

