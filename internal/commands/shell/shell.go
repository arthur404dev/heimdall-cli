package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/config/manager"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// Command creates the shell command
func Command() *cobra.Command {
	var (
		daemon   bool
		stop     bool
		list     bool
		kill     bool
		show     bool
		log      bool
		logRules string
	)

	cmd := &cobra.Command{
		Use:   "shell [message...]",
		Short: "Start or communicate with the shell daemon",
		Long: `Start the shell daemon or send messages to it.
		
The shell daemon runs attached by default (showing logs in terminal).
Use -d flag to run in detached/daemon mode.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration (old system for backward compatibility)
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Ensure config is saved with all defaults merged
			// This creates the file if it doesn't exist and adds any missing defaults
			// while preserving existing user settings
			if err := config.EnsureConfigSaved(); err != nil {
				// Log warning but don't fail - the command can still run
				logger.Warn("Failed to save config with defaults", "error", err)
			}

			// Initialize the new configuration manager to ensure domain configs exist
			configMgr := manager.NewManager()

			// Try to load paths from main config
			configPath := os.Getenv("HEIMDALL_CONFIG")
			if configPath == "" {
				configPath = filepath.Join(paths.HeimdallConfigDir, "config.json")
			}

			if _, err := os.Stat(configPath); err == nil {
				if err := configMgr.LoadPathsFromConfig(configPath); err != nil {
					logger.Debug("Failed to load paths from config", "error", err)
				}
			}

			// Initialize the manager (this creates default providers for cli and shell domains)
			if err := configMgr.Initialize(); err != nil {
				logger.Warn("Failed to initialize config manager", "error", err)
			} else {
				// Load all configurations (this will create files with defaults if they don't exist)
				if err := configMgr.LoadAll(); err != nil {
					logger.Debug("Failed to load all configs", "error", err)
				}

				// Save all configurations to ensure defaults are persisted
				if err := configMgr.SaveAll(); err != nil {
					logger.Warn("Failed to save domain configs", "error", err)
				}
			}

			cfg := config.Get()

			// Handle control flags
			if stop {
				return StopDaemon()
			}

			if kill {
				return KillDaemon()
			}

			if list {
				return ListDaemon()
			}

			if show {
				return ShowIPCCommands(cfg)
			}

			if log {
				return ShowShellLog()
			}

			// Check if daemon is running
			pidFile := filepath.Join(paths.StateDir, "shell.pid")

			// If args provided, send as message
			if len(args) > 0 {
				message := strings.Join(args, " ")
				return sendMessage(cfg, message)
			}

			// Check if daemon is already running
			if isDaemonRunning(pidFile) {
				return fmt.Errorf("shell daemon is already running")
			}

			// Set log rules
			if logRules != "" {
				os.Setenv("RUST_LOG", logRules)
			} else if cfg.Shell.LogRules != "" {
				os.Setenv("RUST_LOG", cfg.Shell.LogRules)
			}

			// Start shell
			if daemon {
				// Start in daemon mode (detached)
				return startDaemon(cfg, pidFile)
			} else {
				// Start in attached mode (default)
				return startAttached(cfg, pidFile)
			}
		},
	}

	cmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "Run in daemon mode (detached)")
	cmd.Flags().BoolVar(&stop, "stop", false, "Stop the running daemon")
	cmd.Flags().BoolVar(&list, "list", false, "List running daemon status")
	cmd.Flags().BoolVarP(&kill, "kill", "k", false, "Force kill the daemon")
	cmd.Flags().BoolVarP(&show, "show", "s", false, "Print all shell IPC commands")
	cmd.Flags().BoolVarP(&log, "log", "l", false, "Print the shell log")
	cmd.Flags().StringVar(&logRules, "log-rules", "", "Set RUST_LOG environment variable")

	return cmd
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

// startAttached starts the shell in attached mode (default)
func startAttached(cfg *config.Config, pidFile string) error {
	logger.Info("Starting shell in attached mode", "command", cfg.Shell.Command)

	// Build command with args
	args := cfg.Shell.Args
	if len(args) == 0 {
		// If no args configured, try to parse from command string for backward compatibility
		parts := strings.Fields(cfg.Shell.Command)
		if len(parts) > 1 {
			args = parts[1:]
			cfg.Shell.Command = parts[0]
		}
	}

	// Create command
	cmd := exec.Command(cfg.Shell.Command, args...)

	// Set up environment
	cmd.Env = os.Environ()

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Connect stdin
	cmd.Stdin = os.Stdin

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Write PID file
	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		// Try to kill the process if we can't write the PID file
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Stream logs
	go streamLogs(stdout, "stdout")
	go streamLogs(stderr, "stderr")

	// Wait for process to exit or signal
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-sigChan:
		// Received interrupt signal
		logger.Info("Received interrupt signal, stopping shell...")
		cmd.Process.Signal(syscall.SIGTERM)
		time.Sleep(100 * time.Millisecond)
		if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
			cmd.Process.Kill()
		}
		os.Remove(pidFile)
		return nil
	case err := <-done:
		// Process exited
		os.Remove(pidFile)
		if err != nil {
			return fmt.Errorf("shell exited with error: %w", err)
		}
		return nil
	}
}

// streamLogs streams logs from a reader with optional filtering
func streamLogs(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Apply any log filtering here if needed
		if shouldLogLine(line) {
			if prefix == "stderr" {
				fmt.Fprintf(os.Stderr, "%s\n", line)
			} else {
				fmt.Println(line)
			}
		}
	}
}

// shouldLogLine determines if a log line should be displayed
func shouldLogLine(_ string) bool {
	// Add any filtering logic here
	// For now, show all lines
	return true
}

// startDaemon starts the shell daemon in detached mode
func startDaemon(cfg *config.Config, pidFile string) error {
	logger.Info("Starting shell daemon", "command", cfg.Shell.Command)

	// Build command with args
	args := cfg.Shell.Args
	if len(args) == 0 {
		// If no args configured, try to parse from command string for backward compatibility
		parts := strings.Fields(cfg.Shell.Command)
		if len(parts) > 1 {
			args = parts[1:]
			cfg.Shell.Command = parts[0]
		}
	}

	// Create command
	cmd := exec.Command(cfg.Shell.Command, args...)

	// Set up environment
	cmd.Env = os.Environ()

	// Set up logging
	logFile := filepath.Join(paths.StateDir, "shell.log")
	logOut, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logOut.Close()

	cmd.Stdout = logOut
	cmd.Stderr = logOut

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start shell daemon: %w", err)
	}

	// Write PID file
	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		// Try to kill the process if we can't write the PID file
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Send notification
	notifier := notify.NewNotifier()
	notifier.Send(&notify.Notification{
		Summary: "Shell Daemon",
		Body:    "Shell daemon started successfully",
		Urgency: notify.UrgencyNormal,
	})

	logger.Info("Shell daemon started", "pid", pid, "log", logFile)

	// Detach from the process
	cmd.Process.Release()

	return nil
}

// sendMessage sends a message to the running daemon
func sendMessage(cfg *config.Config, message string) error {
	// Create IPC client
	client, err := NewIPCClient(cfg.Shell.DaemonPort)
	if err != nil {
		return fmt.Errorf("failed to create IPC client: %w", err)
	}
	defer client.Close()

	// Send message
	response, err := client.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Print response
	fmt.Println(response)

	return nil
}

// StopDaemon stops the running shell daemon gracefully
func StopDaemon() error {
	pidFile := filepath.Join(paths.StateDir, "shell.pid")

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("shell daemon is not running")
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
		Summary: "Shell Daemon",
		Body:    "Shell daemon stopped",
		Urgency: notify.UrgencyNormal,
	})

	logger.Info("Shell daemon stopped")

	return nil
}

// KillDaemon force kills the running shell daemon
func KillDaemon() error {
	pidFile := filepath.Join(paths.StateDir, "shell.pid")

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("shell daemon is not running")
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

	// Force kill
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill daemon: %w", err)
	}

	// Remove PID file
	os.Remove(pidFile)

	logger.Info("Shell daemon killed")

	return nil
}

// ListDaemon lists the status of the shell daemon
func ListDaemon() error {
	pidFile := filepath.Join(paths.StateDir, "shell.pid")

	if !isDaemonRunning(pidFile) {
		fmt.Println("Shell daemon is not running")
		return nil
	}

	// Read PID
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid := strings.TrimSpace(string(data))
	fmt.Printf("Shell daemon is running (PID: %s)\n", pid)

	// Show log file location
	logFile := filepath.Join(paths.StateDir, "shell.log")
	if paths.Exists(logFile) {
		info, err := os.Stat(logFile)
		if err == nil {
			fmt.Printf("Log file: %s (size: %d bytes)\n", logFile, info.Size())
		}
	}

	return nil
}

// ShowIPCCommands shows all available IPC commands from the shell
func ShowIPCCommands(cfg *config.Config) error {
	// Create IPC client
	client, err := NewIPCClient(cfg.Shell.DaemonPort)
	if err != nil {
		return fmt.Errorf("failed to create IPC client: %w", err)
	}
	defer client.Close()

	// Send "ipc show" command to get all available IPC commands
	response, err := client.SendMessage("ipc show")
	if err != nil {
		return fmt.Errorf("failed to get IPC commands: %w", err)
	}

	// Print response
	fmt.Print(response)

	return nil
}

// ShowShellLog shows the shell log output
func ShowShellLog() error {
	logFile := filepath.Join(paths.StateDir, "shell.log")

	// Check if log file exists
	if !paths.Exists(logFile) {
		return fmt.Errorf("shell log file not found: %s", logFile)
	}

	// Read and print log file contents
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	fmt.Print(string(content))

	return nil
}
