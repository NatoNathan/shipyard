package consignment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConsignment(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantID      string
		wantPackages []string
		wantType    string
		wantSummary string
		wantErr     bool
	}{
		{
			name: "valid consignment with all fields",
			content: `---
id: "20260130-143022-a1b2c3"
timestamp: "2026-01-30T14:30:22Z"
packages:
  - "core"
  - "api-client"
changeType: "minor"
metadata:
  author: "dev@example.com"
  issue: "JIRA-123"
---

# Added OAuth2 support

This change implements OAuth2 authentication.

## Details

- Implemented OAuth2 client
- Added token refresh logic
`,
			wantID:       "20260130-143022-a1b2c3",
			wantPackages: []string{"core", "api-client"},
			wantType:     "minor",
			wantSummary:  "# Added OAuth2 support",
			wantErr:      false,
		},
		{
			name: "minimal consignment",
			content: `---
id: "20260130-150000-xyz123"
timestamp: "2026-01-30T15:00:00Z"
packages:
  - "core"
changeType: "patch"
---

Fixed bug in validation
`,
			wantID:       "20260130-150000-xyz123",
			wantPackages: []string{"core"},
			wantType:     "patch",
			wantSummary:  "Fixed bug in validation",
			wantErr:      false,
		},
		{
			name: "missing frontmatter",
			content: `This is just plain text without frontmatter`,
			wantErr: true,
		},
		{
			name: "invalid YAML frontmatter",
			content: `---
id: "test"
packages: invalid syntax here
---
Content`,
			wantErr: true,
		},
		{
			name: "missing required field (packages)",
			content: `---
id: "20260130-150000-xyz123"
timestamp: "2026-01-30T15:00:00Z"
changeType: "patch"
---
Content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with consignment content
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.md")
			require.NoError(t, os.WriteFile(filePath, []byte(tt.content), 0644))

			// Read consignment
			consignment, err := ReadConsignment(filePath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, consignment.ID)
			assert.Equal(t, tt.wantPackages, consignment.Packages)
			assert.Equal(t, tt.wantType, string(consignment.ChangeType))
			assert.Contains(t, consignment.Summary, tt.wantSummary)
		})
	}
}

func TestReadConsignmentWithMetadata(t *testing.T) {
	content := `---
id: "test-id"
timestamp: "2026-01-30T14:30:22Z"
packages:
  - "core"
changeType: "minor"
metadata:
  author: "alice@example.com"
  issue: "FEAT-123"
  breaking: false
  tags:
    - "authentication"
    - "security"
---
Content here
`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	consignment, err := ReadConsignment(filePath)
	require.NoError(t, err)

	// Check metadata
	assert.Equal(t, "alice@example.com", consignment.Metadata["author"])
	assert.Equal(t, "FEAT-123", consignment.Metadata["issue"])
	assert.Equal(t, false, consignment.Metadata["breaking"])

	tags, ok := consignment.Metadata["tags"].([]interface{})
	require.True(t, ok, "tags should be array")
	assert.Len(t, tags, 2)
}

func TestReadConsignmentTimestampParsing(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		wantErr   bool
	}{
		{
			name:      "valid ISO 8601",
			timestamp: "2026-01-30T14:30:22Z",
			wantErr:   false,
		},
		{
			name:      "valid with timezone",
			timestamp: "2026-01-30T14:30:22-08:00",
			wantErr:   false,
		},
		{
			name:      "invalid format",
			timestamp: "2026/01/30 14:30:22",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `---
id: "test-id"
timestamp: "` + tt.timestamp + `"
packages:
  - "core"
changeType: "patch"
---
Content
`
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.md")
			require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

			consignment, err := ReadConsignment(filePath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.False(t, consignment.Timestamp.IsZero())
		})
	}
}

func TestReadAllConsignments(t *testing.T) {
	tmpDir := t.TempDir()
	consignmentDir := filepath.Join(tmpDir, ".shipyard", "consignments")
	require.NoError(t, os.MkdirAll(consignmentDir, 0755))

	// Create multiple consignments
	consignments := []string{
		"20260130-143022-a1b2c3.md",
		"20260130-150000-xyz123.md",
		"20260130-160000-def456.md",
	}

	for _, name := range consignments {
		content := `---
id: "` + name[:len(name)-3] + `"
timestamp: "2026-01-30T14:30:22Z"
packages:
  - "core"
changeType: "patch"
---
Test content for ` + name
		filePath := filepath.Join(consignmentDir, name)
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	// Read all consignments
	result, err := ReadAllConsignments(consignmentDir)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify consignments are sorted by timestamp (or ID)
	assert.Equal(t, consignments[0][:len(consignments[0])-3], result[0].ID)
}

func TestReadAllConsignmentsWithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	consignmentDir := filepath.Join(tmpDir, ".shipyard", "consignments")
	require.NoError(t, os.MkdirAll(consignmentDir, 0755))

	// Create consignments for different packages
	testCases := []struct {
		filename string
		packages []string
	}{
		{"c1.md", []string{"core"}},
		{"c2.md", []string{"api"}},
		{"c3.md", []string{"core", "api"}},
		{"c4.md", []string{"web"}},
	}

	for _, tc := range testCases {
		content := `---
id: "` + tc.filename[:len(tc.filename)-3] + `"
timestamp: "2026-01-30T14:30:22Z"
packages:`
		for _, pkg := range tc.packages {
			content += "\n  - \"" + pkg + "\""
		}
		content += `
changeType: "patch"
---
Content`
		filePath := filepath.Join(consignmentDir, tc.filename)
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	// Read with package filter
	result, err := ReadAllConsignmentsFiltered(consignmentDir, []string{"core"})
	require.NoError(t, err)

	// Should include c1 (core only) and c3 (core+api)
	assert.Len(t, result, 2)
}

func TestReadConsignmentErrorHandling(t *testing.T) {
	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ReadConsignment("/nonexistent/path.md")
		assert.Error(t, err)
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := ReadConsignment(tmpDir)
		assert.Error(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.md")
		require.NoError(t, os.WriteFile(filePath, []byte(""), 0644))

		_, err := ReadConsignment(filePath)
		assert.Error(t, err)
	})
}
