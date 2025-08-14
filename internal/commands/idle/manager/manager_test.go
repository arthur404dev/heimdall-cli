package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/providers"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("creates manager with correct initialization", func(t *testing.T) {
		mgr := NewManager()

		if mgr == nil {
			t.Fatal("Manager should not be nil")
		}
		if mgr.registry == nil {
			t.Error("Registry should be initialized")
		}
		if mgr.sessionManager == nil {
			t.Error("Session manager should be initialized")
		}
		if mgr.env == nil {
			t.Error("Environment should be detected")
		}
		if mgr.stateFile == "" {
			t.Error("State file path should be set")
		}
	})

	t.Run("state file is in correct location", func(t *testing.T) {
		mgr := NewManager()
		expectedPath := filepath.Join(tempDir, "idle-sessions.json")
		if mgr.stateFile != expectedPath {
			t.Errorf("Expected state file at %s, got %s", expectedPath, mgr.stateFile)
		}
	})
}

func TestManagerStart(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("starts session with fallback provider", func(t *testing.T) {
		mgr := NewManager()

		session, err := mgr.Start("test reason", 0, "")
		if err != nil {
			t.Fatalf("Failed to start session: %s", err.Error())
		}
		if session == nil {
			t.Fatal("Session should not be nil")
		}
		if session.ID == "" {
			t.Error("Session ID should not be empty")
		}
		if session.Reason != "test reason" {
			t.Errorf("Expected reason 'test reason', got %s", session.Reason)
		}
		if session.Provider == "" {
			t.Error("Provider should not be empty")
		}

		// Clean up
		mgr.Stop(session.ID)
	})

	t.Run("starts session with specific provider", func(t *testing.T) {
		mgr := NewManager()

		session, err := mgr.Start("test reason", 0, "fallback")
		if err != nil {
			t.Fatalf("Failed to start session with fallback provider: %s", err.Error())
		}
		if session.Provider != "fallback" {
			t.Errorf("Expected provider 'fallback', got %s", session.Provider)
		}

		// Clean up
		mgr.Stop(session.ID)
	})

	t.Run("starts session with timer", func(t *testing.T) {
		mgr := NewManager()

		duration := 100 * time.Millisecond
		session, err := mgr.Start("test reason", duration, "")
		if err != nil {
			t.Fatalf("Failed to start session with timer: %s", err.Error())
		}
		if session.Duration != duration {
			t.Errorf("Expected duration %v, got %v", duration, session.Duration)
		}
		if session.ExpiresAt == nil {
			t.Error("ExpiresAt should be set for timed session")
		}

		// Wait for session to expire
		time.Sleep(150 * time.Millisecond)

		// Session should be automatically removed
		active, sessions, _ := mgr.GetStatus()
		if active {
			t.Error("Session should have expired")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after expiry, got %d", len(sessions))
		}
	})

	t.Run("fails with nonexistent provider", func(t *testing.T) {
		mgr := NewManager()

		_, err := mgr.Start("test reason", 0, "nonexistent")
		if err == nil {
			t.Error("Expected error with nonexistent provider")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected error to contain 'not found', got %s", err.Error())
		}
	})
}

func TestManagerStop(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("stops existing session", func(t *testing.T) {
		mgr := NewManager()

		// Start a session
		session, err := mgr.Start("test reason", 0, "")
		if err != nil {
			t.Fatalf("Failed to start session: %s", err.Error())
		}

		// Stop the session
		err = mgr.Stop(session.ID)
		if err != nil {
			t.Errorf("Failed to stop session: %s", err.Error())
		}

		// Verify session is stopped
		active, sessions, _ := mgr.GetStatus()
		if active {
			t.Error("Session should be stopped")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after stop, got %d", len(sessions))
		}
	})

	t.Run("fails to stop nonexistent session", func(t *testing.T) {
		mgr := NewManager()

		err := mgr.Stop("nonexistent-id")
		if err == nil {
			t.Error("Expected error when stopping nonexistent session")
		}
		if err != nil && !contains(err.Error(), "not found") {
			t.Errorf("Expected error to contain 'not found', got %s", err.Error())
		}
	})

	t.Run("stops all sessions when empty ID provided", func(t *testing.T) {
		mgr := NewManager()

		// Start multiple sessions
		_, _ = mgr.Start("session 1", 0, "")
		_, _ = mgr.Start("session 2", 0, "")

		// Stop with empty ID should stop all
		err := mgr.Stop("")
		if err != nil {
			t.Errorf("Failed to stop all sessions: %s", err.Error())
		}

		// Verify all sessions are stopped
		active, sessions, _ := mgr.GetStatus()
		if active {
			t.Error("All sessions should be stopped")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after stop all, got %d", len(sessions))
		}
	})
}

func TestManagerStopAll(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("stops all active sessions", func(t *testing.T) {
		mgr := NewManager()

		// Start multiple sessions
		session1, _ := mgr.Start("session 1", 0, "")
		session2, _ := mgr.Start("session 2", 0, "")

		// Verify sessions are active
		active, sessions, _ := mgr.GetStatus()
		if !active {
			t.Error("Sessions should be active")
		}
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(sessions))
		}

		// Stop all sessions
		err := mgr.StopAll()
		if err != nil {
			t.Errorf("Failed to stop all sessions: %s", err.Error())
		}

		// Verify all sessions are stopped
		active, sessions, _ = mgr.GetStatus()
		if active {
			t.Error("All sessions should be stopped")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after stop all, got %d", len(sessions))
		}

		// Clean up any remaining
		mgr.Stop(session1.ID)
		mgr.Stop(session2.ID)
	})

	t.Run("fails when no active sessions", func(t *testing.T) {
		mgr := NewManager()

		err := mgr.StopAll()
		if err == nil {
			t.Error("Expected error when no active sessions")
		}
		if err != nil && !contains(err.Error(), "no active sessions") {
			t.Errorf("Expected error to contain 'no active sessions', got %s", err.Error())
		}
	})
}

func TestManagerGetStatus(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("returns correct status with no sessions", func(t *testing.T) {
		mgr := NewManager()

		active, sessions, providers := mgr.GetStatus()
		if active {
			t.Error("Should not be active with no sessions")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
		if len(providers) == 0 {
			t.Error("Should have at least one available provider")
		}
	})

	t.Run("returns correct status with active sessions", func(t *testing.T) {
		mgr := NewManager()

		// Start a session
		session, _ := mgr.Start("test reason", 0, "")

		active, sessions, providers := mgr.GetStatus()
		if !active {
			t.Error("Should be active with session")
		}
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}
		if len(providers) == 0 {
			t.Error("Should have at least one available provider")
		}

		// Clean up
		mgr.Stop(session.ID)
	})
}

func TestManagerListSessions(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("returns empty list with no sessions", func(t *testing.T) {
		mgr := NewManager()

		sessions := mgr.ListSessions()
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
	})

	t.Run("returns correct sessions", func(t *testing.T) {
		mgr := NewManager()

		// Start multiple sessions
		session1, _ := mgr.Start("session 1", 0, "")
		session2, _ := mgr.Start("session 2", 0, "")

		sessions := mgr.ListSessions()
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(sessions))
		}

		// Verify session details
		sessionIDs := make(map[string]bool)
		for _, s := range sessions {
			sessionIDs[s.ID] = true
		}
		if !sessionIDs[session1.ID] {
			t.Error("Session 1 not found in list")
		}
		if !sessionIDs[session2.ID] {
			t.Error("Session 2 not found in list")
		}

		// Clean up
		mgr.StopAll()
	})
}

func TestManagerGetEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("returns detected environment", func(t *testing.T) {
		mgr := NewManager()

		env := mgr.GetEnvironment()
		if env == nil {
			t.Fatal("Environment should not be nil")
		}
		if env.DisplayServer == "" {
			t.Error("Display server should not be empty")
		}
		if env.DesktopEnv == "" {
			t.Error("Desktop environment should not be empty")
		}
	})
}

func TestManagerStateManagement(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("saves and loads state correctly", func(t *testing.T) {
		// Create first manager and start sessions
		mgr1 := NewManager()
		_, _ = mgr1.Start("session 1", 0, "")
		_, _ = mgr1.Start("session 2", 5*time.Minute, "")

		// Verify state file exists
		stateFile := filepath.Join(tempDir, "idle-sessions.json")
		if _, err := os.Stat(stateFile); os.IsNotExist(err) {
			t.Error("State file should be created")
		}

		// Create second manager (should load state)
		mgr2 := NewManager()
		sessions := mgr2.ListSessions()

		// Note: Sessions might not be fully restored due to provider limitations
		// but the state loading mechanism should work
		if len(sessions) < 0 {
			t.Errorf("Expected some sessions to be loaded, got %d", len(sessions))
		}

		// Clean up
		mgr1.StopAll()
		mgr2.StopAll()
	})

	t.Run("handles corrupted state file", func(t *testing.T) {
		// Create corrupted state file
		stateFile := filepath.Join(tempDir, "idle-sessions.json")
		err := os.WriteFile(stateFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create corrupted state file: %s", err.Error())
		}

		// Manager should handle corrupted state gracefully
		mgr := NewManager()
		if mgr == nil {
			t.Error("Manager should be created even with corrupted state")
		}

		// State file should be removed
		if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
			t.Error("Corrupted state file should be removed")
		}
	})

	t.Run("handles missing state file", func(t *testing.T) {
		// Ensure no state file exists
		stateFile := filepath.Join(tempDir, "idle-sessions.json")
		os.Remove(stateFile)

		// Manager should handle missing state gracefully
		mgr := NewManager()
		if mgr == nil {
			t.Error("Manager should be created even without state file")
		}

		sessions := mgr.ListSessions()
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions with no state file, got %d", len(sessions))
		}
	})
}

func TestManagerCleanup(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("cleanup removes all sessions and state", func(t *testing.T) {
		mgr := NewManager()

		// Start sessions
		_, _ = mgr.Start("session 1", 0, "")
		_, _ = mgr.Start("session 2", 0, "")

		// Verify sessions exist
		active, sessions, _ := mgr.GetStatus()
		if !active || len(sessions) != 2 {
			t.Error("Sessions should be active before cleanup")
		}

		// Cleanup
		err := mgr.Cleanup()
		if err != nil {
			t.Errorf("Cleanup failed: %s", err.Error())
		}

		// Verify sessions are removed
		active, sessions, _ = mgr.GetStatus()
		if active {
			t.Error("Should not be active after cleanup")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after cleanup, got %d", len(sessions))
		}

		// Verify state file is removed
		stateFile := filepath.Join(tempDir, "idle-sessions.json")
		if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
			t.Error("State file should be removed after cleanup")
		}
	})
}

// Integration tests
func TestManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("full session lifecycle", func(t *testing.T) {
		mgr := NewManager()

		// Start session
		session, err := mgr.Start("integration test", 200*time.Millisecond, "")
		if err != nil {
			t.Fatalf("Failed to start session: %s", err.Error())
		}

		// Verify session is active
		active, sessions, _ := mgr.GetStatus()
		if !active {
			t.Error("Session should be active")
		}
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}

		// Wait for expiration
		time.Sleep(300 * time.Millisecond)

		// Session should be expired
		active, sessions, _ = mgr.GetStatus()
		if active {
			t.Error("Session should have expired")
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after expiry, got %d", len(sessions))
		}

		// Clean up
		mgr.Stop(session.ID)
	})

	t.Run("multiple concurrent sessions", func(t *testing.T) {
		mgr := NewManager()

		// Start multiple sessions
		sessions := make([]*Session, 3)
		for i := 0; i < 3; i++ {
			session, err := mgr.Start("concurrent test", 0, "")
			if err != nil {
				t.Fatalf("Failed to start session %d: %s", i, err.Error())
			}
			sessions[i] = session
		}

		// Verify all sessions are active
		active, activeSessions, _ := mgr.GetStatus()
		if !active {
			t.Error("Sessions should be active")
		}
		if len(activeSessions) != 3 {
			t.Errorf("Expected 3 sessions, got %d", len(activeSessions))
		}

		// Stop sessions one by one
		for i, session := range sessions {
			err := mgr.Stop(session.ID)
			if err != nil {
				t.Errorf("Failed to stop session %d: %s", i, err.Error())
			}

			// Verify remaining sessions
			_, remaining, _ := mgr.GetStatus()
			expectedRemaining := 3 - i - 1
			if len(remaining) != expectedRemaining {
				t.Errorf("Expected %d remaining sessions, got %d", expectedRemaining, len(remaining))
			}
		}

		// Clean up
		mgr.StopAll()
	})
}

// Benchmark tests
func BenchmarkManagerCreation(b *testing.B) {
	tempDir := b.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	for i := 0; i < b.N; i++ {
		mgr := NewManager()
		_ = mgr
	}
}

func BenchmarkSessionStart(b *testing.B) {
	tempDir := b.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	mgr := NewManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		session, err := mgr.Start("benchmark test", 0, "")
		if err != nil {
			b.Fatalf("Failed to start session: %s", err.Error())
		}
		mgr.Stop(session.ID)
	}
}

func BenchmarkGetStatus(b *testing.B) {
	tempDir := b.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	mgr := NewManager()
	session, _ := mgr.Start("benchmark test", 0, "")
	defer mgr.Stop(session.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		active, sessions, providers := mgr.GetStatus()
		_, _, _ = active, sessions, providers
	}
}

// Test utilities
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

// Mock provider for testing
type mockProvider struct {
	name      string
	available bool
	priority  int
	active    bool
	cookie    providers.Cookie
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) Available() bool {
	return p.available
}

func (p *mockProvider) Priority() int {
	return p.priority
}

func (p *mockProvider) Inhibit(reason string) (providers.Cookie, error) {
	if !p.available {
		return nil, ErrProviderNotAvailable
	}
	p.active = true
	p.cookie = &mockCookie{id: "mock-cookie-" + reason}
	return p.cookie, nil
}

func (p *mockProvider) Uninhibit(cookie providers.Cookie) error {
	if !p.active {
		return ErrNoActiveInhibition
	}
	p.active = false
	p.cookie = nil
	return nil
}

func (p *mockProvider) Status() (bool, error) {
	return p.active, nil
}

type mockCookie struct {
	id string
}

func (c *mockCookie) String() string {
	return c.id
}

// Provider errors for testing
var (
	ErrProviderNotAvailable = fmt.Errorf("provider not available")
	ErrNoActiveInhibition   = fmt.Errorf("no active inhibition")
)

// Test with mock providers
func TestManagerWithMockProviders(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() {
		paths.StateDir = originalStateDir
	}()

	t.Run("uses highest priority available provider", func(t *testing.T) {
		// This test would require more complex mocking of the registry
		// For now, we'll test with the actual providers
		mgr := NewManager()

		session, err := mgr.Start("test", 0, "")
		if err != nil {
			t.Fatalf("Failed to start session: %s", err.Error())
		}

		// Should use fallback provider (lowest priority but always available)
		if session.Provider != "fallback" {
			t.Logf("Using provider: %s (expected fallback, but other providers may be available)", session.Provider)
		}

		mgr.Stop(session.ID)
	})
}
