// Package theme provides the core theme engine for applying color schemes
// to various applications and managing theme-related operations.
package theme

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/discord"
	"github.com/arthur404dev/heimdall-cli/internal/terminal"
	"github.com/arthur404dev/heimdall-cli/internal/theme/appthemes"
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
	// Special handling for Discord (uses Discord client manager)
	if app == "discord" {
		return a.ApplyDiscordThemes(colors)
	}

	// Get template from registry
	templateContent, err := appthemes.Get(app)
	if err != nil {
		// Try checking for custom template file as fallback
		templatePath := filepath.Join(a.templateDir, app+".tmpl")
		if _, statErr := os.Stat(templatePath); statErr == nil {
			contentBytes, readErr := os.ReadFile(templatePath)
			if readErr == nil {
				templateContent = string(contentBytes)
			} else {
				return fmt.Errorf("template not found for %s: %w", app, err)
			}
		} else {
			return fmt.Errorf("template not found for %s: %w", app, err)
		}
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
	// Special handlers that need custom logic
	specialHandlers := map[string]func(map[string]string) error{
		"discord": a.ApplyDiscordThemes,
	}

	// Mode-aware handlers
	modeHandlers := map[string]func(map[string]string, string) error{
		"gtk": func(c map[string]string, m string) error {
			return NewGTKHandler().Apply(c, m)
		},
		"qt": func(c map[string]string, m string) error {
			return NewQtHandler().Apply(c, m)
		},
	}

	// Apply themes to all registered templates
	for _, name := range appthemes.List() {
		// Skip if handled by special or mode handlers
		if _, isSpecial := specialHandlers[name]; isSpecial {
			continue
		}
		if _, hasMode := modeHandlers[name]; hasMode {
			continue
		}

		// Apply generic theme
		if err := a.ApplyTheme(name, colors, mode); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to apply %s theme: %v\n", name, err)
		}
	}

	// Apply special handlers
	for name, handler := range specialHandlers {
		if err := handler(colors); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to apply %s theme: %v\n", name, err)
		}
	}

	// Apply mode-aware handlers
	for name, handler := range modeHandlers {
		if err := handler(colors, mode); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to apply %s theme: %v\n", name, err)
		}
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
			return a.ApplyTheme("btop", colors, mode)
		}},
		{"fuzzel", func() error {
			return a.ApplyTheme("fuzzel", colors, mode)
		}},
		{"spicetify", func() error {
			return a.ApplyTheme("spicetify", colors, mode)
		}},
		{"discord", func() error {
			return a.ApplyDiscordThemes(colors)
		}},
		{"terminal", func() error {
			return a.ApplyTerminalSequences(colors, schemeName)
		}},
		{"kitty", func() error {
			return a.ApplyTheme("kitty", colors, mode)
		}},
		{"alacritty", func() error {
			return a.ApplyTheme("alacritty", colors, mode)
		}},
		{"wezterm", func() error {
			return a.ApplyTheme("wezterm", colors, mode)
		}},
		{"nvim", func() error {
			return a.ApplyTheme("nvim", colors, mode)
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

// GetOutputPath returns the output path for a themed application
// This first checks if the template has registered its own path,
// otherwise falls back to config paths
func (a *Applier) GetOutputPath(app string) string {
	// Try to get path from template registry first
	if path, err := appthemes.GetOutputPath(app); err == nil {
		return path
	}

	// Fall back to config for non-template apps (like discord variants)
	cfg := config.Get()
	if cfg == nil {
		// This should never happen, but if it does, return a sensible default
		return filepath.Join(a.configDir, app, "heimdall.theme")
	}

	// Handle special cases that don't have templates
	switch app {
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
