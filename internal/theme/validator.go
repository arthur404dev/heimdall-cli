package theme

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates color schemes and theme configurations
type Validator struct {
	hexPattern     *regexp.Regexp
	rgbPattern     *regexp.Regexp
	rgbaPattern    *regexp.Regexp
	hslPattern     *regexp.Regexp
	hslaPattern    *regexp.Regexp
	requiredColors []string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		hexPattern:  regexp.MustCompile(`^#?([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`),
		rgbPattern:  regexp.MustCompile(`^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$`),
		rgbaPattern: regexp.MustCompile(`^rgba\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(0?\.\d+|1\.0|0|1)\s*\)$`),
		hslPattern:  regexp.MustCompile(`^hsl\(\s*(\d{1,3})\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*\)$`),
		hslaPattern: regexp.MustCompile(`^hsla\(\s*(\d{1,3})\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*,\s*(0?\.\d+|1\.0|0|1)\s*\)$`),
		requiredColors: []string{
			"background",
			"foreground",
			"colour0", "colour1", "colour2", "colour3",
			"colour4", "colour5", "colour6", "colour7",
			"colour8", "colour9", "colour10", "colour11",
			"colour12", "colour13", "colour14", "colour15",
		},
	}
}

// ValidateScheme validates a complete color scheme
func (v *Validator) ValidateScheme(scheme *ColorScheme) error {
	if scheme == nil {
		return fmt.Errorf("scheme is nil")
	}

	// Validate scheme metadata
	if scheme.Name == "" {
		return fmt.Errorf("scheme name is required")
	}

	if scheme.Mode != "dark" && scheme.Mode != "light" {
		return fmt.Errorf("scheme mode must be 'dark' or 'light', got '%s'", scheme.Mode)
	}

	// Validate required colors
	for _, colorKey := range v.requiredColors {
		if color, exists := scheme.Colors[colorKey]; !exists {
			return fmt.Errorf("required color '%s' is missing", colorKey)
		} else if err := v.ValidateColor(color); err != nil {
			return fmt.Errorf("invalid color for '%s': %w", colorKey, err)
		}
	}

	// Validate special colors if present
	for key, color := range scheme.Special {
		if err := v.ValidateColor(color); err != nil {
			return fmt.Errorf("invalid special color for '%s': %w", key, err)
		}
	}

	return nil
}

// ValidateColor validates a single color value
func (v *Validator) ValidateColor(color string) error {
	if color == "" {
		return fmt.Errorf("color is empty")
	}

	// Check various color formats
	if v.IsHexColor(color) {
		return v.ValidateHexColor(color)
	}

	if v.IsRGBColor(color) {
		return v.ValidateRGBColor(color)
	}

	if v.IsRGBAColor(color) {
		return v.ValidateRGBAColor(color)
	}

	if v.IsHSLColor(color) {
		return v.ValidateHSLColor(color)
	}

	if v.IsHSLAColor(color) {
		return v.ValidateHSLAColor(color)
	}

	return fmt.Errorf("unrecognized color format: %s", color)
}

// IsHexColor checks if a color is in hex format
func (v *Validator) IsHexColor(color string) bool {
	return v.hexPattern.MatchString(color)
}

// ValidateHexColor validates a hex color
func (v *Validator) ValidateHexColor(color string) error {
	if !v.hexPattern.MatchString(color) {
		return fmt.Errorf("invalid hex color format: %s", color)
	}
	return nil
}

// IsRGBColor checks if a color is in RGB format
func (v *Validator) IsRGBColor(color string) bool {
	return v.rgbPattern.MatchString(color)
}

// ValidateRGBColor validates an RGB color
func (v *Validator) ValidateRGBColor(color string) error {
	matches := v.rgbPattern.FindStringSubmatch(color)
	if matches == nil {
		return fmt.Errorf("invalid RGB color format: %s", color)
	}

	// Validate RGB values are in range 0-255
	for i := 1; i <= 3; i++ {
		var val int
		fmt.Sscanf(matches[i], "%d", &val)
		if val < 0 || val > 255 {
			return fmt.Errorf("RGB value out of range (0-255): %d", val)
		}
	}

	return nil
}

// IsRGBAColor checks if a color is in RGBA format
func (v *Validator) IsRGBAColor(color string) bool {
	return v.rgbaPattern.MatchString(color)
}

// ValidateRGBAColor validates an RGBA color
func (v *Validator) ValidateRGBAColor(color string) error {
	matches := v.rgbaPattern.FindStringSubmatch(color)
	if matches == nil {
		return fmt.Errorf("invalid RGBA color format: %s", color)
	}

	// Validate RGB values are in range 0-255
	for i := 1; i <= 3; i++ {
		var val int
		fmt.Sscanf(matches[i], "%d", &val)
		if val < 0 || val > 255 {
			return fmt.Errorf("RGBA value out of range (0-255): %d", val)
		}
	}

	// Alpha is already validated by regex (0-1)
	return nil
}

// IsHSLColor checks if a color is in HSL format
func (v *Validator) IsHSLColor(color string) bool {
	return v.hslPattern.MatchString(color)
}

// ValidateHSLColor validates an HSL color
func (v *Validator) ValidateHSLColor(color string) error {
	matches := v.hslPattern.FindStringSubmatch(color)
	if matches == nil {
		return fmt.Errorf("invalid HSL color format: %s", color)
	}

	// Validate hue (0-360)
	var hue int
	fmt.Sscanf(matches[1], "%d", &hue)
	if hue < 0 || hue > 360 {
		return fmt.Errorf("HSL hue out of range (0-360): %d", hue)
	}

	// Validate saturation and lightness (0-100)
	for i := 2; i <= 3; i++ {
		var val int
		fmt.Sscanf(matches[i], "%d", &val)
		if val < 0 || val > 100 {
			return fmt.Errorf("HSL value out of range (0-100): %d", val)
		}
	}

	return nil
}

// IsHSLAColor checks if a color is in HSLA format
func (v *Validator) IsHSLAColor(color string) bool {
	return v.hslaPattern.MatchString(color)
}

// ValidateHSLAColor validates an HSLA color
func (v *Validator) ValidateHSLAColor(color string) error {
	matches := v.hslaPattern.FindStringSubmatch(color)
	if matches == nil {
		return fmt.Errorf("invalid HSLA color format: %s", color)
	}

	// Validate hue (0-360)
	var hue int
	fmt.Sscanf(matches[1], "%d", &hue)
	if hue < 0 || hue > 360 {
		return fmt.Errorf("HSLA hue out of range (0-360): %d", hue)
	}

	// Validate saturation and lightness (0-100)
	for i := 2; i <= 3; i++ {
		var val int
		fmt.Sscanf(matches[i], "%d", &val)
		if val < 0 || val > 100 {
			return fmt.Errorf("HSLA value out of range (0-100): %d", val)
		}
	}

	// Alpha is already validated by regex (0-1)
	return nil
}

// ValidateTemplate validates a template string
func (v *Validator) ValidateTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("template is empty")
	}

	// Check for basic template syntax errors
	openCount := strings.Count(template, "{{")
	closeCount := strings.Count(template, "}}")

	if openCount != closeCount {
		return fmt.Errorf("mismatched template delimiters: %d '{{' vs %d '}}'", openCount, closeCount)
	}

	// Check for empty placeholders
	if strings.Contains(template, "{{}}") {
		return fmt.Errorf("empty template placeholder found")
	}

	return nil
}

// ValidateApplicationConfig validates configuration for a specific application
func (v *Validator) ValidateApplicationConfig(app string, config map[string]interface{}) error {
	if app == "" {
		return fmt.Errorf("application name is empty")
	}

	if config == nil {
		return fmt.Errorf("configuration is nil for %s", app)
	}

	// Application-specific validation can be added here
	switch app {
	case "discord":
		// Discord-specific validation
		if clients, ok := config["clients"].([]string); ok {
			if len(clients) == 0 {
				return fmt.Errorf("no Discord clients specified")
			}
		}
	case "terminal":
		// Terminal-specific validation
		// Check for required terminal colors
	case "gtk":
		// GTK-specific validation
	case "qt":
		// Qt-specific validation
	}

	return nil
}

// SuggestFix suggests a fix for common validation errors
func (v *Validator) SuggestFix(err error) string {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "hex color"):
		return "Hex colors should be in format #RRGGBB or #RGB"
	case strings.Contains(errStr, "RGB"):
		return "RGB colors should be in format rgb(r, g, b) where r,g,b are 0-255"
	case strings.Contains(errStr, "required color"):
		return "Ensure all 16 terminal colors (colour0-colour15) plus background and foreground are defined"
	case strings.Contains(errStr, "template"):
		return "Check that all {{ and }} are properly matched and not empty"
	default:
		return ""
	}
}
