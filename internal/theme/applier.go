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
	outputPath := a.getOutputPath(app)
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
			outputPath := a.getOutputPath(app)
			return paths.AtomicWrite(outputPath, []byte(rendered))
		}
	}

	// Not in cache, render normally
	err := a.ApplyTheme(app, colors, mode)
	if err != nil {
		return err
	}

	// Read rendered file and cache it
	outputPath := a.getOutputPath(app)
	if content, err := os.ReadFile(outputPath); err == nil {
		// Store in cache (estimate size as byte length)
		a.cache.Set(cacheKey, string(content), int64(len(content)))
	}

	return nil
}

// ApplyTerminalSequences generates and applies ANSI terminal sequences
func (a *Applier) ApplyTerminalSequences(colors map[string]string, schemeName string) error {
	builder := terminal.NewSequenceBuilder()
	applier := terminal.NewApplier()

	// Generate sequences
	sequences, err := builder.GenerateSequences(colors)
	if err != nil {
		return fmt.Errorf("failed to generate terminal sequences: %w", err)
	}

	// Apply sequences to active terminals immediately (like caelestia)
	if err := applier.ApplySequencesWithFallback(colors, schemeName); err != nil {
		// Log warning but don't fail the entire operation
		fmt.Fprintf(os.Stderr, "Warning: failed to apply sequences to terminals: %v\n", err)
	}

	// Format for shell sourcing
	shellScript := builder.FormatSequencesForShell(sequences, schemeName)

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

// ApplyBtopTheme applies theme to btop
func (a *Applier) ApplyBtopTheme(colors map[string]string) error {
	content := a.generateBtopTheme(colors)
	btopPath := filepath.Join(a.configDir, "btop", "themes", "heimdall.theme")

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
	fuzzelPath := filepath.Join(a.configDir, "fuzzel", "fuzzel.ini")

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
	spicetifyPath := filepath.Join(a.configDir, "spicetify", "Themes", "heimdall", "color.ini")

	// Ensure directory exists
	dir := filepath.Dir(spicetifyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create spicetify themes directory: %w", err)
	}

	return paths.AtomicWrite(spicetifyPath, []byte(content))
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
	fg := strings.TrimPrefix(colors["foreground"], "#")
	primary := strings.TrimPrefix(colors["colour4"], "#")
	surface := strings.TrimPrefix(colors["colour0"], "#")
	outline := strings.TrimPrefix(colors["colour8"], "#")

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
	// Spicetify uses colors without # prefix
	bg := strings.TrimPrefix(colors["background"], "#")
	fg := strings.TrimPrefix(colors["foreground"], "#")
	surface := strings.TrimPrefix(colors["colour0"], "#")
	surfaceVar := strings.TrimPrefix(colors["colour8"], "#")
	primary := strings.TrimPrefix(colors["colour4"], "#")
	secondary := strings.TrimPrefix(colors["colour7"], "#")

	return fmt.Sprintf(`# Heimdall theme for Spicetify
# Generated automatically

[Base]
main_bg = %s
sidebar_bg = %s
player_bg = %s
card_bg = %s
shadow = 000000
main_fg = %s
sidebar_fg = %s
secondary_fg = %s
selected_button = %s
pressing_button_bg = %s
pressing_button_fg = %s
miscellaneous_bg = %s
miscellaneous_hover_bg = %s
preserve_1 = ffffff
`, bg, surface, surfaceVar, surface, fg, fg, secondary, primary, surface, fg, surfaceVar, surface)
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
