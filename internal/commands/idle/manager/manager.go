package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/detector"
	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/providers"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Manager is the main idle prevention manager
type Manager struct {
	mu             sync.RWMutex
	registry       *providers.Registry
	sessionManager *SessionManager
	env            *detector.Environment
	stateFile      string
}

// NewManager creates a new idle manager
func NewManager() *Manager {
	// Detect environment
	env := detector.Detect()

	// Create provider registry
	registry := providers.NewRegistry()

	// Register providers based on environment
	registerProviders(registry, env)

	// Create session manager
	sessionManager := NewSessionManager(registry)

	// Determine state file path
	stateDir := paths.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(os.Getenv("HOME"), ".local", "state", "heimdall")
	}
	os.MkdirAll(stateDir, 0755)
	stateFile := filepath.Join(stateDir, "idle-sessions.json")

	manager := &Manager{
		registry:       registry,
		sessionManager: sessionManager,
		env:            env,
		stateFile:      stateFile,
	}

	// Load existing sessions
	manager.loadState()

	return manager
}

// registerProviders registers all available providers
func registerProviders(registry *providers.Registry, env *detector.Environment) {
	// Register D-Bus provider
	dbusProvider := providers.NewDBusProvider()
	if err := registry.Register(dbusProvider); err != nil {
		logger.Debug("Failed to register D-Bus provider", "error", err)
	}

	// Register X11 provider
	x11Provider := providers.NewX11Provider()
	if err := registry.Register(x11Provider); err != nil {
		logger.Debug("Failed to register X11 provider", "error", err)
	}

	// Register systemd provider
	systemdProvider := providers.NewSystemdProvider()
	if err := registry.Register(systemdProvider); err != nil {
		logger.Debug("Failed to register systemd provider", "error", err)
	}

	// Always register fallback provider
	fallbackProvider := providers.NewFallbackProvider()
	if err := registry.Register(fallbackProvider); err != nil {
		logger.Debug("Failed to register fallback provider", "error", err)
	}
}

// Start creates a new idle inhibition session
func (m *Manager) Start(reason string, duration time.Duration, providerName string) (*Session, error) {
	session, err := m.sessionManager.CreateSession(providerName, reason, duration)
	if err != nil {
		return nil, err
	}

	// Save state
	m.saveState()

	// Log the action
	if duration > 0 {
		logger.Info("Started idle prevention",
			"provider", session.Provider,
			"duration", FormatDuration(duration),
			"session", session.ID[:8])
	} else {
		logger.Info("Started idle prevention",
			"provider", session.Provider,
			"session", session.ID[:8])
	}

	return session, nil
}

// Stop stops an idle inhibition session
func (m *Manager) Stop(sessionID string) error {
	// If no session ID provided, stop all sessions
	if sessionID == "" {
		return m.StopAll()
	}

	// Get session info before removing
	session, exists := m.sessionManager.GetSession(sessionID)
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	provider := session.Provider

	// Remove the session
	if err := m.sessionManager.RemoveSession(sessionID); err != nil {
		return err
	}

	// Save state
	m.saveState()

	logger.Info("Stopped idle prevention",
		"provider", provider,
		"session", sessionID[:8])

	return nil
}

// StopAll stops all idle inhibition sessions
func (m *Manager) StopAll() error {
	count := m.sessionManager.GetSessionCount()

	if count == 0 {
		return fmt.Errorf("no active sessions")
	}

	if err := m.sessionManager.RemoveAllSessions(); err != nil {
		return err
	}

	// Save state
	m.saveState()

	logger.Info("Stopped all idle prevention sessions", "count", count)

	return nil
}

// GetStatus returns the current status
func (m *Manager) GetStatus() (bool, []*Session, []string) {
	sessions := m.sessionManager.ListSessions()
	active := len(sessions) > 0

	// Get available providers
	availableProviders := []string{}
	for _, provider := range m.registry.GetAvailable() {
		availableProviders = append(availableProviders, provider.Name())
	}

	return active, sessions, availableProviders
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	return m.sessionManager.ListSessions()
}

// GetEnvironment returns the detected environment
func (m *Manager) GetEnvironment() *detector.Environment {
	return m.env
}

// saveState saves the current state to disk
func (m *Manager) saveState() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessions := m.sessionManager.ListSessions()

	// Create state structure
	state := struct {
		Version  string     `json:"version"`
		SavedAt  time.Time  `json:"saved_at"`
		Sessions []*Session `json:"sessions"`
	}{
		Version:  "1.0",
		SavedAt:  time.Now(),
		Sessions: sessions,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write atomically
	tempFile := m.stateFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tempFile, m.stateFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// loadState loads the state from disk
func (m *Manager) loadState() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if state file exists
	if _, err := os.Stat(m.stateFile); os.IsNotExist(err) {
		return nil // No state to load
	}

	// Read state file
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal state
	var state struct {
		Version  string     `json:"version"`
		SavedAt  time.Time  `json:"saved_at"`
		Sessions []*Session `json:"sessions"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupted state file, remove it
		os.Remove(m.stateFile)
		return fmt.Errorf("corrupted state file: %w", err)
	}

	// Check for stale sessions (> 24 hours old)
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, session := range state.Sessions {
		// Skip expired sessions
		if session.IsExpired() {
			logger.Debug("Skipping expired session", "id", session.ID[:8])
			continue
		}

		// Skip stale sessions
		if session.StartTime.Before(cutoff) {
			logger.Debug("Skipping stale session", "id", session.ID[:8])
			continue
		}

		// Restore the session to our session manager
		// We can't restore the actual inhibition cookie, but we track the session
		m.sessionManager.sessions[session.ID] = session

		// For sessions with timers, recalculate the timer
		if session.ExpiresAt != nil && !session.IsExpired() {
			remaining := time.Until(*session.ExpiresAt)
			if remaining > 0 {
				session.Timer = time.AfterFunc(remaining, func() {
					m.sessionManager.RemoveSession(session.ID)
					m.saveState()
				})
			}
		}

		logger.Debug("Restored session from state",
			"id", session.ID[:8],
			"provider", session.Provider,
			"age", time.Since(session.StartTime))
	}

	return nil
}

// Cleanup performs cleanup operations
func (m *Manager) Cleanup() error {
	// Stop all sessions
	if err := m.sessionManager.RemoveAllSessions(); err != nil {
		logger.Error("Failed to remove all sessions during cleanup", "error", err)
	}

	// Remove state file
	if err := os.Remove(m.stateFile); err != nil && !os.IsNotExist(err) {
		logger.Error("Failed to remove state file during cleanup", "error", err)
	}

	return nil
}
