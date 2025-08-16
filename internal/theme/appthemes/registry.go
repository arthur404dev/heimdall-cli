// Package appthemes contains all the template constants for theming various applications
package appthemes

import (
	"fmt"
	"sync"
)

// Template represents an application theme template
type Template struct {
	// Name is the identifier for this template (e.g., "kitty", "alacritty")
	Name string

	// Aliases are alternative names for this template (e.g., "gtk3", "gtk4" for "gtk")
	Aliases []string

	// Content is the actual template string
	Content string

	// Description provides information about what this template is for
	Description string

	// GetOutputPath returns the path where this theme should be written
	// If nil, uses the default logic from config
	GetOutputPath func() string

	// CustomApply is an optional custom application function
	// If nil, uses the standard template replacement logic
	CustomApply func(colors map[string]string, mode string) error
}

// Registry holds all registered templates
type Registry struct {
	mu        sync.RWMutex
	templates map[string]*Template
}

// globalRegistry is the singleton registry instance
var globalRegistry = &Registry{
	templates: make(map[string]*Template),
}

// Register adds a template to the global registry
func Register(template *Template) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	// Register with primary name
	globalRegistry.templates[template.Name] = template

	// Register aliases
	for _, alias := range template.Aliases {
		globalRegistry.templates[alias] = template
	}
}

// Get retrieves a template by name
func Get(name string) (string, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if template, ok := globalRegistry.templates[name]; ok {
		return template.Content, nil
	}

	return "", fmt.Errorf("no template registered for %s", name)
}

// GetTemplate retrieves the full Template struct by name
func GetTemplate(name string) (*Template, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if template, ok := globalRegistry.templates[name]; ok {
		return template, nil
	}

	return nil, fmt.Errorf("no template registered for %s", name)
}

// List returns all registered template names (excluding aliases)
func List() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	seen := make(map[*Template]bool)
	var names []string

	for _, template := range globalRegistry.templates {
		if !seen[template] {
			seen[template] = true
			names = append(names, template.Name)
		}
	}

	return names
}

// Exists checks if a template is registered
func Exists(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, ok := globalRegistry.templates[name]
	return ok
}

// GetOutputPath returns the output path for a given template
func GetOutputPath(name string) (string, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if template, ok := globalRegistry.templates[name]; ok {
		if template.GetOutputPath != nil {
			return template.GetOutputPath(), nil
		}
	}

	return "", fmt.Errorf("no output path defined for %s", name)
}

// HasCustomApply checks if a template has a custom apply function
func HasCustomApply(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if template, ok := globalRegistry.templates[name]; ok {
		return template.CustomApply != nil
	}
	return false
}

// ApplyCustom runs the custom apply function for a template
func ApplyCustom(name string, colors map[string]string, mode string) error {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if template, ok := globalRegistry.templates[name]; ok {
		if template.CustomApply != nil {
			return template.CustomApply(colors, mode)
		}
		return fmt.Errorf("no custom apply function for %s", name)
	}

	return fmt.Errorf("template not found: %s", name)
}
