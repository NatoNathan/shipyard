package upgrade

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitHubClient(t *testing.T) {
	t.Run("creates client with defaults", func(t *testing.T) {
		client := NewGitHubClient()
		require.NotNil(t, client)
		assert.Equal(t, defaultBaseURL, client.baseURL)
		assert.NotNil(t, client.httpClient)
	})

	t.Run("reads GITHUB_TOKEN from environment", func(t *testing.T) {
		origToken := os.Getenv("GITHUB_TOKEN")
		defer os.Setenv("GITHUB_TOKEN", origToken)

		os.Setenv("GITHUB_TOKEN", "test-token")
		client := NewGitHubClient()
		assert.Equal(t, "test-token", client.authToken)
	})
}

func TestGitHubClient_GetLatestRelease(t *testing.T) {
	t.Run("successfully fetches release", func(t *testing.T) {
		mockRelease := githubRelease{
			TagName:     "v1.2.3",
			Name:        "Release 1.2.3",
			Body:        "Release notes",
			PublishedAt: time.Now(),
			Prerelease:  false,
			Assets: []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
				Size               int64  `json:"size"`
			}{
				{
					Name:               "shipyard_v1.2.3_darwin_arm64.tar.gz",
					BrowserDownloadURL: "https://example.com/download",
					Size:               1024,
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/test/repo/releases/latest", r.URL.Path)
			assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockRelease)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		release, err := client.GetLatestRelease(context.Background(), "test", "repo")
		require.NoError(t, err)
		require.NotNil(t, release)

		assert.Equal(t, "v1.2.3", release.TagName)
		assert.Equal(t, "Release 1.2.3", release.Name)
		assert.Equal(t, "Release notes", release.Body)
		assert.False(t, release.Prerelease)
		assert.Len(t, release.Assets, 1)
		assert.Equal(t, "shipyard_v1.2.3_darwin_arm64.tar.gz", release.Assets[0].Name)
	})

	t.Run("includes auth token when set", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token", auth)

			mockRelease := githubRelease{TagName: "v1.0.0"}
			json.NewEncoder(w).Encode(mockRelease)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL
		client.authToken = "test-token"

		_, err := client.GetLatestRelease(context.Background(), "test", "repo")
		require.NoError(t, err)
	})

	t.Run("handles rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", "1234567890")
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease(context.Background(), "test", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
		assert.Contains(t, err.Error(), "GITHUB_TOKEN")
	})

	t.Run("handles non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease(context.Background(), "test", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		_, err := client.GetLatestRelease(context.Background(), "test", "repo")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		client := NewGitHubClient()
		client.baseURL = server.URL

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.GetLatestRelease(ctx, "test", "repo")
		require.Error(t, err)
	})
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		releaseVersion string
		expected       bool
		expectError    bool
	}{
		{
			name:           "newer major version",
			currentVersion: "1.0.0",
			releaseVersion: "2.0.0",
			expected:       true,
		},
		{
			name:           "newer minor version",
			currentVersion: "1.0.0",
			releaseVersion: "1.1.0",
			expected:       true,
		},
		{
			name:           "newer patch version",
			currentVersion: "1.0.0",
			releaseVersion: "1.0.1",
			expected:       true,
		},
		{
			name:           "same version",
			currentVersion: "1.0.0",
			releaseVersion: "1.0.0",
			expected:       false,
		},
		{
			name:           "older version",
			currentVersion: "2.0.0",
			releaseVersion: "1.0.0",
			expected:       false,
		},
		{
			name:           "strips v prefix from current",
			currentVersion: "v1.0.0",
			releaseVersion: "1.1.0",
			expected:       true,
		},
		{
			name:           "strips v prefix from release",
			currentVersion: "1.0.0",
			releaseVersion: "v1.1.0",
			expected:       true,
		},
		{
			name:           "strips v prefix from both",
			currentVersion: "v1.0.0",
			releaseVersion: "v1.1.0",
			expected:       true,
		},
		{
			name:           "invalid current version",
			currentVersion: "invalid",
			releaseVersion: "1.0.0",
			expectError:    true,
		},
		{
			name:           "invalid release version",
			currentVersion: "1.0.0",
			releaseVersion: "invalid",
			expectError:    true,
		},
		{
			name:           "dev version",
			currentVersion: "dev",
			releaseVersion: "1.0.0",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsNewer(tt.currentVersion, tt.releaseVersion)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
