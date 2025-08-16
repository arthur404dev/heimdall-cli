package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UpdateConfig stores update-related configuration
type UpdateConfig struct {
	CheckEnabled      bool      `json:"check_enabled"`
	CheckFrequency    string    `json:"check_frequency"` // daily, weekly, monthly
	LastCheck         time.Time `json:"last_check"`
	Channel           string    `json:"channel"`
	NotifyOnAvailable bool      `json:"notify_on_available"`
}

// DefaultUpdateConfig returns the default update configuration
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		CheckEnabled:      true,
		CheckFrequency:    "daily",
		LastCheck:         time.Time{},
		Channel:           "stable",
		NotifyOnAvailable: true,
	}
}

// LoadUpdateConfig loads the update configuration
func LoadUpdateConfig() (*UpdateConfig, error) {
	configPath := getUpdateConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultUpdateConfig(), nil
		}
		return nil, err
	}

	var config UpdateConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveUpdateConfig saves the update configuration
func SaveUpdateConfig(config *UpdateConfig) error {
	configPath := getUpdateConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// getUpdateConfigPath returns the path to the update config file
func getUpdateConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "heimdall", "update.json")
}

// ShouldCheckForUpdates determines if an update check should be performed
func ShouldCheckForUpdates(config *UpdateConfig) bool {
	if !config.CheckEnabled {
		return false
	}

	now := time.Now()
	duration := getCheckDuration(config.CheckFrequency)

	return now.Sub(config.LastCheck) >= duration
}

// getCheckDuration returns the duration for the check frequency
func getCheckDuration(frequency string) time.Duration {
	switch frequency {
	case "hourly":
		return time.Hour
	case "daily":
		return 24 * time.Hour
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour // Default to daily
	}
}

// PassiveUpdateCheck performs a passive update check in the background
func PassiveUpdateCheck() {
	// Load config
	config, err := LoadUpdateConfig()
	if err != nil {
		return // Silently fail
	}

	// Check if we should check for updates
	if !ShouldCheckForUpdates(config) {
		return
	}

	// Create GitHub client
	client := NewGitHubClient()

	// Check for updates
	release, err := client.GetLatestRelease(config.Channel)
	if err != nil {
		return // Silently fail
	}

	// Get current version
	currentMeta, err := GetCurrentVersion()
	if err != nil {
		return
	}

	// Parse latest version
	latestVersion, err := ParseVersion(release.TagName)
	if err != nil {
		return
	}

	// Check if update is available
	if latestVersion.IsNewer(currentMeta.Version) && config.NotifyOnAvailable {
		// Show notification
		fmt.Fprintf(os.Stderr, "\n╭─────────────────────────────────────────╮\n")
		fmt.Fprintf(os.Stderr, "│  Update Available: %s → %s  │\n",
			currentMeta.Version.String(), latestVersion.String())
		fmt.Fprintf(os.Stderr, "│  Run 'heimdall update' to upgrade      │\n")
		fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────╯\n\n")
	}

	// Update last check time
	config.LastCheck = time.Now()
	SaveUpdateConfig(config)
}

// ConfigureUpdateSettings configures update settings
func ConfigureUpdateSettings(enabled bool, frequency string, channel string, notify bool) error {
	config, err := LoadUpdateConfig()
	if err != nil {
		config = DefaultUpdateConfig()
	}

	config.CheckEnabled = enabled
	config.CheckFrequency = frequency
	config.Channel = channel
	config.NotifyOnAvailable = notify

	return SaveUpdateConfig(config)
}

// ShowUpdateSettings displays current update settings
func ShowUpdateSettings() error {
	config, err := LoadUpdateConfig()
	if err != nil {
		return err
	}

	fmt.Println("Update Settings:")
	fmt.Printf("  Automatic checks: %v\n", config.CheckEnabled)
	fmt.Printf("  Check frequency: %s\n", config.CheckFrequency)
	fmt.Printf("  Release channel: %s\n", config.Channel)
	fmt.Printf("  Notifications: %v\n", config.NotifyOnAvailable)
	fmt.Printf("  Last check: %s\n", formatLastCheck(config.LastCheck))

	return nil
}

// formatLastCheck formats the last check time
func formatLastCheck(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	duration := time.Since(t)
	if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}
