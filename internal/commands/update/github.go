package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubAPIBase = "https://api.github.com"
	githubOwner   = "arthur404dev"
	githubRepo    = "heimdall-cli"
	cacheTimeout  = 1 * time.Hour
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	ContentType        string    `json:"content_type"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UpdateCache represents cached update information
type UpdateCache struct {
	CheckedAt time.Time      `json:"checked_at"`
	Release   *GitHubRelease `json:"release"`
}

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	httpClient *http.Client
	cacheDir   string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient() *GitHubClient {
	cacheDir := getCacheDir()
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cacheDir: cacheDir,
	}
}

// GetLatestRelease fetches the latest release for the specified channel
func (c *GitHubClient) GetLatestRelease(channel string) (*GitHubRelease, error) {
	// Check cache first
	cached, err := c.loadCache(channel)
	if err == nil && cached != nil && time.Since(cached.CheckedAt) < cacheTimeout {
		return cached.Release, nil
	}

	// Fetch from GitHub
	var release *GitHubRelease
	var fetchErr error

	switch channel {
	case "stable":
		release, fetchErr = c.fetchLatestStableRelease()
	case "beta", "alpha", "nightly":
		release, fetchErr = c.fetchLatestPrereleaseByChannel(channel)
	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}

	if fetchErr != nil {
		// If fetch fails but we have cache, return cached version
		if cached != nil {
			return cached.Release, nil
		}
		return nil, fetchErr
	}

	// Update cache
	c.saveCache(channel, release)
	return release, nil
}

// fetchLatestStableRelease fetches the latest stable release
func (c *GitHubClient) fetchLatestStableRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBase, githubOwner, githubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("heimdall-cli/%s", getCurrentVersionString()))

	// Add GitHub token if available (helps with rate limits)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// fetchLatestPrereleaseByChannel fetches the latest prerelease for a specific channel
func (c *GitHubClient) fetchLatestPrereleaseByChannel(channel string) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", githubAPIBase, githubOwner, githubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("heimdall-cli/%s", getCurrentVersionString()))

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	// Find the latest release matching the channel
	for _, release := range releases {
		if release.Draft {
			continue
		}

		version, err := ParseVersion(release.TagName)
		if err != nil {
			continue
		}

		releaseChannel := detectChannel(version)
		if releaseChannel == channel {
			return &release, nil
		}
	}

	return nil, fmt.Errorf("no %s releases found", channel)
}

// GetReleaseAsset finds the appropriate asset for the current platform
func (c *GitHubClient) GetReleaseAsset(release *GitHubRelease) (*Asset, error) {
	if release == nil {
		return nil, fmt.Errorf("release is nil")
	}

	// Construct expected asset name
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map common variations
	if arch == "amd64" {
		arch = "x86_64"
	}

	// Expected patterns for asset names
	patterns := []string{
		fmt.Sprintf("heimdall-%s-%s", platform, arch),
		fmt.Sprintf("heimdall_%s_%s", platform, arch),
		fmt.Sprintf("heimdall-%s-%s.tar.gz", platform, arch),
		fmt.Sprintf("heimdall_%s_%s.tar.gz", platform, arch),
		fmt.Sprintf("heimdall-%s-%s.zip", platform, arch),
	}

	// Special case for Windows
	if platform == "windows" {
		patterns = append(patterns,
			fmt.Sprintf("heimdall-%s-%s.exe", platform, arch),
			fmt.Sprintf("heimdall_%s_%s.exe", platform, arch),
		)
	}

	// Find matching asset
	for _, asset := range release.Assets {
		assetNameLower := strings.ToLower(asset.Name)
		for _, pattern := range patterns {
			if strings.Contains(assetNameLower, strings.ToLower(pattern)) {
				return &asset, nil
			}
		}
	}

	// If no platform-specific asset found, look for universal binary
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), "heimdall") &&
			!strings.Contains(strings.ToLower(asset.Name), "sha256") &&
			!strings.Contains(strings.ToLower(asset.Name), "checksum") {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no suitable asset found for %s/%s", platform, arch)
}

// GetChecksum fetches the checksum for a release asset
func (c *GitHubClient) GetChecksum(release *GitHubRelease, asset *Asset) (string, error) {
	// Look for checksum file
	var checksumAsset *Asset
	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, "sha256") || strings.Contains(name, "checksum") {
			checksumAsset = &a
			break
		}
	}

	if checksumAsset == nil {
		// No checksum file available
		return "", nil
	}

	// Download checksum file
	resp, err := c.httpClient.Get(checksumAsset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum: %w", err)
	}

	// Parse checksum file (format: "hash  filename")
	lines := strings.Split(string(content), "\n")
	assetName := asset.Name

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(line, assetName) {
			return parts[0], nil
		}
	}

	return "", nil
}

// Cache management functions

func getCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "heimdall", "updates")
}

func (c *GitHubClient) getCacheFile(channel string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("%s.json", channel))
}

func (c *GitHubClient) loadCache(channel string) (*UpdateCache, error) {
	cacheFile := c.getCacheFile(channel)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func (c *GitHubClient) saveCache(channel string, release *GitHubRelease) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return err
	}

	cache := UpdateCache{
		CheckedAt: time.Now(),
		Release:   release,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	cacheFile := c.getCacheFile(channel)
	return os.WriteFile(cacheFile, data, 0644)
}

// ClearCache removes all cached update information
func (c *GitHubClient) ClearCache() error {
	return os.RemoveAll(c.cacheDir)
}
