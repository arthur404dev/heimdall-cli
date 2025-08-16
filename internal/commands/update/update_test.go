package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateInfo(t *testing.T) {
	info := &UpdateInfo{
		Available:      true,
		CurrentVersion: "1.0.0",
		LatestVersion:  "1.1.0",
		ReleaseURL:     "https://github.com/test/test/releases/tag/v1.1.0",
		DownloadURL:    "https://github.com/test/test/releases/download/v1.1.0/test.tar.gz",
		Checksum:       "abc123",
		IsGitInstall:   false,
		Channel:        "stable",
	}

	if !info.Available {
		t.Error("Expected update to be available")
	}

	if info.CurrentVersion >= info.LatestVersion {
		t.Error("Latest version should be newer than current")
	}
}

func TestUpdateConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Test default config
	config := DefaultUpdateConfig()
	if !config.CheckEnabled {
		t.Error("Check should be enabled by default")
	}
	if config.CheckFrequency != "daily" {
		t.Error("Default frequency should be daily")
	}

	// Test save and load
	config.Channel = "beta"
	config.LastCheck = time.Now()

	if err := SaveUpdateConfig(config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := LoadUpdateConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.Channel != "beta" {
		t.Error("Channel not preserved after save/load")
	}
}

func TestShouldCheckForUpdates(t *testing.T) {
	tests := []struct {
		name     string
		config   *UpdateConfig
		expected bool
	}{
		{
			name: "disabled checks",
			config: &UpdateConfig{
				CheckEnabled: false,
				LastCheck:    time.Now().Add(-48 * time.Hour),
			},
			expected: false,
		},
		{
			name: "recent check",
			config: &UpdateConfig{
				CheckEnabled:   true,
				CheckFrequency: "daily",
				LastCheck:      time.Now().Add(-1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "old check",
			config: &UpdateConfig{
				CheckEnabled:   true,
				CheckFrequency: "daily",
				LastCheck:      time.Now().Add(-48 * time.Hour),
			},
			expected: true,
		},
		{
			name: "never checked",
			config: &UpdateConfig{
				CheckEnabled:   true,
				CheckFrequency: "daily",
				LastCheck:      time.Time{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldCheckForUpdates(tt.config)
			if result != tt.expected {
				t.Errorf("ShouldCheckForUpdates() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBackupPath(t *testing.T) {
	path := getBackupPath("/usr/local/bin/heimdall")

	if !strings.Contains(path, "heimdall") {
		t.Error("Backup path should contain 'heimdall'")
	}

	if !strings.Contains(path, "backups") {
		t.Error("Backup path should be in backups directory")
	}

	// Check timestamp format
	name := filepath.Base(path)
	if !strings.HasPrefix(name, "heimdall-") {
		t.Error("Backup name should start with 'heimdall-'")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestGitDetection(t *testing.T) {
	// This test will vary based on where it's run
	isGit, info := CheckGitInstallation()

	if isGit {
		if info == nil {
			t.Error("Git info should not be nil when in git repo")
		}
		if info.Branch == "" {
			t.Error("Branch should be set when in git repo")
		}
	}
}

func TestChecksumVerification(t *testing.T) {
	// Create a temp file
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate expected checksum (SHA256 of "test content")
	expectedChecksum := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	// Test with correct checksum
	err := verifyChecksum(tmpFile, expectedChecksum)
	if err != nil {
		t.Errorf("Checksum verification failed with correct checksum: %v", err)
	}

	// Test with incorrect checksum
	err = verifyChecksum(tmpFile, "wrongchecksum")
	if err == nil {
		t.Error("Checksum verification should fail with incorrect checksum")
	}
}
