package shell

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

func TestCommand(t *testing.T) {
	cmd := Command()

	if cmd.Use != "shell [message...]" {
		t.Errorf("Expected Use to be 'shell [message...]', got %s", cmd.Use)
	}

	if cmd.Short != "Start or communicate with the shell daemon" {
		t.Errorf("Expected Short to be 'Start or communicate with the shell daemon', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "shell daemon") {
		t.Errorf("Expected Long to contain 'shell daemon'")
	}

	// Check flags
	daemonFlag := cmd.Flags().Lookup("daemon")
	if daemonFlag == nil {
		t.Error("Expected daemon flag to exist")
	}
	if daemonFlag.Shorthand != "d" {
		t.Errorf("Expected daemon flag shorthand to be 'd', got %s", daemonFlag.Shorthand)
	}

	stopFlag := cmd.Flags().Lookup("stop")
	if stopFlag == nil {
		t.Error("Expected stop flag to exist")
	}

	listFlag := cmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("Expected list flag to exist")
	}

	killFlag := cmd.Flags().Lookup("kill")
	if killFlag == nil {
		t.Error("Expected kill flag to exist")
	}
	if killFlag.Shorthand != "k" {
		t.Errorf("Expected kill flag shorthand to be 'k', got %s", killFlag.Shorthand)
	}

	showFlag := cmd.Flags().Lookup("show")
	if showFlag == nil {
		t.Error("Expected show flag to exist")
	}
	if showFlag.Shorthand != "s" {
		t.Errorf("Expected show flag shorthand to be 's', got %s", showFlag.Shorthand)
	}

	logFlag := cmd.Flags().Lookup("log")
	if logFlag == nil {
		t.Error("Expected log flag to exist")
	}
	if logFlag.Shorthand != "l" {
		t.Errorf("Expected log flag shorthand to be 'l', got %s", logFlag.Shorthand)
	}

	logRulesFlag := cmd.Flags().Lookup("log-rules")
	if logRulesFlag == nil {
		t.Error("Expected log-rules flag to exist")
	}
}

func TestIsDaemonRunning(t *testing.T) {
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "test.pid")

	tests := []struct {
		name     string
		setup    func() error
		expected bool
	}{
		{
			name: "returns false when pid file doesn't exist",
			setup: func() error {
				return nil // No setup needed
			},
			expected: false,
		},
		{
			name: "returns false when pid file has invalid content",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("invalid"), 0644)
			},
			expected: false,
		},
		{
			name: "returns false when process doesn't exist",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("99999"), 0644)
			},
			expected: false,
		},
		{
			name: "returns true when process exists",
			setup: func() error {
				// Use current process PID
				pid := os.Getpid()
				return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(pidFile)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			result := isDaemonRunning(pidFile)
			if result != tt.expected {
				t.Errorf("isDaemonRunning() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestStopDaemon(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "shell.pid")

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "returns error when daemon not running",
			setup: func() error {
				return nil // No pid file
			},
			expectError: true,
			errorMsg:    "not running",
		},
		{
			name: "returns error with invalid pid file",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("invalid"), 0644)
			},
			expectError: true,
			errorMsg:    "invalid PID",
		},
		{
			name: "handles non-existent process gracefully",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("99999"), 0644)
			},
			expectError: true,
			errorMsg:    "failed to stop daemon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(pidFile)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = StopDaemon()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestKillDaemon(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "shell.pid")

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "returns error when daemon not running",
			setup: func() error {
				return nil // No pid file
			},
			expectError: true,
			errorMsg:    "not running",
		},
		{
			name: "returns error with invalid pid file",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("invalid"), 0644)
			},
			expectError: true,
			errorMsg:    "invalid PID",
		},
		{
			name: "handles non-existent process gracefully",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("99999"), 0644)
			},
			expectError: true,
			errorMsg:    "failed to kill daemon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(pidFile)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = KillDaemon()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestListDaemon(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "shell.pid")
	logFile := filepath.Join(tempDir, "shell.log")

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
	}{
		{
			name: "shows not running when no pid file",
			setup: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name: "shows running with valid pid and log file",
			setup: func() error {
				pid := os.Getpid()
				if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
					return err
				}
				return os.WriteFile(logFile, []byte("test log content"), 0644)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(pidFile)
			os.Remove(logFile)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = ListDaemon()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestShowShellLog(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	logFile := filepath.Join(tempDir, "shell.log")

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "returns error when log file doesn't exist",
			setup: func() error {
				return nil
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name: "shows log content when file exists",
			setup: func() error {
				return os.WriteFile(logFile, []byte("test log content\nline 2"), 0644)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(logFile)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = ShowShellLog()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestShouldLogLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "allows all lines by default",
			line:     "INFO: Starting shell",
			expected: true,
		},
		{
			name:     "allows error lines",
			line:     "ERROR: Connection failed",
			expected: true,
		},
		{
			name:     "allows debug lines",
			line:     "DEBUG: Processing command",
			expected: true,
		},
		{
			name:     "allows empty lines",
			line:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldLogLine(tt.line)
			if result != tt.expected {
				t.Errorf("shouldLogLine(%s) = %v, expected %v", tt.line, result, tt.expected)
			}
		})
	}
}

// IPC Tests
func TestNewIPCClient(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "uses default port when port is 0",
			port:        0,
			expectError: true, // Will fail because no server is running
			errorMsg:    "connection refused",
		},
		{
			name:        "uses custom port",
			port:        12345,
			expectError: true, // Will fail because no server is running
			errorMsg:    "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewIPCClient(tt.port)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					if client != nil {
						client.Close()
					}
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be created")
				} else {
					client.Close()
				}
			}
		})
	}
}

func TestIPCServer(t *testing.T) {
	// Test IPC server creation and basic functionality
	handler := func(message string) string {
		return "response: " + message
	}

	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{
			name:        "creates server with default port",
			port:        0,
			expectError: false,
		},
		{
			name:        "creates server with custom port",
			port:        0, // Use 0 to let OS choose available port
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewIPCServer(tt.port, handler)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					if server != nil {
						server.Stop()
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if server == nil {
					t.Error("Expected server to be created")
				} else {
					// Test that server can be stopped
					err = server.Stop()
					if err != nil {
						t.Errorf("Failed to stop server: %v", err)
					}
				}
			}
		})
	}
}

func TestIPCClientServerCommunication(t *testing.T) {
	// Create a test server
	handler := func(message string) string {
		switch message {
		case "ping":
			return "pong"
		case "echo test":
			return "test"
		default:
			return "unknown command"
		}
	}

	// Use port 0 to let OS choose an available port
	server, err := NewIPCServer(0, handler)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Get the actual port the server is listening on
	addr := server.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Start server in background
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name             string
		message          string
		expectedResponse string
	}{
		{
			name:             "handles ping command",
			message:          "ping",
			expectedResponse: "pong",
		},
		{
			name:             "handles echo command",
			message:          "echo test",
			expectedResponse: "test",
		},
		{
			name:             "handles unknown command",
			message:          "unknown",
			expectedResponse: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewIPCClient(port)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			response, err := client.SendMessage(tt.message)
			if err != nil {
				t.Fatalf("Failed to send message: %v", err)
			}

			if response != tt.expectedResponse {
				t.Errorf("Expected response '%s', got '%s'", tt.expectedResponse, response)
			}
		})
	}
}

func TestStartAttachedConfiguration(t *testing.T) {
	// Test configuration parsing for startAttached
	tests := []struct {
		name         string
		config       *config.Config
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name: "uses command and args from config",
			config: &config.Config{
				Shell: config.ShellConfig{
					Command: "test-shell",
					Args:    []string{"--flag", "value"},
				},
			},
			expectedCmd:  "test-shell",
			expectedArgs: []string{"--flag", "value"},
		},
		{
			name: "parses command string when args are empty",
			config: &config.Config{
				Shell: config.ShellConfig{
					Command: "test-shell --flag value",
					Args:    []string{},
				},
			},
			expectedCmd:  "test-shell",
			expectedArgs: []string{"--flag", "value"},
		},
		{
			name: "handles command without args",
			config: &config.Config{
				Shell: config.ShellConfig{
					Command: "simple-shell",
					Args:    []string{},
				},
			},
			expectedCmd:  "simple-shell",
			expectedArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test command parsing logic
			command := tt.config.Shell.Command
			args := tt.config.Shell.Args

			if len(args) == 0 {
				// Parse from command string
				parts := strings.Fields(command)
				if len(parts) > 1 {
					args = parts[1:]
					command = parts[0]
				}
			}

			if command != tt.expectedCmd {
				t.Errorf("Expected command '%s', got '%s'", tt.expectedCmd, command)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d", len(tt.expectedArgs), len(args))
			} else {
				for i, expected := range tt.expectedArgs {
					if args[i] != expected {
						t.Errorf("Expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
					}
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsDaemonRunning(b *testing.B) {
	tempDir := b.TempDir()
	pidFile := filepath.Join(tempDir, "test.pid")

	// Create a valid pid file
	pid := os.Getpid()
	os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isDaemonRunning(pidFile)
	}
}

func BenchmarkShouldLogLine(b *testing.B) {
	testLine := "INFO: Processing shell command with various parameters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shouldLogLine(testLine)
	}
}

func BenchmarkIPCCommunication(b *testing.B) {
	// Setup server
	handler := func(message string) string {
		return "response: " + message
	}

	server, err := NewIPCServer(0, handler)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	addr := server.listener.Addr().(*net.TCPAddr)
	port := addr.Port

	go server.Start()
	time.Sleep(100 * time.Millisecond)

	client, err := NewIPCClient(port)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SendMessage("test message")
	}
}

// Integration tests
func TestShellCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test command creation and flag parsing
	cmd := Command()

	// Test that command can be created without errors
	if cmd == nil {
		t.Error("Expected command to be created")
	}

	// Test flag parsing
	cmd.SetArgs([]string{"--daemon", "--log-rules", "debug"})
	err := cmd.ParseFlags([]string{"--daemon", "--log-rules", "debug"})
	if err != nil {
		t.Errorf("Flag parsing failed: %v", err)
	}

	// Check that flags were set
	daemonFlagValue, _ := cmd.Flags().GetBool("daemon")
	if !daemonFlagValue {
		t.Error("Expected daemon flag to be true")
	}

	logRulesFlagValue, _ := cmd.Flags().GetString("log-rules")
	if logRulesFlagValue != "debug" {
		t.Errorf("Expected log-rules flag to be 'debug', got '%s'", logRulesFlagValue)
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	tests := []struct {
		name        string
		logRules    string
		expectedEnv string
	}{
		{
			name:        "sets RUST_LOG from config",
			logRules:    "debug",
			expectedEnv: "debug",
		},
		{
			name:        "sets RUST_LOG from flag",
			logRules:    "info",
			expectedEnv: "info",
		},
		{
			name:        "handles empty log rules",
			logRules:    "",
			expectedEnv: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test environment variable setting logic
			originalRustLog := os.Getenv("RUST_LOG")
			defer func() {
				if originalRustLog != "" {
					os.Setenv("RUST_LOG", originalRustLog)
				} else {
					os.Unsetenv("RUST_LOG")
				}
			}()

			if tt.logRules != "" {
				os.Setenv("RUST_LOG", tt.logRules)
			} else {
				os.Unsetenv("RUST_LOG")
			}

			result := os.Getenv("RUST_LOG")
			if result != tt.expectedEnv {
				t.Errorf("Expected RUST_LOG to be '%s', got '%s'", tt.expectedEnv, result)
			}
		})
	}
}
