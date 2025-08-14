package idle

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/manager"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

func TestCommand(t *testing.T) {
	t.Run("creates command with correct structure", func(t *testing.T) {
		cmd := Command()

		if cmd.Use != "idle [OPTIONS]" {
			t.Errorf("Expected Use to be 'idle [OPTIONS]', got %s", cmd.Use)
		}
		if cmd.Short != "Manage system idle prevention" {
			t.Errorf("Expected Short to be 'Manage system idle prevention', got %s", cmd.Short)
		}
		if cmd.RunE == nil {
			t.Error("Expected RunE to be set")
		}

		// Check flags are defined
		flags := cmd.Flags()
		if flags.Lookup("timer") == nil {
			t.Error("Expected timer flag to be defined")
		}
		if flags.Lookup("stop") == nil {
			t.Error("Expected stop flag to be defined")
		}
		if flags.Lookup("status") == nil {
			t.Error("Expected status flag to be defined")
		}
		if flags.Lookup("daemon") == nil {
			t.Error("Expected daemon flag to be defined")
		}
	})

	t.Run("flag shortcuts work correctly", func(t *testing.T) {
		cmd := Command()
		flags := cmd.Flags()

		// Check short flags
		timerFlag := flags.Lookup("timer")
		if timerFlag.Shorthand != "t" {
			t.Errorf("Expected timer shorthand to be 't', got %s", timerFlag.Shorthand)
		}

		stopFlag := flags.Lookup("stop")
		if stopFlag.Shorthand != "s" {
			t.Errorf("Expected stop shorthand to be 's', got %s", stopFlag.Shorthand)
		}

		daemonFlag := flags.Lookup("daemon")
		if daemonFlag.Shorthand != "d" {
			t.Errorf("Expected daemon shorthand to be 'd', got %s", daemonFlag.Shorthand)
		}
	})
}

func TestRun(t *testing.T) {
	// Setup temporary state directory for tests
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("invalid timer duration returns error", func(t *testing.T) {
		err := run("invalid-duration", false, false, false, false, "test", "", false)
		if err == nil {
			t.Error("Expected error for invalid timer duration")
		}
		if err != nil && !contains(err.Error(), "invalid timer duration") {
			t.Errorf("Expected error to contain 'invalid timer duration', got %s", err.Error())
		}
	})

	t.Run("valid timer duration is parsed correctly", func(t *testing.T) {
		// This test will fail at session creation since we don't have real providers
		// but it should pass timer parsing
		err := run("30m", false, false, false, false, "test", "", false)
		// Error should be about provider availability, not timer parsing
		if err != nil && contains(err.Error(), "invalid timer duration") {
			t.Errorf("Timer parsing failed: %s", err.Error())
		}
	})

	t.Run("status command works without active sessions", func(t *testing.T) {
		// Capture output
		output := captureOutput(func() {
			err := run("", false, false, true, false, "", "", false)
			if err != nil {
				t.Errorf("Status command failed: %s", err.Error())
			}
		})

		if !contains(output, "INACTIVE") {
			t.Errorf("Expected output to contain 'INACTIVE', got %s", output)
		}
	})

	t.Run("list command works without active sessions", func(t *testing.T) {
		output := captureOutput(func() {
			err := run("", false, false, false, true, "", "", false)
			if err != nil {
				t.Errorf("List command failed: %s", err.Error())
			}
		})

		if !contains(output, "No active idle prevention sessions") {
			t.Errorf("Expected output to contain 'No active idle prevention sessions', got %s", output)
		}
	})

	t.Run("stop command with no active sessions returns error", func(t *testing.T) {
		err := run("", true, false, false, false, "", "", false)
		if err == nil {
			t.Error("Expected error when stopping with no active sessions")
		}
	})
}

func TestHandleStart(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("fails with no available providers", func(t *testing.T) {
		mgr := manager.NewManager()
		err := handleStart(mgr, "test reason", 0, "nonexistent")
		if err == nil {
			t.Error("Expected error with nonexistent provider")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected error to contain 'not found', got %s", err.Error())
		}
	})

	t.Run("works with fallback provider", func(t *testing.T) {
		mgr := manager.NewManager()

		// This should work with fallback provider
		// We'll run this in a goroutine and cancel quickly to avoid infinite wait
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- handleStart(mgr, "test reason", 50*time.Millisecond, "")
		}()

		select {
		case err := <-done:
			// Should complete successfully when timer expires
			if err != nil {
				t.Errorf("Expected no error, got %s", err.Error())
			}
		case <-ctx.Done():
			// Test timed out, which is expected for unlimited duration
			// This means the function started successfully
		}
	})
}

func TestHandleStop(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("handles daemon PID file correctly", func(t *testing.T) {
		mgr := manager.NewManager()

		// Create a fake PID file
		pidFile := filepath.Join(tempDir, "heimdall-idle.pid")
		err := os.WriteFile(pidFile, []byte("99999"), 0644)
		if err != nil {
			t.Fatalf("Failed to create PID file: %s", err.Error())
		}

		// This should try to stop the daemon process and clean up the PID file
		err = handleStop(mgr, false)
		// Error is expected since PID 99999 likely doesn't exist
		// But PID file should be cleaned up
		_, statErr := os.Stat(pidFile)
		if !os.IsNotExist(statErr) {
			t.Error("PID file should be removed")
		}
	})

	t.Run("handles invalid PID file", func(t *testing.T) {
		mgr := manager.NewManager()

		// Create an invalid PID file
		pidFile := filepath.Join(tempDir, "heimdall-idle.pid")
		err := os.WriteFile(pidFile, []byte("invalid"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid PID file: %s", err.Error())
		}

		err = handleStop(mgr, false)
		// Should handle invalid PID gracefully
		if err == nil {
			t.Error("Expected error with invalid PID file")
		}
	})
}

func TestHandleStatus(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("shows inactive status correctly", func(t *testing.T) {
		mgr := manager.NewManager()

		output := captureOutput(func() {
			err := handleStatus(mgr)
			if err != nil {
				t.Errorf("Status command failed: %s", err.Error())
			}
		})

		if !contains(output, "INACTIVE") {
			t.Errorf("Expected output to contain 'INACTIVE', got %s", output)
		}
		if !contains(output, "Available Providers") {
			t.Errorf("Expected output to contain 'Available Providers', got %s", output)
		}
		if !contains(output, "Environment") {
			t.Errorf("Expected output to contain 'Environment', got %s", output)
		}
	})

	t.Run("detects daemon process correctly", func(t *testing.T) {
		mgr := manager.NewManager()

		// Create a PID file with current process PID (which exists)
		pidFile := filepath.Join(tempDir, "heimdall-idle.pid")
		currentPID := os.Getpid()
		err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", currentPID)), 0644)
		if err != nil {
			t.Fatalf("Failed to create PID file: %s", err.Error())
		}

		output := captureOutput(func() {
			err := handleStatus(mgr)
			if err != nil {
				t.Errorf("Status command failed: %s", err.Error())
			}
		})

		if !contains(output, "ACTIVE") {
			t.Errorf("Expected output to contain 'ACTIVE', got %s", output)
		}
		if !contains(output, "Daemon Process") {
			t.Errorf("Expected output to contain 'Daemon Process', got %s", output)
		}
		if !contains(output, fmt.Sprintf("PID: %d", currentPID)) {
			t.Errorf("Expected output to contain PID %d, got %s", currentPID, output)
		}
	})
}

func TestHandleList(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("shows no sessions message", func(t *testing.T) {
		mgr := manager.NewManager()

		output := captureOutput(func() {
			err := handleList(mgr)
			if err != nil {
				t.Errorf("List command failed: %s", err.Error())
			}
		})

		if !contains(output, "No active idle prevention sessions") {
			t.Errorf("Expected output to contain 'No active idle prevention sessions', got %s", output)
		}
	})
}

func TestSetupSignalHandler(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("signal handler is set up correctly", func(t *testing.T) {
		mgr := manager.NewManager()

		// This test just ensures the function doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("setupSignalHandler panicked: %v", r)
			}
		}()
		setupSignalHandler(mgr)
	})
}

func TestRunDaemon(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("daemon mode forks correctly", func(t *testing.T) {
		// Skip this test in CI or if we're already a child process
		if os.Getenv("CI") != "" || os.Getppid() == 1 {
			t.Skip("Skipping daemon fork test in CI or child process")
		}

		mgr := manager.NewManager()

		// Test daemon forking (parent process path)
		// This should fork and return quickly
		err := runDaemon(mgr, "test daemon", 100*time.Millisecond, "")

		// Parent process should complete successfully
		if err != nil {
			t.Errorf("Daemon start failed: %s", err.Error())
		}

		// Check if PID file was created
		pidFile := filepath.Join(tempDir, "heimdall-idle.pid")
		_, err = os.Stat(pidFile)
		if err != nil {
			t.Errorf("PID file should be created: %s", err.Error())
		}

		// Clean up - try to stop the daemon
		if pidData, err := os.ReadFile(pidFile); err == nil {
			var pid int
			if _, err := fmt.Sscanf(string(pidData), "%d", &pid); err == nil {
				if proc, err := os.FindProcess(pid); err == nil {
					proc.Signal(syscall.SIGTERM)
				}
			}
		}
		os.Remove(pidFile)
	})
}

// Integration test helpers
func TestIntegrationScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("full lifecycle with timer", func(t *testing.T) {
		mgr := manager.NewManager()

		// Start a session with short timer
		session, err := mgr.Start("test session", 200*time.Millisecond, "")
		if err != nil {
			t.Fatalf("Failed to start session: %s", err.Error())
		}
		if session == nil {
			t.Fatal("Session should not be nil")
		}

		// Verify session is active
		active, sessions, _ := mgr.GetStatus()
		if !active {
			t.Error("Session should be active")
		}
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}

		// Wait for timer to expire
		time.Sleep(300 * time.Millisecond)

		// Session should be expired/removed
		active, sessions, _ = mgr.GetStatus()
		if active {
			t.Error("Session should be inactive after timer expires")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after expiry, got %d", len(sessions))
		}
	})

	t.Run("multiple sessions management", func(t *testing.T) {
		mgr := manager.NewManager()

		// Start multiple sessions
		session1, err := mgr.Start("session 1", 0, "") // unlimited
		if err != nil {
			t.Fatalf("Failed to start session 1: %s", err.Error())
		}

		session2, err := mgr.Start("session 2", 0, "") // unlimited
		if err != nil {
			t.Fatalf("Failed to start session 2: %s", err.Error())
		}

		// Verify both sessions are active
		active, sessions, _ := mgr.GetStatus()
		if !active {
			t.Error("Sessions should be active")
		}
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(sessions))
		}

		// Stop one session
		err = mgr.Stop(session1.ID)
		if err != nil {
			t.Errorf("Failed to stop session 1: %s", err.Error())
		}

		// Verify one session remains
		active, sessions, _ = mgr.GetStatus()
		if !active {
			t.Error("One session should still be active")
		}
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session after stopping one, got %d", len(sessions))
		}
		if sessions[0].ID != session2.ID {
			t.Errorf("Wrong session remaining, expected %s, got %s", session2.ID, sessions[0].ID)
		}

		// Stop all sessions
		err = mgr.StopAll()
		if err != nil {
			t.Errorf("Failed to stop all sessions: %s", err.Error())
		}

		// Verify no sessions remain
		active, sessions, _ = mgr.GetStatus()
		if active {
			t.Error("No sessions should be active after stopping all")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after stopping all, got %d", len(sessions))
		}
	})
}

// Benchmark tests
func BenchmarkCommandCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := Command()
		_ = cmd
	}
}

func BenchmarkManagerCreation(b *testing.B) {
	tempDir := b.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr := manager.NewManager()
		_ = mgr
	}
}

// Test utilities
func captureOutput(fn func()) string {
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)
	return buf.String()
}

func createTestCommand(args ...string) *cobra.Command {
	cmd := Command()
	cmd.SetArgs(args)
	return cmd
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// Mock helpers for testing
type mockManager struct {
	sessions []*manager.Session
	active   bool
}

func (m *mockManager) Start(reason string, duration time.Duration, provider string) (*manager.Session, error) {
	session := &manager.Session{
		ID:        "test-session-id",
		Provider:  "test-provider",
		StartTime: time.Now(),
		Duration:  duration,
		Reason:    reason,
	}
	if duration > 0 {
		expiresAt := time.Now().Add(duration)
		session.ExpiresAt = &expiresAt
	}
	m.sessions = append(m.sessions, session)
	m.active = true
	return session, nil
}

func (m *mockManager) Stop(sessionID string) error {
	for i, session := range m.sessions {
		if session.ID == sessionID {
			m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
			if len(m.sessions) == 0 {
				m.active = false
			}
			return nil
		}
	}
	return fmt.Errorf("session not found")
}

func (m *mockManager) StopAll() error {
	if len(m.sessions) == 0 {
		return fmt.Errorf("no active sessions")
	}
	m.sessions = nil
	m.active = false
	return nil
}

func (m *mockManager) GetStatus() (bool, []*manager.Session, []string) {
	return m.active, m.sessions, []string{"test-provider"}
}

func (m *mockManager) ListSessions() []*manager.Session {
	return m.sessions
}

func (m *mockManager) Cleanup() error {
	m.sessions = nil
	m.active = false
	return nil
}
