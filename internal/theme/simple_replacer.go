package theme

import (
	"fmt"
	"strings"
)

// SimpleReplacer handles simple string replacement for templates
// This replaces the complex Go template engine with caelestia-style simple substitution
type SimpleReplacer struct{}

// NewSimpleReplacer creates a new simple replacer
func NewSimpleReplacer() *SimpleReplacer {
	return &SimpleReplacer{}
}

// ReplaceString performs simple string replacement on template content
// Replaces patterns like {{colour0}}, {{colour1}}, etc. with actual color values
func (r *SimpleReplacer) ReplaceString(templateStr string, colors map[string]string) string {
	result := templateStr

	// Create extended color map with aliases for compatibility
	extendedColors := make(map[string]string)
	for key, value := range colors {
		extendedColors[key] = value
	}

	// Add aliases for terminal colors (support term0, color0, colour0 formats)
	for i := 0; i < 16; i++ {
		// Check which format the scheme uses and create aliases
		if val, ok := colors[fmt.Sprintf("term%d", i)]; ok {
			extendedColors[fmt.Sprintf("color%d", i)] = val
			extendedColors[fmt.Sprintf("colour%d", i)] = val
		} else if val, ok := colors[fmt.Sprintf("color%d", i)]; ok {
			extendedColors[fmt.Sprintf("term%d", i)] = val
			extendedColors[fmt.Sprintf("colour%d", i)] = val
		} else if val, ok := colors[fmt.Sprintf("colour%d", i)]; ok {
			extendedColors[fmt.Sprintf("term%d", i)] = val
			extendedColors[fmt.Sprintf("color%d", i)] = val
		}
	}

	// Add text/foreground aliases
	if val, ok := colors["text"]; ok {
		extendedColors["foreground"] = val
	} else if val, ok := colors["foreground"]; ok {
		extendedColors["text"] = val
	}

	// Replace all color variables with their values
	for key, value := range extendedColors {
		// Standard replacement with hash
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)

		// Also support .raw suffix to get color without hash prefix
		placeholderRaw := "{{" + key + ".raw}}"
		valueRaw := strings.TrimPrefix(value, "#")
		result = strings.ReplaceAll(result, placeholderRaw, valueRaw)
	}

	return result
}

// ReplaceTemplate processes a template string with color replacements
// This is the main method that replaces the complex template engine
func (r *SimpleReplacer) ReplaceTemplate(templateContent string, colors map[string]string) (string, error) {
	// Simple string replacement - no complex logic, conditionals, or loops
	// Just direct {{key}} → value substitution like caelestia
	return r.ReplaceString(templateContent, colors), nil
}
