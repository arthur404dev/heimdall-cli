package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ReplaceBinary replaces the current binary with the new one
func ReplaceBinary(newBinaryPath string, createBackup bool) error {
	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable: %w", err)
	}

	// Resolve any symlinks
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Create backup if requested
	if createBackup {
		backupPath := getBackupPath(currentExe)
		if err := createBinaryBackup(currentExe, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Backup created: %s\n", backupPath)
	}

	// Platform-specific replacement
	switch runtime.GOOS {
	case "windows":
		return replaceWindowsBinary(currentExe, newBinaryPath)
	default:
		return replaceUnixBinary(currentExe, newBinaryPath)
	}
}

// replaceUnixBinary replaces the binary on Unix-like systems
func replaceUnixBinary(currentPath, newPath string) error {
	// Read the new binary
	newBinary, err := os.Open(newPath)
	if err != nil {
		return fmt.Errorf("failed to open new binary: %w", err)
	}
	defer newBinary.Close()

	// Get file info for permissions
	info, err := newBinary.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat new binary: %w", err)
	}

	// Create temporary file in the same directory as current binary
	dir := filepath.Dir(currentPath)
	tmpFile, err := os.CreateTemp(dir, ".heimdall-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Copy new binary to temp file
	_, err = io.Copy(tmpFile, newBinary)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(tmpPath, info.Mode()|0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, currentPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// replaceWindowsBinary replaces the binary on Windows
func replaceWindowsBinary(currentPath, newPath string) error {
	// On Windows, we can't replace a running executable directly
	// We need to use a batch script that runs after the process exits

	// Create update script
	scriptPath := currentPath + ".update.bat"
	script := fmt.Sprintf(`@echo off
echo Updating heimdall...
ping 127.0.0.1 -n 2 > nul
move /Y "%s" "%s.old" > nul 2>&1
move /Y "%s" "%s"
if %%errorlevel%% == 0 (
    echo Update complete!
    del "%s.old" > nul 2>&1
    del "%%~f0"
) else (
    echo Update failed!
    move /Y "%s.old" "%s" > nul 2>&1
)
`, currentPath, currentPath, newPath, currentPath, currentPath, currentPath, currentPath)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create update script: %w", err)
	}

	fmt.Println("Update prepared. Please run the following command to complete:")
	fmt.Printf("  %s\n", scriptPath)
	fmt.Println("\nOr restart heimdall and it will auto-update.")

	return nil
}

// createBinaryBackup creates a backup of the current binary
func createBinaryBackup(currentPath, backupPath string) error {
	// Ensure backup directory exists
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(currentPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	// Copy permissions
	info, err := src.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(backupPath, info.Mode())
}

// getBackupPath returns the path for the backup binary
func getBackupPath(currentPath string) string {
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".cache", "heimdall", "backups")

	// Use timestamp in backup name
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("heimdall-%s", timestamp)

	if runtime.GOOS == "windows" {
		backupName += ".exe"
	}

	return filepath.Join(backupDir, backupName)
}

// Rollback rolls back to a previous version
func Rollback() error {
	// Get current executable
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable: %w", err)
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Find latest backup
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".cache", "heimdall", "backups")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("no backups found: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no backups available")
	}

	// Find the most recent backup
	var latestBackup os.DirEntry
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestBackup = entry
		}
	}

	if latestBackup == nil {
		return fmt.Errorf("no valid backups found")
	}

	backupPath := filepath.Join(backupDir, latestBackup.Name())

	// Get version info from backup
	fmt.Printf("Rolling back to backup: %s\n", latestBackup.Name())

	// Replace current binary with backup
	if err := ReplaceBinary(backupPath, false); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback complete! Please restart heimdall.")
	return nil
}

// CleanupBackups removes old backup files
func CleanupBackups(keepCount int) error {
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".cache", "heimdall", "backups")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		// No backup directory, nothing to clean
		return nil
	}

	// Sort by modification time
	type backupFile struct {
		path string
		time time.Time
	}

	var backups []backupFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, backupFile{
			path: filepath.Join(backupDir, entry.Name()),
			time: info.ModTime(),
		})
	}

	// Keep only the most recent backups
	if len(backups) > keepCount {
		// Sort by time (oldest first)
		for i := 0; i < len(backups)-1; i++ {
			for j := i + 1; j < len(backups); j++ {
				if backups[j].time.Before(backups[i].time) {
					backups[i], backups[j] = backups[j], backups[i]
				}
			}
		}

		// Remove old backups
		for i := 0; i < len(backups)-keepCount; i++ {
			os.Remove(backups[i].path)
		}
	}

	return nil
}
