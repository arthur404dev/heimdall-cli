package theme

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
)

// Engine is the modular theme engine that orchestrates theme application
type Engine struct {
	handlers      map[string]ApplicationHandler
	lazyHandlers  map[string]func() ApplicationHandler // Lazy-loaded handlers
	processor     TemplateProcessor
	mapper        ColorMapper
	validator     *Validator
	backup        BackupManager
	transaction   TransactionManager
	templates     map[string]*template.Template
	templateCache *TemplateCache
	colorCache    *ColorConversionCache
	funcs         template.FuncMap
	mu            sync.RWMutex
	workerPool    int
}

// NewEngine creates a new modular theme engine
func NewEngine() *Engine {
	// Initialize caches
	cacheDir := "/tmp/heimdall-cache" // Can be configured
	templateCache := NewTemplateCache(10, true, cacheDir)
	colorCache := NewColorConversionCache(1000)

	e := &Engine{
		handlers:      make(map[string]ApplicationHandler),
		lazyHandlers:  make(map[string]func() ApplicationHandler),
		templates:     make(map[string]*template.Template),
		templateCache: templateCache,
		colorCache:    colorCache,
		validator:     NewValidator(),
		workerPool:    8, // Default worker pool size
	}

	// Set up template functions
	e.funcs = template.FuncMap{
		// Color manipulation functions
		"hex":     e.toHex,
		"rgb":     e.toRGB,
		"rgba":    e.toRGBA,
		"hsl":     e.toHSL,
		"hsla":    e.toHSLA,
		"lighten": e.lighten,
		"darken":  e.darken,
		"alpha":   e.alpha,

		// String manipulation
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"replace": strings.ReplaceAll,

		// Conditional helpers
		"isDark":  e.isDark,
		"isLight": e.isLight,
	}

	// Initialize components
	e.processor = NewTemplateProcessor()
	e.mapper = NewColorMapper()

	return e
}

// RegisterHandler registers a new application handler
func (e *Engine) RegisterHandler(name string, handler ApplicationHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.handlers[name]; exists {
		return fmt.Errorf("handler %s already registered", name)
	}

	e.handlers[name] = handler
	logger.Info("Registered handler", "application", name)
	return nil
}

// RegisterLazyHandler registers a handler factory for lazy initialization
func (e *Engine) RegisterLazyHandler(name string, factory func() ApplicationHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.handlers[name]; exists {
		return fmt.Errorf("handler %s already registered", name)
	}
	if _, exists := e.lazyHandlers[name]; exists {
		return fmt.Errorf("lazy handler %s already registered", name)
	}

	e.lazyHandlers[name] = factory
	logger.Info("Registered lazy handler", "application", name)
	return nil
}

// getHandler retrieves a handler, initializing it if needed (lazy loading)
func (e *Engine) getHandler(name string) (ApplicationHandler, error) {
	e.mu.RLock()
	handler, exists := e.handlers[name]
	e.mu.RUnlock()

	if exists {
		return handler, nil
	}

	// Check for lazy handler
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check after acquiring write lock
	if handler, exists = e.handlers[name]; exists {
		return handler, nil
	}

	if factory, ok := e.lazyHandlers[name]; ok {
		// Measure initialization time
		start := time.Now()
		handler = factory()
		elapsed := time.Since(start)

		// Store initialized handler
		e.handlers[name] = handler
		delete(e.lazyHandlers, name) // Remove factory after initialization

		// Log if initialization was slow
		if elapsed > 10*time.Millisecond {
			logger.Warn("Slow handler initialization", "application", name, "duration", elapsed)
		} else {
			logger.Info("Handler initialized", "application", name, "duration", elapsed)
		}

		return handler, nil
	}

	return nil, fmt.Errorf("no handler registered for %s", name)
}

// SetWorkerPoolSize sets the number of workers for parallel processing
func (e *Engine) SetWorkerPoolSize(size int) {
	if size > 0 {
		e.workerPool = size
	}
}

// ApplyTheme applies a theme to specified applications
func (e *Engine) ApplyTheme(ctx context.Context, scheme *ColorScheme, options ApplyOptions) error {
	// Validate the theme first
	if err := e.ValidateTheme(scheme); err != nil {
		return fmt.Errorf("theme validation failed: %w", err)
	}

	// Get target applications
	apps := options.Applications
	if len(apps) == 0 {
		apps = e.GetSupportedApplications()
	}

	// Create backup if not disabled
	var backupID string
	if !options.NoBackup && e.backup != nil {
		files := e.getTargetFiles(apps)
		id, err := e.backup.CreateBackup(files)
		if err != nil {
			logger.Warn("Failed to create backup", "error", err)
		} else {
			backupID = id
			logger.Info("Created backup", "id", backupID)
		}
	}

	// Start transaction if available
	var tx Transaction
	if e.transaction != nil {
		var err error
		tx, err = e.transaction.Begin()
		if err != nil {
			logger.Warn("Failed to start transaction", "error", err)
		}
	}

	// Apply theme to each application
	var errors []error
	if options.Parallel {
		errors = e.applyParallel(ctx, scheme, apps, options)
	} else {
		errors = e.applySequential(ctx, scheme, apps, options)
	}

	// Handle errors
	if len(errors) > 0 {
		// Rollback transaction if available
		if tx != nil {
			if err := e.transaction.Rollback(tx); err != nil {
				logger.Error("Failed to rollback transaction", "error", err)
			}
		}

		// Restore from backup if available
		if backupID != "" && e.backup != nil {
			if err := e.backup.RestoreBackup(backupID); err != nil {
				logger.Error("Failed to restore backup", "error", err)
			}
		}

		return fmt.Errorf("theme application failed: %v", errors)
	}

	// Commit transaction if available
	if tx != nil {
		if err := e.transaction.Commit(tx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	logger.Info("Theme applied successfully", "applications", len(apps))
	return nil
}

// ValidateTheme validates a theme before application
func (e *Engine) ValidateTheme(scheme *ColorScheme) error {
	return e.validator.ValidateScheme(scheme)
}

// GetSupportedApplications returns a list of supported applications
func (e *Engine) GetSupportedApplications() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	apps := make([]string, 0, len(e.handlers))
	for name := range e.handlers {
		apps = append(apps, name)
	}
	return apps
}

// applySequential applies theme to applications sequentially
func (e *Engine) applySequential(ctx context.Context, scheme *ColorScheme, apps []string, options ApplyOptions) []error {
	var errors []error

	for _, app := range apps {
		select {
		case <-ctx.Done():
			errors = append(errors, ctx.Err())
			return errors
		default:
			if err := e.applyToApp(app, scheme, options); err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", app, err))
				if options.Verbose {
					logger.Error("Failed to apply theme", "app", app, "error", err)
				}
			} else if options.Verbose {
				logger.Info("Applied theme", "app", app)
			}
		}
	}

	return errors
}

// applyParallel applies theme to applications in parallel
func (e *Engine) applyParallel(ctx context.Context, scheme *ColorScheme, apps []string, options ApplyOptions) []error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(apps))

	for _, app := range apps {
		wg.Add(1)
		go func(appName string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- fmt.Errorf("%s: %w", appName, ctx.Err())
				return
			default:
				if err := e.applyToApp(appName, scheme, options); err != nil {
					errChan <- fmt.Errorf("%s: %w", appName, err)
					if options.Verbose {
						logger.Error("Failed to apply theme", "app", appName, "error", err)
					}
				} else if options.Verbose {
					logger.Info("Applied theme", "app", appName)
				}
			}
		}(app)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	return errors
}

// applyToApp applies theme to a single application
func (e *Engine) applyToApp(app string, scheme *ColorScheme, options ApplyOptions) error {
	// Use lazy loading to get handler
	handler, err := e.getHandler(app)
	if err != nil {
		return err
	}

	// Check if application is installed
	if !handler.IsInstalled() {
		if options.Verbose {
			logger.Info("Application not installed, skipping", "app", app)
		}
		return nil
	}

	// Map colors for the application
	colors, err := e.mapper.MapColors(scheme, app)
	if err != nil {
		return fmt.Errorf("failed to map colors: %w", err)
	}

	// Validate colors for the handler
	if err := handler.Validate(colors); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Apply the theme
	handlerOpts := HandlerOptions{
		Mode:    scheme.Mode,
		Verbose: options.Verbose,
	}

	if options.TemplateDir != "" {
		handlerOpts.TemplateOverride = options.TemplateDir
	}

	return handler.Apply(colors, handlerOpts)
}

// getTargetFiles returns the list of files that will be modified
func (e *Engine) getTargetFiles(apps []string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var files []string
	for _, app := range apps {
		if handler, exists := e.handlers[app]; exists {
			files = append(files, handler.GetOutputPath())
		}
	}

	return files
}

// LoadTemplate loads a template from string
func (e *Engine) LoadTemplate(name, content string) error {
	tmpl, err := template.New(name).Funcs(e.funcs).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	e.templates[name] = tmpl
	return nil
}

// LoadTemplateFile loads a template from a file
func (e *Engine) LoadTemplateFile(name, path string) error {
	tmpl, err := template.New(name).Funcs(e.funcs).ParseFiles(path)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", path, err)
	}

	e.templates[name] = tmpl
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(name string, data interface{}) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderString renders a template string directly
func (e *Engine) RenderString(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("inline").Funcs(e.funcs).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse inline template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute inline template: %w", err)
	}

	return buf.String(), nil
}

// Template function implementations

// toHex converts a color to hex format
func (e *Engine) toHex(color interface{}) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			return c
		}
		return "#" + c
	case map[string]interface{}:
		if hex, ok := c["hex"].(string); ok {
			return hex
		}
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("#%02x%02x%02x", r, g, b)
		}
	}
	return "#000000"
}

// toRGB converts a color to RGB format
func (e *Engine) toRGB(color interface{}) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
			}
		}
	case map[string]interface{}:
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
		}
	}
	return "rgb(0, 0, 0)"
}

// toRGBA converts a color to RGBA format with alpha
func (e *Engine) toRGBA(color interface{}, alpha float64) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
			}
		}
	case map[string]interface{}:
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
		}
	}
	return fmt.Sprintf("rgba(0, 0, 0, %.2f)", alpha)
}

// toHSL converts a color to HSL format
func (e *Engine) toHSL(color interface{}) string {
	// Implementation would convert to HSL
	// For now, return a placeholder
	return "hsl(0, 0%, 0%)"
}

// toHSLA converts a color to HSLA format with alpha
func (e *Engine) toHSLA(color interface{}, alpha float64) string {
	// Implementation would convert to HSLA
	// For now, return a placeholder
	return fmt.Sprintf("hsla(0, 0%%, 0%%, %.2f)", alpha)
}

// lighten lightens a color by a percentage
func (e *Engine) lighten(color interface{}, percent float64) string {
	// Implementation would lighten the color
	// For now, return the original color
	return e.toHex(color)
}

// darken darkens a color by a percentage
func (e *Engine) darken(color interface{}, percent float64) string {
	// Implementation would darken the color
	// For now, return the original color
	return e.toHex(color)
}

// alpha adds alpha channel to a color
func (e *Engine) alpha(color interface{}, alpha float64) string {
	return e.toRGBA(color, alpha)
}

// isDark checks if a color is dark
func (e *Engine) isDark(color interface{}) bool {
	// Simple luminance check
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
				return luminance < 128
			}
		}
	}
	return false
}

// isLight checks if a color is light
func (e *Engine) isLight(color interface{}) bool {
	return !e.isDark(color)
}

// Helper functions

func parseHexByte(s string) (uint8, error) {
	var b uint8
	_, err := fmt.Sscanf(s, "%02x", &b)
	return b, err
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case uint8:
		return int(val)
	default:
		return 0
	}
}
