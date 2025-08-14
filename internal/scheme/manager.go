package scheme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/assets/schemes"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Scheme represents a color scheme (aligned with caelestia format)
type Scheme struct {
	Name    string            `json:"name"`
	Flavour string            `json:"flavour"`
	Mode    string            `json:"mode"`
	Variant string            `json:"variant"`
	Colours map[string]string `json:"colours"` // British spelling, simple strings
}

// Manager manages color schemes with caelestia's simple approach
type Manager struct {
	schemesDir string
	stateDir   string
}

// NewManager creates a new scheme manager
func NewManager() *Manager {
	return &Manager{
		schemesDir: paths.SchemeDataDir,
		stateDir:   paths.StateDir,
	}
}

// GetCurrent returns the current active scheme
func (m *Manager) GetCurrent() (*Scheme, error) {
	statePath := filepath.Join(m.stateDir, "scheme.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default scheme like caelestia
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

	return &scheme, nil
}

// SetScheme sets the active scheme with triple-write for QuickShell integration
func (m *Manager) SetScheme(scheme *Scheme) error {
	// Prepare Heimdall format data (with # prefix on colors)
	heimdallScheme := &Scheme{
		Name:    scheme.Name,
		Flavour: scheme.Flavour,
		Mode:    scheme.Mode,
		Variant: scheme.Variant,
		Colours: scheme.Colours, // Already has # prefix
	}

	// 1. Primary write to Heimdall config location
	configPath := filepath.Join(paths.ConfigDir, "scheme.json")
	if err := paths.EnsureDir(paths.ConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := paths.AtomicWriteJSON(configPath, heimdallScheme); err != nil {
		return fmt.Errorf("failed to write config scheme: %w", err)
	}

	// 2. Secondary write to Heimdall state location (matching Caelestia pattern)
	statePath := filepath.Join(m.stateDir, "scheme.json")
	if err := paths.EnsureDir(m.stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	if err := paths.AtomicWriteJSON(statePath, heimdallScheme); err != nil {
		return fmt.Errorf("failed to write state scheme: %w", err)
	}

	// 3. CRITICAL: QuickShell-specific format (no # prefix, "colours" key)
	// This bridges the gap that Caelestia missed!
	quickshellScheme := m.prepareQuickShellFormat(scheme)
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
		fmt.Fprintf(os.Stderr, "Info: Updated QuickShell colors (bridging Caelestia gap) at %s\n", quickshellPath)
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

	// First, add bundled schemes from embedded assets
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

	// Then, add user schemes from filesystem directories
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

	// First check embedded assets
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

	// First check embedded assets
	embeddedPath := filepath.Join(schemeName, flavour)
	entries, err := fs.ReadDir(schemes.Content, embeddedPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
				mode := strings.TrimSuffix(entry.Name(), ".txt")
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
	// Try multiple locations for scheme files
	schemePaths := []string{
		// Primary location (data dir)
		filepath.Join(m.schemesDir, name, flavour, mode+".json"),
		// Legacy cache location with extra "schemes" level
		filepath.Join(paths.SchemeCacheDir, "schemes", name, flavour, mode+".json"),
		// Direct cache location
		filepath.Join(paths.SchemeCacheDir, name, flavour, mode+".json"),
	}

	var data []byte
	var err error

	// First try loading from disk (existing behavior)
	for _, schemePath := range schemePaths {
		data, err = os.ReadFile(schemePath)
		if err == nil {
			// Parse JSON format from disk
			scheme := &Scheme{
				Name:    name,
				Flavour: flavour,
				Mode:    mode,
				Colours: make(map[string]string),
			}

			// Parse JSON directly into a map to handle flexible formats
			var rawData map[string]interface{}
			if err := json.Unmarshal(data, &rawData); err != nil {
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

			return scheme, nil
		}
	}

	// If not found on disk, try loading from embedded assets
	embeddedPath := fmt.Sprintf("%s/%s/%s.txt", name, flavour, mode)
	embeddedData, err := schemes.Content.ReadFile(embeddedPath)
	if err == nil {
		// Parse embedded .txt format (space-separated key-value pairs)
		scheme := &Scheme{
			Name:    name,
			Flavour: flavour,
			Mode:    mode,
			Colours: make(map[string]string),
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

	// Write scheme file
	filePath := filepath.Join(schemePath, scheme.Mode+".json")
	if err := paths.AtomicWriteJSON(filePath, scheme); err != nil {
		return fmt.Errorf("failed to write scheme file: %w", err)
	}

	return nil
}

// GetColors returns the colors of a scheme as a simple string map
func (s *Scheme) GetColors() map[string]string {
	return s.Colours
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
