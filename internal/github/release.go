package github

import (
	"fmt"
	"net/http"
)

// Release represents a GitHub release.
type Release struct {
	ID         int    `json:"id"`
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	HTMLURL    string `json:"html_url"`
}

// createReleaseRequest is the payload for creating a GitHub release.
type createReleaseRequest struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// CreateRelease creates a new GitHub release for the given tag.
func (c *Client) CreateRelease(tag, name, body string, draft bool) (*Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases", c.owner, c.repo)

	req := createReleaseRequest{
		TagName: tag,
		Name:    name,
		Body:    body,
		Draft:   draft,
	}

	resp, err := c.do(http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := decodeResponse(resp, &release); err != nil {
		return nil, fmt.Errorf("failed to create release: %w", err)
	}

	return &release, nil
}

// GetReleaseByTag fetches a release by its tag name.
// Returns nil, nil if the release does not exist (404).
func (c *Client) GetReleaseByTag(tag string) (*Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases/tags/%s", c.owner, c.repo, tag)

	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil
	}

	var release Release
	if err := decodeResponse(resp, &release); err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	return &release, nil
}

// DeleteRelease deletes a release by ID.
func (c *Client) DeleteRelease(id int) error {
	path := fmt.Sprintf("/repos/%s/%s/releases/%d", c.owner, c.repo, id)

	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete release: HTTP %d", resp.StatusCode)
	}

	return nil
}

// UpdateRelease updates an existing release (e.g. to publish a draft).
func (c *Client) UpdateRelease(id int, name, body string, draft bool) (*Release, error) {
	path := fmt.Sprintf("/repos/%s/%s/releases/%d", c.owner, c.repo, id)

	req := createReleaseRequest{
		Name:  name,
		Body:  body,
		Draft: draft,
	}

	resp, err := c.do(http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := decodeResponse(resp, &release); err != nil {
		return nil, fmt.Errorf("failed to update release: %w", err)
	}

	return &release, nil
}
