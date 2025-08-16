package scheme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/assets/schemes"
	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// SchemeSource represents where a scheme comes from
type SchemeSource string

const (
	SourceBundled   SchemeSource = "bundled"
	SourceUser      SchemeSource = "user"
	SourceGenerated SchemeSource = "generated"
)

// Scheme represents a color scheme
type Scheme struct {
	Name    string            `json:"name"`
	Flavour string            `json:"flavour"`
	Mode    string            `json:"mode"`
	Variant string            `json:"variant"`
	Colours map[string]string `json:"colours"` // British spelling, simple strings
	Source  SchemeSource      `json:"-"`       // Not persisted, runtime only
}

// Manager manages color schemes
type Manager struct {
	schemesDir string
	stateDir   string
}

// NewManager creates a new scheme manager
func NewManager() *Manager {
	m := &Manager{
		schemesDir: paths.SchemeDataDir, // Default, will be overridden
		stateDir:   paths.StateDir,
	}

	// Use configured generated path if available
	m.schemesDir = m.getGeneratedSchemePath()

	return m
}

// getUserSchemePaths returns the configured user scheme paths
func (m *Manager) getUserSchemePaths() []string {
	cfg := config.Get()
	if cfg == nil || len(cfg.Scheme.UserPaths) == 0 {
		// Return default if config not loaded or no paths configured
		return []string{paths.UserSchemeDir}
	}

	// Expand ~ to home directory for each path
	expandedPaths := make([]string, 0, len(cfg.Scheme.UserPaths))
	for _, p := range cfg.Scheme.UserPaths {
		if strings.HasPrefix(p, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				p = filepath.Join(home, p[2:])
			}
		}
		expandedPaths = append(expandedPaths, p)
	}

	return expandedPaths
}

// getGeneratedSchemePath returns the configured generated scheme path
func (m *Manager) getGeneratedSchemePath() string {
	cfg := config.Get()
	if cfg == nil || cfg.Scheme.GeneratedPath == "" {
		// Return default if config not loaded or no path configured
		return paths.SchemeDataDir
	}

	// Expand ~ to home directory if needed
	p := cfg.Scheme.GeneratedPath
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, p[2:])
		}
	}

	return p
}

// GetCurrent returns the current active scheme
func (m *Manager) GetCurrent() (*Scheme, error) {
	statePath := filepath.Join(m.stateDir, "scheme.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default scheme
			return &Scheme{
				Name:    "catppuccin",
				Flavour: "mocha",
				Mode:    "dark",
				Variant: "tonalspot",
				Colours: getDefaultColours(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read current scheme: %w", err)
	}

	var scheme Scheme
	if err := json.Unmarshal(data, &scheme); err != nil {
		return nil, fmt.Errorf("failed to parse current scheme: %w", err)
	}

	// Normalize colors to ensure consistency (remove # prefix if present)
	// This handles cases where the state file might have inconsistent formats
	for key, value := range scheme.Colours {
		scheme.Colours[key] = strings.TrimPrefix(value, "#")
	}

	return &scheme, nil
}

// SetScheme sets the active scheme with triple-write for QuickShell integration
func (m *Manager) SetScheme(scheme *Scheme) error {
	// Normalize colors to ensure they don't have # prefix internally
	// This ensures consistency regardless of source
	normalizedColors := make(map[string]string)
	for key, value := range scheme.Colours {
		normalizedColors[key] = strings.TrimPrefix(value, "#")
	}

	// Prepare Heimdall format data (colors stored without # prefix)
	heimdallScheme := &Scheme{
		Name:    scheme.Name,
		Flavour: scheme.Flavour,
		Mode:    scheme.Mode,
		Variant: scheme.Variant,
		Colours: normalizedColors,
	}

	// 1. Primary write to Heimdall config location
	configPath := filepath.Join(paths.ConfigDir, "scheme.json")
	if err := paths.EnsureDir(paths.ConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := paths.AtomicWriteJSON(configPath, heimdallScheme); err != nil {
		return fmt.Errorf("failed to write config scheme: %w", err)
	}

	// 2. Secondary write to Heimdall state location
	statePath := filepath.Join(m.stateDir, "scheme.json")
	if err := paths.EnsureDir(m.stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	if err := paths.AtomicWriteJSON(statePath, heimdallScheme); err != nil {
		return fmt.Errorf("failed to write state scheme: %w", err)
	}

	// 3. CRITICAL: QuickShell-specific format (no # prefix, "colours" key)
	quickshellScheme := m.prepareQuickShellFormat(heimdallScheme)
	quickshellDir := filepath.Join(os.Getenv("HOME"), ".local", "state", "quickshell", "user", "generated")

	// Create QuickShell directory if it doesn't exist
	if err := os.MkdirAll(quickshellDir, 0755); err != nil {
		// Log warning but don't fail - QuickShell might not be installed
		fmt.Fprintf(os.Stderr, "Warning: Failed to create QuickShell directory: %v\n", err)
		return nil
	}

	quickshellPath := filepath.Join(quickshellDir, "scheme.json")
	if err := paths.AtomicWriteJSON(quickshellPath, quickshellScheme); err != nil {
		// Log warning but don't fail the primary operation
		fmt.Fprintf(os.Stderr, "Warning: Failed to write QuickShell colors: %v\n", err)
	} else {
		// Log success for QuickShell integration
		fmt.Fprintf(os.Stderr, "Info: Updated QuickShell colors at %s\n", quickshellPath)
	}

	return nil
}

// prepareQuickShellFormat converts scheme to QuickShell's expected format
func (m *Manager) prepareQuickShellFormat(scheme *Scheme) map[string]interface{} {
	// Strip # prefix from all colors for QuickShell
	colours := make(map[string]string)
	for key, value := range scheme.Colours {
		colours[key] = strings.TrimPrefix(value, "#")
	}

	// QuickShell also expects special colors
	special := make(map[string]string)
	if cursor, ok := scheme.Colours["cursor"]; ok {
		special["cursor"] = strings.TrimPrefix(cursor, "#")
	}
	if cursorText, ok := scheme.Colours["cursor_text"]; ok {
		special["cursor_text"] = strings.TrimPrefix(cursorText, "#")
	}

	return map[string]interface{}{
		"name":    scheme.Name,
		"flavour": scheme.Flavour,
		"mode":    scheme.Mode,
		"variant": scheme.Variant,
		"colours": colours, // Note: British spelling for QuickShell
		"special": special,
	}
}

// ListSchemes returns available scheme names from the schemes directory and embedded assets
func (m *Manager) ListSchemes() ([]string, error) {
	schemeMap := make(map[string]bool) // Use map to avoid duplicates

	// First, add user schemes from configured paths (higher priority)
	userPaths := m.getUserSchemePaths()
	for _, userPath := range userPaths {
		entries, err := os.ReadDir(userPath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() {
				schemeMap[entry.Name()] = true
			}
		}
	}

	// Then, add schemes from filesystem directories (legacy locations)
	schemeDirs := []string{
		m.schemesDir, // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes"), // Legacy cache location with extra "schemes" level
		paths.SchemeCacheDir,                           // Direct cache location
	}

	for _, schemeDir := range schemeDirs {
		entries, err := os.ReadDir(schemeDir)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() {
				schemeMap[entry.Name()] = true
			}
		}
	}

	// Finally, add bundled schemes from embedded assets (lowest priority)
	err := fs.WalkDir(schemes.Content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root and non-directories
		if path == "." || !d.IsDir() {
			return nil
		}

		// Only add top-level directories (scheme names)
		parts := strings.Split(path, "/")
		if len(parts) == 1 {
			schemeMap[parts[0]] = true
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk embedded schemes: %w", err)
	}

	// Convert map to slice
	var schemesList []string
	for scheme := range schemeMap {
		schemesList = append(schemesList, scheme)
	}

	return schemesList, nil
}

// ListFlavours returns available flavours for a scheme
func (m *Manager) ListFlavours(schemeName string) ([]string, error) {
	flavourMap := make(map[string]bool) // Use map to avoid duplicates

	// First check user scheme paths (highest priority)
	userPaths := m.getUserSchemePaths()
	for _, userPath := range userPaths {
		schemePath := filepath.Join(userPath, schemeName)
		entries, err := os.ReadDir(schemePath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() {
				flavourMap[entry.Name()] = true
			}
		}
	}

	// Then check filesystem locations
	schemePaths := []string{
		filepath.Join(m.schemesDir, schemeName),                    // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes", schemeName), // Legacy cache location with extra "schemes" level
		filepath.Join(paths.SchemeCacheDir, schemeName),            // Direct cache location
	}

	for _, schemePath := range schemePaths {
		entries, err := os.ReadDir(schemePath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() {
				flavourMap[entry.Name()] = true
			}
		}
	}

	// Finally check embedded assets (lowest priority)
	err := fs.WalkDir(schemes.Content, schemeName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip if scheme doesn't exist in embedded assets
		}

		// Skip the root directory and files
		if path == schemeName || !d.IsDir() {
			return nil
		}

		// Extract flavour name (first level subdirectory)
		relativePath := strings.TrimPrefix(path, schemeName+"/")
		parts := strings.Split(relativePath, "/")
		if len(parts) >= 1 && parts[0] != "" {
			flavourMap[parts[0]] = true
		}

		return nil
	})
	// Ignore errors from embedded assets walk (scheme might not exist there)
	_ = err

	if len(flavourMap) == 0 {
		return nil, fmt.Errorf("no flavours found for scheme %s", schemeName)
	}

	// Convert map to slice
	var flavours []string
	for flavour := range flavourMap {
		flavours = append(flavours, flavour)
	}

	return flavours, nil
}

// ListModes returns available modes for a scheme flavour
func (m *Manager) ListModes(schemeName, flavour string) ([]string, error) {
	modeMap := make(map[string]bool) // Use map to avoid duplicates

	// First check user scheme paths (highest priority)
	userPaths := m.getUserSchemePaths()
	for _, userPath := range userPaths {
		flavourPath := filepath.Join(userPath, schemeName, flavour)
		entries, err := os.ReadDir(flavourPath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				mode := strings.TrimSuffix(entry.Name(), ".json")
				modeMap[mode] = true
			}
		}
	}

	// Then check filesystem locations
	flavourPaths := []string{
		filepath.Join(m.schemesDir, schemeName, flavour),                    // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes", schemeName, flavour), // Legacy cache location with extra "schemes" level
		filepath.Join(paths.SchemeCacheDir, schemeName, flavour),            // Direct cache location
	}

	for _, flavourPath := range flavourPaths {
		entries, err := os.ReadDir(flavourPath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				mode := strings.TrimSuffix(entry.Name(), ".json")
				modeMap[mode] = true
			}
		}
	}

	// Finally check embedded assets (lowest priority)
	embeddedPath := filepath.Join(schemeName, flavour)
	entries, err := fs.ReadDir(schemes.Content, embeddedPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				mode := strings.TrimSuffix(entry.Name(), ".json")
				modeMap[mode] = true
			}
		}
	}

	if len(modeMap) == 0 {
		return nil, fmt.Errorf("no modes found for flavour %s/%s", schemeName, flavour)
	}

	// Convert map to slice
	var modes []string
	for mode := range modeMap {
		modes = append(modes, mode)
	}

	return modes, nil
}

// LoadScheme loads a specific scheme
func (m *Manager) LoadScheme(name, flavour, mode string) (*Scheme, error) {
	var data []byte
	var err error
	var source SchemeSource

	// First try user scheme paths (highest priority)
	userPaths := m.getUserSchemePaths()
	for _, userPath := range userPaths {
		schemePath := filepath.Join(userPath, name, flavour, mode+".json")
		data, err = os.ReadFile(schemePath)
		if err == nil {
			// Determine source based on path location
			generatedPath := m.getGeneratedSchemePath()
			if strings.HasPrefix(schemePath, generatedPath) {
				source = SourceGenerated
			} else {
				source = SourceUser
			}
			break
		}
	}

	// Then try other filesystem locations
	if data == nil {
		schemePaths := []string{
			// Primary location (data dir)
			filepath.Join(m.schemesDir, name, flavour, mode+".json"),
			// Skip cache locations - they have old format
		}

		for _, schemePath := range schemePaths {
			data, err = os.ReadFile(schemePath)
			if err == nil {
				// Determine source based on path location
				generatedPath := m.getGeneratedSchemePath()
				if strings.HasPrefix(schemePath, generatedPath) {
					source = SourceGenerated
				} else {
					source = SourceUser
				}
				break
			}
		}
	}

	// If found on disk, parse it
	if data != nil {
		// Parse JSON format from disk
		scheme := &Scheme{
			Name:    name,
			Flavour: flavour,
			Mode:    mode,
			Colours: make(map[string]string),
			Source:  source,
		}

		// Parse JSON directly into a map to handle flexible formats
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err != nil {
			// Try to provide a more helpful error message
			if validationErr, ok := err.(*ValidationError); ok {
				return nil, validationErr
			}
			return nil, fmt.Errorf("failed to parse scheme JSON: %w", err)
		}

		// Extract colours (handle both British and American spelling)
		if colours, ok := rawData["colours"].(map[string]interface{}); ok {
			for key, value := range colours {
				if colorStr, ok := value.(string); ok {
					scheme.Colours[key] = strings.TrimPrefix(colorStr, "#")
				}
			}
		} else if colors, ok := rawData["colors"].(map[string]interface{}); ok {
			for key, value := range colors {
				if colorStr, ok := value.(string); ok {
					scheme.Colours[key] = strings.TrimPrefix(colorStr, "#")
				}
			}
		}

		// Extract variant if present
		if variant, ok := rawData["variant"].(string); ok {
			scheme.Variant = variant
		}

		// Always use the detected source based on file location
		// The source should be determined by WHERE the file is, not what's IN the file
		scheme.Source = source

		// Validate ALL loaded schemes regardless of source
		// All schemes should meet the same standards
		// Sanitize first to fix common issues
		SanitizeScheme(scheme)

		// Then validate
		if err := ValidateScheme(scheme); err != nil {
			// Log the validation error but don't fail - allow partial schemes
			// This is more user-friendly for all schemes
			if validationErrs, ok := err.(ValidationErrors); ok && len(validationErrs) > 0 {
				// Only fail on critical errors (missing colors map)
				for _, vErr := range validationErrs {
					if vErr.Field == "colours" && strings.Contains(vErr.Message, "colors map is required") {
						return nil, fmt.Errorf("invalid scheme %s: %w", name, err)
					}
				}
				// For non-critical errors, just log them (in production, you'd use a logger)
				// fmt.Printf("Warning: Scheme %s has validation issues: %v\n", name, err)
			}
		}

		return scheme, nil
	}

	// If not found on disk, try loading from embedded assets
	// Try JSON format first
	embeddedPath := fmt.Sprintf("%s/%s/%s.json", name, flavour, mode)
	embeddedData, err := schemes.Content.ReadFile(embeddedPath)
	if err == nil {
		// Parse embedded JSON format
		var scheme Scheme
		if err := json.Unmarshal(embeddedData, &scheme); err != nil {
			return nil, fmt.Errorf("failed to parse embedded scheme JSON: %w", err)
		}

		// Source is always bundled for embedded schemes
		scheme.Source = SourceBundled

		// Ensure colors don't have # prefix in storage (add it when needed)
		for key, value := range scheme.Colours {
			scheme.Colours[key] = strings.TrimPrefix(value, "#")
		}

		return &scheme, nil
	}

	// Fallback to .txt format for backward compatibility
	embeddedPath = fmt.Sprintf("%s/%s/%s.txt", name, flavour, mode)
	embeddedData, err = schemes.Content.ReadFile(embeddedPath)
	if err == nil {
		// Parse embedded .txt format (space-separated key-value pairs)
		scheme := &Scheme{
			Name:    name,
			Flavour: flavour,
			Mode:    mode,
			Colours: make(map[string]string),
			Source:  SourceBundled,
		}

		// Parse the space-separated format
		lines := strings.Split(string(embeddedData), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) >= 2 {
				key := parts[0]
				value := strings.TrimPrefix(parts[1], "#")
				scheme.Colours[key] = value
			}
		}

		return scheme, nil
	}

	return nil, fmt.Errorf("scheme %s/%s/%s not found in user or bundled schemes", name, flavour, mode)
}

// SaveScheme saves a scheme to the schemes directory
func (m *Manager) SaveScheme(scheme *Scheme) error {
	schemePath := filepath.Join(m.schemesDir, scheme.Name, scheme.Flavour)

	// Ensure directory exists
	if err := paths.EnsureDir(schemePath); err != nil {
		return fmt.Errorf("failed to create scheme directory: %w", err)
	}

	// Safety-net: Ensure source is set when saving
	if scheme.Source == "" {
		// Determine source based on save location
		// If saving to the generated directory, it's a generated scheme
		// Otherwise it's a user scheme
		generatedPath := m.getGeneratedSchemePath()
		if m.schemesDir == generatedPath {
			scheme.Source = SourceGenerated
		} else {
			scheme.Source = SourceUser
		}
	}

	// Write scheme file
	filePath := filepath.Join(schemePath, scheme.Mode+".json")
	if err := paths.AtomicWriteJSON(filePath, scheme); err != nil {
		return fmt.Errorf("failed to write scheme file: %w", err)
	}

	return nil
}

// SaveSchemeToUser saves a scheme to the user schemes directory
func (m *Manager) SaveSchemeToUser(scheme *Scheme) error {
	// Get the first user path (default)
	userPaths := m.getUserSchemePaths()
	if len(userPaths) == 0 {
		return fmt.Errorf("no user scheme paths configured")
	}

	userPath := userPaths[0]
	schemePath := filepath.Join(userPath, scheme.Name, scheme.Flavour)

	// Ensure directory exists
	if err := paths.EnsureDir(schemePath); err != nil {
		return fmt.Errorf("failed to create user scheme directory: %w", err)
	}

	// Safety-net: Ensure source is set to user when saving to user directory
	if scheme.Source == "" {
		scheme.Source = SourceUser
	}

	// Write scheme file
	filePath := filepath.Join(schemePath, scheme.Mode+".json")
	if err := paths.AtomicWriteJSON(filePath, scheme); err != nil {
		return fmt.Errorf("failed to write user scheme file: %w", err)
	}

	return nil
}

// GetColors returns the colors of a scheme as a simple string map
func (s *Scheme) GetColors() map[string]string {
	return s.Colours
}

// GetSchemeSource determines the source of a scheme by checking where it exists
func (m *Manager) GetSchemeSource(schemeName string) SchemeSource {
	// Check user paths first (highest priority)
	userPaths := m.getUserSchemePaths()
	for _, userPath := range userPaths {
		schemePath := filepath.Join(userPath, schemeName)
		if paths.IsDir(schemePath) {
			return SourceUser
		}
	}

	// Check data directory (generated schemes location)
	schemePath := filepath.Join(m.schemesDir, schemeName)
	if paths.IsDir(schemePath) {
		// If it's in the data directory, it's a generated scheme
		// The data directory is specifically for generated schemes
		return SourceGenerated
	}

	// Check if it exists in embedded assets
	entries, err := fs.ReadDir(schemes.Content, ".")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == schemeName {
				return SourceBundled
			}
		}
	}

	// Default to bundled if not found (shouldn't happen)
	return SourceBundled
}

// getDefaultColours returns default catppuccin mocha colors
func getDefaultColours() map[string]string {
	return map[string]string{
		"base":      "1e1e2e",
		"mantle":    "181825",
		"crust":     "11111b",
		"text":      "cdd6f4",
		"subtext0":  "a6adc8",
		"subtext1":  "bac2de",
		"surface0":  "313244",
		"surface1":  "45475a",
		"surface2":  "585b70",
		"overlay0":  "6c7086",
		"overlay1":  "7f849c",
		"overlay2":  "9399b2",
		"blue":      "89b4fa",
		"lavender":  "b4befe",
		"sapphire":  "74c7ec",
		"sky":       "89dceb",
		"teal":      "94e2d5",
		"green":     "a6e3a1",
		"yellow":    "f9e2af",
		"peach":     "fab387",
		"maroon":    "eba0ac",
		"red":       "f38ba8",
		"mauve":     "cba6f7",
		"pink":      "f5c2e7",
		"flamingo":  "f2cdcd",
		"rosewater": "f5e0dc",
	}
}
