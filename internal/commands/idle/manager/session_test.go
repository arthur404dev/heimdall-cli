package manager

import (
	"testing"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/providers"
)

func TestNewSession(t *testing.T) {
	t.Run("creates session with correct properties", func(t *testing.T) {
		provider := "test-provider"
		reason := "test reason"
		duration := 30 * time.Minute

		session := NewSession(provider, reason, duration)

		if session == nil {
			t.Fatal("Session should not be nil")
		}
		if session.ID == "" {
			t.Error("Session ID should not be empty")
		}
		if session.Provider != provider {
			t.Errorf("Expected provider %s, got %s", provider, session.Provider)
		}
		if session.Reason != reason {
			t.Errorf("Expected reason %s, got %s", reason, session.Reason)
		}
		if session.Duration != duration {
			t.Errorf("Expected duration %v, got %v", duration, session.Duration)
		}
		if session.StartTime.IsZero() {
			t.Error("Start time should be set")
		}
	})

	t.Run("creates session with unlimited duration", func(t *testing.T) {
		session := NewSession("provider", "reason", 0)

		if session.Duration != 0 {
			t.Errorf("Expected duration 0, got %v", session.Duration)
		}
		if session.ExpiresAt != nil {
			t.Error("ExpiresAt should be nil for unlimited duration")
		}
	})

	t.Run("creates session with expiration time", func(t *testing.T) {
		duration := 1 * time.Hour
		session := NewSession("provider", "reason", duration)

		if session.ExpiresAt == nil {
			t.Error("ExpiresAt should be set for timed session")
		}
		if session.ExpiresAt.Before(time.Now()) {
			t.Error("ExpiresAt should be in the future")
		}

		expectedExpiry := session.StartTime.Add(duration)
		if session.ExpiresAt.Sub(expectedExpiry) > time.Second {
			t.Errorf("ExpiresAt should be approximately %v, got %v", expectedExpiry, *session.ExpiresAt)
		}
	})

	t.Run("generates unique session IDs", func(t *testing.T) {
		session1 := NewSession("provider", "reason", 0)
		session2 := NewSession("provider", "reason", 0)

		if session1.ID == session2.ID {
			t.Error("Session IDs should be unique")
		}
	})
}

func TestSessionIsExpired(t *testing.T) {
	t.Run("unlimited session never expires", func(t *testing.T) {
		session := NewSession("provider", "reason", 0)

		if session.IsExpired() {
			t.Error("Unlimited session should never expire")
		}
	})

	t.Run("future expiration is not expired", func(t *testing.T) {
		session := NewSession("provider", "reason", 1*time.Hour)

		if session.IsExpired() {
			t.Error("Future session should not be expired")
		}
	})

	t.Run("past expiration is expired", func(t *testing.T) {
		session := NewSession("provider", "reason", 1*time.Millisecond)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		if !session.IsExpired() {
			t.Error("Past session should be expired")
		}
	})

	t.Run("manual expiration time", func(t *testing.T) {
		session := NewSession("provider", "reason", 0)
		pastTime := time.Now().Add(-1 * time.Hour)
		session.ExpiresAt = &pastTime

		if !session.IsExpired() {
			t.Error("Session with past expiration should be expired")
		}
	})
}

func TestSessionTimeRemaining(t *testing.T) {
	t.Run("unlimited session has zero remaining", func(t *testing.T) {
		session := NewSession("provider", "reason", 0)

		remaining := session.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 remaining for unlimited session, got %v", remaining)
		}
	})

	t.Run("future session has positive remaining", func(t *testing.T) {
		duration := 1 * time.Hour
		session := NewSession("provider", "reason", duration)

		remaining := session.TimeRemaining()
		if remaining <= 0 {
			t.Errorf("Expected positive remaining time, got %v", remaining)
		}
		if remaining > duration {
			t.Errorf("Remaining time should not exceed duration, got %v", remaining)
		}
	})

	t.Run("expired session has zero remaining", func(t *testing.T) {
		session := NewSession("provider", "reason", 1*time.Millisecond)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		remaining := session.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 remaining for expired session, got %v", remaining)
		}
	})
}

func TestSessionFormatTimeRemaining(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "unlimited session",
			duration: 0,
			expected: "unlimited",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute + 45*time.Second,
			expected: "2h 30m 45s",
		},
		{
			name:     "minutes and seconds",
			duration: 15*time.Minute + 30*time.Second,
			expected: "15m 30s",
		},
		{
			name:     "only seconds",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "exactly one hour",
			duration: 1 * time.Hour,
			expected: "1h 0m 0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("provider", "reason", tt.duration)

			result := session.FormatTimeRemaining()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}

	t.Run("expired session shows expired", func(t *testing.T) {
		session := NewSession("provider", "reason", 1*time.Millisecond)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		result := session.FormatTimeRemaining()
		if result != "expired" {
			t.Errorf("Expected 'expired', got %s", result)
		}
	})
}

func TestNewSessionManager(t *testing.T) {
	t.Run("creates session manager with registry", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		if mgr == nil {
			t.Fatal("Session manager should not be nil")
		}
		if mgr.providers != registry {
			t.Error("Session manager should use provided registry")
		}
		if mgr.sessions == nil {
			t.Error("Sessions map should be initialized")
		}
	})
}

func TestSessionManagerCreateSession(t *testing.T) {
	t.Run("creates session with fallback provider", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, err := mgr.CreateSession("", "test reason", 0)
		if err != nil {
			t.Fatalf("Failed to create session: %s", err.Error())
		}
		if session == nil {
			t.Fatal("Session should not be nil")
		}
		if session.Provider != "fallback" {
			t.Errorf("Expected fallback provider, got %s", session.Provider)
		}
		if session.Reason != "test reason" {
			t.Errorf("Expected reason 'test reason', got %s", session.Reason)
		}

		// Clean up
		mgr.RemoveSession(session.ID)
	})

	t.Run("creates session with specific provider", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, err := mgr.CreateSession("fallback", "test reason", 0)
		if err != nil {
			t.Fatalf("Failed to create session: %s", err.Error())
		}
		if session.Provider != "fallback" {
			t.Errorf("Expected fallback provider, got %s", session.Provider)
		}

		// Clean up
		mgr.RemoveSession(session.ID)
	})

	t.Run("fails with nonexistent provider", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		_, err := mgr.CreateSession("nonexistent", "test reason", 0)
		if err == nil {
			t.Error("Expected error with nonexistent provider")
		}
		if err != nil && !containsSubstring(err.Error(), "not found") {
			t.Errorf("Expected error to contain 'not found', got %s", err.Error())
		}
	})

	t.Run("fails with unavailable provider", func(t *testing.T) {
		registry := providers.NewRegistry()
		mockProvider := &mockProvider{
			name:      "mock",
			available: false,
			priority:  50,
		}
		registry.Register(mockProvider)

		mgr := NewSessionManager(registry)

		_, err := mgr.CreateSession("mock", "test reason", 0)
		if err == nil {
			t.Error("Expected error with unavailable provider")
		}
		if err != nil && !containsSubstring(err.Error(), "not available") {
			t.Errorf("Expected error to contain 'not available', got %s", err.Error())
		}
	})

	t.Run("creates session with timer", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		duration := 100 * time.Millisecond
		session, err := mgr.CreateSession("", "test reason", duration)
		if err != nil {
			t.Fatalf("Failed to create session: %s", err.Error())
		}
		if session.Duration != duration {
			t.Errorf("Expected duration %v, got %v", duration, session.Duration)
		}
		if session.Timer == nil {
			t.Error("Timer should be set for timed session")
		}

		// Wait for timer to expire and auto-remove session
		time.Sleep(150 * time.Millisecond)

		// Session should be automatically removed
		_, exists := mgr.GetSession(session.ID)
		if exists {
			t.Error("Session should be automatically removed after timer expires")
		}
	})
}

func TestSessionManagerGetSession(t *testing.T) {
	t.Run("returns existing session", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, _ := mgr.CreateSession("", "test reason", 0)

		retrieved, exists := mgr.GetSession(session.ID)
		if !exists {
			t.Error("Session should exist")
		}
		if retrieved.ID != session.ID {
			t.Errorf("Expected session ID %s, got %s", session.ID, retrieved.ID)
		}

		// Clean up
		mgr.RemoveSession(session.ID)
	})

	t.Run("returns false for nonexistent session", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		_, exists := mgr.GetSession("nonexistent")
		if exists {
			t.Error("Nonexistent session should not exist")
		}
	})
}

func TestSessionManagerListSessions(t *testing.T) {
	t.Run("returns empty list with no sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		sessions := mgr.ListSessions()
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
	})

	t.Run("returns all active sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session1, _ := mgr.CreateSession("", "session 1", 0)
		session2, _ := mgr.CreateSession("", "session 2", 0)

		sessions := mgr.ListSessions()
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions, got %d", len(sessions))
		}

		// Verify session IDs
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
		mgr.RemoveAllSessions()
	})
}

func TestSessionManagerRemoveSession(t *testing.T) {
	t.Run("removes existing session", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, _ := mgr.CreateSession("", "test reason", 0)

		err := mgr.RemoveSession(session.ID)
		if err != nil {
			t.Errorf("Failed to remove session: %s", err.Error())
		}

		// Verify session is removed
		_, exists := mgr.GetSession(session.ID)
		if exists {
			t.Error("Session should be removed")
		}
	})

	t.Run("fails to remove nonexistent session", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		err := mgr.RemoveSession("nonexistent")
		if err == nil {
			t.Error("Expected error when removing nonexistent session")
		}
		if err != nil && !containsSubstring(err.Error(), "not found") {
			t.Errorf("Expected error to contain 'not found', got %s", err.Error())
		}
	})

	t.Run("cancels timer when removing timed session", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, _ := mgr.CreateSession("", "test reason", 1*time.Hour)

		// Verify timer is set
		if session.Timer == nil {
			t.Error("Timer should be set")
		}

		err := mgr.RemoveSession(session.ID)
		if err != nil {
			t.Errorf("Failed to remove session: %s", err.Error())
		}

		// Timer should be stopped (we can't easily verify this without exposing internals)
	})
}

func TestSessionManagerRemoveAllSessions(t *testing.T) {
	t.Run("removes all sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session1, _ := mgr.CreateSession("", "session 1", 0)
		session2, _ := mgr.CreateSession("", "session 2", 0)

		// Verify sessions exist
		sessions := mgr.ListSessions()
		if len(sessions) != 2 {
			t.Errorf("Expected 2 sessions before removal, got %d", len(sessions))
		}

		err := mgr.RemoveAllSessions()
		if err != nil {
			t.Errorf("Failed to remove all sessions: %s", err.Error())
		}

		// Verify all sessions are removed
		sessions = mgr.ListSessions()
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions after removal, got %d", len(sessions))
		}

		// Verify individual sessions don't exist
		_, exists1 := mgr.GetSession(session1.ID)
		_, exists2 := mgr.GetSession(session2.ID)
		if exists1 || exists2 {
			t.Error("Individual sessions should not exist after remove all")
		}
	})

	t.Run("handles empty session list", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		err := mgr.RemoveAllSessions()
		// Should not error on empty list
		if err != nil {
			t.Errorf("Should not error on empty session list: %s", err.Error())
		}
	})
}

func TestSessionManagerHasActiveSessions(t *testing.T) {
	t.Run("returns false with no sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		if mgr.HasActiveSessions() {
			t.Error("Should not have active sessions")
		}
	})

	t.Run("returns true with active sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session, _ := mgr.CreateSession("", "test reason", 0)

		if !mgr.HasActiveSessions() {
			t.Error("Should have active sessions")
		}

		// Clean up
		mgr.RemoveSession(session.ID)

		if mgr.HasActiveSessions() {
			t.Error("Should not have active sessions after removal")
		}
	})
}

func TestSessionManagerGetSessionCount(t *testing.T) {
	t.Run("returns zero with no sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		mgr := NewSessionManager(registry)

		count := mgr.GetSessionCount()
		if count != 0 {
			t.Errorf("Expected 0 sessions, got %d", count)
		}
	})

	t.Run("returns correct count with sessions", func(t *testing.T) {
		registry := providers.NewRegistry()
		fallbackProvider := providers.NewFallbackProvider()
		registry.Register(fallbackProvider)

		mgr := NewSessionManager(registry)

		session1, _ := mgr.CreateSession("", "session 1", 0)
		count := mgr.GetSessionCount()
		if count != 1 {
			t.Errorf("Expected 1 session, got %d", count)
		}

		session2, _ := mgr.CreateSession("", "session 2", 0)
		count = mgr.GetSessionCount()
		if count != 2 {
			t.Errorf("Expected 2 sessions, got %d", count)
		}

		mgr.RemoveSession(session1.ID)
		count = mgr.GetSessionCount()
		if count != 1 {
			t.Errorf("Expected 1 session after removal, got %d", count)
		}

		// Clean up
		mgr.RemoveSession(session2.ID)
	})
}

// Benchmark tests
func BenchmarkNewSession(b *testing.B) {
	for i := 0; i < b.N; i++ {
		session := NewSession("provider", "reason", 30*time.Minute)
		_ = session
	}
}

func BenchmarkSessionIsExpired(b *testing.B) {
	session := NewSession("provider", "reason", 1*time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		expired := session.IsExpired()
		_ = expired
	}
}

func BenchmarkSessionTimeRemaining(b *testing.B) {
	session := NewSession("provider", "reason", 1*time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		remaining := session.TimeRemaining()
		_ = remaining
	}
}

func BenchmarkSessionFormatTimeRemaining(b *testing.B) {
	session := NewSession("provider", "reason", 2*time.Hour+30*time.Minute+45*time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		formatted := session.FormatTimeRemaining()
		_ = formatted
	}
}

func BenchmarkSessionManagerCreateSession(b *testing.B) {
	registry := providers.NewRegistry()
	fallbackProvider := providers.NewFallbackProvider()
	registry.Register(fallbackProvider)

	mgr := NewSessionManager(registry)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		session, err := mgr.CreateSession("", "benchmark", 0)
		if err != nil {
			b.Fatalf("Failed to create session: %s", err.Error())
		}
		mgr.RemoveSession(session.ID)
	}
}

// Test utilities
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
