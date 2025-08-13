package theme

import (
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

	// Replace all color variables with their values
	for key, value := range colors {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
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
