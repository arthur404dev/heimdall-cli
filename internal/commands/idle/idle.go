package idle

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/manager"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// Command creates the idle command
func Command() *cobra.Command {
	var (
		timer    string
		stop     bool
		status   bool
		list     bool
		reason   string
		provider string
		stopAll  bool
		daemon   bool
	)

	cmd := &cobra.Command{
		Use:   "idle [OPTIONS]",
		Short: "Manage system idle prevention",
		Long: `Prevent system from going idle/sleep, similar to caffeinate/caffeine.
		
This command prevents your system from entering idle state, activating the
screensaver, or going to sleep. It works across different desktop environments
and display servers (X11, Wayland) using the most appropriate method available.

Note: By default, idle prevention is active only while this command is running.
Use the -d/--daemon flag to run in the background. The inhibition is automatically
released when the process exits (or timer expires).

Examples:
  heimdall idle                    # Start idle prevention (runs until Ctrl+C)
  heimdall idle -d                 # Run in background (daemon mode)
  heimdall idle -t 30m             # Prevent idle for 30 minutes
  heimdall idle -d -t 2h           # Prevent idle for 2 hours in background
  heimdall idle --status           # Check current status (including daemons)
  heimdall idle --stop             # Stop idle prevention (including daemons)
  heimdall idle --list             # List all active sessions
  heimdall idle --provider systemd # Use specific provider`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(timer, stop, stopAll, status, list, reason, provider, daemon)
		},
	}

	cmd.Flags().StringVarP(&timer, "timer", "t", "", "Set timer duration (e.g., 30m, 2h, 1h30m)")
	cmd.Flags().BoolVarP(&stop, "stop", "s", false, "Stop idle prevention (last session)")
	cmd.Flags().BoolVar(&stopAll, "stop-all", false, "Stop all idle prevention sessions")
	cmd.Flags().BoolVar(&status, "status", false, "Show current status")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List active sessions")
	cmd.Flags().StringVarP(&reason, "reason", "r", "Heimdall idle prevention", "Reason for preventing idle")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "Force specific provider (auto-detect if not specified)")
	cmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Run in background (daemon mode)")

	return cmd
}

func run(timer string, stop, stopAll, status, list bool, reason, provider string, daemon bool) error {
	// Create manager
	mgr := manager.NewManager()

	// Handle cleanup on exit
	setupSignalHandler(mgr)

	// Handle stop command
	if stop || stopAll {
		return handleStop(mgr, stopAll)
	}

	// Handle status command
	if status {
		return handleStatus(mgr)
	}

	// Handle list command
	if list {
		return handleList(mgr)
	}

	// Parse duration if timer is specified
	var duration time.Duration
	if timer != "" {
		d, err := manager.ParseDuration(timer)
		if err != nil {
			return fmt.Errorf("invalid timer duration: %w", err)
		}
		duration = d
	}

	// Handle daemon mode
	if daemon {
		return runDaemon(mgr, reason, duration, provider)
	}

	// Start idle prevention
	return handleStart(mgr, reason, duration, provider)
}

func handleStart(mgr *manager.Manager, reason string, duration time.Duration, provider string) error {
	// Check if we already have active sessions
	sessions := mgr.ListSessions()
	if len(sessions) > 0 {
		fmt.Printf("Note: %d session(s) already active\n", len(sessions))
	}

	// Start new session
	session, err := mgr.Start(reason, duration, provider)
	if err != nil {
		return fmt.Errorf("failed to start idle prevention: %w", err)
	}

	// Display session info
	fmt.Printf("✓ Idle prevention started\n")
	fmt.Printf("  Session ID: %s\n", session.ID[:8])
	fmt.Printf("  Provider:   %s\n", session.Provider)
	fmt.Printf("  Reason:     %s\n", session.Reason)

	if duration > 0 {
		fmt.Printf("  Duration:   %s\n", manager.FormatDuration(duration))
		fmt.Printf("  Expires at: %s\n", session.ExpiresAt.Format("15:04:05"))

		// Send notification
		notifier := notify.NewNotifier()
		notifier.Send(&notify.Notification{
			Summary: "Idle Prevention Started",
			Body:    fmt.Sprintf("Will expire in %s", manager.FormatDuration(duration)),
			Urgency: notify.UrgencyNormal,
		})

		// If duration is more than 1 minute, schedule a warning notification
		if duration > time.Minute {
			warningTime := duration - time.Minute
			time.AfterFunc(warningTime, func() {
				notifier.Send(&notify.Notification{
					Summary: "Idle Prevention Expiring Soon",
					Body:    "Will expire in 1 minute",
					Urgency: notify.UrgencyNormal,
				})
			})
		}

		// Wait for the timer to expire
		fmt.Printf("\nPress Ctrl+C to stop early...\n")
		time.Sleep(duration)

		// Timer expired, stop the session
		mgr.Stop(session.ID)
		fmt.Printf("\n✓ Idle prevention expired\n")

	} else {
		fmt.Printf("  Duration:   unlimited (press Ctrl+C to stop)\n")

		// Send notification
		notifier := notify.NewNotifier()
		notifier.Send(&notify.Notification{
			Summary: "Idle Prevention Started",
			Body:    "Running indefinitely",
			Urgency: notify.UrgencyNormal,
		})

		// For unlimited duration, keep the process running
		fmt.Printf("\nPress Ctrl+C to stop...\n")

		// Wait indefinitely (will be interrupted by signal)
		select {}
	}

	// Show environment info in debug mode
	env := mgr.GetEnvironment()
	logger.Debug("Environment detected",
		"display_server", env.DisplayServer,
		"desktop", env.DesktopEnv,
		"dbus", env.HasDBus,
		"systemd", env.HasSystemd)

	return nil
}

func handleStop(mgr *manager.Manager, stopAll bool) error {
	// First check for daemon process
	pidFile := filepath.Join(paths.StateDir, "heimdall-idle.pid")
	if pidData, err := os.ReadFile(pidFile); err == nil {
		var pid int
		if _, err := fmt.Sscanf(string(pidData), "%d", &pid); err == nil {
			// Try to stop the daemon
			if proc, err := os.FindProcess(pid); err == nil {
				if err := proc.Signal(syscall.SIGTERM); err == nil {
					fmt.Printf("✓ Stopped daemon process (PID: %d)\n", pid)
					// Remove PID file
					os.Remove(pidFile)

					// Send notification
					notifier := notify.NewNotifier()
					notifier.Send(&notify.Notification{
						Summary: "Idle Prevention Stopped",
						Body:    "Daemon process terminated",
						Urgency: notify.UrgencyNormal,
					})
					return nil
				}
			}
			// Process not found or couldn't signal, remove stale PID file
			os.Remove(pidFile)
		}
	}

	// Handle regular session stop
	if stopAll {
		if err := mgr.StopAll(); err != nil {
			// No sessions in current process, but that's okay if we stopped a daemon
			if pidFile != "" {
				return nil
			}
			return fmt.Errorf("failed to stop sessions: %w", err)
		}
		fmt.Println("✓ All idle prevention sessions stopped")
	} else {
		// Stop the most recent session
		sessions := mgr.ListSessions()
		if len(sessions) == 0 {
			return fmt.Errorf("no active sessions to stop")
		}

		// Find the most recent session
		var mostRecent *manager.Session
		for _, s := range sessions {
			if mostRecent == nil || s.StartTime.After(mostRecent.StartTime) {
				mostRecent = s
			}
		}

		if err := mgr.Stop(mostRecent.ID); err != nil {
			return fmt.Errorf("failed to stop session: %w", err)
		}

		fmt.Printf("✓ Stopped session %s\n", mostRecent.ID[:8])

		// Check if there are more sessions
		remaining := len(sessions) - 1
		if remaining > 0 {
			fmt.Printf("Note: %d session(s) still active\n", remaining)
		}
	}

	// Send notification
	notifier := notify.NewNotifier()
	notifier.Send(&notify.Notification{
		Summary: "Idle Prevention Stopped",
		Body:    "System can now go idle",
		Urgency: notify.UrgencyNormal,
	})

	return nil
}

func handleStatus(mgr *manager.Manager) error {
	active, sessions, providers := mgr.GetStatus()

	// Check for daemon process
	daemonActive := false
	var daemonPID int
	pidFile := filepath.Join(paths.StateDir, "heimdall-idle.pid")
	if pidData, err := os.ReadFile(pidFile); err == nil {
		if _, err := fmt.Sscanf(string(pidData), "%d", &daemonPID); err == nil {
			// Check if process is still running
			if proc, err := os.FindProcess(daemonPID); err == nil {
				// Try sending signal 0 to check if process exists
				if err := proc.Signal(syscall.Signal(0)); err == nil {
					daemonActive = true
				}
			}
		}
	}

	if active || daemonActive {
		fmt.Println("✓ Idle prevention is ACTIVE")

		if daemonActive {
			fmt.Printf("\nDaemon Process:\n")
			fmt.Printf("  PID: %d\n", daemonPID)
			fmt.Printf("  Status: Running in background\n")
			fmt.Printf("  Stop with: heimdall idle --stop\n")
		}

		if len(sessions) > 0 {
			fmt.Printf("\nActive Sessions (%d):\n", len(sessions))

			for _, session := range sessions {
				fmt.Printf("\n  Session %s:\n", session.ID[:8])
				fmt.Printf("    Provider:  %s\n", session.Provider)
				fmt.Printf("    Started:   %s\n", session.StartTime.Format("15:04:05"))
				fmt.Printf("    Reason:    %s\n", session.Reason)

				if session.Duration > 0 {
					fmt.Printf("    Remaining: %s\n", session.FormatTimeRemaining())
				} else {
					fmt.Printf("    Duration:  unlimited\n")
				}
			}
		}
	} else {
		fmt.Println("✗ Idle prevention is INACTIVE")
	}

	fmt.Printf("\nAvailable Providers:\n")
	for _, p := range providers {
		fmt.Printf("  - %s\n", p)
	}

	// Show environment info
	env := mgr.GetEnvironment()
	fmt.Printf("\nEnvironment:\n")
	fmt.Printf("  Display Server: %s\n", env.DisplayServer)
	fmt.Printf("  Desktop:        %s\n", env.DesktopEnv)

	return nil
}

func handleList(mgr *manager.Manager) error {
	sessions := mgr.ListSessions()

	if len(sessions) == 0 {
		fmt.Println("No active idle prevention sessions")
		return nil
	}

	fmt.Printf("Active Sessions (%d):\n\n", len(sessions))

	// Create table header
	fmt.Printf("%-10s %-12s %-20s %-15s %s\n",
		"ID", "Provider", "Started", "Remaining", "Reason")
	fmt.Println(strings.Repeat("-", 80))

	for _, session := range sessions {
		startTime := session.StartTime.Format("15:04:05")
		remaining := session.FormatTimeRemaining()

		// Truncate reason if too long
		reason := session.Reason
		if len(reason) > 25 {
			reason = reason[:22] + "..."
		}

		fmt.Printf("%-10s %-12s %-20s %-15s %s\n",
			session.ID[:8],
			session.Provider,
			startTime,
			remaining,
			reason)
	}

	return nil
}

func setupSignalHandler(mgr *manager.Manager) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, cleaning up...")
		mgr.Cleanup()
		os.Exit(0)
	}()
}

func runDaemon(mgr *manager.Manager, reason string, duration time.Duration, provider string) error {
	// Fork to background
	if os.Getppid() != 1 {
		// We are the parent process
		args := os.Args[1:]
		cmd := exec.Command(os.Args[0], args...)

		// Detach from parent
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		// Redirect output to /dev/null or log file
		logFile := filepath.Join(paths.StateDir, "heimdall-idle.log")
		if output, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			cmd.Stdout = output
			cmd.Stderr = output
		}

		// Start the daemon process
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Write PID file
		pidFile := filepath.Join(paths.StateDir, "heimdall-idle.pid")
		if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
			logger.Error("Failed to write PID file", "error", err)
		}

		fmt.Printf("✓ Idle prevention daemon started (PID: %d)\n", cmd.Process.Pid)
		fmt.Printf("  Session will run in background\n")
		if duration > 0 {
			fmt.Printf("  Duration: %s\n", manager.FormatDuration(duration))
		} else {
			fmt.Printf("  Duration: unlimited\n")
		}
		fmt.Printf("  To stop: heimdall idle --stop or kill %d\n", cmd.Process.Pid)

		return nil
	}

	// We are the daemon process
	// Close stdin
	os.Stdin.Close()

	// Start the session
	session, err := mgr.Start(reason, duration, provider)
	if err != nil {
		return fmt.Errorf("daemon failed to start idle prevention: %w", err)
	}

	logger.Info("Daemon started idle prevention",
		"session", session.ID[:8],
		"provider", session.Provider,
		"duration", duration)

	// If duration is specified, wait for it
	if duration > 0 {
		time.Sleep(duration)
		mgr.Stop(session.ID)
		logger.Info("Daemon session expired", "session", session.ID[:8])

		// Clean up PID file
		pidFile := filepath.Join(paths.StateDir, "heimdall-idle.pid")
		os.Remove(pidFile)
	} else {
		// Run indefinitely
		select {}
	}

	return nil
}
