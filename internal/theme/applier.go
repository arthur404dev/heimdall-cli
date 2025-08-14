// Package theme provides the core theme engine for applying color schemes
// to various applications and managing theme-related operations.
package theme

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/discord"
	"github.com/arthur404dev/heimdall-cli/internal/terminal"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Applier applies themes to various applications
type Applier struct {
	replacer     *SimpleReplacer
	configDir    string
	dataDir      string
	templateDir  string
	cache        *TemplateCache
	colorCache   *ColorConversionCache
	workerPool   int
	lazyHandlers map[string]func() ApplicationHandler
	handlers     map[string]ApplicationHandler
	handlersMu   sync.RWMutex
}

// NewApplier creates a new theme applier
func NewApplier(configDir, dataDir string) *Applier {
	// Initialize caches
	cacheDir := filepath.Join(dataDir, "cache")
	templateCache := NewTemplateCache(10, true, cacheDir) // 10MB cache with disk persistence
	colorCache := NewColorConversionCache(1000)           // Cache up to 1000 color conversions

	return &Applier{
		replacer:     NewSimpleReplacer(),
		configDir:    configDir,
		dataDir:      dataDir,
		templateDir:  filepath.Join(dataDir, "templates"),
		cache:        templateCache,
		colorCache:   colorCache,
		workerPool:   8, // Default to 8 workers for parallel application
		lazyHandlers: make(map[string]func() ApplicationHandler),
		handlers:     make(map[string]ApplicationHandler),
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
	outputPath := a.GetOutputPath(app)
	if err := paths.AtomicWrite(outputPath, []byte(rendered)); err != nil {
		return fmt.Errorf("failed to write theme for %s: %w", app, err)
	}

	return nil
}

// SetWorkerPoolSize sets the number of workers for parallel application
func (a *Applier) SetWorkerPoolSize(size int) {
	if size > 0 {
		a.workerPool = size
	}
}

// RegisterLazyHandler registers a handler to be loaded on first use
func (a *Applier) RegisterLazyHandler(name string, factory func() ApplicationHandler) {
	a.handlersMu.Lock()
	defer a.handlersMu.Unlock()
	a.lazyHandlers[name] = factory
}

// getHandler retrieves a handler, initializing it if needed (lazy loading)
func (a *Applier) getHandler(name string) ApplicationHandler {
	a.handlersMu.RLock()
	handler, exists := a.handlers[name]
	a.handlersMu.RUnlock()

	if exists {
		return handler
	}

	// Check for lazy handler
	a.handlersMu.Lock()
	defer a.handlersMu.Unlock()

	// Double-check after acquiring write lock
	if handler, exists = a.handlers[name]; exists {
		return handler
	}

	if factory, ok := a.lazyHandlers[name]; ok {
		start := time.Now()
		handler = factory()
		a.handlers[name] = handler
		delete(a.lazyHandlers, name) // Remove factory after initialization

		// Log initialization time
		elapsed := time.Since(start)
		if elapsed > 10*time.Millisecond {
			fmt.Fprintf(os.Stderr, "Warning: Handler %s took %v to initialize\n", name, elapsed)
		}

		return handler
	}

	return nil
}

// ApplyAllThemes applies themes to all supported applications
func (a *Applier) ApplyAllThemes(colors map[string]string, mode string, schemeName string) error {
	// Apply GTK theme
	gtkHandler := NewGTKHandler()
	if err := gtkHandler.Apply(colors, mode); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply GTK theme: %v\n", err)
	}

	// Apply Qt theme
	qtHandler := NewQtHandler()
	if err := qtHandler.Apply(colors, mode); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply Qt theme: %v\n", err)
	}

	// Apply tool-specific themes
	if err := a.ApplyBtopTheme(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply btop theme: %v\n", err)
	}

	if err := a.ApplyFuzzelTheme(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply fuzzel theme: %v\n", err)
	}

	if err := a.ApplySpicetifyTheme(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply spicetify theme: %v\n", err)
	}

	// Apply Discord themes to all detected clients
	if err := a.ApplyDiscordThemes(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply Discord themes: %v\n", err)
	}

	// Generate and save terminal sequences
	if err := a.ApplyTerminalSequences(colors, schemeName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to apply terminal sequences: %v\n", err)
	}

	return nil
}

// ApplyAllThemesParallel applies themes to all applications concurrently
func (a *Applier) ApplyAllThemesParallel(ctx context.Context, colors map[string]string, mode string, schemeName string) error {
	// Create a semaphore to limit concurrent workers
	sem := make(chan struct{}, a.workerPool)

	// Error channel to collect errors
	errChan := make(chan error, 10) // Buffer for up to 10 errors

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Result tracking
	type result struct {
		app string
		err error
	}
	results := make(chan result, 10)

	// Start result collector
	var allErrors []error
	go func() {
		for r := range results {
			if r.err != nil {
				allErrors = append(allErrors, fmt.Errorf("%s: %w", r.app, r.err))
				fmt.Fprintf(os.Stderr, "Warning: failed to apply %s theme: %v\n", r.app, r.err)
			}
		}
	}()

	// Define all application tasks
	tasks := []struct {
		name string
		fn   func() error
	}{
		{"gtk", func() error {
			gtkHandler := NewGTKHandler()
			return gtkHandler.Apply(colors, mode)
		}},
		{"qt", func() error {
			qtHandler := NewQtHandler()
			return qtHandler.Apply(colors, mode)
		}},
		{"btop", func() error {
			return a.ApplyBtopTheme(colors)
		}},
		{"fuzzel", func() error {
			return a.ApplyFuzzelTheme(colors)
		}},
		{"spicetify", func() error {
			return a.ApplySpicetifyTheme(colors)
		}},
		{"discord", func() error {
			return a.ApplyDiscordThemes(colors)
		}},
		{"terminal", func() error {
			return a.ApplyTerminalSequences(colors, schemeName)
		}},
		{"kitty", func() error {
			return a.ApplyKittyTheme(colors)
		}},
		{"alacritty", func() error {
			return a.ApplyAlacrittyTheme(colors)
		}},
		{"wezterm", func() error {
			return a.ApplyWeztermTheme(colors)
		}},
	}

	// Launch workers for each task
	for _, task := range tasks {
		wg.Add(1)

		// Copy task to avoid closure issues
		t := task

		go func() {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release semaphore
			case <-ctx.Done():
				results <- result{app: t.name, err: ctx.Err()}
				return
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				results <- result{app: t.name, err: ctx.Err()}
				return
			default:
			}

			// Execute task with timing
			start := time.Now()
			err := t.fn()
			elapsed := time.Since(start)

			// Log slow operations
			if elapsed > 200*time.Millisecond {
				fmt.Fprintf(os.Stderr, "Warning: %s theme took %v to apply\n", t.name, elapsed)
			}

			results <- result{app: t.name, err: err}
		}()
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(results)
	close(errChan)

	// Return aggregated errors if any
	if len(allErrors) > 0 {
		return fmt.Errorf("parallel theme application had %d errors", len(allErrors))
	}

	return nil
}

// ApplyThemeWithCache applies a theme using cached templates
func (a *Applier) ApplyThemeWithCache(app string, colors map[string]string, mode string) error {
	// Generate cache key
	cacheKey := CacheKey(app, mode, colors)

	// Check cache first
	if cached, ok := a.cache.Get(cacheKey); ok {
		// Use cached rendered template
		if rendered, ok := cached.(string); ok {
			outputPath := a.GetOutputPath(app)
			return paths.AtomicWrite(outputPath, []byte(rendered))
		}
	}

	// Not in cache, render normally
	err := a.ApplyTheme(app, colors, mode)
	if err != nil {
		return err
	}

	// Read rendered file and cache it
	outputPath := a.GetOutputPath(app)
	if content, err := os.ReadFile(outputPath); err == nil {
		// Store in cache (estimate size as byte length)
		a.cache.Set(cacheKey, string(content), int64(len(content)))
	}

	return nil
}

// ApplyTerminalSequences generates and saves ANSI terminal sequences to a file
func (a *Applier) ApplyTerminalSequences(colors map[string]string, schemeName string) error {
	builder := terminal.NewSequenceBuilder()

	// Generate sequences
	sequences, err := builder.GenerateSequences(colors)
	if err != nil {
		return fmt.Errorf("failed to generate terminal sequences: %w", err)
	}

	// DISABLED: Direct terminal application causes issues with modern terminals like Kitty
	// Modern terminals should use their config files (kitty.conf, alacritty.toml, etc.)
	// Only generate the sequences file for manual sourcing if needed

	// Format for shell sourcing
	shellScript := builder.FormatSequencesForShell(sequences, schemeName)

	// Write to sequences file for manual sourcing if needed
	sequencesPath := a.GetOutputPath("terminal")
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

// ApplyBtopTheme applies theme to btop
func (a *Applier) ApplyBtopTheme(colors map[string]string) error {
	content := a.generateBtopTheme(colors)
	btopPath := a.GetOutputPath("btop")

	// Ensure directory exists
	dir := filepath.Dir(btopPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create btop themes directory: %w", err)
	}

	return paths.AtomicWrite(btopPath, []byte(content))
}

// ApplyFuzzelTheme applies theme to fuzzel
func (a *Applier) ApplyFuzzelTheme(colors map[string]string) error {
	content := a.generateFuzzelTheme(colors)
	fuzzelPath := a.GetOutputPath("fuzzel")

	// Ensure directory exists
	dir := filepath.Dir(fuzzelPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create fuzzel directory: %w", err)
	}

	return paths.AtomicWrite(fuzzelPath, []byte(content))
}

// ApplySpicetifyTheme applies theme to Spicetify
func (a *Applier) ApplySpicetifyTheme(colors map[string]string) error {
	content := a.generateSpicetifyTheme(colors)
	spicetifyPath := a.GetOutputPath("spicetify")

	// Ensure directory exists
	dir := filepath.Dir(spicetifyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create spicetify themes directory: %w", err)
	}

	return paths.AtomicWrite(spicetifyPath, []byte(content))
}

// ApplyKittyTheme applies theme to Kitty terminal
func (a *Applier) ApplyKittyTheme(colors map[string]string) error {
	content, err := a.replacer.ReplaceTemplate(kittyTemplate, colors)
	if err != nil {
		return fmt.Errorf("failed to process kitty template: %w", err)
	}

	kittyPath := a.GetOutputPath("kitty")

	// Ensure directory exists
	dir := filepath.Dir(kittyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create kitty themes directory: %w", err)
	}

	return paths.AtomicWrite(kittyPath, []byte(content))
}

// ApplyAlacrittyTheme applies theme to Alacritty terminal
func (a *Applier) ApplyAlacrittyTheme(colors map[string]string) error {
	content, err := a.replacer.ReplaceTemplate(alacrittyTemplate, colors)
	if err != nil {
		return fmt.Errorf("failed to process alacritty template: %w", err)
	}

	alacrittyPath := a.GetOutputPath("alacritty")

	// Ensure directory exists
	dir := filepath.Dir(alacrittyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create alacritty themes directory: %w", err)
	}

	return paths.AtomicWrite(alacrittyPath, []byte(content))
}

// ApplyWeztermTheme applies theme to Wezterm terminal
func (a *Applier) ApplyWeztermTheme(colors map[string]string) error {
	content, err := a.replacer.ReplaceTemplate(weztermTemplate, colors)
	if err != nil {
		return fmt.Errorf("failed to process wezterm template: %w", err)
	}

	weztermPath := a.GetOutputPath("wezterm")

	// Ensure directory exists
	dir := filepath.Dir(weztermPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create wezterm themes directory: %w", err)
	}

	return paths.AtomicWrite(weztermPath, []byte(content))
}

// generateBtopTheme generates btop theme content
func (a *Applier) generateBtopTheme(colors map[string]string) string {
	// Use the replacer to process the template
	content, _ := a.replacer.ReplaceTemplate(btopTemplate, colors)
	return content
}

// generateFuzzelTheme generates fuzzel theme content
func (a *Applier) generateFuzzelTheme(colors map[string]string) string {
	// Fuzzel uses RGBA format, need to convert colors
	bg := strings.TrimPrefix(colors["background"], "#")
	fg := strings.TrimPrefix(colors["text"], "#")
	primary := strings.TrimPrefix(colors["term4"], "#")
	surface := strings.TrimPrefix(colors["term0"], "#")
	outline := strings.TrimPrefix(colors["term8"], "#")

	return fmt.Sprintf(`# Heimdall theme for fuzzel
# Generated automatically

[main]
font=monospace:size=10
dpi-aware=yes
width=30
horizontal-pad=20
vertical-pad=10
inner-pad=10

[colors]
background=%sdd
text=%sff
match=%sff
selection=%sff
selection-text=%sff
selection-match=%sff
border=%sff
`, bg, fg, primary, surface, fg, primary, outline)
}

// generateSpicetifyTheme generates Spicetify theme content
func (a *Applier) generateSpicetifyTheme(colors map[string]string) string {
	// Use the replacer to process the template
	content, _ := a.replacer.ReplaceTemplate(spicetifyTemplate, colors)
	return content
}

// GetOutputPath returns the output path for a themed application
// This is the single source of truth for all application theme paths
func (a *Applier) GetOutputPath(app string) string {
	cfg := config.Get()
	if cfg == nil {
		// This should never happen, but if it does, return a sensible default
		return filepath.Join(a.configDir, app, "heimdall.theme")
	}

	// All paths come from config (which has defaults set)
	switch app {
	case "btop":
		return cfg.Theme.Paths.Btop
	case "fuzzel":
		return cfg.Theme.Paths.Fuzzel
	case "gtk", "gtk3":
		return cfg.Theme.Paths.Gtk3
	case "gtk4":
		return cfg.Theme.Paths.Gtk4
	case "qt", "qt5":
		return cfg.Theme.Paths.Qt5
	case "qt6":
		return cfg.Theme.Paths.Qt6
	case "spicetify":
		return cfg.Theme.Paths.Spicetify
	case "kitty":
		return cfg.Theme.Paths.Kitty
	case "alacritty":
		return cfg.Theme.Paths.Alacritty
	case "wezterm":
		return cfg.Theme.Paths.Wezterm
	case "terminal":
		return cfg.Theme.Paths.Terminal
	case "vesktop":
		return cfg.Theme.Paths.Vesktop
	case "discord":
		return cfg.Theme.Paths.Discord
	case "discordcanary":
		return cfg.Theme.Paths.DiscordCanary
	case "vencord":
		return cfg.Theme.Paths.Vencord
	case "equicord":
		return cfg.Theme.Paths.Equicord
	case "betterdiscord":
		return cfg.Theme.Paths.BetterDiscord
	default:
		// Unknown app - use a generic path
		return filepath.Join(a.configDir, app, "heimdall.theme")
	}
}

// GetDiscordPaths returns all Discord-related paths from config
func (a *Applier) GetDiscordPaths() map[string]string {
	cfg := config.Get()
	if cfg == nil {
		return map[string]string{}
	}

	return map[string]string{
		"vesktop":       cfg.Theme.Paths.Vesktop,
		"discord":       cfg.Theme.Paths.Discord,
		"discordcanary": cfg.Theme.Paths.DiscordCanary,
		"vencord":       cfg.Theme.Paths.Vencord,
		"equicord":      cfg.Theme.Paths.Equicord,
		"betterdiscord": cfg.Theme.Paths.BetterDiscord,
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
	case "gtk", "gtk3", "gtk4":
		return gtkTemplate, nil
	case "qt", "qt5", "qt6":
		return qtTemplate, nil
	case "spicetify":
		return spicetifyTemplate, nil
	case "kitty":
		return kittyTemplate, nil
	case "alacritty":
		return alacrittyTemplate, nil
	case "wezterm":
		return weztermTemplate, nil
	default:
		return "", fmt.Errorf("no embedded template for %s", app)
	}
}

// Embedded template strings (simplified versions)
const btopTemplate = `# Heimdall theme for btop
# Generated automatically

# Main background and foreground
theme[main_bg]="#{{background.raw}}"
theme[main_fg]="#{{text.raw}}"

# Title
theme[title]="#{{text.raw}}"

# Highlight
theme[hi_fg]="#{{term4.raw}}"

# Selected
theme[selected_bg]="#{{term8.raw}}"
theme[selected_fg]="#{{term7.raw}}"

# Status
theme[inactive_fg]="#{{term8.raw}}"
theme[graph_text]="#{{text.raw}}"

# Process box
theme[proc_misc]="#{{term5.raw}}"

# CPU box
theme[cpu_box]="#{{term4.raw}}"
theme[cpu_text]="#{{term7.raw}}"

# Memory/Disk box
theme[mem_box]="#{{term5.raw}}"
theme[mem_text]="#{{term7.raw}}"

# Network box
theme[net_box]="#{{term6.raw}}"
theme[net_text]="#{{term7.raw}}"

# Process list
theme[proc_box]="#{{term0.raw}}"
theme[proc_text]="#{{text.raw}}"
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
main_bg = {{background.raw}}
sidebar_bg = {{term0.raw}}
player_bg = {{term8.raw}}
card_bg = {{term0.raw}}
shadow = 000000
main_fg = {{text.raw}}
sidebar_fg = {{text.raw}}
secondary_fg = {{term7.raw}}
selected_button = {{term4.raw}}
pressing_button_bg = {{term0.raw}}
pressing_button_fg = {{text.raw}}
miscellaneous_bg = {{term8.raw}}
miscellaneous_hover_bg = {{term0.raw}}
preserve_1 = ffffff
`

const kittyTemplate = `# Heimdall theme for Kitty
# Generated automatically

foreground #{{text.raw}}
background #{{background.raw}}
cursor #{{text.raw}}

# Black
color0 #{{term0.raw}}
color8 #{{term8.raw}}

# Red
color1 #{{term1.raw}}
color9 #{{term9.raw}}

# Green
color2 #{{term2.raw}}
color10 #{{term10.raw}}

# Yellow
color3 #{{term3.raw}}
color11 #{{term11.raw}}

# Blue
color4 #{{term4.raw}}
color12 #{{term12.raw}}

# Magenta
color5 #{{term5.raw}}
color13 #{{term13.raw}}

# Cyan
color6 #{{term6.raw}}
color14 #{{term14.raw}}

# White
color7 #{{term7.raw}}
color15 #{{term15.raw}}
`

const alacrittyTemplate = `# Heimdall theme for Alacritty
# Generated automatically

[colors.primary]
background = "#{{background.raw}}"
foreground = "#{{text.raw}}"

[colors.normal]
black = "#{{term0.raw}}"
red = "#{{term1.raw}}"
green = "#{{term2.raw}}"
yellow = "#{{term3.raw}}"
blue = "#{{term4.raw}}"
magenta = "#{{term5.raw}}"
cyan = "#{{term6.raw}}"
white = "#{{term7.raw}}"

[colors.bright]
black = "#{{term8.raw}}"
red = "#{{term9.raw}}"
green = "#{{term10.raw}}"
yellow = "#{{term11.raw}}"
blue = "#{{term12.raw}}"
magenta = "#{{term13.raw}}"
cyan = "#{{term14.raw}}"
white = "#{{term15.raw}}"
`

const weztermTemplate = `-- Heimdall theme for WezTerm
-- Generated automatically

return {
  color_scheme = "Heimdall",
  color_schemes = {
    ["Heimdall"] = {
      background = "#{{background.raw}}",
      foreground = "#{{text.raw}}",
      cursor_bg = "#{{text.raw}}",
      cursor_fg = "#{{background.raw}}",
      ansi = {
        "#{{term0.raw}}", -- black
        "#{{term1.raw}}", -- red
        "#{{term2.raw}}", -- green
        "#{{term3.raw}}", -- yellow
        "#{{term4.raw}}", -- blue
        "#{{term5.raw}}", -- magenta
        "#{{term6.raw}}", -- cyan
        "#{{term7.raw}}", -- white
      },
      brights = {
        "#{{term8.raw}}",  -- bright black
        "#{{term9.raw}}",  -- bright red
        "#{{term10.raw}}", -- bright green
        "#{{term11.raw}}", -- bright yellow
        "#{{term12.raw}}", -- bright blue
        "#{{term13.raw}}", -- bright magenta
        "#{{term14.raw}}", -- bright cyan
        "#{{term15.raw}}", -- bright white
      },
    },
  },
}
`
