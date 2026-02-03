package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

const (
	defaultBaseURL = "https://api.github.com"
	defaultTimeout = 10 * time.Second
)

// GitHubClient handles communication with GitHub API
type GitHubClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		authToken: os.Getenv("GITHUB_TOKEN"),
	}
}

// githubRelease represents the GitHub API release response
type githubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// GetLatestRelease fetches the latest release from GitHub
func (c *GitHubClient) GetLatestRelease(ctx context.Context, owner, repo string) (*ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusForbidden {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Reset at %s. Set GITHUB_TOKEN environment variable to increase rate limit", resetTime)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var ghRelease githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to ReleaseInfo
	release := &ReleaseInfo{
		TagName:     ghRelease.TagName,
		Name:        ghRelease.Name,
		Body:        ghRelease.Body,
		PublishedAt: ghRelease.PublishedAt,
		Prerelease:  ghRelease.Prerelease,
		Assets:      make([]ReleaseAsset, len(ghRelease.Assets)),
	}

	for i, asset := range ghRelease.Assets {
		release.Assets[i] = ReleaseAsset{
			Name:        asset.Name,
			DownloadURL: asset.BrowserDownloadURL,
			Size:        asset.Size,
		}
	}

	return release, nil
}

// IsNewer compares the release version with the current version
// Returns true if the release is newer than the current version
func IsNewer(currentVersion, releaseVersion string) (bool, error) {
	// Strip "v" prefix if present
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	releaseVersion = strings.TrimPrefix(releaseVersion, "v")

	current, err := semver.Parse(currentVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %w", err)
	}

	release, err := semver.Parse(releaseVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse release version: %w", err)
	}

	return release.Compare(current) > 0, nil
}
