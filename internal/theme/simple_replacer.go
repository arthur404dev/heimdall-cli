package theme

import (
	"fmt"
	"strings"
)

// SimpleReplacer handles simple string replacement for templates
// This replaces the complex Go template engine with simple substitution
type SimpleReplacer struct{}

// NewSimpleReplacer creates a new simple replacer
func NewSimpleReplacer() *SimpleReplacer {
	return &SimpleReplacer{}
}

// isHexColor checks if a string is a valid hex color (without # prefix)
func isHexColor(s string) bool {
	if len(s) != 6 && len(s) != 8 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
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

	// Add cursor default if not present
	if _, ok := extendedColors["cursor"]; !ok {
		if val, ok := extendedColors["foreground"]; ok {
			extendedColors["cursor"] = val
		}
	}

	// First handle placeholders with default values like {{cursor|default:foreground}}
	// Process all occurrences
	startPos := 0
	for startPos < len(result) {
		start := strings.Index(result[startPos:], "{{")
		if start == -1 {
			break
		}
		start += startPos

		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		placeholder := result[start:end]
		inner := strings.Trim(placeholder, "{}")

		// Check for default value syntax
		if strings.Contains(inner, "|default:") {
			parts := strings.Split(inner, "|default:")
			if len(parts) == 2 {
				key := parts[0]
				defaultKey := parts[1]

				// Try to get the value for the key
				if val, ok := extendedColors[key]; ok {
					// Ensure color has # prefix if it's a hex color
					if len(val) == 6 && isHexColor(val) {
						val = "#" + val
					}
					result = result[:start] + val + result[end:]
					startPos = start + len(val)
				} else if defaultVal, ok := extendedColors[defaultKey]; ok {
					// Use the default value
					if len(defaultVal) == 6 && isHexColor(defaultVal) {
						defaultVal = "#" + defaultVal
					}
					result = result[:start] + defaultVal + result[end:]
					startPos = start + len(defaultVal)
				} else {
					// Skip this placeholder if we can't resolve it
					startPos = end
				}
			} else {
				startPos = end
			}
		} else {
			startPos = end
		}
	}

	// Replace all remaining color variables with their values
	for key, value := range extendedColors {
		// Ensure color has # prefix if it's a hex color
		if len(value) == 6 && isHexColor(value) {
			value = "#" + value
		}

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
	// Just direct {{key}} → value substitution
	return r.ReplaceString(templateContent, colors), nil
}
