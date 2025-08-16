package update

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitInfo contains information about the git repository
type GitInfo struct {
	IsRepo       bool
	Branch       string
	Commit       string
	IsDirty      bool
	HasRemote    bool
	RemoteURL    string
	BehindRemote int
	AheadRemote  int
}

// GetGitInfo returns information about the current git repository
func GetGitInfo() (*GitInfo, error) {
	info := &GitInfo{}

	// Check if we're in a git repository
	repoRoot, err := findGitRoot()
	if err != nil {
		return info, nil // Not a git repo, return empty info
	}

	info.IsRepo = true

	// Get current branch
	branch, err := gitCommand(repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		info.Branch = strings.TrimSpace(branch)
	}

	// Get current commit
	commit, err := gitCommand(repoRoot, "rev-parse", "HEAD")
	if err == nil {
		info.Commit = strings.TrimSpace(commit)[:8] // Short commit hash
	}

	// Check if working directory is dirty
	status, err := gitCommand(repoRoot, "status", "--porcelain")
	if err == nil {
		info.IsDirty = len(strings.TrimSpace(status)) > 0
	}

	// Check for remote
	remote, err := gitCommand(repoRoot, "config", "--get", fmt.Sprintf("branch.%s.remote", info.Branch))
	if err == nil && strings.TrimSpace(remote) != "" {
		info.HasRemote = true

		// Get remote URL
		remoteURL, err := gitCommand(repoRoot, "config", "--get", fmt.Sprintf("remote.%s.url", strings.TrimSpace(remote)))
		if err == nil {
			info.RemoteURL = strings.TrimSpace(remoteURL)
		}

		// Check if we're behind/ahead of remote
		// First, fetch to get latest remote info (but don't merge)
		gitCommand(repoRoot, "fetch", "--quiet")

		// Get behind/ahead counts
		revList, err := gitCommand(repoRoot, "rev-list", "--left-right", "--count", fmt.Sprintf("%s...%s/%s", info.Branch, strings.TrimSpace(remote), info.Branch))
		if err == nil {
			parts := strings.Fields(strings.TrimSpace(revList))
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &info.AheadRemote)
				fmt.Sscanf(parts[1], "%d", &info.BehindRemote)
			}
		}
	}

	return info, nil
}

// findGitRoot finds the root of the git repository
func findGitRoot() (string, error) {
	// Get the executable path
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks to get the real path
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	// Check if this is a local installation in ~/.local/bin
	homeDir, _ := os.UserHomeDir()
	localBinPath := filepath.Join(homeDir, ".local", "bin")

	if filepath.Dir(exe) == localBinPath {
		// This is a local installation, look for the source repository
		// Try common locations for the heimdall-cli repository
		possiblePaths := []string{
			filepath.Join(homeDir, "software-development", "heimdall-cli"),
			filepath.Join(homeDir, "projects", "heimdall-cli"),
			filepath.Join(homeDir, "code", "heimdall-cli"),
			filepath.Join(homeDir, "src", "heimdall-cli"),
			filepath.Join(homeDir, "dev", "heimdall-cli"),
			filepath.Join(homeDir, "heimdall-cli"),
		}

		// Also check GOPATH if set
		if gopath := os.Getenv("GOPATH"); gopath != "" {
			possiblePaths = append(possiblePaths,
				filepath.Join(gopath, "src", "github.com", "arthur404dev", "heimdall-cli"))
		}

		// Check each possible path for a git repository
		for _, path := range possiblePaths {
			gitDir := filepath.Join(path, ".git")
			if _, err := os.Stat(gitDir); err == nil {
				// Verify this is actually the heimdall-cli repository
				if isHeimdallRepo(path) {
					return path, nil
				}
			}
		}

		// If we couldn't find the repository, return an error
		return "", fmt.Errorf("local installation detected but source repository not found")
	}

	// Not in ~/.local/bin, walk up the directory tree looking for .git
	dir := filepath.Dir(exe)
	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not a git repository")
}

// isHeimdallRepo checks if the given path is the heimdall-cli repository
func isHeimdallRepo(path string) bool {
	// Check for key files that indicate this is the heimdall-cli repo
	requiredFiles := []string{
		"go.mod",
		"Makefile",
		filepath.Join("internal", "commands", "root.go"),
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(path, file)); err != nil {
			return false
		}
	}

	// Check go.mod contains the correct module name
	goModPath := filepath.Join(path, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "module github.com/arthur404dev/heimdall-cli")
}

// gitCommand executes a git command in the specified directory
func gitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String(), nil
}

// PerformGitUpdate is deprecated - use handleLocalUpdate from update.go instead
// Keeping for backward compatibility
func PerformGitUpdate(verbose bool) error {
	return fmt.Errorf("PerformGitUpdate is deprecated, please use the update command directly")
}

// rebuildBinary rebuilds the heimdall binary
func rebuildBinary(repoRoot string, verbose bool) error {
	// Check if we have a Makefile
	makefilePath := filepath.Join(repoRoot, "Makefile")
	if _, err := os.Stat(makefilePath); err == nil {
		// For local installations, always use install-local target
		// This is simpler and more predictable
		makeTarget := "install-local"
		fmt.Println("Building and installing to ~/.local/bin...")

		cmd := exec.Command("make", makeTarget)
		cmd.Dir = repoRoot

		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("make %s failed: %w", makeTarget, err)
		}
		return nil
	}

	// Otherwise use go build
	fmt.Println("Building with go build...")

	// Find the main.go file
	mainPath := filepath.Join(repoRoot, "cmd", "heimdall", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		// Try alternative location
		mainPath = filepath.Join(repoRoot, "main.go")
		if _, err := os.Stat(mainPath); err != nil {
			return fmt.Errorf("cannot find main.go")
		}
	}

	// Get the current executable path
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build command
	buildArgs := []string{"build", "-o", exe}

	// Add version information via ldflags
	gitCommit, _ := gitCommand(repoRoot, "rev-parse", "HEAD")
	gitCommit = strings.TrimSpace(gitCommit)

	gitTag, _ := gitCommand(repoRoot, "describe", "--tags", "--abbrev=0")
	gitTag = strings.TrimSpace(gitTag)
	if gitTag == "" {
		gitTag = "dev"
	}

	ldflags := fmt.Sprintf("-X github.com/arthur404dev/heimdall-cli/internal/commands.Version=%s -X github.com/arthur404dev/heimdall-cli/internal/commands.Commit=%s",
		gitTag, gitCommit[:8])

	buildArgs = append(buildArgs, "-ldflags", ldflags, mainPath)

	cmd := exec.Command("go", buildArgs...)
	cmd.Dir = repoRoot

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("Running: go %s\n", strings.Join(buildArgs, " "))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}

// CheckGitInstallation checks if the current binary is running from a git installation
func CheckGitInstallation() (bool, *GitInfo) {
	info, err := GetGitInfo()
	if err != nil || !info.IsRepo {
		return false, nil
	}
	return true, info
}
