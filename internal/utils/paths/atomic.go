package paths

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicWriteJSON writes JSON data to a file atomically
func AtomicWriteJSON(path string, data interface{}) error {
	// Ensure parent directory exists
	if err := EnsureParentDir(path); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Marshal data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to target
	if err := os.Rename(tmpFile, path); err != nil {
		os.Remove(tmpFile) // Clean up on failure
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// AtomicWrite writes data to a file atomically
func AtomicWrite(path string, data []byte) error {
	// Ensure parent directory exists
	if err := EnsureParentDir(path); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Write to temporary file
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to target
	if err := os.Rename(tmpFile, path); err != nil {
		os.Remove(tmpFile) // Clean up on failure
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// ComputeHash computes the SHA256 hash of a file
func ComputeHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ReadJSON reads JSON data from a file
func ReadJSON(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := EnsureParentDir(dst); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create temporary destination file
	tmpDst := dst + ".tmp"
	dstFile, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(tmpDst)
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close destination file
	if err := dstFile.Close(); err != nil {
		os.Remove(tmpDst)
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	// Atomically rename to final destination
	if err := os.Rename(tmpDst, dst); err != nil {
		os.Remove(tmpDst)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// CreateSymlink creates a symbolic link
func CreateSymlink(target, link string) error {
	// Ensure parent directory exists
	if err := EnsureParentDir(link); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Remove existing link if it exists
	if Exists(link) {
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove existing link: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// BackupFile creates a backup of a file
func BackupFile(path string) error {
	if !Exists(path) {
		return nil // Nothing to backup
	}

	backupPath := path + ".backup"

	// Find a unique backup name
	for i := 1; Exists(backupPath); i++ {
		backupPath = fmt.Sprintf("%s.backup.%d", path, i)
	}

	return CopyFile(path, backupPath)
}

// CleanPath returns the absolute path with ~ expanded
func CleanPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}
	abs, _ := filepath.Abs(path)
	return abs
}
