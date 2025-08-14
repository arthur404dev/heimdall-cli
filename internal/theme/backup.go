package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// FileBackupManager handles file backups for theme operations
type FileBackupManager struct {
	backupDir  string
	maxBackups int
}

// NewFileBackupManager creates a new backup manager
func NewFileBackupManager(backupDir string) *FileBackupManager {
	return &FileBackupManager{
		backupDir:  backupDir,
		maxBackups: 5, // Default retention
	}
}

// SetMaxBackups sets the maximum number of backups to retain
func (bm *FileBackupManager) SetMaxBackups(max int) {
	if max > 0 {
		bm.maxBackups = max
	}
}

// Backup creates a backup of the specified files
func (bm *FileBackupManager) Backup(files []string) (string, error) {
	// Generate unique backup ID with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupID := fmt.Sprintf("theme-backup-%s", timestamp)
	backupPath := filepath.Join(bm.backupDir, backupID)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Track backed up files for rollback on error
	backedUp := []string{}

	// Backup each file
	for _, file := range files {
		if err := bm.backupFile(file, backupPath); err != nil {
			// Rollback partial backup on failure
			logger.Warn("Backup failed, cleaning up partial backup", "error", err)
			os.RemoveAll(backupPath)
			return "", fmt.Errorf("failed to backup %s: %w", file, err)
		}
		backedUp = append(backedUp, file)
	}

	// Create backup manifest
	if err := bm.createManifest(backupPath, backedUp); err != nil {
		logger.Warn("Failed to create backup manifest", "error", err)
		// Non-fatal: backup is still valid without manifest
	}

	// Clean old backups
	if err := bm.cleanOldBackups(); err != nil {
		logger.Warn("Failed to clean old backups", "error", err)
		// Non-fatal: current backup is still valid
	}

	logger.Info("Backup created", "id", backupID, "files", len(backedUp))
	return backupID, nil
}

// backupFile backs up a single file
func (bm *FileBackupManager) backupFile(src, backupDir string) error {
	// Skip if file doesn't exist
	if !paths.Exists(src) {
		logger.Debug("Skipping non-existent file", "file", src)
		return nil
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Get file info for permissions
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Preserve directory structure in backup
	// Remove leading slash for relative path
	relPath := strings.TrimPrefix(src, "/")
	destPath := filepath.Join(backupDir, relPath)

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory structure: %w", err)
	}

	// Write backup file with original permissions
	if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// createManifest creates a manifest file listing all backed up files
func (bm *FileBackupManager) createManifest(backupPath string, files []string) error {
	manifest := "# Heimdall Theme Backup Manifest\n"
	manifest += fmt.Sprintf("# Created: %s\n", time.Now().Format(time.RFC3339))
	manifest += fmt.Sprintf("# Files: %d\n\n", len(files))

	for _, file := range files {
		manifest += file + "\n"
	}

	manifestPath := filepath.Join(backupPath, "MANIFEST.txt")
	return os.WriteFile(manifestPath, []byte(manifest), 0644)
}

// Restore restores files from a backup
func (bm *FileBackupManager) Restore(backupID string) error {
	backupPath := filepath.Join(bm.backupDir, backupID)

	// Verify backup exists
	if !paths.Exists(backupPath) {
		return fmt.Errorf("backup %s not found", backupID)
	}

	logger.Info("Starting restore", "backup", backupID)

	// Track restore progress
	restored := 0
	failed := 0

	// Walk through backup directory
	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and manifest
		if info.IsDir() || filepath.Base(path) == "MANIFEST.txt" {
			return nil
		}

		// Calculate original path
		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			logger.Warn("Failed to calculate relative path", "path", path, "error", err)
			failed++
			return nil // Continue with other files
		}

		// Restore to original location
		destPath := filepath.Join("/", relPath)

		// Read backup file
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Warn("Failed to read backup file", "file", path, "error", err)
			failed++
			return nil // Continue with other files
		}

		// Use atomic write for restore
		if err := paths.AtomicWrite(destPath, data); err != nil {
			logger.Warn("Failed to restore file", "file", destPath, "error", err)
			failed++
			return nil // Continue with other files
		}

		// Restore permissions
		if err := os.Chmod(destPath, info.Mode()); err != nil {
			logger.Warn("Failed to restore permissions", "file", destPath, "error", err)
			// Non-fatal: file is restored even if permissions fail
		}

		restored++
		logger.Debug("Restored file", "file", destPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("restore walk failed: %w", err)
	}

	logger.Info("Restore completed", "restored", restored, "failed", failed)

	if failed > 0 {
		return fmt.Errorf("restore completed with %d failures", failed)
	}

	return nil
}

// cleanOldBackups removes old backups exceeding retention limit
func (bm *FileBackupManager) cleanOldBackups() error {
	// List all backup directories
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No backup directory yet
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Filter and collect theme backups
	var backups []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "theme-backup-") {
			backups = append(backups, entry)
		}
	}

	// Skip if within retention limit
	if len(backups) <= bm.maxBackups {
		return nil
	}

	// Sort by name (which includes timestamp)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	// Remove oldest backups
	toRemove := len(backups) - bm.maxBackups
	for i := range toRemove {
		backupPath := filepath.Join(bm.backupDir, backups[i].Name())
		logger.Debug("Removing old backup", "backup", backups[i].Name())
		if err := os.RemoveAll(backupPath); err != nil {
			logger.Warn("Failed to remove old backup", "backup", backups[i].Name(), "error", err)
			// Continue with other backups
		}
	}

	logger.Info("Cleaned old backups", "removed", toRemove)
	return nil
}

// ListBackups returns a list of available backups
func (bm *FileBackupManager) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "theme-backup-") {
			backups = append(backups, entry.Name())
		}
	}

	// Sort newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	return backups, nil
}

// GetBackupInfo returns information about a specific backup
func (bm *FileBackupManager) GetBackupInfo(backupID string) (map[string]any, error) {
	backupPath := filepath.Join(bm.backupDir, backupID)

	if !paths.Exists(backupPath) {
		return nil, fmt.Errorf("backup %s not found", backupID)
	}

	info := make(map[string]any)
	info["id"] = backupID

	// Extract timestamp from backup ID
	if timestamp, ok := strings.CutPrefix(backupID, "theme-backup-"); ok {
		info["timestamp"] = timestamp
	}

	// Count files in backup
	fileCount := 0
	totalSize := int64(0)

	filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Base(path) != "MANIFEST.txt" {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	info["files"] = fileCount
	info["size"] = totalSize

	// Read manifest if available
	manifestPath := filepath.Join(backupPath, "MANIFEST.txt")
	if paths.Exists(manifestPath) {
		info["has_manifest"] = true
	} else {
		info["has_manifest"] = false
	}

	return info, nil
}

// RemoveBackup removes a specific backup
func (bm *FileBackupManager) RemoveBackup(backupID string) error {
	backupPath := filepath.Join(bm.backupDir, backupID)

	if !paths.Exists(backupPath) {
		return fmt.Errorf("backup %s not found", backupID)
	}

	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	logger.Info("Backup removed", "id", backupID)
	return nil
}
