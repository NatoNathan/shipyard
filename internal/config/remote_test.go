package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const remoteBaseConfig = `
packages:
  - name: platform-core
    path: ./core
    ecosystem: go
templates:
  changelog:
    source: builtin:default
  tagName:
    inline: "{{.Package}}/v{{.Version}}"
metadata:
  fields:
    - name: author
      required: true
      type: string
`

func writeLocalConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".shipyard"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".shipyard", "shipyard.yaml"), []byte(content), 0644))
	return dir
}

func TestResolveExtends_HTTPSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, remoteBaseConfig)
	}))
	defer server.Close()

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
packages:
  - name: web-app
    path: ./apps/web
    ecosystem: npm
templates:
  tagName:
    inline: "frontend/{{.Package}}/v{{.Version}}"
metadata:
  fields:
    - name: reviewer
      type: string
`, server.URL))

	cfg, err := LoadFromDir(dir)
	require.NoError(t, err)

	// Packages appended: base first, then local
	require.Len(t, cfg.Packages, 2)
	assert.Equal(t, "platform-core", cfg.Packages[0].Name)
	assert.Equal(t, "web-app", cfg.Packages[1].Name)

	// Templates merged per key: changelog inherited, tagName overridden
	require.NotNil(t, cfg.Templates.Changelog)
	assert.Equal(t, "builtin:default", cfg.Templates.Changelog.Source)
	require.NotNil(t, cfg.Templates.TagName)
	assert.Equal(t, "frontend/{{.Package}}/v{{.Version}}", cfg.Templates.TagName.Inline)

	// Metadata fields appended
	require.Len(t, cfg.Metadata.Fields, 2)
	assert.Equal(t, "author", cfg.Metadata.Fields[0].Name)
	assert.Equal(t, "reviewer", cfg.Metadata.Fields[1].Name)
}

func TestResolveExtends_HTTPAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, remoteBaseConfig)
	}))
	defer server.Close()

	t.Setenv("SHIPYARD_TEST_CONFIG_TOKEN", "secret-token")

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - url: %q
    auth: "env:SHIPYARD_TEST_CONFIG_TOKEN"
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL))

	cfg, err := LoadFromDir(dir)
	require.NoError(t, err)
	assert.Len(t, cfg.Packages, 2)
}

func TestResolveExtends_HTTPAuthMissingEnv(t *testing.T) {
	dir := writeLocalConfig(t, `
extends:
  - url: "https://example.invalid/base.yaml"
    auth: "env:SHIPYARD_TEST_UNSET_TOKEN"
packages:
  - name: app
    path: ./
    ecosystem: go
`)

	_, err := LoadFromDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SHIPYARD_TEST_UNSET_TOKEN")
}

func TestResolveExtends_HTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL))

	_, err := LoadFromDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch remote config")
}

func TestResolveExtends_ChainedRemotes(t *testing.T) {
	// grandparent <- parent <- local
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/grandparent.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
packages:
  - name: org-core
    path: ./core
    ecosystem: go
`)
	})
	mux.HandleFunc("/parent.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
extends:
  - %q
packages:
  - name: team-lib
    path: ./lib
    ecosystem: go
`, server.URL+"/grandparent.yaml")
	})

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL+"/parent.yaml"))

	cfg, err := LoadFromDir(dir)
	require.NoError(t, err)
	require.Len(t, cfg.Packages, 3)
	assert.Equal(t, "org-core", cfg.Packages[0].Name)
	assert.Equal(t, "team-lib", cfg.Packages[1].Name)
	assert.Equal(t, "app", cfg.Packages[2].Name)
}

func TestResolveExtends_CircularDependency(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "extends:\n  - %q\n", server.URL+"/b.yaml")
	})
	mux.HandleFunc("/b.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "extends:\n  - %q\n", server.URL+"/a.yaml")
	})

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL+"/a.yaml"))

	_, err := LoadFromDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestResolveExtends_DiamondIsNotACycle(t *testing.T) {
	// Two extends entries sharing the same base must not be reported as a cycle.
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/shared.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
metadata:
  fields:
    - name: author
      type: string
`)
	})
	mux.HandleFunc("/left.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
extends:
  - %q
packages:
  - name: left
    path: ./left
    ecosystem: go
`, server.URL+"/shared.yaml")
	})
	mux.HandleFunc("/right.yaml", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
extends:
  - %q
packages:
  - name: right
    path: ./right
    ecosystem: go
`, server.URL+"/shared.yaml")
	})

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
  - %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL+"/left.yaml", server.URL+"/right.yaml"))

	cfg, err := LoadFromDir(dir)
	require.NoError(t, err)
	assert.Len(t, cfg.Packages, 3)
	assert.Len(t, cfg.Metadata.Fields, 1, "shared field should not duplicate")
}

func TestResolveExtends_RemoteExtendsMustStayOnSameHost(t *testing.T) {
	// A fetched remote config may only extend sources on its own host; the
	// check runs before any fetch, so the foreign URL is never contacted.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "extends:\n  - \"https://evil.example.invalid/pivot.yaml\"\n")
	}))
	defer server.Close()

	dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, server.URL+"/parent.yaml"))

	_, err := LoadFromDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must stay on its own host")
}

func TestRemoteConfig_ObjectFormGitDefaults(t *testing.T) {
	// Object-form git entries get the same documented path/ref defaults as
	// string-form entries parsed via parseGitURL.
	var rc RemoteConfig
	require.NoError(t, yaml.Unmarshal([]byte(`git: "https://example.invalid/org/repo.git"`), &rc))
	assert.Equal(t, "shipyard.yaml", rc.Path)
	assert.Equal(t, "main", rc.Ref)

	// Explicit values are kept, and URL-form entries stay untouched.
	rc = RemoteConfig{}
	require.NoError(t, yaml.Unmarshal([]byte("git: \"https://example.invalid/org/repo.git\"\npath: custom.yaml\nref: v2"), &rc))
	assert.Equal(t, "custom.yaml", rc.Path)
	assert.Equal(t, "v2", rc.Ref)

	rc = RemoteConfig{}
	require.NoError(t, yaml.Unmarshal([]byte(`url: "https://example.invalid/base.yaml"`), &rc))
	assert.Empty(t, rc.Path)
	assert.Empty(t, rc.Ref)
}

func TestResolveExtends_GitSource(t *testing.T) {
	// Create a local git repo serving as the remote config source.
	repoDir := t.TempDir()
	repo, err := gogit.PlainInit(repoDir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "base.yaml"), []byte(remoteBaseConfig), 0644))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = wt.Add(".")
	require.NoError(t, err)
	commit, err := wt.Commit("add base config", &gogit.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	require.NoError(t, err)
	_, err = repo.CreateTag("v1.0.0", commit, nil)
	require.NoError(t, err)

	head, err := repo.Head()
	require.NoError(t, err)
	branch := head.Name().Short()

	for _, ref := range []string{branch, "v1.0.0", commit.String()} {
		t.Run("ref="+ref, func(t *testing.T) {
			dir := writeLocalConfig(t, fmt.Sprintf(`
extends:
  - git: %q
    path: base.yaml
    ref: %q
packages:
  - name: app
    path: ./
    ecosystem: go
`, repoDir, ref))

			cfg, err := LoadFromDir(dir)
			require.NoError(t, err)
			require.Len(t, cfg.Packages, 2)
			assert.Equal(t, "platform-core", cfg.Packages[0].Name)
		})
	}
}

func TestLoadFromDir_StringFormExtendsParses(t *testing.T) {
	// String-form extends entries must decode through viper (regression: they
	// previously failed with "expected a map or struct, got string").
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, remoteBaseConfig)
	}))
	defer server.Close()

	dir := writeLocalConfig(t, fmt.Sprintf("extends:\n  - %q\npackages:\n  - name: app\n    path: ./\n    ecosystem: go\n", server.URL))

	cfg, err := LoadFromDir(dir)
	require.NoError(t, err)
	require.Len(t, cfg.Extends, 1)
	assert.Equal(t, server.URL, cfg.Extends[0].URL)
}

func TestMerge_DocumentedSemantics(t *testing.T) {
	base := &Config{
		Packages: []Package{
			{Name: "core", Path: "./core", Ecosystem: "go"},
			{Name: "shared", Path: "./shared", Ecosystem: "go"},
		},
		Templates: TemplateConfig{
			Changelog: &TemplateSource{Source: "builtin:default"},
			TagName:   &TemplateSource{Inline: "v{{.Version}}"},
		},
		Metadata: MetadataConfig{Fields: []MetadataField{
			{Name: "author", Required: true},
		}},
	}
	overlay := &Config{
		Packages: []Package{
			{Name: "shared", Path: "./replaced", Ecosystem: "npm"},
			{Name: "app", Path: "./app", Ecosystem: "go"},
		},
		Templates: TemplateConfig{
			TagName: &TemplateSource{Inline: "app/v{{.Version}}"},
		},
		Metadata: MetadataConfig{Fields: []MetadataField{
			{Name: "author", Required: false},
			{Name: "issue"},
		}},
	}

	merged := base.Merge(overlay)

	// Same-name package replaced in place, new package appended
	require.Len(t, merged.Packages, 3)
	assert.Equal(t, "./replaced", merged.Packages[1].Path)
	assert.Equal(t, "npm", merged.Packages[1].Ecosystem)
	assert.Equal(t, "app", merged.Packages[2].Name)

	// Templates merged per key
	require.NotNil(t, merged.Templates.Changelog)
	assert.Equal(t, "builtin:default", merged.Templates.Changelog.Source)
	assert.Equal(t, "app/v{{.Version}}", merged.Templates.TagName.Inline)

	// Metadata fields: same-name replaced, new appended
	require.Len(t, merged.Metadata.Fields, 2)
	assert.False(t, merged.Metadata.Fields[0].Required)
	assert.Equal(t, "issue", merged.Metadata.Fields[1].Name)

	// Base is not mutated
	assert.True(t, base.Metadata.Fields[0].Required)
	assert.Equal(t, "./shared", base.Packages[1].Path)
}
