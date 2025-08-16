package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DownloadProgress tracks download progress
type DownloadProgress struct {
	Total      int64
	Downloaded int64
	StartTime  time.Time
}

// ProgressWriter wraps an io.Writer and tracks progress
type ProgressWriter struct {
	io.Writer
	Progress   *DownloadProgress
	OnProgress func(downloaded, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.Progress.Downloaded += int64(n)
	if pw.OnProgress != nil {
		pw.OnProgress(pw.Progress.Downloaded, pw.Progress.Total)
	}

	return n, nil
}

// DownloadRelease downloads a release asset
func DownloadRelease(url string, checksum string, verbose bool) (string, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "heimdall-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Parse filename from URL
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	tmpFile := filepath.Join(tmpDir, filename)

	// Create the file
	out, err := os.Create(tmpFile)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download the file
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Set up progress tracking
	progress := &DownloadProgress{
		Total:     resp.ContentLength,
		StartTime: time.Now(),
	}

	var writer io.Writer = out

	if verbose && resp.ContentLength > 0 {
		writer = &ProgressWriter{
			Writer:   out,
			Progress: progress,
			OnProgress: func(downloaded, total int64) {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\rDownloading: %.1f%% (%s / %s)",
					percent,
					formatBytes(downloaded),
					formatBytes(total))
			},
		}
	}

	// Copy the response body to the file
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	if verbose {
		fmt.Println() // New line after progress
	}

	// Verify checksum if provided
	if checksum != "" {
		if verbose {
			fmt.Println("Verifying checksum...")
		}

		if err := verifyChecksum(tmpFile, checksum); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("checksum verification failed: %w", err)
		}

		if verbose {
			fmt.Println("Checksum verified ✓")
		}
	}

	// Extract if it's an archive
	extractedPath, err := extractIfArchive(tmpFile, tmpDir, verbose)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract: %w", err)
	}

	return extractedPath, nil
}

// verifyChecksum verifies the SHA256 checksum of a file
func verifyChecksum(filePath string, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// extractIfArchive extracts an archive file if necessary
func extractIfArchive(filePath string, destDir string, verbose bool) (string, error) {
	// Check file extension
	ext := strings.ToLower(filePath)

	switch {
	case strings.HasSuffix(ext, ".tar.gz") || strings.HasSuffix(ext, ".tgz"):
		if verbose {
			fmt.Println("Extracting tar.gz archive...")
		}
		return extractTarGz(filePath, destDir)

	case strings.HasSuffix(ext, ".zip"):
		if verbose {
			fmt.Println("Extracting zip archive...")
		}
		return extractZip(filePath, destDir)

	case strings.HasSuffix(ext, ".exe") || !strings.Contains(ext, "."):
		// Binary file, no extraction needed
		return filePath, nil

	default:
		// Assume it's a binary
		return filePath, nil
	}
}

// extractTarGz extracts a tar.gz file
func extractTarGz(filePath string, destDir string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var binaryPath string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return "", err
			}

		case tar.TypeReg:
			outFile, err := os.Create(targetPath)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			// Set permissions
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return "", err
			}

			// Check if this is the heimdall binary
			if strings.Contains(header.Name, "heimdall") && !strings.Contains(header.Name, ".") {
				binaryPath = targetPath
			}
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("heimdall binary not found in archive")
	}

	return binaryPath, nil
}

// extractZip extracts a zip file
func extractZip(filePath string, destDir string) (string, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var binaryPath string

	for _, file := range reader.File {
		targetPath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(targetPath, file.Mode())
			continue
		}

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return "", err
		}

		// Extract file
		fileReader, err := file.Open()
		if err != nil {
			return "", err
		}
		defer fileReader.Close()

		targetFile, err := os.Create(targetPath)
		if err != nil {
			return "", err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return "", err
		}

		// Set permissions
		if err := os.Chmod(targetPath, file.Mode()); err != nil {
			return "", err
		}

		// Check if this is the heimdall binary
		if strings.Contains(file.Name, "heimdall") &&
			(runtime.GOOS == "windows" && strings.HasSuffix(file.Name, ".exe") ||
				runtime.GOOS != "windows" && !strings.Contains(file.Name, ".")) {
			binaryPath = targetPath
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("heimdall binary not found in archive")
	}

	return binaryPath, nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
