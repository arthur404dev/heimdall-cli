package scheme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heimdall-cli/heimdall/internal/utils/color"
	"github.com/heimdall-cli/heimdall/internal/utils/paths"
)

// Scheme represents a color scheme
type Scheme struct {
	Name     string                  `json:"name"`
	Flavour  string                  `json:"flavour"`
	Mode     string                  `json:"mode"`
	Variant  string                  `json:"variant"`
	Colors   map[string]*color.Color `json:"colors"`
	Metadata SchemeMetadata          `json:"metadata,omitempty"`
}

// SchemeMetadata contains metadata about a scheme
type SchemeMetadata struct {
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
	Generated   bool   `json:"generated,omitempty"`
}

// Manager manages color schemes
type Manager struct {
	dataDir  string
	cacheDir string
	stateDir string
}

// NewManager creates a new scheme manager
func NewManager() *Manager {
	return &Manager{
		dataDir:  paths.SchemeDataDir,
		cacheDir: paths.SchemeCacheDir,
		stateDir: paths.StateDir,
	}
}

// GetCurrent returns the current active scheme
func (m *Manager) GetCurrent() (*Scheme, error) {
	statePath := filepath.Join(m.stateDir, "current_scheme.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no scheme currently set")
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
	statePath := filepath.Join(m.stateDir, "current_scheme.json")

	// Ensure state directory exists
	if err := paths.EnsureDir(m.stateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal scheme to JSON
	data, err := json.MarshalIndent(scheme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scheme: %w", err)
	}

	// Write to state file
	if err := paths.AtomicWrite(statePath, data); err != nil {
		return fmt.Errorf("failed to write scheme state: %w", err)
	}

	return nil
}

// ListSchemes returns a list of available scheme names
func (m *Manager) ListSchemes() ([]string, error) {
	schemes := make(map[string]bool)

	// List schemes from data directory
	dataPath := filepath.Join(m.dataDir, "schemes")
	if entries, err := os.ReadDir(dataPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				schemes[entry.Name()] = true
			}
		}
	}

	// List schemes from cache directory (generated schemes)
	cachePath := filepath.Join(m.cacheDir, "schemes")
	if entries, err := os.ReadDir(cachePath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				schemes[entry.Name()] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(schemes))
	for name := range schemes {
		result = append(result, name)
	}

	return result, nil
}

// ListFlavours returns available flavours for a scheme
func (m *Manager) ListFlavours(schemeName string) ([]string, error) {
	flavours := make(map[string]bool)

	// Check data directory
	dataPath := filepath.Join(m.dataDir, "schemes", schemeName)
	if entries, err := os.ReadDir(dataPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				flavours[entry.Name()] = true
			}
		}
	}

	// Check cache directory
	cachePath := filepath.Join(m.cacheDir, "schemes", schemeName)
	if entries, err := os.ReadDir(cachePath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				flavours[entry.Name()] = true
			}
		}
	}

	if len(flavours) == 0 {
		return nil, fmt.Errorf("scheme %s not found", schemeName)
	}

	// Convert map to slice
	result := make([]string, 0, len(flavours))
	for flavour := range flavours {
		result = append(result, flavour)
	}

	return result, nil
}

// ListModes returns available modes for a scheme flavour
func (m *Manager) ListModes(schemeName, flavour string) ([]string, error) {
	modes := make(map[string]bool)

	// Check data directory
	dataPath := filepath.Join(m.dataDir, "schemes", schemeName, flavour)
	if entries, err := os.ReadDir(dataPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				mode := strings.TrimSuffix(entry.Name(), ".json")
				modes[mode] = true
			}
		}
	}

	// Check cache directory
	cachePath := filepath.Join(m.cacheDir, "schemes", schemeName, flavour)
	if entries, err := os.ReadDir(cachePath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				mode := strings.TrimSuffix(entry.Name(), ".json")
				modes[mode] = true
			}
		}
	}

	if len(modes) == 0 {
		return nil, fmt.Errorf("flavour %s/%s not found", schemeName, flavour)
	}

	// Convert map to slice
	result := make([]string, 0, len(modes))
	for mode := range modes {
		result = append(result, mode)
	}

	return result, nil
}

// LoadScheme loads a specific scheme
func (m *Manager) LoadScheme(name, flavour, mode string) (*Scheme, error) {
	// Try data directory first
	dataPath := filepath.Join(m.dataDir, "schemes", name, flavour, mode+".json")
	if data, err := os.ReadFile(dataPath); err == nil {
		return m.parseScheme(data, name, flavour, mode)
	}

	// Try cache directory
	cachePath := filepath.Join(m.cacheDir, "schemes", name, flavour, mode+".json")
	if data, err := os.ReadFile(cachePath); err == nil {
		return m.parseScheme(data, name, flavour, mode)
	}

	return nil, fmt.Errorf("scheme %s/%s/%s not found", name, flavour, mode)
}

// parseScheme parses scheme data from JSON
func (m *Manager) parseScheme(data []byte, name, flavour, mode string) (*Scheme, error) {
	var rawScheme map[string]interface{}
	if err := json.Unmarshal(data, &rawScheme); err != nil {
		return nil, fmt.Errorf("failed to parse scheme JSON: %w", err)
	}

	scheme := &Scheme{
		Name:    name,
		Flavour: flavour,
		Mode:    mode,
		Colors:  make(map[string]*color.Color),
	}

	// Parse colors
	if colorsRaw, ok := rawScheme["colors"].(map[string]interface{}); ok {
		for key, value := range colorsRaw {
			if colorStr, ok := value.(string); ok {
				if c, err := color.NewFromHex(colorStr); err == nil {
					scheme.Colors[key] = c
				}
			}
		}
	}

	// Parse metadata if present
	if metaRaw, ok := rawScheme["metadata"].(map[string]interface{}); ok {
		if author, ok := metaRaw["author"].(string); ok {
			scheme.Metadata.Author = author
		}
		if desc, ok := metaRaw["description"].(string); ok {
			scheme.Metadata.Description = desc
		}
		if source, ok := metaRaw["source"].(string); ok {
			scheme.Metadata.Source = source
		}
		if generated, ok := metaRaw["generated"].(bool); ok {
			scheme.Metadata.Generated = generated
		}
	}

	// Parse variant if present
	if variant, ok := rawScheme["variant"].(string); ok {
		scheme.Variant = variant
	}

	return scheme, nil
}

// SaveScheme saves a scheme to the cache directory
func (m *Manager) SaveScheme(scheme *Scheme) error {
	// Create path in cache directory
	schemePath := filepath.Join(m.cacheDir, "schemes", scheme.Name, scheme.Flavour)

	// Ensure directory exists
	if err := paths.EnsureDir(schemePath); err != nil {
		return fmt.Errorf("failed to create scheme directory: %w", err)
	}

	// Prepare data for JSON
	data := map[string]interface{}{
		"name":     scheme.Name,
		"flavour":  scheme.Flavour,
		"mode":     scheme.Mode,
		"variant":  scheme.Variant,
		"metadata": scheme.Metadata,
		"colors":   make(map[string]string),
	}

	// Convert colors to hex strings
	for key, c := range scheme.Colors {
		data["colors"].(map[string]string)[key] = c.Hex
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scheme: %w", err)
	}

	// Write to file
	filePath := filepath.Join(schemePath, scheme.Mode+".json")
	if err := paths.AtomicWrite(filePath, jsonData); err != nil {
		return fmt.Errorf("failed to write scheme file: %w", err)
	}

	return nil
}

// GenerateFromWallpaper generates a Material You scheme from a wallpaper
func (m *Manager) GenerateFromWallpaper(wallpaperPath string) (*Scheme, error) {
	// This would use the Material You generator
	// For now, return a placeholder
	return nil, fmt.Errorf("wallpaper generation not yet implemented")
}

// GetSchemeColors returns the colors of a scheme as a simple string map
func (s *Scheme) GetColors() map[string]string {
	colors := make(map[string]string)
	for key, c := range s.Colors {
		colors[key] = c.Hex
	}
	return colors
}
