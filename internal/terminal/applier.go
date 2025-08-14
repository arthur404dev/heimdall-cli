// Package terminal provides functionality for applying color schemes to terminal emulators
// through ANSI escape sequences and configuration files.
package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
)

// Applier handles applying ANSI sequences to active terminals
type Applier struct {
	sequenceBuilder *SequenceBuilder
}

// NewApplier creates a new terminal applier
func NewApplier() *Applier {
	return &Applier{
		sequenceBuilder: NewSequenceBuilder(),
	}
}

// TerminalDevice represents an active terminal device
type TerminalDevice struct {
	Path     string
	PID      int
	Writable bool
}

// ApplyToTerminals applies ANSI sequences to all active terminal devices and writes to file
func (a *Applier) ApplyToTerminals(colours map[string]string, schemeName string) error {
	// Generate sequences
	sequences, err := a.sequenceBuilder.GenerateSequences(colours)
	if err != nil {
		return fmt.Errorf("failed to generate sequences: %w", err)
	}

	// Write sequences to file for sourcing
	if err := a.writeSequencesToFile(sequences, schemeName); err != nil {
		logger.Warn("Failed to write sequences to file", "error", err)
		// Don't fail the operation, continue with PTY application
	}

	// Check if we're in a PTY environment
	if !a.isInPTY() {
		logger.Info("Not in a PTY environment, sequences written to file only")
		return nil
	}

	// Detect active terminals
	terminals, err := a.detectActiveTerminals()
	if err != nil {
		logger.Warn("Failed to detect terminals, sequences written to file", "error", err)
		return nil // Don't fail, file was written
	}

	if len(terminals) == 0 {
		logger.Info("No active terminals detected, sequences written to file")
		return nil
	}

	// Apply sequences to each terminal
	var errors []string
	successCount := 0

	for _, terminal := range terminals {
		if err := a.applySequencesToTerminal(terminal, sequences); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", terminal.Path, err))
			logger.Debug("Failed to apply sequences to terminal",
				"path", terminal.Path,
				"pid", terminal.PID,
				"error", err)
		} else {
			successCount++
			logger.Debug("Applied sequences to terminal",
				"path", terminal.Path,
				"pid", terminal.PID)
		}
	}

	logger.Info("Terminal application complete",
		"total", len(terminals),
		"success", successCount,
		"failed", len(errors))

	// Don't fail the entire operation if some terminals can't be written to
	if len(errors) > 0 {
		logger.Debug("Some terminals could not be updated", "errors", strings.Join(errors, "; "))
	}

	return nil
}

// writeSequencesToFile writes sequences to ~/.config/heimdall/sequences.txt
func (a *Applier) writeSequencesToFile(sequences []string, schemeName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "heimdall")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	sequencesPath := filepath.Join(configDir, "sequences.txt")
	content := a.sequenceBuilder.FormatSequencesForShell(sequences, schemeName)

	if err := os.WriteFile(sequencesPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write sequences file: %w", err)
	}

	logger.Info("Terminal sequences written to file", "path", sequencesPath)
	return nil
}

// isInPTY checks if we're running in a pseudo-terminal
func (a *Applier) isInPTY() bool {
	// Check if stdout is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	// Check if it's a character device (terminal)
	return fileInfo.Mode()&os.ModeCharDevice != 0
}

// detectActiveTerminals scans /dev/pts/ for active terminal devices
func (a *Applier) detectActiveTerminals() ([]*TerminalDevice, error) {
	ptsDir := "/dev/pts"

	// Check if /dev/pts exists
	if _, err := os.Stat(ptsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("/dev/pts directory not found")
	}

	// Read directory contents
	entries, err := os.ReadDir(ptsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read /dev/pts: %w", err)
	}

	var terminals []*TerminalDevice

	for _, entry := range entries {
		// Skip non-numeric entries (like ptmx)
		if !isNumeric(entry.Name()) {
			continue
		}

		devicePath := filepath.Join(ptsDir, entry.Name())

		// Check if device is active and get associated PID
		terminal, err := a.checkTerminalDevice(devicePath)
		if err != nil {
			logger.Debug("Skipping terminal device", "path", devicePath, "reason", err)
			continue
		}

		if terminal != nil {
			terminals = append(terminals, terminal)
		}
	}

	return terminals, nil
}

// checkTerminalDevice checks if a terminal device is active and writable
func (a *Applier) checkTerminalDevice(devicePath string) (*TerminalDevice, error) {
	// Get file info
	info, err := os.Stat(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat device: %w", err)
	}

	// Check if it's a character device
	if info.Mode()&os.ModeCharDevice == 0 {
		return nil, fmt.Errorf("not a character device")
	}

	// Try to find associated process
	pid, err := a.findTerminalPID(devicePath)
	if err != nil {
		return nil, fmt.Errorf("no active process found: %w", err)
	}

	// Check write permissions
	writable := a.checkWritePermission(devicePath)

	return &TerminalDevice{
		Path:     devicePath,
		PID:      pid,
		Writable: writable,
	}, nil
}

// findTerminalPID finds the PID of the process using the terminal
func (a *Applier) findTerminalPID(devicePath string) (int, error) {
	// Extract pts number from path
	ptsNum := filepath.Base(devicePath)

	// Look for processes with this terminal
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, entry := range entries {
		if !isNumeric(entry.Name()) {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Check if this process is using our terminal
		if a.processUsesTerminal(pid, ptsNum) {
			return pid, nil
		}
	}

	return 0, fmt.Errorf("no process found using terminal")
}

// processUsesTerminal checks if a process is using the specified terminal
func (a *Applier) processUsesTerminal(pid int, ptsNum string) bool {
	// Check the process's controlling terminal
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return false
	}

	// Parse stat file to get tty_nr (field 7)
	fields := strings.Fields(string(statData))
	if len(fields) < 7 {
		return false
	}

	ttyNr, err := strconv.Atoi(fields[6])
	if err != nil {
		return false
	}

	// Convert pts number to tty_nr
	// pts/N has tty_nr = 136 * 256 + N (for most systems)
	ptsNumInt, err := strconv.Atoi(ptsNum)
	if err != nil {
		return false
	}

	expectedTtyNr := 136*256 + ptsNumInt
	return ttyNr == expectedTtyNr
}

// checkWritePermission checks if we can write to the terminal device
func (a *Applier) checkWritePermission(devicePath string) bool {
	// Try to open for writing
	file, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

// applySequencesToTerminal writes ANSI sequences to a specific terminal
func (a *Applier) applySequencesToTerminal(terminal *TerminalDevice, sequences []string) error {
	if !terminal.Writable {
		return fmt.Errorf("terminal not writable")
	}

	// Open terminal for writing
	file, err := os.OpenFile(terminal.Path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open terminal: %w", err)
	}
	defer file.Close()

	// Write each sequence
	for _, sequence := range sequences {
		// Convert escape sequences to actual escape characters
		actualSequence := strings.ReplaceAll(sequence, "\\033", "\033")
		actualSequence = strings.ReplaceAll(actualSequence, "\\\\", "\\")

		if _, err := file.WriteString(actualSequence); err != nil {
			return fmt.Errorf("failed to write sequence: %w", err)
		}
	}

	// Note: Sync() is not needed for terminal devices and may cause errors
	// The sequences are applied immediately when written

	return nil
}

// GetActiveTerminalCount returns the number of active terminals
func (a *Applier) GetActiveTerminalCount() (int, error) {
	terminals, err := a.detectActiveTerminals()
	if err != nil {
		return 0, err
	}
	return len(terminals), nil
}

// ListActiveTerminals returns information about active terminals
func (a *Applier) ListActiveTerminals() ([]*TerminalDevice, error) {
	return a.detectActiveTerminals()
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ApplySequencesWithFallback applies sequences with graceful error handling
func (a *Applier) ApplySequencesWithFallback(colours map[string]string, schemeName string) error {
	err := a.ApplyToTerminals(colours, schemeName)
	if err != nil {
		// Log the error but don't fail the entire operation
		logger.Warn("Terminal application failed, continuing with other theme applications", "error", err)
		return nil
	}
	return nil
}
