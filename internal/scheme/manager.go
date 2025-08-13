package scheme

import (
	"encoding/json"
	"fmt"
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

// SetScheme sets the active scheme
func (m *Manager) SetScheme(scheme *Scheme) error {
	statePath := filepath.Join(m.stateDir, "scheme.json")

	// Ensure state directory exists
	if err := paths.EnsureDir(m.stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write scheme directly using atomic write
	if err := paths.AtomicWriteJSON(statePath, scheme); err != nil {
		return fmt.Errorf("failed to write scheme state: %w", err)
	}

	return nil
}

// ListSchemes returns available scheme names from the schemes directory
func (m *Manager) ListSchemes() ([]string, error) {
	// Try multiple locations for scheme directories
	schemeDirs := []string{
		m.schemesDir, // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes"), // Legacy cache location with extra "schemes" level
		paths.SchemeCacheDir,                           // Direct cache location
	}

	schemeMap := make(map[string]bool) // Use map to avoid duplicates

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
	var schemes []string
	for scheme := range schemeMap {
		schemes = append(schemes, scheme)
	}

	return schemes, nil
}

// ListFlavours returns available flavours for a scheme
func (m *Manager) ListFlavours(schemeName string) ([]string, error) {
	// Try multiple locations for scheme directories
	schemePaths := []string{
		filepath.Join(m.schemesDir, schemeName),                    // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes", schemeName), // Legacy cache location with extra "schemes" level
		filepath.Join(paths.SchemeCacheDir, schemeName),            // Direct cache location
	}

	flavourMap := make(map[string]bool) // Use map to avoid duplicates

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
	// Try multiple locations for flavour directories
	flavourPaths := []string{
		filepath.Join(m.schemesDir, schemeName, flavour),                    // Primary location (data dir)
		filepath.Join(paths.SchemeCacheDir, "schemes", schemeName, flavour), // Legacy cache location with extra "schemes" level
		filepath.Join(paths.SchemeCacheDir, schemeName, flavour),            // Direct cache location
	}

	modeMap := make(map[string]bool) // Use map to avoid duplicates

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
