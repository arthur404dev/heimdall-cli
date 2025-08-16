package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// StateManager manages theme state persistence and tracking
type StateManager struct {
	statePath string
	state     *ThemeState
}

// ThemeState represents the current theme state
type ThemeState struct {
	Current     CurrentTheme    `json:"current"`
	History     []ThemeHistory  `json:"history"`
	Generated   GeneratedInfo   `json:"generated"`
	Preferences UserPreferences `json:"preferences"`
	Version     string          `json:"version"`
}

// CurrentTheme represents the currently active theme
type CurrentTheme struct {
	Name      string              `json:"name"`
	Flavour   string              `json:"flavour,omitempty"`
	Mode      string              `json:"mode,omitempty"`
	Variant   string              `json:"variant,omitempty"`
	Source    scheme.SchemeSource `json:"source"`
	AppliedAt time.Time           `json:"applied_at"`
	Metadata  map[string]string   `json:"metadata,omitempty"`
}

// ThemeHistory represents a previously applied theme
type ThemeHistory struct {
	Name      string              `json:"name"`
	Flavour   string              `json:"flavour,omitempty"`
	Mode      string              `json:"mode,omitempty"`
	Variant   string              `json:"variant,omitempty"`
	Source    scheme.SchemeSource `json:"source"`
	AppliedAt time.Time           `json:"applied_at"`
}

// GeneratedInfo tracks information about generated themes
type GeneratedInfo struct {
	Available        bool              `json:"available"`
	WallpaperPath    string            `json:"wallpaper_path,omitempty"`
	GeneratedAt      time.Time         `json:"generated_at,omitempty"`
	PreferredVariant string            `json:"preferred_variant,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// UserPreferences stores user preferences for theme behavior
type UserPreferences struct {
	AutoApplyGenerated bool   `json:"auto_apply_generated"`
	AutoApplyUser      bool   `json:"auto_apply_user"`
	AutoApplyBundled   bool   `json:"auto_apply_bundled"`
	PreferredVariant   string `json:"preferred_variant,omitempty"`
	PreferredMode      string `json:"preferred_mode,omitempty"`
	NotifyOnGeneration bool   `json:"notify_on_generation"`
}

// NewStateManager creates a new theme state manager
func NewStateManager() *StateManager {
	cfg := config.Get()
	statePath := filepath.Join(paths.HeimdallStateDir, "theme-state.json")

	// Allow override from config
	if cfg != nil && cfg.Paths.StateDir != "" {
		statePath = filepath.Join(cfg.Paths.StateDir, "theme-state.json")
	}

	manager := &StateManager{
		statePath: statePath,
	}

	// Load existing state or create new
	if err := manager.Load(); err != nil {
		// Initialize with defaults if load fails
		manager.state = manager.getDefaultState()
	}

	return manager
}

// Load loads the theme state from disk
func (sm *StateManager) Load() error {
	// Ensure state directory exists
	stateDir := filepath.Dir(sm.statePath)
	if err := paths.EnsureDir(stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if state file exists
	if !paths.Exists(sm.statePath) {
		sm.state = sm.getDefaultState()
		return sm.Save() // Save default state
	}

	// Read state file
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse state
	var state ThemeState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}

	// Migrate if needed
	if state.Version != "1.0.0" {
		state = sm.migrateState(state)
	}

	sm.state = &state
	return nil
}

// Save saves the theme state to disk atomically
func (sm *StateManager) Save() error {
	// Ensure state directory exists
	stateDir := filepath.Dir(sm.statePath)
	if err := paths.EnsureDir(stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Update version
	sm.state.Version = "1.0.0"

	// Write atomically
	if err := paths.AtomicWriteJSON(sm.statePath, sm.state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetCurrent returns the current theme information
func (sm *StateManager) GetCurrent() CurrentTheme {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}
	return sm.state.Current
}

// SetCurrent updates the current theme and adds previous to history
func (sm *StateManager) SetCurrent(theme CurrentTheme) error {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}

	// Add current theme to history if it's different
	if sm.state.Current.Name != "" &&
		(sm.state.Current.Name != theme.Name ||
			sm.state.Current.Variant != theme.Variant) {
		sm.addToHistory(sm.state.Current)
	}

	// Update current theme
	theme.AppliedAt = time.Now()
	sm.state.Current = theme

	// Save state
	return sm.Save()
}

// GetHistory returns the theme history
func (sm *StateManager) GetHistory() []ThemeHistory {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}
	return sm.state.History
}

// RevertToPrevious reverts to the previous theme in history
func (sm *StateManager) RevertToPrevious() (*ThemeHistory, error) {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}

	if len(sm.state.History) == 0 {
		return nil, fmt.Errorf("no previous theme in history")
	}

	// Get the previous theme
	previous := sm.state.History[0]

	// Remove it from history
	sm.state.History = sm.state.History[1:]

	// Set it as current (without adding current to history again)
	sm.state.Current = CurrentTheme{
		Name:      previous.Name,
		Flavour:   previous.Flavour,
		Mode:      previous.Mode,
		Variant:   previous.Variant,
		Source:    previous.Source,
		AppliedAt: time.Now(),
	}

	// Save state
	if err := sm.Save(); err != nil {
		return nil, err
	}

	return &previous, nil
}

// SetGeneratedAvailable marks that a new generated theme is available
func (sm *StateManager) SetGeneratedAvailable(wallpaperPath string, metadata map[string]string) error {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}

	sm.state.Generated = GeneratedInfo{
		Available:        true,
		WallpaperPath:    wallpaperPath,
		GeneratedAt:      time.Now(),
		PreferredVariant: sm.state.Preferences.PreferredVariant,
		Metadata:         metadata,
	}

	return sm.Save()
}

// ClearGeneratedAvailable clears the generated theme availability
func (sm *StateManager) ClearGeneratedAvailable() error {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}

	sm.state.Generated = GeneratedInfo{
		Available: false,
	}

	return sm.Save()
}

// GetPreferences returns user preferences
func (sm *StateManager) GetPreferences() UserPreferences {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}
	return sm.state.Preferences
}

// UpdatePreferences updates user preferences
func (sm *StateManager) UpdatePreferences(prefs UserPreferences) error {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}

	sm.state.Preferences = prefs
	return sm.Save()
}

// ShouldAutoApply checks if a theme from the given source should be auto-applied
func (sm *StateManager) ShouldAutoApply(source scheme.SchemeSource) bool {
	prefs := sm.GetPreferences()

	switch source {
	case scheme.SourceGenerated:
		return prefs.AutoApplyGenerated
	case scheme.SourceUser:
		return prefs.AutoApplyUser
	case scheme.SourceBundled:
		return prefs.AutoApplyBundled
	default:
		return false
	}
}

// addToHistory adds a theme to the history, maintaining max size
func (sm *StateManager) addToHistory(theme CurrentTheme) {
	history := ThemeHistory{
		Name:      theme.Name,
		Flavour:   theme.Flavour,
		Mode:      theme.Mode,
		Variant:   theme.Variant,
		Source:    theme.Source,
		AppliedAt: theme.AppliedAt,
	}

	// Prepend to history
	sm.state.History = append([]ThemeHistory{history}, sm.state.History...)

	// Maintain max history size (5 items)
	if len(sm.state.History) > 5 {
		sm.state.History = sm.state.History[:5]
	}
}

// getDefaultState returns the default theme state
func (sm *StateManager) getDefaultState() *ThemeState {
	return &ThemeState{
		Current: CurrentTheme{
			Name:      "catppuccin",
			Flavour:   "mocha",
			Mode:      "dark",
			Source:    scheme.SourceBundled,
			AppliedAt: time.Now(),
		},
		History: []ThemeHistory{},
		Generated: GeneratedInfo{
			Available: false,
		},
		Preferences: UserPreferences{
			AutoApplyGenerated: false, // Don't auto-apply by default
			AutoApplyUser:      false,
			AutoApplyBundled:   false,
			PreferredVariant:   "tonal",
			PreferredMode:      "dark",
			NotifyOnGeneration: true,
		},
		Version: "1.0.0",
	}
}

// migrateState migrates old state formats to the current version
func (sm *StateManager) migrateState(old ThemeState) ThemeState {
	// For now, just update the version
	// Add migration logic here as needed
	old.Version = "1.0.0"

	// Ensure preferences have defaults
	if old.Preferences.PreferredVariant == "" {
		old.Preferences.PreferredVariant = "tonal"
	}
	if old.Preferences.PreferredMode == "" {
		old.Preferences.PreferredMode = "dark"
	}

	return old
}

// GetState returns the full state (for debugging/inspection)
func (sm *StateManager) GetState() *ThemeState {
	if sm.state == nil {
		sm.state = sm.getDefaultState()
	}
	return sm.state
}
