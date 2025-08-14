package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur404dev/heimdall-cli/internal/discord"
	"github.com/arthur404dev/heimdall-cli/internal/terminal"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Applier applies themes to various applications
type Applier struct {
	replacer    *SimpleReplacer
	configDir   string
	dataDir     string
	templateDir string
}

// NewApplier creates a new theme applier
func NewApplier(configDir, dataDir string) *Applier {
	return &Applier{
		replacer:    NewSimpleReplacer(),
		configDir:   configDir,
		dataDir:     dataDir,
		templateDir: filepath.Join(dataDir, "templates"),
	}
}

// ApplyTheme applies a theme to a specific application
func (a *Applier) ApplyTheme(app string, colors map[string]string, mode string) error {
	// Special handling for Discord
	if app == "discord" {
		return a.ApplyDiscordThemes(colors)
	}

	// Load the template for the application
	templatePath := filepath.Join(a.templateDir, app+".tmpl")

	var templateContent string
	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try embedded templates
		content, err := a.getEmbeddedTemplate(app)
		if err != nil {
			return fmt.Errorf("template not found for %s: %w", app, err)
		}
		templateContent = content
	} else {
		// Read template file
		contentBytes, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
		}
		templateContent = string(contentBytes)
	}

	// Render the template using simple string replacement
	rendered, err := a.replacer.ReplaceTemplate(templateContent, colors)
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

	// Apply Discord themes to all detected clients
	if err := a.ApplyDiscordThemes(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply Discord themes: %v\n", err)
	}

	// Generate and save terminal sequences
	if err := a.ApplyTerminalSequences(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply terminal sequences: %v\n", err)
	}

	return nil
}

// ApplyTerminalSequences generates and applies ANSI terminal sequences
func (a *Applier) ApplyTerminalSequences(colors map[string]string) error {
	builder := terminal.NewSequenceBuilder()
	applier := terminal.NewApplier()

	// Generate sequences
	sequences, err := builder.GenerateSequences(colors)
	if err != nil {
		return fmt.Errorf("failed to generate terminal sequences: %w", err)
	}

	// Apply sequences to active terminals immediately (like caelestia)
	if err := applier.ApplySequencesWithFallback(colors); err != nil {
		// Log warning but don't fail the entire operation
		fmt.Fprintf(os.Stderr, "Warning: failed to apply sequences to terminals: %v\n", err)
	}

	// Format for shell sourcing
	shellScript := builder.FormatSequencesForShell(sequences)

	// Write to sequences file
	sequencesPath := filepath.Join(a.configDir, "sequences.txt")
	if err := paths.AtomicWrite(sequencesPath, []byte(shellScript)); err != nil {
		return fmt.Errorf("failed to write terminal sequences: %w", err)
	}

	return nil
}

// ApplyDiscordThemes applies themes to all detected Discord clients
func (a *Applier) ApplyDiscordThemes(colors map[string]string) error {
	clientManager := discord.NewClientManager()

	// Get templates
	cssTemplate := discord.GetTemplate("css")
	betterDiscordTemplate := discord.GetTemplate("betterdiscord")

	// Apply themes to all detected Discord clients
	return clientManager.ApplyThemeToAll(colors, cssTemplate, betterDiscordTemplate)
}

// getOutputPath returns the output path for a themed application
func (a *Applier) getOutputPath(app string) string {
	switch app {
	case "btop":
		return filepath.Join(a.configDir, "btop", "themes", "heimdall.theme")
	// Discord clients are now handled by ApplyDiscordThemes method
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
	// Discord templates are now handled by the Discord client manager
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
theme[main_bg]="{{background}}"
theme[main_fg]="{{foreground}}"

# Title
theme[title]="{{foreground}}"

# Highlight
theme[hi_fg]="{{colour4}}"

# Selected
theme[selected_bg]="{{colour8}}"
theme[selected_fg]="{{colour7}}"

# Status
theme[inactive_fg]="{{colour8}}"
theme[graph_text]="{{foreground}}"

# Process box
theme[proc_misc]="{{colour5}}"

# CPU box
theme[cpu_box]="{{colour4}}"
theme[cpu_text]="{{colour7}}"

# Memory/Disk box
theme[mem_box]="{{colour5}}"
theme[mem_text]="{{colour7}}"

# Network box
theme[net_box]="{{colour6}}"
theme[net_text]="{{colour7}}"

# Process list
theme[proc_box]="{{colour0}}"
theme[proc_text]="{{foreground}}"
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
