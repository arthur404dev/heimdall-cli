package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorMapper maps color schemes to application-specific formats
type colorMapper struct {
	validator *Validator
}

// NewColorMapper creates a new color mapper
func NewColorMapper() ColorMapper {
	return &colorMapper{
		validator: NewValidator(),
	}
}

// MapColors maps a color scheme to application-specific color names
func (m *colorMapper) MapColors(scheme *ColorScheme, targetApp string) (map[string]string, error) {
	if scheme == nil {
		return nil, fmt.Errorf("scheme is nil")
	}

	// Start with the base colors
	colors := make(map[string]string)
	for k, v := range scheme.Colors {
		colors[k] = v
	}

	// Add special colors if present
	for k, v := range scheme.Special {
		colors[k] = v
	}

	// Apply application-specific mappings
	switch targetApp {
	case "discord":
		return m.mapDiscordColors(colors), nil
	case "gtk":
		return m.mapGTKColors(colors), nil
	case "qt":
		return m.mapQtColors(colors), nil
	case "terminal":
		return m.mapTerminalColors(colors), nil
	case "quickshell":
		return m.mapQuickShellColors(colors), nil
	case "hyprland":
		return m.mapHyprlandColors(colors), nil
	default:
		// Return colors as-is for unknown applications
		return colors, nil
	}
}

// ConvertFormat converts a color to a specific format
func (m *colorMapper) ConvertFormat(color string, format ColorFormat) (string, error) {
	// Validate the input color
	if err := m.validator.ValidateColor(color); err != nil {
		return "", fmt.Errorf("invalid color: %w", err)
	}

	// Normalize to hex first
	hexColor := m.normalizeToHex(color)

	switch format {
	case ColorFormatHex:
		return hexColor, nil
	case ColorFormatHexNoHash:
		return strings.TrimPrefix(hexColor, "#"), nil
	case ColorFormatRGB:
		return m.hexToRGB(hexColor), nil
	case ColorFormatRGBA:
		return m.hexToRGBA(hexColor, 1.0), nil
	case ColorFormatHSL:
		return m.hexToHSL(hexColor), nil
	case ColorFormatHSLA:
		return m.hexToHSLA(hexColor, 1.0), nil
	default:
		return "", fmt.Errorf("unsupported color format: %s", format)
	}
}

// ValidateColor validates a color string
func (m *colorMapper) ValidateColor(color string) error {
	return m.validator.ValidateColor(color)
}

// mapDiscordColors maps colors for Discord themes
func (m *colorMapper) mapDiscordColors(colors map[string]string) map[string]string {
	mapped := make(map[string]string)

	// Copy original colors
	for k, v := range colors {
		mapped[k] = v
	}

	// Add Discord-specific mappings
	if bg, ok := colors["background"]; ok {
		mapped["background-primary"] = bg
		mapped["background-secondary"] = m.darkenColor(bg, 5)
		mapped["background-tertiary"] = m.darkenColor(bg, 10)
		mapped["background-floating"] = bg
		mapped["background-mobile-primary"] = bg
		mapped["background-mobile-secondary"] = m.darkenColor(bg, 5)
		mapped["channeltextarea-background"] = m.lightenColor(bg, 5)
	}

	if fg, ok := colors["foreground"]; ok {
		mapped["text-normal"] = fg
		mapped["text-muted"] = m.darkenColor(fg, 30)
		mapped["header-primary"] = fg
		mapped["header-secondary"] = colors["colour7"]
	}

	if blue, ok := colors["colour4"]; ok {
		mapped["text-link"] = blue
	}

	mapped["interactive-normal"] = colors["colour7"]
	mapped["interactive-hover"] = colors["foreground"]
	mapped["interactive-active"] = colors["colour15"]
	mapped["interactive-muted"] = colors["colour8"]

	return mapped
}

// mapGTKColors maps colors for GTK themes
func (m *colorMapper) mapGTKColors(colors map[string]string) map[string]string {
	mapped := make(map[string]string)

	// Copy original colors
	for k, v := range colors {
		mapped[k] = v
	}

	// GTK Material Design mappings
	if bg, ok := colors["background"]; ok {
		mapped["surface"] = m.darkenColor(bg, 5)
		mapped["surface_variant"] = m.lightenColor(bg, 10)
	}

	mapped["primary"] = colors["colour4"] // Blue
	mapped["primary_container"] = m.lightenColor(colors["background"], 10)
	mapped["secondary"] = colors["colour5"] // Magenta
	mapped["secondary_container"] = m.lightenColor(colors["background"], 15)

	mapped["error"] = colors["colour1"]   // Red
	mapped["warning"] = colors["colour3"] // Yellow
	mapped["success"] = colors["colour2"] // Green

	mapped["on_background"] = colors["foreground"]
	mapped["on_surface"] = colors["foreground"]
	mapped["on_surface_variant"] = m.darkenColor(colors["foreground"], 20)

	mapped["outline"] = colors["colour8"]         // Bright Black
	mapped["outline_variant"] = colors["colour7"] // White

	return mapped
}

// mapQtColors maps colors for Qt themes
func (m *colorMapper) mapQtColors(colors map[string]string) map[string]string {
	// Qt uses the same color names, just ensure they're all present
	return colors
}

// mapTerminalColors ensures terminal colors are properly formatted
func (m *colorMapper) mapTerminalColors(colors map[string]string) map[string]string {
	mapped := make(map[string]string)

	// Ensure all colors have # prefix
	for k, v := range colors {
		if !strings.HasPrefix(v, "#") && m.validator.IsHexColor(v) {
			mapped[k] = "#" + v
		} else {
			mapped[k] = v
		}
	}

	return mapped
}

// mapQuickShellColors maps colors for QuickShell (no # prefix, British spelling)
func (m *colorMapper) mapQuickShellColors(colors map[string]string) map[string]string {
	mapped := make(map[string]string)

	// QuickShell uses "colours" (British spelling) and no # prefix
	for k, v := range colors {
		// Remove # prefix for QuickShell
		cleanColor := strings.TrimPrefix(v, "#")
		mapped[k] = cleanColor
	}

	return mapped
}

// mapHyprlandColors maps colors for Hyprland window manager
func (m *colorMapper) mapHyprlandColors(colors map[string]string) map[string]string {
	mapped := make(map[string]string)

	// Hyprland uses rgb() format without #
	for k, v := range colors {
		// Convert to RGB format for Hyprland
		mapped[k] = m.hexToRGB(v)
	}

	// Add accent colors
	if blue, ok := colors["colour4"]; ok {
		mapped["accent"] = m.hexToRGB(blue)
	}
	if magenta, ok := colors["colour5"]; ok {
		mapped["accent_alt"] = m.hexToRGB(magenta)
	}

	return mapped
}

// Color manipulation helpers

func (m *colorMapper) normalizeToHex(color string) string {
	// Already hex
	if m.validator.IsHexColor(color) {
		if !strings.HasPrefix(color, "#") {
			return "#" + color
		}
		return color
	}

	// Convert from other formats
	// This is simplified - full implementation would parse RGB/HSL
	return color
}

func (m *colorMapper) hexToRGB(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "rgb(0, 0, 0)"
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

func (m *colorMapper) hexToRGBA(hex string, alpha float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return fmt.Sprintf("rgba(0, 0, 0, %.2f)", alpha)
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
}

func (m *colorMapper) hexToHSL(hex string) string {
	// Simplified - full implementation would do proper conversion
	return "hsl(0, 0%, 0%)"
}

func (m *colorMapper) hexToHSLA(hex string, alpha float64) string {
	// Simplified - full implementation would do proper conversion
	return fmt.Sprintf("hsla(0, 0%%, 0%%, %.2f)", alpha)
}

func (m *colorMapper) lightenColor(hex string, percent float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "#" + hex
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Simple lightening - increase each component
	factor := 1.0 + (percent / 100.0)
	r = int64(float64(r) * factor)
	g = int64(float64(g) * factor)
	b = int64(float64(b) * factor)

	// Clamp to 255
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (m *colorMapper) darkenColor(hex string, percent float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "#" + hex
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	// Simple darkening - decrease each component
	factor := 1.0 - (percent / 100.0)
	r = int64(float64(r) * factor)
	g = int64(float64(g) * factor)
	b = int64(float64(b) * factor)

	// Clamp to 0
	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
