package scheme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/arthur404dev/heimdall-cli/assets/schemes"
)

// BundledScheme represents a bundled color scheme
type BundledScheme struct {
	Name    string            `json:"name"`
	Author  string            `json:"author"`
	Variant string            `json:"variant"`
	Colors  map[string]string `json:"colors"`
	Family  string            `json:"-"` // e.g., catppuccin, gruvbox
	Flavour string            `json:"-"` // e.g., mocha, dark
}

// GetBundledSchemes returns all bundled color schemes from embedded assets
func GetBundledSchemes() ([]BundledScheme, error) {
	var schemeList []BundledScheme

	// Walk through the embedded filesystem
	err := fs.WalkDir(schemes.Content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process .json files
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			data, err := schemes.Content.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read embedded scheme %s: %w", path, err)
			}

			var scheme BundledScheme
			if err := json.Unmarshal(data, &scheme); err != nil {
				return fmt.Errorf("failed to unmarshal scheme %s: %w", path, err)
			}

			// Extract family and flavour from path
			// e.g., catppuccin/mocha/dark.json
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				scheme.Family = parts[0]  // catppuccin
				scheme.Flavour = parts[1] // mocha
			}

			schemeList = append(schemeList, scheme)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk embedded schemes: %w", err)
	}

	return schemeList, nil
}

// GetBundledScheme returns a specific bundled scheme by name
func GetBundledScheme(name string) (*BundledScheme, error) {
	schemes, err := GetBundledSchemes()
	if err != nil {
		return nil, err
	}

	for _, scheme := range schemes {
		// Check if the name matches (case-insensitive)
		if strings.EqualFold(scheme.Name, name) {
			return &scheme, nil
		}

		// Also check if it matches the family/flavour pattern
		// e.g., "catppuccin-mocha" or "gruvbox-dark"
		fullName := fmt.Sprintf("%s-%s", scheme.Family, scheme.Flavour)
		if strings.EqualFold(fullName, name) {
			return &scheme, nil
		}

		// Also check family/flavour separated by space
		fullNameSpace := fmt.Sprintf("%s %s", scheme.Family, scheme.Flavour)
		if strings.EqualFold(fullNameSpace, name) {
			return &scheme, nil
		}
	}

	return nil, fmt.Errorf("bundled scheme '%s' not found", name)
}

// ListBundledSchemeNames returns the names of all bundled schemes
func ListBundledSchemeNames() ([]string, error) {
	schemes, err := GetBundledSchemes()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(schemes))
	for _, scheme := range schemes {
		names = append(names, scheme.Name)
	}

	return names, nil
}

// InstallBundledScheme installs a bundled scheme to the user's scheme directory
func (m *Manager) InstallBundledScheme(name string) error {
	bundled, err := GetBundledScheme(name)
	if err != nil {
		return err
	}

	// Convert BundledScheme to Scheme
	scheme := &Scheme{
		Name:    bundled.Family,
		Flavour: bundled.Flavour,
		Mode:    bundled.Variant, // dark or light
		Variant: bundled.Variant,
		Colours: make(map[string]string),
	}

	// Convert color strings (remove # prefix if present)
	for key, hexColor := range bundled.Colors {
		// Remove # prefix if present
		hexColor = strings.TrimPrefix(hexColor, "#")
		scheme.Colours[key] = hexColor
	}

	// Save the scheme
	return m.SaveScheme(scheme)
}

// InstallBundledSchemeToUser installs a bundled scheme to the user scheme directory
func (m *Manager) InstallBundledSchemeToUser(name string) error {
	bundled, err := GetBundledScheme(name)
	if err != nil {
		return err
	}

	// Convert BundledScheme to Scheme
	scheme := &Scheme{
		Name:    bundled.Family,
		Flavour: bundled.Flavour,
		Mode:    bundled.Variant, // dark or light
		Variant: bundled.Variant,
		Colours: make(map[string]string),
	}

	// Convert color strings (remove # prefix if present)
	for key, hexColor := range bundled.Colors {
		// Remove # prefix if present
		hexColor = strings.TrimPrefix(hexColor, "#")
		scheme.Colours[key] = hexColor
	}

	// Save the scheme to user directory
	return m.SaveSchemeToUser(scheme)
}

// InstallAllBundledSchemes installs all bundled schemes
func (m *Manager) InstallAllBundledSchemes() error {
	schemes, err := GetBundledSchemes()
	if err != nil {
		return err
	}

	for _, scheme := range schemes {
		if err := m.InstallBundledScheme(scheme.Name); err != nil {
			return fmt.Errorf("failed to install scheme %s: %w", scheme.Name, err)
		}
	}

	return nil
}

// InstallAllBundledSchemesToUser installs all bundled schemes to user directory
func (m *Manager) InstallAllBundledSchemesToUser() error {
	schemes, err := GetBundledSchemes()
	if err != nil {
		return err
	}

	for _, scheme := range schemes {
		if err := m.InstallBundledSchemeToUser(scheme.Name); err != nil {
			return fmt.Errorf("failed to install scheme %s: %w", scheme.Name, err)
		}
	}

	return nil
}

// ListAllSchemes returns both user schemes and bundled schemes
func (m *Manager) ListAllSchemes() ([]string, error) {
	// Get user schemes
	userSchemes, err := m.ListSchemes()
	if err != nil {
		return nil, err
	}

	// Get bundled scheme names
	bundledNames, err := ListBundledSchemeNames()
	if err != nil {
		// If no bundled schemes, just return user schemes
		return userSchemes, nil
	}

	// Combine and deduplicate
	schemeMap := make(map[string]bool)
	for _, s := range userSchemes {
		schemeMap[s] = true
	}
	for _, s := range bundledNames {
		schemeMap[fmt.Sprintf("[bundled] %s", s)] = true
	}

	// Convert back to slice
	allSchemes := make([]string, 0, len(schemeMap))
	for s := range schemeMap {
		allSchemes = append(allSchemes, s)
	}

	return allSchemes, nil
}

// LoadSchemeWithFallback tries to load a user scheme first, then falls back to bundled
func (m *Manager) LoadSchemeWithFallback(name, flavour, mode string) (*Scheme, error) {
	// Try to load user scheme first
	scheme, err := m.LoadScheme(name, flavour, mode)
	if err == nil {
		return scheme, nil
	}

	// Try bundled scheme - construct possible names
	possibleNames := []string{
		name,                                // Direct name match
		fmt.Sprintf("%s-%s", name, flavour), // family-flavour pattern
		fmt.Sprintf("%s %s", name, flavour), // family flavour with space
	}

	// Add title case versions
	if len(name) > 0 && len(flavour) > 0 {
		titleName := strings.ToUpper(name[:1]) + name[1:]
		titleFlavour := strings.ToUpper(flavour[:1]) + flavour[1:]
		possibleNames = append(possibleNames,
			fmt.Sprintf("%s %s", titleName, titleFlavour),
			fmt.Sprintf("%s-%s", titleName, titleFlavour),
		)
	}

	for _, schemeName := range possibleNames {
		bundled, err := GetBundledScheme(schemeName)
		if err != nil {
			continue
		}

		// Convert to Scheme
		scheme := &Scheme{
			Name:    name,
			Flavour: flavour,
			Mode:    mode,
			Variant: bundled.Variant,
			Colours: make(map[string]string),
		}

		// Convert colors (remove # prefix if present)
		for key, hexColor := range bundled.Colors {
			// Remove # prefix if present
			hexColor = strings.TrimPrefix(hexColor, "#")
			scheme.Colours[key] = hexColor
		}

		return scheme, nil
	}

	return nil, fmt.Errorf("scheme %s/%s/%s not found in user or bundled schemes", name, flavour, mode)
}
