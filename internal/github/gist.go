package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/kasperbasse/skel/internal/version"
)

// APIBase is the GitHub API base URL. Override in tests.
var APIBase = "https://api.github.com"

// GistFile represents a single file in a GitHub Gist.
type GistFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	RawURL   string `json:"raw_url"`
	Content  string `json:"content"`
}

// Gist represents the relevant fields of a GitHub Gist API response.
type Gist struct {
	ID          string              `json:"id"`
	HTMLURL     string              `json:"html_url"`
	Description string              `json:"description"`
	Files       map[string]GistFile `json:"files"`
}

// CreateGistRequest is the payload for creating a new gist.
type CreateGistRequest struct {
	Description string                    `json:"description"`
	Public      bool                      `json:"public"`
	Files       map[string]CreateGistFile `json:"files"`
}

// CreateGistFile is a file entry in a create-gist request.
type CreateGistFile struct {
	Content string `json:"content"`
}

// ParseSource extracts a gist ID from a URL or github:user/id shorthand.
//
// Accepted formats:
//   - https://gist.github.com/user/abc123
//   - https://gist.github.com/abc123
//   - https://gist.github.com/user/abc123/revision
//   - github:user/abc123
func ParseSource(source string) (string, error) {
	// Shorthand: github:user/id
	if strings.HasPrefix(source, "github:") {
		parts := strings.SplitN(strings.TrimPrefix(source, "github:"), "/", 2)
		if len(parts) != 2 || parts[1] == "" {
			return "", fmt.Errorf("invalid shorthand %q, expected github:user/gist-id", source)
		}
		return parts[1], nil
	}

	// Full URL
	if strings.Contains(source, "gist.github.com") {
		trimmed := strings.TrimRight(source, "/")
		segments := strings.Split(trimmed, "/")
		// URL forms: .../user/id, .../id, .../user/id/revision
		for i := len(segments) - 1; i >= 0; i-- {
			seg := segments[i]
			if seg == "" || seg == "gist.github.com" {
				break
			}
			// Gist IDs are hex strings, 20-32 chars. Revision hashes are 40 chars.
			// Take the first segment from the end that looks like a gist ID (not a 40-char revision).
			if len(seg) >= 20 && len(seg) <= 32 && isHex(seg) {
				return seg, nil
			}
		}
		// Fallback: just take the last meaningful segment (handles non-hex IDs).
		for i := len(segments) - 1; i >= 0; i-- {
			if segments[i] != "" && segments[i] != "gist.github.com" {
				// Skip 40-char revision hashes at the end.
				if len(segments[i]) == 40 && isHex(segments[i]) && i > 0 {
					continue
				}
				return segments[i], nil
			}
		}
		return "", fmt.Errorf("could not extract gist ID from URL %q", source)
	}

	return "", fmt.Errorf("unrecognized source %q\n  Expected: a gist URL or github:user/gist-id", source)
}

// FetchGist retrieves a public gist by ID. No authentication required.
func FetchGist(gistID string) (gist *Gist, err error) {
	req, err := http.NewRequest("GET", APIBase+"/gists/"+gistID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "skel/"+version.Version)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not connect to GitHub API: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// Using a 60MB safety cap is a smart "skel" move for stability
	body, err := io.ReadAll(io.LimitReader(resp.Body, 60*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Handle Status Codes
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("gist not found (404) - check the URL or ID")
	}
	if resp.StatusCode == 403 {
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return nil, fmt.Errorf("GitHub API rate limit exceeded - set GITHUB_TOKEN")
		}
		return nil, fmt.Errorf("GitHub API access denied (403)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var result Gist
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing gist response: %w", err)
	}
	return &result, nil
}

// FindProfileJSON finds the .json profile file in a gist and returns its content.
// If the file content was truncated by the API (>1MB), it fetches the raw URL.
func FindProfileJSON(gist *Gist, maxSize int64) (string, error) {
	var match *GistFile
	for _, f := range gist.Files {
		f := f
		if strings.HasSuffix(f.Filename, ".json") {
			if match != nil {
				return "", fmt.Errorf("gist contains multiple .json files - expected exactly one profile")
			}
			match = &f
		}
	}
	if match == nil {
		return "", fmt.Errorf("gist contains no .json files - not an skel profile")
	}

	if match.Size > maxSize {
		return "", fmt.Errorf("profile file too large (%d bytes, max %d)", match.Size, maxSize)
	}

	// The API includes content inline for files <1MB. For larger files, fetch raw_url.
	if match.Content != "" {
		return match.Content, nil
	}
	if match.RawURL == "" {
		return "", fmt.Errorf("gist file %q has no content or raw URL", match.Filename)
	}

	resp, err := http.Get(match.RawURL)
	if err != nil {
		return "", fmt.Errorf("fetching raw content: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return "", fmt.Errorf("reading raw content: %w", err)
	}
	if int64(len(data)) > maxSize {
		return "", fmt.Errorf("profile file too large (max %d bytes)", maxSize)
	}
	return string(data), nil
}

// CreateGist creates a new gist. Requires a valid GitHub token.
func CreateGist(token string, req *CreateGistRequest) (*Gist, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding gist request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", APIBase+"/gists", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("User-Agent", "skel/"+version.Version)
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("could not connect to GitHub API: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed - check your GITHUB_TOKEN or run 'gh auth login'")
	}
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var gist Gist
	if err := json.Unmarshal(respBody, &gist); err != nil {
		return nil, fmt.Errorf("parsing gist response: %w", err)
	}
	return &gist, nil
}

// ResolveToken finds a GitHub token from GITHUB_TOKEN env var or `gh auth token`.
func ResolveToken() (string, error) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("no GITHUB_TOKEN set and 'gh' CLI not found\n\nSet GITHUB_TOKEN or install gh: https://cli.github.com")
	}

	out, err := exec.Command(ghPath, "auth", "token").Output()
	if err != nil {
		return "", fmt.Errorf("'gh auth token' failed - run 'gh auth login' first")
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("'gh auth token' returned empty - run 'gh auth login' first")
	}
	return token, nil
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
