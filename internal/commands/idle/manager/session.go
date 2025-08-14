package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/providers"
	"github.com/google/uuid"
)

// Session represents an idle inhibition session
type Session struct {
	ID        string           `json:"id"`
	Provider  string           `json:"provider"`
	StartTime time.Time        `json:"start_time"`
	Timer     *time.Timer      `json:"-"`
	Duration  time.Duration    `json:"duration,omitempty"`
	Reason    string           `json:"reason"`
	Cookie    providers.Cookie `json:"-"`
	ExpiresAt *time.Time       `json:"expires_at,omitempty"`
}

// NewSession creates a new session
func NewSession(provider string, reason string, duration time.Duration) *Session {
	session := &Session{
		ID:        uuid.New().String(),
		Provider:  provider,
		StartTime: time.Now(),
		Duration:  duration,
		Reason:    reason,
	}

	if duration > 0 {
		expiresAt := time.Now().Add(duration)
		session.ExpiresAt = &expiresAt
	}

	return session
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// TimeRemaining returns the time remaining for the session
func (s *Session) TimeRemaining() time.Duration {
	if s.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(*s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// FormatTimeRemaining returns a human-readable time remaining
func (s *Session) FormatTimeRemaining() string {
	if s.ExpiresAt == nil {
		return "unlimited"
	}

	remaining := s.TimeRemaining()
	if remaining == 0 {
		return "expired"
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// SessionManager manages idle inhibition sessions
type SessionManager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session
	providers *providers.Registry
}

// NewSessionManager creates a new session manager
func NewSessionManager(registry *providers.Registry) *SessionManager {
	return &SessionManager{
		sessions:  make(map[string]*Session),
		providers: registry,
	}
}

// CreateSession creates a new idle inhibition session
func (m *SessionManager) CreateSession(providerName string, reason string, duration time.Duration) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get the provider
	var provider providers.IdleProvider
	var err error

	if providerName != "" {
		// Use specified provider
		p, exists := m.providers.Get(providerName)
		if !exists {
			return nil, fmt.Errorf("provider %s not found", providerName)
		}
		if !p.Available() {
			return nil, fmt.Errorf("provider %s is not available", providerName)
		}
		provider = p
	} else {
		// Auto-select best provider
		provider, err = m.providers.GetBest()
		if err != nil {
			return nil, err
		}
	}

	// Create the inhibition
	cookie, err := provider.Inhibit(reason)
	if err != nil {
		return nil, fmt.Errorf("failed to create inhibition: %w", err)
	}

	// Create session
	session := NewSession(provider.Name(), reason, duration)
	session.Cookie = cookie

	// Set up timer if duration is specified
	if duration > 0 {
		session.Timer = time.AfterFunc(duration, func() {
			m.RemoveSession(session.ID)
		})
	}

	// Store session
	m.sessions[session.ID] = session

	return session, nil
}

// GetSession returns a session by ID
func (m *SessionManager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	return session, exists
}

// ListSessions returns all active sessions
func (m *SessionManager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// RemoveSession removes a session by ID
func (m *SessionManager) RemoveSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session %s not found", id)
	}

	// Get the provider
	provider, exists := m.providers.Get(session.Provider)
	if !exists {
		return fmt.Errorf("provider %s not found", session.Provider)
	}

	// Release the inhibition
	if err := provider.Uninhibit(session.Cookie); err != nil {
		return fmt.Errorf("failed to release inhibition: %w", err)
	}

	// Cancel timer if exists
	if session.Timer != nil {
		session.Timer.Stop()
	}

	// Remove from map
	delete(m.sessions, id)

	return nil
}

// RemoveAllSessions removes all active sessions
func (m *SessionManager) RemoveAllSessions() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	for id, session := range m.sessions {
		// Get the provider
		provider, exists := m.providers.Get(session.Provider)
		if exists {
			// Try to release the inhibition
			if err := provider.Uninhibit(session.Cookie); err != nil {
				errors = append(errors, fmt.Errorf("session %s: %w", id, err))
			}
		}

		// Cancel timer if exists
		if session.Timer != nil {
			session.Timer.Stop()
		}

		// Remove from map
		delete(m.sessions, id)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove some sessions: %v", errors)
	}

	return nil
}

// HasActiveSessions returns true if there are any active sessions
func (m *SessionManager) HasActiveSessions() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessions) > 0
}

// GetSessionCount returns the number of active sessions
func (m *SessionManager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessions)
}
