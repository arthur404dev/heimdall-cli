package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// GTKHandler handles GTK theme generation
type GTKHandler struct {
	configDir string
}

// NewGTKHandler creates a new GTK theme handler
func NewGTKHandler() *GTKHandler {
	homeDir, _ := os.UserHomeDir()
	return &GTKHandler{
		configDir: homeDir,
	}
}

// Apply generates and applies GTK theme
func (h *GTKHandler) Apply(colors map[string]string, mode string) error {
	// Generate GTK CSS content
	content := h.generateGTKCSS(colors, mode)

	// Write to GTK3 config
	gtk3Path := filepath.Join(h.configDir, ".config", "gtk-3.0", "gtk.css")
	if err := h.writeThemeFile(gtk3Path, content); err != nil {
		logger.Warn("Failed to write GTK3 theme", "error", err)
	} else {
		logger.Info("GTK3 theme applied", "path", gtk3Path)
	}

	// Write to GTK4 config
	gtk4Path := filepath.Join(h.configDir, ".config", "gtk-4.0", "gtk.css")
	if err := h.writeThemeFile(gtk4Path, content); err != nil {
		logger.Warn("Failed to write GTK4 theme", "error", err)
	} else {
		logger.Info("GTK4 theme applied", "path", gtk4Path)
	}

	return nil
}

// generateGTKCSS creates GTK CSS content from colors
func (h *GTKHandler) generateGTKCSS(colors map[string]string, mode string) string {
	var builder strings.Builder

	// Header
	builder.WriteString("/* Heimdall GTK Theme */\n")
	builder.WriteString(fmt.Sprintf("/* Generated: %s */\n", time.Now().Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("/* Mode: %s */\n\n", mode))

	// Define color variables
	builder.WriteString("/* Color definitions */\n")
	builder.WriteString(fmt.Sprintf("@define-color background %s;\n", colors["background"]))
	builder.WriteString(fmt.Sprintf("@define-color foreground %s;\n", colors["foreground"]))
	builder.WriteString(fmt.Sprintf("@define-color primary %s;\n", colors["colour4"]))
	builder.WriteString(fmt.Sprintf("@define-color primary_container %s;\n", h.lighten(colors["background"], 0.1)))
	builder.WriteString(fmt.Sprintf("@define-color secondary %s;\n", colors["colour5"]))
	builder.WriteString(fmt.Sprintf("@define-color secondary_container %s;\n", h.lighten(colors["background"], 0.15)))
	builder.WriteString(fmt.Sprintf("@define-color error %s;\n", colors["colour1"]))
	builder.WriteString(fmt.Sprintf("@define-color warning %s;\n", colors["colour3"]))
	builder.WriteString(fmt.Sprintf("@define-color success %s;\n", colors["colour2"]))
	builder.WriteString(fmt.Sprintf("@define-color surface %s;\n", h.darken(colors["background"], 0.05)))
	builder.WriteString(fmt.Sprintf("@define-color on_surface %s;\n", colors["foreground"]))
	builder.WriteString(fmt.Sprintf("@define-color outline %s;\n", colors["colour8"]))
	builder.WriteString("\n")

	// Window styling
	builder.WriteString("/* Window styling */\n")
	builder.WriteString("window {\n")
	builder.WriteString("    background-color: @background;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("}\n\n")

	// Button styling
	builder.WriteString("/* Button styling */\n")
	builder.WriteString("button {\n")
	builder.WriteString("    background-color: @primary_container;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("    border: 1px solid @outline;\n")
	builder.WriteString("    border-radius: 4px;\n")
	builder.WriteString("    padding: 6px 12px;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("button:hover {\n")
	builder.WriteString("    background-color: @primary;\n")
	builder.WriteString("    color: @background;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("button:active {\n")
	builder.WriteString("    background-color: @secondary;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("button:disabled {\n")
	builder.WriteString("    opacity: 0.5;\n")
	builder.WriteString("}\n\n")

	// Entry (text input) styling
	builder.WriteString("/* Entry styling */\n")
	builder.WriteString("entry {\n")
	builder.WriteString("    background-color: @surface;\n")
	builder.WriteString("    color: @on_surface;\n")
	builder.WriteString("    border: 1px solid @outline;\n")
	builder.WriteString("    border-radius: 4px;\n")
	builder.WriteString("    padding: 6px;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("entry:focus {\n")
	builder.WriteString("    border-color: @primary;\n")
	builder.WriteString("    box-shadow: 0 0 0 1px @primary;\n")
	builder.WriteString("}\n\n")

	// Headerbar styling
	builder.WriteString("/* Headerbar styling */\n")
	builder.WriteString("headerbar {\n")
	builder.WriteString("    background-color: @surface;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("    border-bottom: 1px solid @outline;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("headerbar button {\n")
	builder.WriteString("    background-color: transparent;\n")
	builder.WriteString("    border: none;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("headerbar button:hover {\n")
	builder.WriteString("    background-color: @primary_container;\n")
	builder.WriteString("}\n\n")

	// Sidebar styling
	builder.WriteString("/* Sidebar styling */\n")
	builder.WriteString(".sidebar {\n")
	builder.WriteString("    background-color: @surface;\n")
	builder.WriteString("    border-right: 1px solid @outline;\n")
	builder.WriteString("}\n\n")

	// List styling
	builder.WriteString("/* List styling */\n")
	builder.WriteString("list {\n")
	builder.WriteString("    background-color: @background;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("list row {\n")
	builder.WriteString("    padding: 8px;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("list row:hover {\n")
	builder.WriteString("    background-color: @primary_container;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("list row:selected {\n")
	builder.WriteString("    background-color: @primary;\n")
	builder.WriteString("    color: @background;\n")
	builder.WriteString("}\n\n")

	// Tooltip styling
	builder.WriteString("/* Tooltip styling */\n")
	builder.WriteString("tooltip {\n")
	builder.WriteString("    background-color: @surface;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("    border: 1px solid @outline;\n")
	builder.WriteString("    border-radius: 4px;\n")
	builder.WriteString("    padding: 4px 8px;\n")
	builder.WriteString("}\n\n")

	// Menu styling
	builder.WriteString("/* Menu styling */\n")
	builder.WriteString("menu, popover {\n")
	builder.WriteString("    background-color: @surface;\n")
	builder.WriteString("    color: @foreground;\n")
	builder.WriteString("    border: 1px solid @outline;\n")
	builder.WriteString("    border-radius: 4px;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("menuitem:hover {\n")
	builder.WriteString("    background-color: @primary_container;\n")
	builder.WriteString("}\n\n")

	// Scrollbar styling
	builder.WriteString("/* Scrollbar styling */\n")
	builder.WriteString("scrollbar {\n")
	builder.WriteString("    background-color: @background;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("scrollbar slider {\n")
	builder.WriteString("    background-color: @outline;\n")
	builder.WriteString("    border-radius: 4px;\n")
	builder.WriteString("    min-width: 8px;\n")
	builder.WriteString("    min-height: 8px;\n")
	builder.WriteString("}\n\n")

	builder.WriteString("scrollbar slider:hover {\n")
	builder.WriteString("    background-color: @primary;\n")
	builder.WriteString("}\n")

	return builder.String()
}

// writeThemeFile writes theme content to file with atomic operations
func (h *GTKHandler) writeThemeFile(path string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write atomically
	return paths.AtomicWrite(path, []byte(content))
}

// lighten makes a color lighter (simple approximation)
func (h *GTKHandler) lighten(hexColor string, factor float64) string {
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
func (h *GTKHandler) darken(hexColor string, factor float64) string {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
