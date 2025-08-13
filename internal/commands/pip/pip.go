package pip

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// Command creates the pip command
func Command() *cobra.Command {
	var (
		daemon bool
		stop   bool
		status bool
	)

	cmd := &cobra.Command{
		Use:   "pip",
		Short: "Picture-in-picture daemon",
		Long: `Manage picture-in-picture mode for windows.
		
The PIP daemon monitors the active window and automatically
enables picture-in-picture mode for supported applications
like video players and browsers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if stop {
				return stopDaemon()
			}

			if status {
				return showStatus()
			}

			if daemon {
				// Start daemon
				return startDaemon()
			}

			// Default behavior - start daemon
			return startDaemon()
		},
	}

	cmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Start PIP daemon")
	cmd.Flags().BoolVar(&stop, "stop", false, "Stop the PIP daemon")
	cmd.Flags().BoolVar(&status, "status", false, "Show daemon status")

	return cmd
}

// startDaemon starts the PIP daemon
func startDaemon() error {
	pidFile := filepath.Join(paths.StateDir, "pip.pid")

	// Check if already running
	if isDaemonRunning(pidFile) {
		return fmt.Errorf("PIP daemon is already running")
	}

	// Fork to background
	if os.Getenv("PIP_DAEMON") != "1" {
		// Re-execute ourselves with daemon flag
		cmd := exec.Command(os.Args[0], "pip")
		cmd.Env = append(os.Environ(), "PIP_DAEMON=1")

		// Detach from parent
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		// Start daemon
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Write PID file
		pid := cmd.Process.Pid
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			cmd.Process.Kill()
			return fmt.Errorf("failed to write PID file: %w", err)
		}

		// Send notification
		notifier := notify.NewNotifier()
		notifier.Send(&notify.Notification{
			Summary: "PIP Daemon",
			Body:    "Picture-in-picture daemon started",
			Urgency: notify.UrgencyNormal,
		})

		fmt.Println("PIP daemon started")
		return nil
	}

	// We are the daemon process
	return runDaemon()
}

// runDaemon runs the main daemon loop
func runDaemon() error {
	logger.Info("PIP daemon starting")

	// Load configuration
	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Hyprland client
	client, err := hypr.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Hyprland client: %w", err)
	}

	// Subscribe to window events
	events, err := client.Subscribe([]string{"activewindow", "closewindow"})
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Track PIP windows
	pipWindows := make(map[string]bool)

	// Main event loop
	for event := range events {
		switch event.Type {
		case "activewindow":
			// Parse window info from event data
			parts := strings.Split(event.Data, ",")
			if len(parts) >= 2 {
				windowClass := parts[0]
				windowTitle := parts[1]

				// Check if this is a video window
				if isVideoWindow(windowClass, windowTitle) {
					if err := enablePIP(client, windowClass); err != nil {
						logger.Error("Failed to enable PIP", "class", windowClass, "error", err)
					} else {
						pipWindows[windowClass] = true
						logger.Info("Enabled PIP", "class", windowClass)
					}
				}
			}

		case "closewindow":
			// Remove from PIP windows if tracked
			if address := event.Data; pipWindows[address] {
				delete(pipWindows, address)
				logger.Info("PIP window closed", "address", address)
			}
		}
	}

	return nil
}

// isVideoWindow checks if a window is likely playing video
func isVideoWindow(class, title string) bool {
	class = strings.ToLower(class)
	title = strings.ToLower(title)

	// Check for common video player applications
	videoApps := []string{
		"mpv", "vlc", "firefox", "chromium", "chrome",
		"brave", "youtube", "netflix", "twitch", "spotify",
	}

	for _, app := range videoApps {
		if strings.Contains(class, app) {
			// Additional checks for browsers
			if strings.Contains(class, "firefox") || strings.Contains(class, "chrom") || strings.Contains(class, "brave") {
				// Check if title indicates video
				videoKeywords := []string{
					"youtube", "netflix", "twitch", "vimeo",
					"- playing", "▶", "►", "video", "stream",
				}
				for _, keyword := range videoKeywords {
					if strings.Contains(title, keyword) {
						return true
					}
				}
				return false
			}
			return true
		}
	}

	return false
}

// enablePIP enables picture-in-picture for a window
func enablePIP(client *hypr.Client, windowClass string) error {
	// Get current window info
	windows, err := client.GetWindows()
	if err != nil {
		return fmt.Errorf("failed to get windows: %w", err)
	}

	// Find the window
	var targetWindow *hypr.Window
	for _, w := range windows {
		if w.Class == windowClass {
			targetWindow = &w
			break
		}
	}

	if targetWindow == nil {
		return fmt.Errorf("window not found: %s", windowClass)
	}

	// Enable PIP mode
	// Float the window
	if err := client.Dispatch("togglefloating", targetWindow.Address); err != nil {
		return fmt.Errorf("failed to float window: %w", err)
	}

	// Resize to PIP size (e.g., 25% of screen)
	if err := client.Dispatch("resizeactive", "25%", "25%"); err != nil {
		return fmt.Errorf("failed to resize window: %w", err)
	}

	// Move to corner (bottom-right)
	if err := client.Dispatch("moveactive", "70%", "70%"); err != nil {
		return fmt.Errorf("failed to move window: %w", err)
	}

	// Pin the window so it stays on all workspaces
	if err := client.Dispatch("pin"); err != nil {
		logger.Error("Failed to pin window", "error", err)
	}

	// Set always on top
	if err := client.Dispatch("bringactivetotop"); err != nil {
		logger.Error("Failed to bring to top", "error", err)
	}

	return nil
}

// stopDaemon stops the running PIP daemon
func stopDaemon() error {
	pidFile := filepath.Join(paths.StateDir, "pip.pid")

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("PIP daemon is not running")
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	// Find process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Check if still running and force kill if necessary
	if err := process.Signal(syscall.Signal(0)); err == nil {
		// Still running, force kill
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill daemon: %w", err)
		}
	}

	// Remove PID file
	os.Remove(pidFile)

	// Send notification
	notifier := notify.NewNotifier()
	notifier.Send(&notify.Notification{
		Summary: "PIP Daemon",
		Body:    "Picture-in-picture daemon stopped",
		Urgency: notify.UrgencyNormal,
	})

	fmt.Println("PIP daemon stopped")
	return nil
}

// showStatus shows the daemon status
func showStatus() error {
	pidFile := filepath.Join(paths.StateDir, "pip.pid")

	if !isDaemonRunning(pidFile) {
		fmt.Println("PIP daemon is not running")
		return nil
	}

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid := strings.TrimSpace(string(data))
	fmt.Printf("PIP daemon is running (PID: %s)\n", pid)

	return nil
}

// isDaemonRunning checks if the daemon is already running
func isDaemonRunning(pidFile string) bool {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is alive
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
