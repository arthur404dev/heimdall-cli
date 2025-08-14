package theme

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Embedded default templates for each application

//go:embed templates/discord.css.tmpl
var embeddedDiscordTemplate string

//go:embed templates/gtk.css.tmpl
var embeddedGtkTemplate string

//go:embed templates/qt.conf.tmpl
var embeddedQtTemplate string

//go:embed templates/btop.theme.tmpl
var embeddedBtopTemplate string

//go:embed templates/fuzzel.ini.tmpl
var embeddedFuzzelTemplate string

//go:embed templates/spicetify.ini.tmpl
var embeddedSpicetifyTemplate string

//go:embed templates/hyprland.conf.tmpl
var embeddedHyprlandTemplate string

//go:embed templates/terminal.sh.tmpl
var embeddedTerminalTemplate string

//go:embed templates/kitty.conf.tmpl
var embeddedKittyTemplate string

//go:embed templates/alacritty.toml.tmpl
var embeddedAlacrittyTemplate string

//go:embed templates/wezterm.lua.tmpl
var embeddedWeztermTemplate string

//go:embed templates/waybar.css.tmpl
var embeddedWaybarTemplate string

//go:embed templates/rofi.rasi.tmpl
var embeddedRofiTemplate string

//go:embed templates/dunst.conf.tmpl
var embeddedDunstTemplate string

// TemplateRegistry manages embedded and custom templates
type TemplateRegistry struct {
	embedded  map[string]string
	customDir string
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry(customDir string) *TemplateRegistry {
	if customDir == "" {
		home, _ := os.UserHomeDir()
		customDir = filepath.Join(home, ".config", "heimdall", "templates")
	}

	return &TemplateRegistry{
		embedded: map[string]string{
			"discord":   embeddedDiscordTemplate,
			"gtk":       embeddedGtkTemplate,
			"qt":        embeddedQtTemplate,
			"btop":      embeddedBtopTemplate,
			"fuzzel":    embeddedFuzzelTemplate,
			"spicetify": embeddedSpicetifyTemplate,
			"hyprland":  embeddedHyprlandTemplate,
			"terminal":  embeddedTerminalTemplate,
			"kitty":     embeddedKittyTemplate,
			"alacritty": embeddedAlacrittyTemplate,
			"wezterm":   embeddedWeztermTemplate,
			"waybar":    embeddedWaybarTemplate,
			"rofi":      embeddedRofiTemplate,
			"dunst":     embeddedDunstTemplate,
		},
		customDir: customDir,
	}
}

// GetTemplate retrieves a template by name, checking custom templates first
func (tr *TemplateRegistry) GetTemplate(app string, templateName string) (string, error) {
	// If a specific template name is provided, look for it
	if templateName != "" && templateName != "default" {
		customPath := filepath.Join(tr.customDir, app, templateName+".tmpl")
		if content, err := os.ReadFile(customPath); err == nil {
			return string(content), nil
		}
	}

	// Check for default custom template
	customPath := filepath.Join(tr.customDir, app, "default.tmpl")
	if content, err := os.ReadFile(customPath); err == nil {
		return string(content), nil
	}

	// Fall back to embedded template
	if template, ok := tr.embedded[app]; ok {
		return template, nil
	}

	return "", fmt.Errorf("no template found for application: %s", app)
}

// ListTemplates returns available templates for an application
func (tr *TemplateRegistry) ListTemplates(app string) ([]string, error) {
	templates := []string{"default"} // Embedded template is always available

	// Check custom templates directory
	customAppDir := filepath.Join(tr.customDir, app)
	if entries, err := os.ReadDir(customAppDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
				name := strings.TrimSuffix(entry.Name(), ".tmpl")
				if name != "default" {
					templates = append(templates, name)
				}
			}
		}
	}

	return templates, nil
}

// ValidateTemplate checks if a template is syntactically valid
func (tr *TemplateRegistry) ValidateTemplate(content string) error {
	// Create a test processor to validate the template
	processor := NewTemplateProcessor()

	// Set test colors for validation
	testColors := map[string]string{
		"background": "#1e1e2e",
		"foreground": "#cdd6f4",
		"colour0":    "#45475a",
		"colour1":    "#f38ba8",
		"colour2":    "#a6e3a1",
		"colour3":    "#f9e2af",
		"colour4":    "#89b4fa",
		"colour5":    "#f5c2e7",
		"colour6":    "#94e2d5",
		"colour7":    "#bac2de",
		"colour8":    "#585b70",
		"colour9":    "#f38ba8",
		"colour10":   "#a6e3a1",
		"colour11":   "#f9e2af",
		"colour12":   "#89b4fa",
		"colour13":   "#f5c2e7",
		"colour14":   "#94e2d5",
		"colour15":   "#a6adc8",
	}

	// Try to process with simple substitution
	if _, err := processor.ProcessSimple(content, testColors); err != nil {
		return fmt.Errorf("simple template validation failed: %w", err)
	}

	// Try to process as advanced template if it contains advanced syntax
	if strings.Contains(content, "{{if") || strings.Contains(content, "{{range") {
		testData := TemplateData{
			Colors: testColors,
			Mode:   "dark",
			Dark:   true,
			Light:  false,
			Custom: map[string]interface{}{
				"scheme": map[string]string{
					"name":    "test",
					"flavour": "default",
					"mode":    "dark",
				},
			},
		}
		if _, err := processor.ProcessAdvanced("test", content, testData); err != nil {
			return fmt.Errorf("advanced template validation failed: %w", err)
		}
	}

	return nil
}

// SaveCustomTemplate saves a custom template
func (tr *TemplateRegistry) SaveCustomTemplate(app, name, content string) error {
	// Validate template first
	if err := tr.ValidateTemplate(content); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Create directory if needed
	appDir := filepath.Join(tr.customDir, app)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Write template file
	templatePath := filepath.Join(appDir, name+".tmpl")
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// GetTemplateVersion returns the version/hash of a template for caching
func (tr *TemplateRegistry) GetTemplateVersion(app, templateName string) (string, error) {
	content, err := tr.GetTemplate(app, templateName)
	if err != nil {
		return "", err
	}

	// Simple hash for version tracking
	hash := fmt.Sprintf("%x", hashString(content))
	return hash[:8], nil
}

// hashString creates a simple hash of a string
func hashString(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h
}
