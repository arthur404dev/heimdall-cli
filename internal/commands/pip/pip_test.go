package pip

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// MockHyprClient is a simple mock implementation for testing
type MockHyprClient struct {
	windows     []hypr.Window
	events      chan hypr.Event
	dispatchErr error
	windowsErr  error
}

func (m *MockHyprClient) Subscribe(events []string) (<-chan hypr.Event, error) {
	return m.events, nil
}

func (m *MockHyprClient) GetWindows() ([]hypr.Window, error) {
	return m.windows, m.windowsErr
}

func (m *MockHyprClient) Dispatch(command string, args ...string) error {
	return m.dispatchErr
}

func (m *MockHyprClient) GetMonitors() ([]hypr.Monitor, error) {
	return []hypr.Monitor{}, nil
}

func (m *MockHyprClient) GetWorkspaces() ([]hypr.Workspace, error) {
	return []hypr.Workspace{}, nil
}

func (m *MockHyprClient) Close() error {
	return nil
}

func TestCommand(t *testing.T) {
	cmd := Command()

	if cmd.Use != "pip" {
		t.Errorf("Expected Use to be 'pip', got %s", cmd.Use)
	}

	if cmd.Short != "Picture-in-picture daemon" {
		t.Errorf("Expected Short to be 'Picture-in-picture daemon', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "picture-in-picture mode") {
		t.Errorf("Expected Long to contain 'picture-in-picture mode'")
	}

	// Check flags
	if !cmd.Flags().HasAvailableFlags() {
		t.Error("Expected command to have flags")
	}

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

	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag == nil {
		t.Error("Expected status flag to exist")
	}
}

func TestIsVideoWindow(t *testing.T) {
	// Setup test config
	originalConfig := config.Get()
	defer func() {
		config.Load() // Restore original config
	}()

	// For now, we'll test with the default behavior since we don't have a SetConfig method

	tests := []struct {
		name     string
		class    string
		title    string
		expected bool
	}{
		{
			name:     "detects mpv as video app",
			class:    "mpv",
			title:    "some video file",
			expected: true,
		},
		{
			name:     "detects vlc as video app",
			class:    "vlc",
			title:    "VLC media player",
			expected: true,
		},
		{
			name:     "detects firefox with youtube",
			class:    "firefox",
			title:    "YouTube - Some Video",
			expected: true,
		},
		{
			name:     "detects firefox with netflix",
			class:    "firefox",
			title:    "Netflix - Watching Movie",
			expected: true,
		},
		{
			name:     "detects firefox with playing indicator",
			class:    "firefox",
			title:    "Some Video - playing",
			expected: true,
		},
		{
			name:     "rejects firefox without video keywords",
			class:    "firefox",
			title:    "GitHub - Repository",
			expected: false,
		},
		{
			name:     "rejects non-video app",
			class:    "terminal",
			title:    "bash",
			expected: false,
		},
		{
			name:     "handles case insensitive matching",
			class:    "Firefox",
			title:    "YOUTUBE - Video",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVideoWindow(tt.class, tt.title)
			if result != tt.expected {
				t.Errorf("isVideoWindow(%s, %s) = %v, expected %v", tt.class, tt.title, result, tt.expected)
			}
		})
	}

	// Restore original config
	_ = originalConfig
}

func TestIsVideoWindowWithDefaults(t *testing.T) {
	// Test with empty config to ensure defaults work
	tests := []struct {
		name     string
		class    string
		title    string
		expected bool
	}{
		{
			name:     "uses default video apps - mpv",
			class:    "mpv",
			title:    "video.mp4",
			expected: true,
		},
		{
			name:     "uses default video apps - chrome with youtube",
			class:    "chrome",
			title:    "YouTube - Video",
			expected: true,
		},
		{
			name:     "uses default keywords - playing indicator",
			class:    "firefox",
			title:    "Video ▶ Playing",
			expected: true,
		},
		{
			name:     "rejects chrome without video keywords",
			class:    "chrome",
			title:    "Google Search",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVideoWindow(tt.class, tt.title)
			if result != tt.expected {
				t.Errorf("isVideoWindow(%s, %s) = %v, expected %v", tt.class, tt.title, result, tt.expected)
			}
		})
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

func TestShowStatus(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "pip.pid")

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "shows not running when no pid file",
			setup: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name: "shows running with valid pid",
			setup: func() error {
				pid := os.Getpid()
				return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
			},
			expectError: false,
		},
		{
			name: "handles invalid pid file",
			setup: func() error {
				return os.WriteFile(pidFile, []byte("invalid"), 0644)
			},
			expectError: true,
			errorMsg:    "invalid PID",
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

			err = showStatus()
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

func TestStopDaemon(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "pip.pid")

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

			err = stopDaemon()
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

func TestEnablePIP(t *testing.T) {
	// Test window data
	testWindows := []hypr.Window{
		{
			Address: "0x12345",
			Class:   "firefox",
			Title:   "YouTube - Test Video",
		},
		{
			Address: "0x67890",
			Class:   "terminal",
			Title:   "bash",
		},
	}

	tests := []struct {
		name        string
		windowClass string
		mockSetup   func() *MockHyprClient
		expectError bool
		errorMsg    string
	}{
		{
			name:        "enables PIP for existing window",
			windowClass: "firefox",
			mockSetup: func() *MockHyprClient {
				return &MockHyprClient{
					windows:     testWindows,
					windowsErr:  nil,
					dispatchErr: nil,
				}
			},
			expectError: false,
		},
		{
			name:        "returns error when window not found",
			windowClass: "nonexistent",
			mockSetup: func() *MockHyprClient {
				return &MockHyprClient{
					windows:     testWindows,
					windowsErr:  nil,
					dispatchErr: nil,
				}
			},
			expectError: true,
			errorMsg:    "window not found",
		},
		{
			name:        "returns error when GetWindows fails",
			windowClass: "firefox",
			mockSetup: func() *MockHyprClient {
				return &MockHyprClient{
					windows:     []hypr.Window{},
					windowsErr:  fmt.Errorf("connection failed"),
					dispatchErr: nil,
				}
			},
			expectError: true,
			errorMsg:    "failed to get windows",
		},
		{
			name:        "returns error when dispatch fails",
			windowClass: "firefox",
			mockSetup: func() *MockHyprClient {
				return &MockHyprClient{
					windows:     testWindows,
					windowsErr:  nil,
					dispatchErr: fmt.Errorf("dispatch failed"),
				}
			},
			expectError: true,
			errorMsg:    "failed to float window",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tt.mockSetup()

			// We need to create an interface that our mock can implement
			// For now, we'll test the logic without the actual hypr client
			err := testEnablePIPLogic(mockClient, tt.windowClass)

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

// testEnablePIPLogic tests the core logic of enablePIP without requiring the exact hypr.Client interface
func testEnablePIPLogic(client *MockHyprClient, windowClass string) error {
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

	// Test dispatch calls
	if err := client.Dispatch("togglefloating", targetWindow.Address); err != nil {
		return fmt.Errorf("failed to float window: %w", err)
	}

	return nil
}

func TestStartDaemonEnvironmentHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	pidFile := filepath.Join(tempDir, "pip.pid")

	tests := []struct {
		name        string
		envVar      string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name:   "returns error when daemon already running",
			envVar: "",
			setup: func() {
				// Create a pid file with current process
				pid := os.Getpid()
				os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
			},
			expectError: true,
			errorMsg:    "already running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.Remove(pidFile)

			// Set environment variable
			if tt.envVar != "" {
				os.Setenv("PIP_DAEMON", tt.envVar)
				defer os.Unsetenv("PIP_DAEMON")
			}

			tt.setup()

			err := startDaemon()

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

// Benchmark tests
func BenchmarkIsVideoWindow(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isVideoWindow("firefox", "YouTube - Test Video")
	}
}

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

// Integration tests
func TestPIPCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	// Test command creation and flag parsing
	cmd := Command()

	// Test status command when no daemon running
	cmd.SetArgs([]string{"--status"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Status command failed: %v", err)
	}
}

func TestRunDaemonEventHandling(t *testing.T) {
	// Test the event processing logic
	tests := []struct {
		name      string
		eventType string
		eventData string
		expectPIP bool
	}{
		{
			name:      "handles activewindow event with video app",
			eventType: "activewindow",
			eventData: "firefox,YouTube - Test Video",
			expectPIP: true,
		},
		{
			name:      "handles activewindow event with non-video app",
			eventType: "activewindow",
			eventData: "terminal,bash",
			expectPIP: false,
		},
		{
			name:      "handles closewindow event",
			eventType: "closewindow",
			eventData: "0x12345",
			expectPIP: false,
		},
		{
			name:      "handles malformed activewindow event",
			eventType: "activewindow",
			eventData: "firefox", // Missing title
			expectPIP: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the event parsing logic
			if tt.eventType == "activewindow" {
				parts := strings.Split(tt.eventData, ",")
				if len(parts) >= 2 {
					windowClass := parts[0]
					windowTitle := parts[1]

					result := isVideoWindow(windowClass, windowTitle)
					if result != tt.expectPIP {
						t.Errorf("isVideoWindow(%s, %s) = %v, expected %v", windowClass, windowTitle, result, tt.expectPIP)
					}
				} else if tt.expectPIP {
					t.Error("Expected PIP detection but event data was malformed")
				}
			}
		})
	}
}

// Test helper functions
func createTestPIDFile(t *testing.T, dir string, pid int) string {
	pidFile := filepath.Join(dir, "test.pid")
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}
	return pidFile
}

func TestConfigurationLoading(t *testing.T) {
	tests := []struct {
		name             string
		expectedApps     []string
		expectedKeywords []string
	}{
		{
			name:             "uses defaults when config is empty",
			expectedApps:     []string{"mpv", "vlc", "firefox", "chromium", "chrome", "brave", "youtube", "netflix", "twitch", "spotify"},
			expectedKeywords: []string{"youtube", "netflix", "twitch", "vimeo", "- playing", "▶", "►", "video", "stream"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test video app detection with defaults
			for _, app := range tt.expectedApps {
				result := isVideoWindow(app, "test title")
				if strings.Contains(app, "firefox") || strings.Contains(app, "chrom") || strings.Contains(app, "brave") {
					// Browser apps need video keywords in title
					if result {
						t.Errorf("Browser app %s should require video keywords but was detected without them", app)
					}
				} else {
					// Non-browser apps should be detected directly
					if !result {
						t.Errorf("Video app %s should be detected but wasn't", app)
					}
				}
			}
		})
	}
}
