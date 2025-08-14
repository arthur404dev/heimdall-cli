package scheme

import (
	"fmt"
	"strings"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
)

// MockSchemeManager provides a mock implementation of scheme.Manager for testing
type MockSchemeManager struct {
	mu                   sync.RWMutex
	currentScheme        *scheme.Scheme
	schemes              map[string][]string                             // scheme -> flavours
	flavours             map[string]map[string][]string                  // scheme -> flavour -> modes
	schemeData           map[string]map[string]map[string]*scheme.Scheme // scheme -> flavour -> mode -> Scheme
	bundledSchemes       []scheme.BundledScheme
	bundledSchemeNames   []string
	errors               map[string]error // method name -> error to return
	installAllCalled     bool
	installBundledCalled []string
	saveSchemesCalled    []*scheme.Scheme
}

// NewMockSchemeManager creates a new mock scheme manager
func NewMockSchemeManager() *MockSchemeManager {
	return &MockSchemeManager{
		schemes:              make(map[string][]string),
		flavours:             make(map[string]map[string][]string),
		schemeData:           make(map[string]map[string]map[string]*scheme.Scheme),
		bundledSchemes:       make([]scheme.BundledScheme, 0),
		bundledSchemeNames:   make([]string, 0),
		errors:               make(map[string]error),
		installBundledCalled: make([]string, 0),
		saveSchemesCalled:    make([]*scheme.Scheme, 0),
	}
}

// SetCurrentScheme sets the current scheme for testing
func (m *MockSchemeManager) SetCurrentScheme(s *scheme.Scheme) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentScheme = s
}

// AddScheme adds a scheme for testing
func (m *MockSchemeManager) AddScheme(name string, flavours []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schemes[name] = flavours
	if m.flavours[name] == nil {
		m.flavours[name] = make(map[string][]string)
	}
}

// AddFlavour adds a flavour to a scheme for testing
func (m *MockSchemeManager) AddFlavour(schemeName, flavour string, modes []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.flavours[schemeName] == nil {
		m.flavours[schemeName] = make(map[string][]string)
	}
	m.flavours[schemeName][flavour] = modes
}

// AddSchemeData adds scheme data for testing
func (m *MockSchemeManager) AddSchemeData(schemeName, flavour, mode string, s *scheme.Scheme) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.schemeData[schemeName] == nil {
		m.schemeData[schemeName] = make(map[string]map[string]*scheme.Scheme)
	}
	if m.schemeData[schemeName][flavour] == nil {
		m.schemeData[schemeName][flavour] = make(map[string]*scheme.Scheme)
	}
	m.schemeData[schemeName][flavour][mode] = s
}

// AddBundledScheme adds a bundled scheme for testing
func (m *MockSchemeManager) AddBundledScheme(s scheme.BundledScheme) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bundledSchemes = append(m.bundledSchemes, s)
	m.bundledSchemeNames = append(m.bundledSchemeNames, s.Name)
}

// SetError sets an error to be returned by a specific method
func (m *MockSchemeManager) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// GetCurrent returns the current active scheme
func (m *MockSchemeManager) GetCurrent() (*scheme.Scheme, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["GetCurrent"]; exists {
		return nil, err
	}

	if m.currentScheme == nil {
		// Return default scheme like the real manager
		return &scheme.Scheme{
			Name:    "catppuccin",
			Flavour: "mocha",
			Mode:    "dark",
			Variant: "tonalspot",
			Colours: map[string]string{
				"base": "1e1e2e",
				"text": "cdd6f4",
			},
		}, nil
	}

	return m.currentScheme, nil
}

// SetScheme sets the active scheme
func (m *MockSchemeManager) SetScheme(s *scheme.Scheme) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.errors["SetScheme"]; exists {
		return err
	}

	m.currentScheme = s
	return nil
}

// ListSchemes returns available scheme names
func (m *MockSchemeManager) ListSchemes() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["ListSchemes"]; exists {
		return nil, err
	}

	var schemes []string
	for name := range m.schemes {
		schemes = append(schemes, name)
	}
	return schemes, nil
}

// ListFlavours returns available flavours for a scheme
func (m *MockSchemeManager) ListFlavours(schemeName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["ListFlavours"]; exists {
		return nil, err
	}

	flavours, exists := m.schemes[schemeName]
	if !exists {
		return nil, fmt.Errorf("scheme %s not found", schemeName)
	}

	return flavours, nil
}

// ListModes returns available modes for a scheme flavour
func (m *MockSchemeManager) ListModes(schemeName, flavour string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["ListModes"]; exists {
		return nil, err
	}

	if m.flavours[schemeName] == nil {
		return nil, fmt.Errorf("scheme %s not found", schemeName)
	}

	modes, exists := m.flavours[schemeName][flavour]
	if !exists {
		return nil, fmt.Errorf("flavour %s not found for scheme %s", flavour, schemeName)
	}

	return modes, nil
}

// LoadScheme loads a specific scheme
func (m *MockSchemeManager) LoadScheme(name, flavour, mode string) (*scheme.Scheme, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["LoadScheme"]; exists {
		return nil, err
	}

	if m.schemeData[name] == nil || m.schemeData[name][flavour] == nil {
		return nil, fmt.Errorf("scheme %s/%s/%s not found", name, flavour, mode)
	}

	s, exists := m.schemeData[name][flavour][mode]
	if !exists {
		return nil, fmt.Errorf("scheme %s/%s/%s not found", name, flavour, mode)
	}

	return s, nil
}

// LoadSchemeWithFallback tries to load a user scheme first, then falls back to bundled
func (m *MockSchemeManager) LoadSchemeWithFallback(name, flavour, mode string) (*scheme.Scheme, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors["LoadSchemeWithFallback"]; exists {
		return nil, err
	}

	// Try to load from schemeData first
	if m.schemeData[name] != nil && m.schemeData[name][flavour] != nil {
		if s, exists := m.schemeData[name][flavour][mode]; exists {
			return s, nil
		}
	}

	// Try bundled schemes
	for _, bundled := range m.bundledSchemes {
		if strings.EqualFold(bundled.Name, name) ||
			strings.EqualFold(bundled.Family, name) ||
			strings.EqualFold(fmt.Sprintf("%s %s", bundled.Family, bundled.Flavour), name) {
			return &scheme.Scheme{
				Name:    name,
				Flavour: flavour,
				Mode:    mode,
				Variant: bundled.Variant,
				Colours: bundled.Colors,
			}, nil
		}
	}

	return nil, fmt.Errorf("scheme %s/%s/%s not found in user or bundled schemes", name, flavour, mode)
}

// SaveScheme saves a scheme to the schemes directory
func (m *MockSchemeManager) SaveScheme(s *scheme.Scheme) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.errors["SaveScheme"]; exists {
		return err
	}

	m.saveSchemesCalled = append(m.saveSchemesCalled, s)
	return nil
}

// InstallBundledScheme installs a bundled scheme
func (m *MockSchemeManager) InstallBundledScheme(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.errors["InstallBundledScheme"]; exists {
		return err
	}

	m.installBundledCalled = append(m.installBundledCalled, name)

	// Check if bundled scheme exists
	for _, bundled := range m.bundledSchemes {
		if strings.EqualFold(bundled.Name, name) ||
			strings.EqualFold(fmt.Sprintf("%s %s", bundled.Family, bundled.Flavour), name) {
			return nil
		}
	}

	return fmt.Errorf("bundled scheme '%s' not found", name)
}

// InstallAllBundledSchemes installs all bundled schemes
func (m *MockSchemeManager) InstallAllBundledSchemes() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, exists := m.errors["InstallAllBundledSchemes"]; exists {
		return err
	}

	m.installAllCalled = true
	return nil
}

// GetInstallAllCalled returns whether InstallAllBundledSchemes was called
func (m *MockSchemeManager) GetInstallAllCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.installAllCalled
}

// GetInstallBundledCalled returns the list of schemes passed to InstallBundledScheme
func (m *MockSchemeManager) GetInstallBundledCalled() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string(nil), m.installBundledCalled...)
}

// GetSaveSchemesCalled returns the list of schemes passed to SaveScheme
func (m *MockSchemeManager) GetSaveSchemesCalled() []*scheme.Scheme {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*scheme.Scheme(nil), m.saveSchemesCalled...)
}

// Clear clears all mock data
func (m *MockSchemeManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentScheme = nil
	m.schemes = make(map[string][]string)
	m.flavours = make(map[string]map[string][]string)
	m.schemeData = make(map[string]map[string]map[string]*scheme.Scheme)
	m.bundledSchemes = make([]scheme.BundledScheme, 0)
	m.bundledSchemeNames = make([]string, 0)
	m.errors = make(map[string]error)
	m.installAllCalled = false
	m.installBundledCalled = make([]string, 0)
	m.saveSchemesCalled = make([]*scheme.Scheme, 0)
}

// MockBundledSchemeProvider provides mock functions for bundled schemes
type MockBundledSchemeProvider struct {
	schemes []scheme.BundledScheme
	names   []string
	errors  map[string]error
}

// NewMockBundledSchemeProvider creates a new mock bundled scheme provider
func NewMockBundledSchemeProvider() *MockBundledSchemeProvider {
	return &MockBundledSchemeProvider{
		schemes: make([]scheme.BundledScheme, 0),
		names:   make([]string, 0),
		errors:  make(map[string]error),
	}
}

// AddBundledScheme adds a bundled scheme for testing
func (m *MockBundledSchemeProvider) AddBundledScheme(s scheme.BundledScheme) {
	m.schemes = append(m.schemes, s)
	m.names = append(m.names, s.Name)
}

// SetError sets an error to be returned by a specific function
func (m *MockBundledSchemeProvider) SetError(function string, err error) {
	m.errors[function] = err
}

// GetBundledSchemes returns all bundled color schemes
func (m *MockBundledSchemeProvider) GetBundledSchemes() ([]scheme.BundledScheme, error) {
	if err, exists := m.errors["GetBundledSchemes"]; exists {
		return nil, err
	}
	return m.schemes, nil
}

// ListBundledSchemeNames returns the names of all bundled schemes
func (m *MockBundledSchemeProvider) ListBundledSchemeNames() ([]string, error) {
	if err, exists := m.errors["ListBundledSchemeNames"]; exists {
		return nil, err
	}
	return m.names, nil
}

// GetBundledScheme returns a specific bundled scheme by name
func (m *MockBundledSchemeProvider) GetBundledScheme(name string) (*scheme.BundledScheme, error) {
	if err, exists := m.errors["GetBundledScheme"]; exists {
		return nil, err
	}

	for _, s := range m.schemes {
		if strings.EqualFold(s.Name, name) ||
			strings.EqualFold(fmt.Sprintf("%s %s", s.Family, s.Flavour), name) {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("bundled scheme '%s' not found", name)
}

// MockThemeApplier provides a mock theme applier for testing
type MockThemeApplier struct {
	mu            sync.RWMutex
	appliedThemes []AppliedTheme
	applyError    error
}

// AppliedTheme represents a theme that was applied
type AppliedTheme struct {
	App    string
	Colors map[string]string
	Mode   string
}

// NewMockThemeApplier creates a new mock theme applier
func NewMockThemeApplier() *MockThemeApplier {
	return &MockThemeApplier{
		appliedThemes: make([]AppliedTheme, 0),
	}
}

// SetApplyError sets an error to be returned when applying themes
func (m *MockThemeApplier) SetApplyError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.applyError = err
}

// ApplyTheme applies a theme to an app
func (m *MockThemeApplier) ApplyTheme(app string, colors map[string]string, mode string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.applyError != nil {
		return m.applyError
	}

	m.appliedThemes = append(m.appliedThemes, AppliedTheme{
		App:    app,
		Colors: colors,
		Mode:   mode,
	})

	return nil
}

// GetAppliedThemes returns all applied themes
func (m *MockThemeApplier) GetAppliedThemes() []AppliedTheme {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]AppliedTheme(nil), m.appliedThemes...)
}

// Clear clears all applied themes
func (m *MockThemeApplier) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appliedThemes = make([]AppliedTheme, 0)
	m.applyError = nil
}

// MockNotifier provides a mock notification system for testing
type MockNotifier struct {
	mu            sync.RWMutex
	notifications []MockNotification
	sendError     error
	isAvailable   bool
}

// MockNotification represents a mock notification
type MockNotification struct {
	Summary string
	Body    string
	Urgency string
}

// NewMockNotifier creates a new mock notifier
func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		notifications: make([]MockNotification, 0),
		isAvailable:   true,
	}
}

// SetSendError sets an error to be returned when sending notifications
func (m *MockNotifier) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendError = err
}

// SetAvailable sets whether the notification system is available
func (m *MockNotifier) SetAvailable(available bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isAvailable = available
}

// Send sends a mock notification
func (m *MockNotifier) Send(notification interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendError != nil {
		return m.sendError
	}

	// This is a simplified mock - in real usage, we'd need to handle the actual notification type
	m.notifications = append(m.notifications, MockNotification{
		Summary: "Mock Notification",
		Body:    "Mock Body",
		Urgency: "normal",
	})

	return nil
}

// GetNotifications returns all sent notifications
func (m *MockNotifier) GetNotifications() []MockNotification {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]MockNotification(nil), m.notifications...)
}

// Clear clears all notifications
func (m *MockNotifier) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications = make([]MockNotification, 0)
	m.sendError = nil
}
