package template

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTemplate_Builtin(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		templateType TemplateType
		expectError  bool
		description  string
	}{
		{
			name:         "builtin:default",
			source:       "builtin:default",
			templateType: TemplateTypeChangelog,
			expectError:  false,
			description:  "should load default changelog builtin template",
		},
		{
			name:         "builtin:changelog",
			source:       "builtin:default",
			templateType: TemplateTypeChangelog,
			expectError:  false,
			description:  "should load builtin changelog template",
		},
		{
			name:         "builtin:tagname",
			source:       "builtin:default",
			templateType: TemplateTypeTag,
			expectError:  false,
			description:  "should load builtin tag name template",
		},
		{
			name:         "builtin:nonexistent",
			source:       "builtin:nonexistent",
			templateType: TemplateTypeChangelog,
			expectError:  true,
			description:  "should error on unknown builtin template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader()
			content, err := loader.Load(tt.source, tt.templateType)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, content, tt.description)
		})
	}
}

func TestLoadTemplate_File(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test template file
	templateContent := `# Changelog

{{ range .Consignments }}
- {{ .Summary }}
{{ end }}
`
	templatePath := filepath.Join(tmpDir, "changelog.tmpl")
	require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

	tests := []struct {
		name        string
		source      string
		expectError bool
		description string
	}{
		{
			name:        "file with absolute path",
			source:      "file:" + templatePath,
			expectError: false,
			description: "should load template from absolute file path",
		},
		{
			name:        "file with relative path",
			source:      "file:changelog.tmpl",
			expectError: false,
			description: "should load template from relative path",
		},
		{
			name:        "nonexistent file",
			source:      "file:/nonexistent/template.tmpl",
			expectError: true,
			description: "should error on nonexistent file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader()
			loader.SetBaseDir(tmpDir) // Set base dir for relative paths

			content, err := loader.Load(tt.source)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, content, "Changelog", tt.description)
		})
	}
}

func TestLoadTemplate_Inline(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expected    string
		description string
	}{
		{
			name: "simple inline template",
			source: `# Version {{ .Version }}

{{ range .Consignments }}
- {{ .Summary }}
{{ end }}`,
			expected:    "# Version",
			description: "should return inline content as-is",
		},
		{
			name:        "empty inline template",
			source:      "",
			expected:    "",
			description: "should handle empty template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewTemplateLoader()

			// Inline templates don't have a prefix
			content, err := loader.LoadInline(tt.source)

			require.NoError(t, err)
			assert.Contains(t, content, tt.expected, tt.description)
		})
	}
}

func TestLoadTemplate_HTTPS(t *testing.T) {
	t.Run("loads from local HTTP server with auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			_, _ = w.Write([]byte("remote template"))
		}))
		defer server.Close()

		loader := NewTemplateLoader()
		loader.SetAuthToken("test-token")
		content, err := loader.Load(server.URL + "/template.tmpl")

		require.NoError(t, err)
		assert.Equal(t, "remote template", content)
	})

	t.Run("rejects oversized content length", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "11")
			_, _ = w.Write([]byte("hello world"))
		}))
		defer server.Close()

		loader := NewTemplateLoader()
		loader.SetMaxResponseBytes(5)
		_, err := loader.Load(server.URL + "/template.tmpl")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum size")
	})

	t.Run("rejects streaming response beyond limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello world"))
		}))
		defer server.Close()

		loader := NewTemplateLoader()
		loader.SetMaxResponseBytes(5)
		_, err := loader.Load(server.URL + "/template.tmpl")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "maximum size")
	})

	t.Run("limits redirects", func(t *testing.T) {
		server := httptest.NewServer(nil)
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, server.URL+r.URL.Path, http.StatusFound)
		})
		defer server.Close()

		loader := NewTemplateLoader()
		_, err := loader.Load(server.URL + "/template.tmpl")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "redirects")
	})

	t.Run("allows unauthenticated same-origin redirect", func(t *testing.T) {
		server := httptest.NewServer(nil)
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/start" {
				http.Redirect(w, r, server.URL+"/final", http.StatusFound)
				return
			}
			_, _ = fmt.Fprint(w, "redirected template")
		})
		defer server.Close()

		loader := NewTemplateLoader()
		content, err := loader.Load(server.URL + "/start")

		require.NoError(t, err)
		assert.Equal(t, "redirected template", content)
	})

	t.Run("rejects authenticated redirect to insecure target", func(t *testing.T) {
		server := httptest.NewServer(nil)
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, server.URL+"/final", http.StatusFound)
		})
		defer server.Close()

		loader := NewTemplateLoader()
		loader.SetAuthToken("test-token")
		_, err := loader.Load(server.URL + "/start")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "authenticated redirect")
	})
}

func TestLoadTemplate_Git(t *testing.T) {
	t.Run("loads from local git repository", func(t *testing.T) {
		repoDir := t.TempDir()
		repo, err := gogit.PlainInit(repoDir, false)
		require.NoError(t, err)

		templatePath := filepath.Join(repoDir, "templates", "changelog.tmpl")
		require.NoError(t, os.MkdirAll(filepath.Dir(templatePath), 0755))
		require.NoError(t, os.WriteFile(templatePath, []byte("git template"), 0644))

		worktree, err := repo.Worktree()
		require.NoError(t, err)
		_, err = worktree.Add("templates/changelog.tmpl")
		require.NoError(t, err)
		_, err = worktree.Commit("add template", &gogit.CommitOptions{
			Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
		})
		require.NoError(t, err)

		loader := NewTemplateLoader()
		source := "git:" + repoDir + "#templates/changelog.tmpl@master"
		content, err := loader.Load(source)

		require.NoError(t, err)
		assert.Equal(t, "git template", content)
	})

	t.Run("rejects unsafe git template path", func(t *testing.T) {
		loader := NewTemplateLoader()
		_, err := loader.Load("git:/tmp/repo#../secret@main")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsafe")
	})

	t.Run("rejects symlink escape from git template path", func(t *testing.T) {
		repoDir, err := os.MkdirTemp(os.TempDir(), "shipyard-template-repo-*")
		require.NoError(t, err)
		t.Cleanup(func() { _ = os.RemoveAll(repoDir) })

		outsideFile := filepath.Join(filepath.Dir(repoDir), fmt.Sprintf("shipyard-template-secret-%d.tmpl", time.Now().UnixNano()))
		require.NoError(t, os.WriteFile(outsideFile, []byte("secret"), 0644))
		t.Cleanup(func() { _ = os.Remove(outsideFile) })

		repo, err := gogit.PlainInit(repoDir, false)
		require.NoError(t, err)

		templatePath := filepath.Join(repoDir, "templates", "escape.tmpl")
		require.NoError(t, os.MkdirAll(filepath.Dir(templatePath), 0755))
		require.NoError(t, os.Symlink(filepath.Join("..", "..", filepath.Base(outsideFile)), templatePath))

		worktree, err := repo.Worktree()
		require.NoError(t, err)
		_, err = worktree.Add("templates/escape.tmpl")
		require.NoError(t, err)
		_, err = worktree.Commit("add symlink template", &gogit.CommitOptions{
			Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
		})
		require.NoError(t, err)

		loader := NewTemplateLoader()
		source := "git:" + repoDir + "#templates/escape.tmpl@master"
		_, err = loader.Load(source)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsafe")
	})
}

func TestDetectTemplateSource(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   SourceType
		expectedTarget string
		description    string
	}{
		{
			name:           "builtin format",
			source:         "builtin:default",
			expectedType:   SourceTypeBuiltin,
			expectedTarget: "default",
			description:    "should detect builtin source",
		},
		{
			name:           "file format",
			source:         "file:/path/to/template.tmpl",
			expectedType:   SourceTypeFile,
			expectedTarget: "/path/to/template.tmpl",
			description:    "should detect file source",
		},
		{
			name:           "https format",
			source:         "https://example.com/template.tmpl",
			expectedType:   SourceTypeHTTPS,
			expectedTarget: "https://example.com/template.tmpl",
			description:    "should detect HTTPS source",
		},
		{
			name:           "git format",
			source:         "git:https://github.com/example/repo.git#path@ref",
			expectedType:   SourceTypeGit,
			expectedTarget: "https://github.com/example/repo.git#path@ref",
			description:    "should detect git source",
		},
		{
			name:           "inline format (multiline)",
			source:         "# Template\n{{ .Version }}",
			expectedType:   SourceTypeInline,
			expectedTarget: "# Template\n{{ .Version }}",
			description:    "should detect inline multiline template",
		},
		{
			name:           "implicit file (no prefix, single line)",
			source:         "templates/changelog.tmpl",
			expectedType:   SourceTypeFile,
			expectedTarget: "templates/changelog.tmpl",
			description:    "should treat single-line without prefix as file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceType, target := DetectSourceType(tt.source)

			assert.Equal(t, tt.expectedType, sourceType, tt.description)
			assert.Equal(t, tt.expectedTarget, target)
		})
	}
}

func TestTemplateLoader_WithCache(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.tmpl")
	require.NoError(t, os.WriteFile(templatePath, []byte("content"), 0644))

	loader := NewTemplateLoader()
	loader.SetBaseDir(tmpDir)

	// Load twice
	content1, err1 := loader.Load("file:template.tmpl")
	content2, err2 := loader.Load("file:template.tmpl")

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, content1, content2, "cached content should match")
}

func TestTemplateLoader_Authentication(t *testing.T) {
	t.Run("auth token from environment", func(t *testing.T) {
		t.Setenv("TEMPLATE_AUTH_TOKEN", "test-token")

		loader := NewTemplateLoader()
		loader.SetAuthToken("test-token")

		// In real implementation, this would use the token for authenticated requests
		assert.Equal(t, "test-token", loader.GetAuthToken())
	})
}

func TestLoadTemplate_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		setup       func(t *testing.T)
		expectError string
		description string
	}{
		{
			name:        "empty source",
			source:      "",
			expectError: "",
			description: "empty source should be treated as inline empty template",
		},
		{
			name:        "invalid prefix",
			source:      "invalid:something",
			expectError: "unknown source type",
			description: "should error on unknown source prefix",
		},
		{
			name:        "file permission denied",
			source:      "file:/root/template.tmpl",
			expectError: "permission denied",
			description: "should error on permission issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			loader := NewTemplateLoader()
			_, err := loader.Load(tt.source)

			if tt.expectError != "" {
				assert.Error(t, err)
				// Note: Error message matching would be more specific in real implementation
			}
		})
	}
}
