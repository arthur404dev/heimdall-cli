package theme

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
)

// ReloadKittyInstances sends a signal to all running kitty instances to reload their configuration
func ReloadKittyInstances() error {
	logger.Debug("Starting kitty reload process")

	// Try multiple methods to reload kitty
	// Method 1: Try direct remote control (works if we're in a kitty terminal with remote control)
	if os.Getenv("KITTY_PID") != "" {
		cmd := exec.Command("kitten", "@", "load-config")
		cmd.Env = os.Environ()

		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()

		select {
		case err := <-done:
			if err == nil {
				logger.Info("Successfully reloaded kitty via direct remote control")
				return nil
			}
			logger.Debug("Direct remote control failed", "error", err)
		case <-time.After(500 * time.Millisecond):
			// Very short timeout for direct control
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			logger.Debug("Direct remote control timed out")
		}
	}

	// Method 2: Try with listening socket if available
	if listenOn := os.Getenv("KITTY_LISTEN_ON"); listenOn != "" {
		cmd := exec.Command("kitten", "@", "--to", listenOn, "load-config")

		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()

		select {
		case err := <-done:
			if err == nil {
				logger.Info("Successfully reloaded kitty via socket", "socket", listenOn)
				return nil
			}
			logger.Debug("Socket-based remote control failed", "error", err)
		case <-time.After(1 * time.Second):
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			logger.Debug("Socket-based remote control timed out")
		}
	}

	// Method 3: Always fall back to signal method which is most reliable
	logger.Debug("Falling back to signal-based reload")
	return reloadKittyViaSignal()
}

// reloadKittyViaSignal sends SIGUSR1 to all kitty processes to trigger config reload
func reloadKittyViaSignal() error {
	// Get all kitty process IDs
	pids, err := getKittyPIDs()
	if err != nil {
		logger.Debug("Could not find kitty processes", "error", err)
		return nil // Not an error if no kitty processes are running
	}

	if len(pids) == 0 {
		logger.Debug("No kitty processes found")
		return nil
	}

	logger.Debug("Found kitty processes", "pids", pids)

	// Send SIGUSR1 to each kitty process
	var errors []string
	successCount := 0
	for _, pid := range pids {
		cmd := exec.Command("kill", "-USR1", pid)
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("PID %s: %v", pid, err))
			logger.Debug("Failed to signal kitty process", "pid", pid, "error", err)
		} else {
			successCount++
			logger.Debug("Successfully signaled kitty process", "pid", pid)
		}
	}

	if successCount > 0 {
		logger.Info("Reloaded kitty instances via SIGUSR1 signal", "count", successCount, "total", len(pids))
	}

	if len(errors) > 0 {
		logger.Debug("Some kitty processes could not be signaled", "errors", strings.Join(errors, ", "))
	}

	return nil
}

// getKittyPIDs returns the PIDs of all running kitty processes
func getKittyPIDs() ([]string, error) {
	// Use pgrep to find kitty processes
	cmd := exec.Command("pgrep", "-x", "kitty")
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns exit code 1 if no processes found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, err
	}

	// Parse PIDs from output
	pids := strings.Fields(string(output))
	return pids, nil
}

// ReloadKittyForApp reloads kitty only if the app being themed is kitty
func ReloadKittyForApp(app string) error {
	if app == "kitty" {
		return ReloadKittyInstances()
	}
	return nil
}
