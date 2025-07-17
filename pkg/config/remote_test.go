package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRemoteConfigFetcher_FetchFromHTTP(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := `type: monorepo
repo: github.com/test/remote-repo
changelog:
  template: keepachangelog
packages:
  - name: shared-api
    path: packages/api
    ecosystem: go
  - name: shared-frontend
    path: packages/frontend
    ecosystem: npm
`
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config))
	}))
	defer server.Close()

	// Create a temporary directory for cache
	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	// Test fetching from HTTP
	v, err := fetcher.fetchFromHTTP(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to fetch from HTTP: %v", err)
	}

	// Verify the fetched config
	if v.GetString("type") != "monorepo" {
		t.Errorf("Expected type 'monorepo', got %s", v.GetString("type"))
	}

	if v.GetString("repo") != "github.com/test/remote-repo" {
		t.Errorf("Expected repo 'github.com/test/remote-repo', got %s", v.GetString("repo"))
	}

	packages := v.Get("packages")
	if packages == nil {
		t.Error("Expected packages to be present")
	}
}

func TestRemoteConfigFetcher_FetchFromHTTP_NotFound(t *testing.T) {
	// Create a test HTTP server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	// Test fetching from HTTP - should fail
	_, err := fetcher.fetchFromHTTP(server.URL, false)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestRemoteConfigFetcher_Cache(t *testing.T) {
	// Create a test HTTP server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		config := `type: monorepo
repo: github.com/test/cached-repo
changelog:
  template: keepachangelog`
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	// First fetch - should hit the server
	v1, err := fetcher.fetchFromHTTP(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to fetch from HTTP: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second fetch - should use cache
	v2, err := fetcher.fetchFromHTTP(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to fetch from HTTP (cached): %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call (cached), got %d", callCount)
	}

	// Verify both configs are identical
	if v1.GetString("repo") != v2.GetString("repo") {
		t.Error("Cached config differs from original")
	}
}

func TestRemoteConfigFetcher_ForceFresh(t *testing.T) {
	// Create a test HTTP server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		config := `type: monorepo
repo: github.com/test/fresh-repo
changelog:
  template: keepachangelog`
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	// First fetch
	_, err := fetcher.fetchFromHTTP(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to fetch from HTTP: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second fetch with force fresh
	_, err = fetcher.fetchFromHTTP(server.URL, true)
	if err != nil {
		t.Fatalf("Failed to fetch from HTTP (force fresh): %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 server calls (force fresh), got %d", callCount)
	}
}

func TestRemoteConfigFetcher_GitHubURL(t *testing.T) {
	tests := []struct {
		name        string
		githubURL   string
		expectedURL string
		shouldError bool
	}{
		{
			name:        "valid GitHub URL with default branch",
			githubURL:   "github:owner/repo/config.yaml",
			expectedURL: "https://raw.githubusercontent.com/owner/repo/main/config.yaml",
			shouldError: false,
		},
		{
			name:        "valid GitHub URL with specific branch",
			githubURL:   "github:owner/repo/config.yaml@develop",
			expectedURL: "https://raw.githubusercontent.com/owner/repo/develop/config.yaml",
			shouldError: false,
		},
		{
			name:        "valid GitHub URL with nested path",
			githubURL:   "github:owner/repo/configs/shipyard/config.yaml",
			expectedURL: "https://raw.githubusercontent.com/owner/repo/main/configs/shipyard/config.yaml",
			shouldError: false,
		},
		{
			name:        "invalid GitHub URL - too few parts",
			githubURL:   "github:owner/repo",
			expectedURL: "",
			shouldError: true,
		},
	}

	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the actual GitHub fetching without external dependencies,
			// but we can test URL parsing by inspecting the error messages
			_, err := fetcher.fetchFromGitHub(tt.githubURL, false)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				// For valid URLs, we expect a network-related error since we're not mocking the HTTP call
				if err == nil {
					t.Error("Expected network error but got none")
				}
				// The error should be network-related, not URL parsing related
				if err.Error() == "invalid GitHub URL format, expected github:owner/repo/path/to/config.yaml[@ref]" {
					t.Error("Got URL parsing error instead of network error")
				}
			}
		})
	}
}

func TestRemoteConfigFetcher_DetectFormat(t *testing.T) {
	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	tests := []struct {
		name           string
		sourceURL      string
		content        string
		expectedFormat string
	}{
		{
			name:           "YAML from URL extension",
			sourceURL:      "https://example.com/config.yaml",
			content:        "type: monorepo",
			expectedFormat: "yaml",
		},
		{
			name:           "JSON from URL extension",
			sourceURL:      "https://example.com/config.json",
			content:        `{"type": "monorepo"}`,
			expectedFormat: "json",
		},
		{
			name:           "JSON from content",
			sourceURL:      "https://example.com/config",
			content:        `{"type": "monorepo"}`,
			expectedFormat: "json",
		},
		{
			name:           "YAML default",
			sourceURL:      "https://example.com/config",
			content:        "type: monorepo",
			expectedFormat: "yaml",
		},
		{
			name:           "TOML from URL extension",
			sourceURL:      "https://example.com/config.toml",
			content:        `type = "monorepo"`,
			expectedFormat: "toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := fetcher.detectFormat(tt.sourceURL, tt.content)
			if format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, format)
			}
		})
	}
}

func TestIsValidRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid GitHub URL",
			url:      "github:owner/repo/config.yaml",
			expected: true,
		},
		{
			name:     "valid Git URL",
			url:      "git+https://github.com/owner/repo.git/config.yaml",
			expected: true,
		},
		{
			name:     "valid Git SSH URL",
			url:      "git+git@github.com:owner/repo.git/config.yaml",
			expected: true,
		},
		{
			name:     "valid HTTPS URL",
			url:      "https://example.com/config.yaml",
			expected: true,
		},
		{
			name:     "valid HTTP URL",
			url:      "http://example.com/config.yaml",
			expected: true,
		},
		{
			name:     "invalid file path",
			url:      "./config.yaml",
			expected: false,
		},
		{
			name:     "invalid absolute path",
			url:      "/path/to/config.yaml",
			expected: false,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRemoteURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsValidRemoteURL(%s) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestRemoteConfigFetcher_ClearCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	// Create fetcher and ensure cache directory exists
	fetcher := NewRemoteConfigFetcher(cacheDir)

	// Create a dummy cache file
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	dummyFile := filepath.Join(cacheDir, "dummy.json")
	if err := os.WriteFile(dummyFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create dummy cache file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dummyFile); os.IsNotExist(err) {
		t.Error("Dummy cache file should exist before clearing")
	}

	// Clear cache
	if err := fetcher.ClearCache(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify cache directory is gone
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("Cache directory should not exist after clearing")
	}
}

func TestLoadRemoteConfig(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := `type: monorepo
repo: github.com/test/remote-config
changelog:
  template: keepachangelog
change_types:
  - name: patch
    display_name: "🔧 Patch"
    semver_bump: patch
  - name: minor
    display_name: "✨ Minor"
    semver_bump: minor
`
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config))
	}))
	defer server.Close()

	// Test loading remote config
	config, err := LoadRemoteConfig(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to load remote config: %v", err)
	}

	// Verify the loaded config
	if config.Type != RepositoryTypeMonorepo {
		t.Errorf("Expected type %s, got %s", RepositoryTypeMonorepo, config.Type)
	}

	if config.Repo != "github.com/test/remote-config" {
		t.Errorf("Expected repo 'github.com/test/remote-config', got %s", config.Repo)
	}

	// Remote configs should not have packages
	if len(config.Packages) != 0 {
		t.Errorf("Expected 0 packages for remote config, got %d", len(config.Packages))
	}

	// Should have change types
	if len(config.ChangeTypes) != 2 {
		t.Errorf("Expected 2 change types, got %d", len(config.ChangeTypes))
	}
}

func TestLoadRemoteConfig_InvalidURL(t *testing.T) {
	_, err := LoadRemoteConfig("./local-config.yaml", false)
	if err == nil {
		t.Error("Expected error for invalid remote URL")
	}
}

func TestRemoteConfigCacheOperations(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := `type: monorepo
repo: github.com/test/cache-operations`
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(config))
	}))
	defer server.Close()

	// Set up temporary cache directory
	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "test-cache"))

	// Initially, there should be no cached configs
	cached, err := fetcher.ListCachedConfigs()
	if err != nil {
		t.Fatalf("Failed to list cached configs: %v", err)
	}
	if len(cached) != 0 {
		t.Errorf("Expected 0 cached configs, got %d", len(cached))
	}

	// Fetch a config to create cache
	_, err = fetcher.fetchFromHTTP(server.URL, false)
	if err != nil {
		t.Fatalf("Failed to fetch config: %v", err)
	}

	// Now there should be one cached config
	cached, err = fetcher.ListCachedConfigs()
	if err != nil {
		t.Fatalf("Failed to list cached configs: %v", err)
	}
	if len(cached) != 1 {
		t.Errorf("Expected 1 cached config, got %d", len(cached))
	}

	// Verify cache entry
	if cached[0].URL != server.URL {
		t.Errorf("Expected cached URL %s, got %s", server.URL, cached[0].URL)
	}

	// Clear cache
	err = fetcher.ClearCache()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Cache should be empty again
	cached, err = fetcher.ListCachedConfigs()
	if err != nil {
		t.Fatalf("Failed to list cached configs after clearing: %v", err)
	}
	if len(cached) != 0 {
		t.Errorf("Expected 0 cached configs after clearing, got %d", len(cached))
	}
}

func TestRemoteConfigFetcher_CacheExpiry(t *testing.T) {
	tempDir := t.TempDir()
	fetcher := NewRemoteConfigFetcher(filepath.Join(tempDir, "cache"))

	// Create an expired cache entry manually
	testURL := "https://example.com/expired-config.yaml"
	expiredContent := "type: monorepo\nrepo: github.com/test/expired"

	// Save with very short TTL and back-date it
	err := fetcher.saveToCache(testURL, expiredContent, 1) // 1 minute TTL
	if err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	// Manually modify the cache file to make it appear expired
	cached, err := fetcher.ListCachedConfigs()
	if err != nil {
		t.Fatalf("Failed to list cached configs: %v", err)
	}
	if len(cached) != 1 {
		t.Fatalf("Expected 1 cached config, got %d", len(cached))
	}

	// Update the cache entry to be expired
	cacheEntry := cached[0]
	cacheEntry.LastFetched = time.Now().Add(-2 * time.Minute) // 2 minutes ago

	// Save the modified cache entry
	data, err := json.Marshal(cacheEntry)
	if err != nil {
		t.Fatalf("Failed to marshal cache entry: %v", err)
	}

	err = os.WriteFile(cacheEntry.CachePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write modified cache entry: %v", err)
	}

	// Try to load from cache - should fail due to expiry
	_, err = fetcher.loadFromCache(testURL)
	if err == nil {
		t.Error("Expected error for expired cache entry")
	}
	if !strings.Contains(err.Error(), "cache expired") {
		t.Errorf("Expected 'cache expired' error, got: %v", err)
	}
}
