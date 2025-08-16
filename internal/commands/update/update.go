package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	force     bool
	channel   string
	rollback  bool
	verbose   bool
	useRemote bool
	useLocal  bool
	latest    bool // New flag for pulling latest from git
)

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update heimdall to the latest version",
		Long: `Update heimdall to the latest version.

For local installations (binary in ~/.local/bin):
  - Default: Rebuilds from current local source code
  - With --latest: Pulls latest changes from git before rebuilding (requires clean working directory)

For system installations:
  - Downloads and installs the latest binary from GitHub releases
  - Supports multiple release channels (stable, beta, nightly)
  - Includes rollback capability in case of issues`,
		Example: `  # Local installation: rebuild from current source
  heimdall update

  # Local installation: pull latest from git and rebuild
  heimdall update --latest

  # Check for updates without installing
  heimdall update --check

  # Force update even if already on latest version
  heimdall update --force

  # Update to beta channel (remote installations)
  heimdall update --channel beta

  # Rollback to previous version
  heimdall update --rollback`,
		RunE: runUpdate,
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without installing")
	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already on latest version")
	cmd.Flags().StringVar(&channel, "channel", "stable", "Release channel (stable, beta, nightly)")
	cmd.Flags().BoolVar(&rollback, "rollback", false, "Rollback to previous version")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed output during update")
	cmd.Flags().BoolVar(&useRemote, "remote", false, "Force update from GitHub releases")
	cmd.Flags().BoolVar(&useLocal, "local", false, "Force update from local git repository")
	cmd.Flags().BoolVar(&latest, "latest", false, "Pull latest changes from git before rebuilding (local installations only)")

	// Mark remote and local as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("remote", "local")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Handle rollback separately
	if rollback {
		return performRollback()
	}

	// Check if this is a local installation
	isLocal := isLocalInstallation()

	// For local installations, handle differently
	if isLocal {
		return handleLocalUpdate()
	}

	// For non-local installations, proceed with normal update flow
	updateInfo, err := checkForUpdates()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if checkOnly {
		if updateInfo.Available {
			fmt.Printf("Update available: %s -> %s\n", updateInfo.CurrentVersion, updateInfo.LatestVersion)
			fmt.Printf("Release notes: %s\n", updateInfo.ReleaseURL)
		} else {
			fmt.Println("You are running the latest version")
		}
		return nil
	}

	if !updateInfo.Available && !force {
		fmt.Println("You are already running the latest version")
		return nil
	}

	// Perform the update
	fmt.Printf("Updating heimdall from %s to %s...\n", updateInfo.CurrentVersion, updateInfo.LatestVersion)

	return performBinaryUpdate(updateInfo)
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	DownloadURL    string
	Checksum       string
	IsGitInstall   bool
	Channel        string
}

func checkForUpdates() (*UpdateInfo, error) {
	// Get current version
	currentMeta, err := GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	// Determine update mode
	isGitInstall := false
	if useLocal {
		// User explicitly requested local update
		isGitInstall = true
	} else if useRemote {
		// User explicitly requested remote update
		isGitInstall = false
	} else {
		// Auto-detect based on installation location and git repository
		isGitInstall = isLocalInstallation()
	}

	// If it's a git installation, check if the repository exists
	if isGitInstall {
		if _, err := findGitRoot(); err != nil {
			if !useLocal {
				// Auto-detected as local but can't find repo, fall back to remote
				fmt.Println("Local installation detected but source repository not found, falling back to remote updates")
				isGitInstall = false
			} else {
				// User explicitly requested local but repo not found
				return nil, fmt.Errorf("local update requested but git repository not found: %w", err)
			}
		}
	}

	// For git installations, we check against remote git tags
	if isGitInstall {
		repoRoot, _ := findGitRoot()
		if repoRoot != "" {
			// Fetch latest tags from remote
			gitCommand(repoRoot, "fetch", "--tags", "--quiet")

			// Get latest tag
			latestTag, err := gitCommand(repoRoot, "describe", "--tags", "--abbrev=0", "origin/main")
			if err != nil {
				// Try without origin/main
				latestTag, err = gitCommand(repoRoot, "describe", "--tags", "--abbrev=0")
			}

			if err == nil {
				latestTag = strings.TrimSpace(latestTag)
				latestVersion, _ := ParseVersion(latestTag)

				return &UpdateInfo{
					Available:      latestVersion.IsNewer(currentMeta.Version),
					CurrentVersion: currentMeta.Version.String(),
					LatestVersion:  latestVersion.String(),
					IsGitInstall:   true,
					Channel:        channel,
				}, nil
			}
		}
	}

	// For remote updates, check GitHub releases
	client := NewGitHubClient()

	// Fetch latest release for the channel
	release, err := client.GetLatestRelease(channel)
	if err != nil {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: currentMeta.Version.String(),
			LatestVersion:  currentMeta.Version.String(),
			IsGitInstall:   false,
			Channel:        channel,
		}, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Parse latest version
	latestVersion, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest version: %w", err)
	}

	// Find appropriate asset
	asset, err := client.GetReleaseAsset(release)
	if err != nil {
		return nil, fmt.Errorf("failed to find release asset: %w", err)
	}

	// Get checksum if available
	var checksum string
	if asset != nil {
		checksum, _ = client.GetChecksum(release, asset)
	}

	// Determine if update is available
	updateAvailable := latestVersion.IsNewer(currentMeta.Version)

	info := &UpdateInfo{
		Available:      updateAvailable,
		CurrentVersion: currentMeta.Version.String(),
		LatestVersion:  latestVersion.String(),
		ReleaseURL:     release.HTMLURL,
		IsGitInstall:   false,
		Channel:        channel,
	}

	if asset != nil {
		info.DownloadURL = asset.BrowserDownloadURL
		info.Checksum = checksum
	}

	return info, nil
}

// isGitInstallation checks if heimdall is running from a git repository
func isGitInstallation() bool {
	isGit, _ := CheckGitInstallation()
	return isGit
}

// isLocalInstallation checks if heimdall is a local installation (in ~/.local/bin)
func isLocalInstallation() bool {
	// Get the executable path
	exe, err := os.Executable()
	if err != nil {
		return false
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}

	// Check if it's in ~/.local/bin
	homeDir, _ := os.UserHomeDir()
	localBinPath := filepath.Join(homeDir, ".local", "bin")

	// Check if the binary is in ~/.local/bin
	if filepath.Dir(exe) == localBinPath {
		// This is likely a local installation
		// Try to find the git repository
		if _, err := findGitRoot(); err == nil {
			return true
		}
	}

	// Also check if we're running directly from a git repository
	isGit, _ := CheckGitInstallation()
	return isGit
}

func performRollback() error {
	return Rollback()
}

// performGitUpdate is now handled by handleLocalUpdate
// Keeping this for backward compatibility if needed
func performGitUpdate() error {
	return handleLocalUpdate()
}

func handleLocalUpdate() error {
	// Find the git repository
	repoRoot, err := findGitRoot()
	if err != nil {
		return fmt.Errorf("local installation detected but source repository not found: %w", err)
	}

	fmt.Printf("Local installation detected at ~/.local/bin\n")
	fmt.Printf("Repository location: %s\n", repoRoot)

	// If --latest flag is set, try to pull from git
	if latest {
		fmt.Println("Checking for upstream updates...")

		// Get git info
		info, err := GetGitInfo()
		if err != nil {
			return fmt.Errorf("failed to get git info: %w", err)
		}

		// Check for uncommitted changes
		if info.IsDirty {
			fmt.Println("\nCannot pull latest: you have uncommitted changes.")
			fmt.Println("Please commit or stash your changes first.")
			return fmt.Errorf("repository has uncommitted changes")
		}

		// Check if we have a remote
		if !info.HasRemote {
			return fmt.Errorf("no remote configured for branch %s", info.Branch)
		}

		// Fetch and check for updates
		fmt.Println("Fetching latest changes from upstream...")
		if _, err := gitCommand(repoRoot, "fetch", "--quiet"); err != nil {
			return fmt.Errorf("failed to fetch: %w", err)
		}

		// Refresh git info after fetch
		info, _ = GetGitInfo()

		if info.BehindRemote > 0 {
			fmt.Printf("Found %d new commit(s) from upstream\n", info.BehindRemote)
			fmt.Println("Pulling changes...")

			output, err := gitCommand(repoRoot, "pull")
			if err != nil {
				return fmt.Errorf("failed to pull: %w", err)
			}
			if verbose && output != "" {
				fmt.Print(output)
			}
		} else {
			fmt.Println("Already up to date with upstream")
		}
	} else {
		// Default behavior: just rebuild from current code
		// But check if there's a newer version upstream and inform the user
		if _, err := gitCommand(repoRoot, "fetch", "--tags", "--quiet"); err == nil {
			// Try to get the latest tag
			latestTag, err := gitCommand(repoRoot, "describe", "--tags", "--abbrev=0", "origin/main")
			if err != nil {
				// Try without origin/main
				latestTag, _ = gitCommand(repoRoot, "describe", "--tags", "--abbrev=0")
			}

			if latestTag != "" {
				latestTag = strings.TrimSpace(latestTag)
				currentMeta, _ := GetCurrentVersion()
				latestVersion, _ := ParseVersion(latestTag)

				if latestVersion != nil && currentMeta != nil && latestVersion.IsNewer(currentMeta.Version) {
					fmt.Printf("\nNote: Version %s is available upstream. Use --latest to update from git\n", latestVersion.String())
				}
			}
		}
	}

	// Rebuild from current source
	fmt.Println("\nRebuilding heimdall from local source...")
	if err := rebuildBinary(repoRoot, verbose); err != nil {
		return fmt.Errorf("failed to rebuild: %w", err)
	}

	fmt.Println("Rebuild complete! Heimdall has been updated from local source.")
	return nil
}

func performBinaryUpdate(info *UpdateInfo) error {
	if info.DownloadURL == "" {
		return fmt.Errorf("no download URL available")
	}

	fmt.Printf("Downloading update for %s/%s...\n", runtime.GOOS, runtime.GOARCH)

	// Download the new binary
	newBinaryPath, err := DownloadRelease(info.DownloadURL, info.Checksum, verbose)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.RemoveAll(filepath.Dir(newBinaryPath)) // Clean up temp directory

	// Replace the binary
	fmt.Println("Installing update...")
	if err := ReplaceBinary(newBinaryPath, true); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Clean up old backups (keep last 3)
	CleanupBackups(3)

	fmt.Println("Update complete! Please restart heimdall to use the new version.")
	return nil
}
