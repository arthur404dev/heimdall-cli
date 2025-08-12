package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Applier applies themes to various applications
type Applier struct {
	engine      *Engine
	configDir   string
	dataDir     string
	templateDir string
}

// NewApplier creates a new theme applier
func NewApplier(configDir, dataDir string) *Applier {
	return &Applier{
		engine:      NewEngine(),
		configDir:   configDir,
		dataDir:     dataDir,
		templateDir: filepath.Join(dataDir, "templates"),
	}
}

// ApplyTheme applies a theme to a specific application
func (a *Applier) ApplyTheme(app string, colors map[string]string, mode string) error {
	// Load the template for the application
	templatePath := filepath.Join(a.templateDir, app+".tmpl")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try embedded templates
		content, err := a.getEmbeddedTemplate(app)
		if err != nil {
			return fmt.Errorf("template not found for %s: %w", app, err)
		}
		if err := a.engine.LoadTemplate(app, content); err != nil {
			return err
		}
	} else {
		if err := a.engine.LoadTemplateFile(app, templatePath); err != nil {
			return err
		}
	}

	// Prepare template data
	data := map[string]interface{}{
		"colors": colors,
		"mode":   mode,
		"isDark": mode == "dark",
	}

	// Render the template
	rendered, err := a.engine.Render(app, data)
	if err != nil {
		return fmt.Errorf("failed to render theme for %s: %w", app, err)
	}

	// Write the rendered theme to the appropriate location
	outputPath := a.getOutputPath(app)
	if err := paths.AtomicWrite(outputPath, []byte(rendered)); err != nil {
		return fmt.Errorf("failed to write theme for %s: %w", app, err)
	}

	return nil
}

// ApplyAllThemes applies themes to all supported applications
func (a *Applier) ApplyAllThemes(colors map[string]string, mode string) error {
	apps := []string{
		"btop",
		"discord",
		"fuzzel",
		"gtk",
		"qt",
		"spicetify",
	}

	for _, app := range apps {
		if err := a.ApplyTheme(app, colors, mode); err != nil {
			// Log error but continue with other apps
			fmt.Fprintf(os.Stderr, "Warning: failed to apply theme for %s: %v\n", app, err)
		}
	}

	return nil
}

// getOutputPath returns the output path for a themed application
func (a *Applier) getOutputPath(app string) string {
	switch app {
	case "btop":
		return filepath.Join(a.configDir, "btop", "themes", "heimdall.theme")
	case "discord":
		return filepath.Join(a.configDir, "vesktop", "themes", "heimdall.css")
	case "fuzzel":
		return filepath.Join(a.configDir, "fuzzel", "fuzzel.ini")
	case "gtk":
		return filepath.Join(a.configDir, "gtk-3.0", "gtk.css")
	case "qt":
		return filepath.Join(a.configDir, "qt5ct", "colors", "heimdall.conf")
	case "spicetify":
		return filepath.Join(a.configDir, "spicetify", "Themes", "heimdall", "color.ini")
	default:
		return filepath.Join(a.configDir, app, "heimdall.theme")
	}
}

// getEmbeddedTemplate returns embedded template content
func (a *Applier) getEmbeddedTemplate(app string) (string, error) {
	// These would normally be embedded with go:embed
	// For now, return basic templates
	switch app {
	case "btop":
		return btopTemplate, nil
	case "discord":
		return discordTemplate, nil
	case "fuzzel":
		return fuzzelTemplate, nil
	case "gtk":
		return gtkTemplate, nil
	case "qt":
		return qtTemplate, nil
	case "spicetify":
		return spicetifyTemplate, nil
	default:
		return "", fmt.Errorf("no embedded template for %s", app)
	}
}

// Embedded template strings (simplified versions)
const btopTemplate = `# Heimdall theme for btop
# Generated automatically

# Main background and foreground
theme[main_bg]="{{.colors.background}}"
theme[main_fg]="{{.colors.on_background}}"

# Title
theme[title]="{{.colors.on_background}}"

# Highlight
theme[hi_fg]="{{.colors.primary}}"

# Selected
theme[selected_bg]="{{.colors.surface_variant}}"
theme[selected_fg]="{{.colors.on_surface_variant}}"

# Status
theme[inactive_fg]="{{.colors.outline}}"
theme[graph_text]="{{.colors.on_surface}}"

# Process box
theme[proc_misc]="{{.colors.secondary}}"

# CPU box
theme[cpu_box]="{{.colors.primary}}"
theme[cpu_text]="{{.colors.on_primary}}"

# Memory/Disk box
theme[mem_box]="{{.colors.secondary}}"
theme[mem_text]="{{.colors.on_secondary}}"

# Network box
theme[net_box]="{{.colors.tertiary}}"
theme[net_text]="{{.colors.on_tertiary}}"

# Process list
theme[proc_box]="{{.colors.surface}}"
theme[proc_text]="{{.colors.on_surface}}"
`

const discordTemplate = `/* Heimdall theme for Discord */
/* Generated automatically */

:root {
    --primary: {{.colors.primary}};
    --primary-container: {{.colors.primary_container}};
    --on-primary: {{.colors.on_primary}};
    --on-primary-container: {{.colors.on_primary_container}};
    
    --secondary: {{.colors.secondary}};
    --secondary-container: {{.colors.secondary_container}};
    --on-secondary: {{.colors.on_secondary}};
    
    --background: {{.colors.background}};
    --surface: {{.colors.surface}};
    --surface-variant: {{.colors.surface_variant}};
    
    --on-background: {{.colors.on_background}};
    --on-surface: {{.colors.on_surface}};
    --on-surface-variant: {{.colors.on_surface_variant}};
    
    --outline: {{.colors.outline}};
    --outline-variant: {{.colors.outline_variant}};
    
    --error: {{.colors.error}};
    --on-error: {{.colors.on_error}};
}

/* Discord specific mappings */
.theme-{{if .isDark}}dark{{else}}light{{end}} {
    --background-primary: var(--background);
    --background-secondary: var(--surface);
    --background-tertiary: var(--surface-variant);
    
    --text-normal: var(--on-background);
    --text-muted: var(--on-surface-variant);
    --interactive-normal: var(--on-surface);
    --interactive-hover: var(--primary);
    --interactive-active: var(--primary-container);
}
`

const fuzzelTemplate = `# Heimdall theme for fuzzel
# Generated automatically

[main]
font=monospace:size=10
dpi-aware=yes
width=30
horizontal-pad=20
vertical-pad=10
inner-pad=10

[colors]
background={{.colors.surface}}dd
text={{.colors.on_surface}}ff
match={{.colors.primary}}ff
selection={{.colors.primary_container}}ff
selection-text={{.colors.on_primary_container}}ff
selection-match={{.colors.primary}}ff
border={{.colors.outline}}ff
`

const gtkTemplate = `/* Heimdall theme for GTK */
/* Generated automatically */

@define-color background {{.colors.background}};
@define-color surface {{.colors.surface}};
@define-color surface_variant {{.colors.surface_variant}};

@define-color primary {{.colors.primary}};
@define-color primary_container {{.colors.primary_container}};
@define-color secondary {{.colors.secondary}};
@define-color secondary_container {{.colors.secondary_container}};

@define-color on_background {{.colors.on_background}};
@define-color on_surface {{.colors.on_surface}};
@define-color on_surface_variant {{.colors.on_surface_variant}};

@define-color outline {{.colors.outline}};
@define-color outline_variant {{.colors.outline_variant}};

@define-color error {{.colors.error}};

/* Apply to GTK widgets */
window {
    background-color: @background;
    color: @on_background;
}

button {
    background-color: @primary;
    color: @on_primary;
}

button:hover {
    background-color: @primary_container;
    color: @on_primary_container;
}

entry {
    background-color: @surface;
    color: @on_surface;
    border-color: @outline;
}
`

const qtTemplate = `# Heimdall theme for Qt
# Generated automatically

[ColorScheme]
active_colors={{.colors.on_surface}}, {{.colors.surface}}, {{.colors.surface_variant}}, {{.colors.outline_variant}}, {{.colors.on_surface_variant}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.surface}}, {{.colors.background}}, {{.colors.on_background}}, {{.colors.primary}}, {{.colors.on_primary}}, {{.colors.primary_container}}, {{.colors.on_primary_container}}, {{.colors.surface_variant}}, {{.colors.on_surface}}, {{.colors.surface}}, {{.colors.on_surface}}, {{.colors.outline}}
disabled_colors={{.colors.outline}}, {{.colors.surface}}, {{.colors.surface_variant}}, {{.colors.outline_variant}}, {{.colors.outline}}, {{.colors.outline}}, {{.colors.outline}}, {{.colors.outline}}, {{.colors.outline}}, {{.colors.surface}}, {{.colors.background}}, {{.colors.outline}}, {{.colors.surface_variant}}, {{.colors.outline}}, {{.colors.primary_container}}, {{.colors.outline}}, {{.colors.surface_variant}}, {{.colors.outline}}, {{.colors.surface}}, {{.colors.outline}}, {{.colors.outline_variant}}
inactive_colors={{.colors.on_surface}}, {{.colors.surface}}, {{.colors.surface_variant}}, {{.colors.outline_variant}}, {{.colors.on_surface_variant}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.on_surface}}, {{.colors.surface}}, {{.colors.background}}, {{.colors.on_background}}, {{.colors.primary}}, {{.colors.on_primary}}, {{.colors.primary_container}}, {{.colors.on_primary_container}}, {{.colors.surface_variant}}, {{.colors.on_surface}}, {{.colors.surface}}, {{.colors.on_surface}}, {{.colors.outline}}
`

const spicetifyTemplate = `# Heimdall theme for Spicetify
# Generated automatically

[Base]
main_bg = {{.colors.background}}
sidebar_bg = {{.colors.surface}}
player_bg = {{.colors.surface_variant}}
card_bg = {{.colors.surface}}
shadow = {{.colors.shadow}}
main_fg = {{.colors.on_background}}
sidebar_fg = {{.colors.on_surface}}
secondary_fg = {{.colors.on_surface_variant}}
selected_button = {{.colors.primary}}
pressing_button_bg = {{.colors.primary_container}}
miscellaneous_bg = {{.colors.surface_variant}}
preserve_1 = {{.colors.on_primary}}
`
