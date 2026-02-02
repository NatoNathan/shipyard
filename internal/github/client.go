package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client provides access to the GitHub API for release management.
type Client struct {
	owner      string
	repo       string
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a GitHub API client from the provided owner, repo, and token spec.
// The tokenSpec supports "env:VAR_NAME" format to read from environment variables,
// or a literal token string.
func NewClient(owner, repo, tokenSpec string) (*Client, error) {
	token, err := resolveToken(tokenSpec)
	if err != nil {
		return nil, err
	}

	return &Client{
		owner:      owner,
		repo:       repo,
		token:      token,
		httpClient: http.DefaultClient,
		baseURL:    "https://api.github.com",
	}, nil
}

// resolveToken resolves a token specification to an actual token value.
// Supports "env:VAR_NAME" format or literal strings.
func resolveToken(spec string) (string, error) {
	if spec == "" {
		// Try common environment variables as fallback
		for _, envVar := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
			if token := os.Getenv(envVar); token != "" {
				return token, nil
			}
		}
		return "", fmt.Errorf("no GitHub token provided; set github.token in config (e.g. \"env:GITHUB_TOKEN\") or set the GITHUB_TOKEN environment variable")
	}

	if after, ok := strings.CutPrefix(spec, "env:"); ok {
		token := os.Getenv(after)
		if token == "" {
			return "", fmt.Errorf("environment variable %s is not set", after)
		}
		return token, nil
	}

	return spec, nil
}

// do executes an HTTP request against the GitHub API.
func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// decodeResponse reads and decodes a JSON response body.
func decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}
