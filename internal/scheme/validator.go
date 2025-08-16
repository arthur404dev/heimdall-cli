package scheme

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RequiredColorKeys defines the core color keys that should be present
// Note: We're lenient with validation to allow various scheme formats
var RequiredColorKeys = []string{
	// Core colors - these are essential
	"background",
	"foreground",

	// Base16 colors - commonly used but not strictly required for user schemes
	// We'll validate these exist OR that the scheme has sufficient other colors
}

// OptionalColorKeys defines additional color keys that may be present
var OptionalColorKeys = []string{
	"base10", // Extended colors for additional customization
	"base11",
	"base12",
	"base13",
	"base14",
	"base15",
	"base16",
	"base17",
}

// ValidationError represents a scheme validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors:\n  - %s", strings.Join(messages, "\n  - "))
}

// ValidateScheme validates a scheme's structure and content
func ValidateScheme(scheme *Scheme) error {
	var errors ValidationErrors

	// Check basic fields
	if scheme.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "scheme name is required",
		})
	}

	// Check colors map exists
	if scheme.Colours == nil {
		errors = append(errors, ValidationError{
			Field:   "colours",
			Message: "colors map is required",
		})
		return errors // Can't check color keys if map is nil
	}

	// For user schemes, we're more lenient
	// Just check that they have a reasonable number of colors
	if len(scheme.Colours) < 8 {
		errors = append(errors, ValidationError{
			Field:   "colours",
			Message: fmt.Sprintf("insufficient colors: found %d, minimum 8 required", len(scheme.Colours)),
		})
	}

	// Check for essential colors (more lenient than before)
	essentialKeys := []string{"background", "foreground"}
	missingKeys := []string{}
	for _, key := range essentialKeys {
		if _, exists := scheme.Colours[key]; !exists {
			// Also check for base00/base05 as alternatives
			if key == "background" && scheme.Colours["base00"] == "" && scheme.Colours["base"] == "" {
				missingKeys = append(missingKeys, key)
			} else if key == "foreground" && scheme.Colours["base05"] == "" && scheme.Colours["text"] == "" {
				missingKeys = append(missingKeys, key)
			}
		}
	}

	if len(missingKeys) > 0 {
		errors = append(errors, ValidationError{
			Field:   "colours",
			Message: fmt.Sprintf("missing essential color keys: %s", strings.Join(missingKeys, ", ")),
		})
	}

	// Validate color format (should be hex colors)
	for key, value := range scheme.Colours {
		if !isValidHexColor(value) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("colours.%s", key),
				Message: fmt.Sprintf("invalid hex color format: %s", value),
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateJSON validates that a byte slice contains valid JSON for a scheme
func ValidateJSON(data []byte) (*Scheme, error) {
	var scheme Scheme

	// Try to unmarshal
	if err := json.Unmarshal(data, &scheme); err != nil {
		// Provide more helpful error message
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			return nil, ValidationError{
				Field:   "json",
				Message: fmt.Sprintf("invalid JSON syntax at position %d: %v", syntaxErr.Offset, err),
			}
		}
		if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
			return nil, ValidationError{
				Field: typeErr.Field,
				Message: fmt.Sprintf("invalid type for field %s: expected %s, got %s",
					typeErr.Field, typeErr.Type, typeErr.Value),
			}
		}
		return nil, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
		}
	}

	// Validate the parsed scheme
	if err := ValidateScheme(&scheme); err != nil {
		return nil, err
	}

	return &scheme, nil
}

// isValidHexColor checks if a string is a valid hex color
func isValidHexColor(color string) bool {
	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// Check length (3, 6, or 8 characters for RGB, RRGGBB, or RRGGBBAA)
	if len(color) != 3 && len(color) != 6 && len(color) != 8 {
		return false
	}

	// Check if all characters are valid hex
	for _, c := range color {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// MigrateOldFormat attempts to convert old scheme formats to the new JSON format
func MigrateOldFormat(data []byte, format string) (*Scheme, error) {
	switch format {
	case "txt", "text":
		return migrateTextFormat(data)
	case "yaml", "yml":
		return migrateYAMLFormat(data)
	case "toml":
		return migrateTOMLFormat(data)
	default:
		return nil, fmt.Errorf("unsupported format for migration: %s", format)
	}
}

// migrateTextFormat converts old text-based color schemes to JSON
func migrateTextFormat(data []byte) (*Scheme, error) {
	scheme := &Scheme{
		Name:    "Migrated Scheme",
		Colours: make(map[string]string),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Try to parse various formats
		// Format 1: base00=#1a1a1a
		// Format 2: base00 #1a1a1a
		// Format 3: base00: #1a1a1a

		parts := strings.Fields(line)
		if len(parts) < 2 {
			// Try splitting by = or :
			if strings.Contains(line, "=") {
				parts = strings.SplitN(line, "=", 2)
			} else if strings.Contains(line, ":") {
				parts = strings.SplitN(line, ":", 2)
			}
		}

		if len(parts) >= 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Ensure the value has a # prefix
			if !strings.HasPrefix(value, "#") {
				value = "#" + value
			}

			// Only add if it's a base color key
			if strings.HasPrefix(key, "base") {
				scheme.Colours[key] = value
			}
		}
	}

	// Validate the migrated scheme
	if err := ValidateScheme(scheme); err != nil {
		return nil, fmt.Errorf("migrated scheme validation failed: %w", err)
	}

	return scheme, nil
}

// migrateYAMLFormat converts YAML schemes to JSON (stub for now)
func migrateYAMLFormat(data []byte) (*Scheme, error) {
	// This would require a YAML parser
	// For now, return an error
	return nil, fmt.Errorf("YAML migration not yet implemented")
}

// migrateTOMLFormat converts TOML schemes to JSON (stub for now)
func migrateTOMLFormat(data []byte) (*Scheme, error) {
	// This would require a TOML parser
	// For now, return an error
	return nil, fmt.Errorf("TOML migration not yet implemented")
}

// SanitizeScheme ensures a scheme has valid values and fills in defaults
func SanitizeScheme(scheme *Scheme) {
	if scheme.Colours == nil {
		scheme.Colours = make(map[string]string)
	}

	// Ensure all color values have # prefix
	for key, value := range scheme.Colours {
		if !strings.HasPrefix(value, "#") {
			scheme.Colours[key] = "#" + value
		}
	}

	// Set default name if empty
	if scheme.Name == "" {
		scheme.Name = "Unnamed Scheme"
	}

	// Set default flavour/mode if empty
	if scheme.Flavour == "" {
		scheme.Flavour = "default"
	}
	if scheme.Mode == "" {
		scheme.Mode = "dark"
	}
}
