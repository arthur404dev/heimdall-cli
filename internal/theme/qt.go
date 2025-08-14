package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// QtHandler handles Qt theme generation
type QtHandler struct {
	configDir string
}

// NewQtHandler creates a new Qt theme handler
func NewQtHandler() *QtHandler {
	homeDir, _ := os.UserHomeDir()
	return &QtHandler{
		configDir: homeDir,
	}
}

// Apply generates and applies Qt theme
func (h *QtHandler) Apply(colors map[string]string, mode string) error {
	// Generate Qt color configuration
	content := h.generateQtColors(colors, mode)

	// Write to Qt5ct config
	qt5Path := filepath.Join(h.configDir, ".config", "qt5ct", "colors", "heimdall.conf")
	if err := h.writeThemeFile(qt5Path, content); err != nil {
		logger.Warn("Failed to write Qt5ct theme", "error", err)
	} else {
		logger.Info("Qt5ct theme applied", "path", qt5Path)
	}

	// Write to Qt6ct config
	qt6Path := filepath.Join(h.configDir, ".config", "qt6ct", "colors", "heimdall.conf")
	if err := h.writeThemeFile(qt6Path, content); err != nil {
		logger.Warn("Failed to write Qt6ct theme", "error", err)
	} else {
		logger.Info("Qt6ct theme applied", "path", qt6Path)
	}

	return nil
}

// generateQtColors creates Qt color configuration from colors
func (h *QtHandler) generateQtColors(colors map[string]string, mode string) string {
	// Qt5ct/Qt6ct uses a specific format for color schemes
	// Format: active_colors=#foreground, #background, ...

	// Build color arrays for different states
	activeColors := h.buildColorArray(colors, "active")
	disabledColors := h.buildColorArray(colors, "disabled")
	inactiveColors := h.buildColorArray(colors, "inactive")

	var builder strings.Builder
	builder.WriteString("[ColorScheme]\n")
	builder.WriteString(fmt.Sprintf("active_colors=%s\n", activeColors))
	builder.WriteString(fmt.Sprintf("disabled_colors=%s\n", disabledColors))
	builder.WriteString(fmt.Sprintf("inactive_colors=%s\n", inactiveColors))

	return builder.String()
}

// buildColorArray builds a color array for Qt5ct/Qt6ct
func (h *QtHandler) buildColorArray(colors map[string]string, state string) string {
	// Qt5ct expects 21 colors in a specific order
	// Reference: https://github.com/desktop-app/qt5ct/blob/master/src/qt5ct/paletteeditdialog.cpp

	var qtColors []string

	switch state {
	case "active":
		qtColors = []string{
			colors["foreground"],                  // 0: WindowText
			h.lighten(colors["background"], 0.1),  // 1: Button
			h.lighten(colors["background"], 0.2),  // 2: Light
			h.lighten(colors["background"], 0.15), // 3: Midlight
			colors["colour8"],                     // 4: Dark
			colors["colour7"],                     // 5: Mid
			colors["foreground"],                  // 6: Text
			"#ffffff",                             // 7: BrightText
			colors["foreground"],                  // 8: ButtonText
			colors["background"],                  // 9: Base
			h.darken(colors["background"], 0.05),  // 10: Window
			colors["colour0"],                     // 11: Shadow
			colors["colour4"],                     // 12: Highlight
			colors["background"],                  // 13: HighlightedText
			colors["colour6"],                     // 14: Link
			colors["colour5"],                     // 15: LinkVisited
			h.lighten(colors["background"], 0.1),  // 16: AlternateBase
			colors["foreground"],                  // 17: NoRole
			colors["colour3"],                     // 18: ToolTipBase
			colors["background"],                  // 19: ToolTipText
			colors["colour7"],                     // 20: PlaceholderText
		}

	case "disabled":
		qtColors = []string{
			colors["colour7"],                     // 0: WindowText
			h.lighten(colors["background"], 0.1),  // 1: Button
			h.lighten(colors["background"], 0.2),  // 2: Light
			h.lighten(colors["background"], 0.15), // 3: Midlight
			colors["colour8"],                     // 4: Dark
			colors["colour7"],                     // 5: Mid
			colors["colour7"],                     // 6: Text
			"#ffffff",                             // 7: BrightText
			colors["colour7"],                     // 8: ButtonText
			colors["background"],                  // 9: Base
			h.darken(colors["background"], 0.05),  // 10: Window
			colors["colour0"],                     // 11: Shadow
			h.lighten(colors["background"], 0.1),  // 12: Highlight
			colors["colour7"],                     // 13: HighlightedText
			colors["colour6"],                     // 14: Link
			colors["colour5"],                     // 15: LinkVisited
			h.lighten(colors["background"], 0.1),  // 16: AlternateBase
			colors["colour7"],                     // 17: NoRole
			colors["colour3"],                     // 18: ToolTipBase
			colors["colour7"],                     // 19: ToolTipText
			colors["colour7"],                     // 20: PlaceholderText
		}

	case "inactive":
		qtColors = []string{
			colors["foreground"],                  // 0: WindowText
			h.lighten(colors["background"], 0.1),  // 1: Button
			h.lighten(colors["background"], 0.2),  // 2: Light
			h.lighten(colors["background"], 0.15), // 3: Midlight
			colors["colour8"],                     // 4: Dark
			colors["colour7"],                     // 5: Mid
			colors["foreground"],                  // 6: Text
			"#ffffff",                             // 7: BrightText
			colors["foreground"],                  // 8: ButtonText
			colors["background"],                  // 9: Base
			h.darken(colors["background"], 0.05),  // 10: Window
			colors["colour0"],                     // 11: Shadow
			colors["colour4"],                     // 12: Highlight
			colors["background"],                  // 13: HighlightedText
			colors["colour6"],                     // 14: Link
			colors["colour5"],                     // 15: LinkVisited
			h.lighten(colors["background"], 0.1),  // 16: AlternateBase
			colors["foreground"],                  // 17: NoRole
			colors["colour3"],                     // 18: ToolTipBase
			colors["background"],                  // 19: ToolTipText
			colors["colour7"],                     // 20: PlaceholderText
		}
	}

	// Join colors with commas
	return strings.Join(qtColors, ", ")
}

// writeThemeFile writes theme content to file with atomic operations
func (h *QtHandler) writeThemeFile(path string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write atomically
	return paths.AtomicWrite(path, []byte(content))
}

// lighten makes a color lighter (simple approximation)
func (h *QtHandler) lighten(hexColor string, factor float64) string {
	// Remove # prefix
	color := strings.TrimPrefix(hexColor, "#")
	if len(color) != 6 {
		return hexColor // Return original if invalid
	}

	// Parse RGB
	var r, g, b int
	fmt.Sscanf(color, "%02x%02x%02x", &r, &g, &b)

	// Lighten
	r = min(255, int(float64(r)*(1+factor)))
	g = min(255, int(float64(g)*(1+factor)))
	b = min(255, int(float64(b)*(1+factor)))

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// darken makes a color darker (simple approximation)
func (h *QtHandler) darken(hexColor string, factor float64) string {
	// Remove # prefix
	color := strings.TrimPrefix(hexColor, "#")
	if len(color) != 6 {
		return hexColor // Return original if invalid
	}

	// Parse RGB
	var r, g, b int
	fmt.Sscanf(color, "%02x%02x%02x", &r, &g, &b)

	// Darken
	r = max(0, int(float64(r)*(1-factor)))
	g = max(0, int(float64(g)*(1-factor)))
	b = max(0, int(float64(b)*(1-factor)))

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
